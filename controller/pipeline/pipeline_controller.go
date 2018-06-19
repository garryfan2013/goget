package pipeline

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strconv"
	"sync"

	"github.com/garryfan2013/goget/config"
	"github.com/garryfan2013/goget/controller"
	"github.com/garryfan2013/goget/sink"
	"github.com/garryfan2013/goget/source"
	"github.com/garryfan2013/goget/util"
)

// Default value for config parameters
const (
	PoolBufferAllocSize = 128 * 1024
)

// Control message definition
const (
	CTRL_MSG_GET_STATS = iota
)

const (
	NOTIFY_MSG_STREAM_INFO = iota + 1000
)

func init() {
	controller.Register(&PipelineControllerCreator{})
}

type PipelineControllerCreator struct{}

func (*PipelineControllerCreator) Create() (controller.ProgressController, error) {
	sw := newPipelineController().(controller.ProgressController)
	return sw, nil
}

func (*PipelineControllerCreator) Scheme() string {
	return controller.SchemePipeline
}

// ReaderTaskGenerator-AsyncExecutor
// Input(string): download url
// Output(ReaderTaskIterator): Iterator for reader task
type UrlRequestHandler struct {
	src    source.StreamReader
	tasks  []ReaderTask
	pos    int
	notify chan<- *Roundtrip
}

type ReaderTask struct {
	offset int64
	size   int64
}

func NewUrlRequestHandler(s source.StreamReader, n int, notify chan<- *Roundtrip) *UrlRequestHandler {
	return &UrlRequestHandler{
		src:    s,
		tasks:  make([]ReaderTask, n),
		pos:    0,
		notify: notify}
}

func (h *UrlRequestHandler) Handle(ctx context.Context, arg interface{}) (interface{}, error) {
	_, ok := arg.(string)
	if !ok {
		fmt.Printf("Unexpected data type %t\n", arg)
		return nil, errors.New("Unexpected data type")
	}

	total, err := h.src.Size(ctx)
	if err != nil {
		return nil, err
	}

	cnt := len(h.tasks)
	size := total / int64(cnt)
	stats := make([]workerStats, cnt)
	for i := 0; i < cnt; i++ {
		h.tasks[i].offset = int64(i) * size
		h.tasks[i].size = size
		if i == cnt-1 {
			h.tasks[i].size = h.tasks[i].size + total%int64(cnt)
		}
		stats[i].offset = h.tasks[i].offset
		stats[i].size = h.tasks[i].size
	}

	// Inform the controller of the stream information
	// This is a RountdTrip message since we dont want to start workers until
	// controller get the stream info
	resp := make(chan *Message)
	h.notify <- &Roundtrip{
		msg: Message{
			cmd: NOTIFY_MSG_STREAM_INFO,
			data: &streamInfo{
				total: total,
				stats: stats,
			},
		},
		resp: resp,
	}
	<-resp

	return h, nil
}

func (h *UrlRequestHandler) Next(ctx context.Context) (interface{}, error) {
	if h.pos < len(h.tasks) {
		var tp *ReaderTask = &h.tasks[h.pos]
		h.pos += 1
		return tp, nil
	}

	return nil, util.ErrIteratorEOF
}

// Reader-AsyncExecutor
// Input(ReaderTask): the specific offset and size of a target file block
// Output([]byte): stream of block data
type ReaderTaskHandler struct {
	src        source.StreamReader // Source crawler
	rc         io.ReadCloser       // io interface obtained from crawler
	bufferPool *sync.Pool          // This BufferPool will live as long as the ReaderTaskHandler
	offset     int64               // Block offset in the whole file
	size       int64               // Block size
	left       int64               // Block data remained to read
}

// It's the writer's duty to put the buffer back to the pool
type WriterTask struct {
	buf        *util.StaticBuffer
	bufferPool *sync.Pool
	offset     int64
}

func NewReaderTaskHandler(s source.StreamReader, bs int64) *ReaderTaskHandler {
	return &ReaderTaskHandler{
		src: s,
		bufferPool: &sync.Pool{
			New: func() interface{} {
				t := new(WriterTask)
				t.buf = util.NewStaticBuffer(bs)
				return t
			},
		}}
}

func (h *ReaderTaskHandler) Handle(ctx context.Context, arg interface{}) (interface{}, error) {
	t, ok := arg.(*ReaderTask)
	if !ok {
		fmt.Printf("ReaderTaskHandler: Unexpected data type %t\n", arg)
		return nil, errors.New("Unexpected data type")
	}

	h.offset = t.offset
	h.size = t.size
	h.left = t.size

	rc, err := h.src.Get(ctx, h.offset, h.size)
	if err != nil {
		return nil, err
	}

	h.rc = rc
	return h, nil
}

// Implement the iterator interface
func (h *ReaderTaskHandler) Next(ctx context.Context) (interface{}, error) {
	task := h.bufferPool.Get().(*WriterTask)

	task.buf.Reset()
	task.offset = h.offset + h.size - h.left
	task.bufferPool = h.bufferPool

	n, err := task.buf.ReadFrom(h.rc)

	if n > 0 {
		h.left -= int64(n)
	}

	if err == io.EOF {
		if h.left != 0 {
			panic("ReaderHandler.Next reach EOF, but got not enough data")
		}
		return task, util.ErrIteratorEOF
	}

	if err != nil {
		h.bufferPool.Put(task)
		return task, err
	}

	if h.left < 0 {
		panic("ReaderTaskHandler.Next: left negative!")
	}

	if h.left == 0 {
		return task, util.ErrIteratorEOF
	}

	return task, nil
}

// Implement the Closer interface
func (h *ReaderTaskHandler) Close() error {
	h.rc.Close()
	return nil
}

// FileWriter-AsyncExecutor
// Input([]byte): the actual data block
// Output(bool): status
type WriterTaskHandler struct {
	sw sink.StreamWriter
}

type writeInfo struct {
	size   int64
	offset int64
}

func NewWriterTaskHandler(s sink.StreamWriter) *WriterTaskHandler {
	return &WriterTaskHandler{sw: s}
}

func (h *WriterTaskHandler) Handle(ctx context.Context, arg interface{}) (interface{}, error) {
	t, ok := arg.(*WriterTask)
	if !ok {
		fmt.Printf("Unexpected data type %t\n", arg)
		return nil, errors.New("Unexpected data type")
	}

	defer t.bufferPool.Put(t)

	wa := h.sw.(io.WriterAt)
	w := util.NewOffsetWriter(wa, t.offset)
	n, err := t.buf.WriteTo(w)
	if err != nil {
		return nil, err
	}

	return &writeInfo{
		offset: t.offset,
		size:   int64(n),
	}, nil
}

// Sync request-response module
type Message struct {
	cmd  int
	data interface{}
}

type Roundtrip struct {
	msg  Message
	resp chan<- *Message
}

// PiplelineController constructor
type PipelineController struct {
	configs map[string]string   // config params map
	src     source.StreamReader // source
	snk     sink.StreamWriter   // sink
	ctrl    chan *Roundtrip     // The ctrl channel for pipeline
	cancel  context.CancelFunc  // Cancel function
	stats   []workerStats       // This is the stats slice for all reader workers
	total   int64               // This indicates the total lengh of the file stream
	done    int64               // This indicates the finished bytes
}

type streamInfo struct {
	total int64
	stats []workerStats
}

// the worker stats indicate the progress of a single worker routine
type workerStats struct {
	size   int64
	offset int64
	Done   int64
}

func newPipelineController() interface{} {
	return new(PipelineController)
}

func (pc *PipelineController) Open(c source.StreamReader, h sink.StreamWriter) error {
	pc.configs = make(map[string]string)
	pc.src = c
	pc.snk = h
	pc.ctrl = make(chan *Roundtrip)

	return nil
}

func (pc *PipelineController) SetConfig(key string, value string) {
	pc.configs[key] = value
}

func (pc *PipelineController) Start() error {
	if pc.src == nil {
		return errors.New("No source set yet")
	}

	if pc.snk == nil {
		return errors.New("No sink set yet")
	}

	url, exists := pc.configs[config.KeyRemoteUrl]
	if exists == false {
		return errors.New("Source url not set!")
	}

	if err := pc.src.Open(url); err != nil {
		return err
	}

	if user, exists := pc.configs[config.KeyUserName]; exists {
		pc.src.SetConfig(config.KeyUserName, user)
	}

	if passwd, exists := pc.configs[config.KeyPasswd]; exists {
		pc.src.SetConfig(config.KeyPasswd, passwd)
	}

	path, exists := pc.configs[config.KeyLocalPath]
	if exists == false {
		return errors.New("Sink path not set!")
	}

	taskCountStr, exists := pc.configs[config.KeyTaskCount]
	if exists == false {
		return errors.New("TaskCount not set!")
	}
	taskCount, err := strconv.Atoi(taskCountStr)
	if err != nil {
		return err
	}

	if err := pc.snk.Open(path); err != nil {
		return err
	}

	ctx := context.Background()
	ctx, pc.cancel = context.WithCancel(ctx)

	// Prepare the start channel and notify channel
	startCh := make(chan interface{})
	notifyCh := make(chan *Roundtrip)

	urlHandler := NewUrlRequestHandler(pc.src, taskCount, notifyCh)
	urlEx := util.NewAsyncExecutor(urlHandler)

	crawlerExes := make([]util.Executor, taskCount)
	for i := 0; i < taskCount; i++ {
		crawlerHandler := NewReaderTaskHandler(pc.src, PoolBufferAllocSize)
		crawlerExes[i] = util.NewAsyncExecutor(crawlerHandler)
	}

	fanOutEx := util.NewFanOutAsyncExecutor()
	fanInEx := util.NewFanInAsyncExecutor()

	writerHandler := NewWriterTaskHandler(pc.snk)
	writerEx := util.NewAsyncExecutor(writerHandler)

	// Start the url executor
	taskCh, urlErrCh := urlEx.Run(ctx, startCh)

	// Start the fanOut executor
	fanOutChs, fanOutErrChs := fanOutEx.Run(ctx, taskCh, crawlerExes)

	// Start the fanIn executor
	fanInCh, fanInErrCh := fanInEx.Run(ctx, fanOutChs, fanOutErrChs)

	// Start the writer executor
	counterCh, writerErrCh := writerEx.Run(ctx, fanInCh)

	var sinfo *streamInfo
	go func() {
		for {
			select {
			// Error handler
			case e := <-urlErrCh:
				fmt.Printf("urlErrch: %s\n", e.Error())
			case e := <-fanInErrCh:
				fmt.Printf("fanInErrCh: %s\n", e.Error())
			case e := <-writerErrCh:
				fmt.Printf("writerErrCh: %s\n", e.Error())

			// Recv the total size of this job
			case rt := <-notifyCh:
				switch rt.msg.cmd {
				case NOTIFY_MSG_STREAM_INFO:
					sinfo = rt.msg.data.(*streamInfo)
					pc.stats = sinfo.stats
					pc.total = sinfo.total

					if rt.resp != nil {
						rt.resp <- &Message{
							cmd:  rt.msg.cmd,
							data: nil,
						}
					}
				}

			// Stats updated
			case ret := <-counterCh:
				wi, ok := ret.(*writeInfo)
				if !ok {
					panic("counterCh ret cant type switch to writeInfo")
				}

				for i, _ := range pc.stats {
					if pc.stats[i].offset+pc.stats[i].size >= wi.offset+wi.size {
						if pc.stats[i].offset <= wi.offset {
							pc.stats[i].Done += wi.size
							pc.done += wi.size
						}
					}
				}

			// Ctrl channel
			case rt := <-pc.ctrl:
				switch rt.msg.cmd {
				case CTRL_MSG_GET_STATS:
					rt.resp <- &Message{
						cmd: rt.msg.cmd,
						data: &controller.Stats{
							Size: pc.total,
							Done: pc.done,
						}}
				}

			// Cancel channel
			case <-ctx.Done():
				return
			}
		}
	}()

	// Start the pipeline
	startCh <- url
	return nil
}

func (pc *PipelineController) Stop() error {
	if pc.cancel != nil {
		pc.cancel()
	}

	return nil
}

func (pc *PipelineController) Close() {
	if pc.cancel != nil {
		pc.cancel()
	}

	pc.snk.Close()
	pc.src.Close()
}

func (pc *PipelineController) Progress() (*controller.Stats, error) {
	ch := make(chan *Message, 1)
	pc.ctrl <- &Roundtrip{
		msg:  Message{cmd: CTRL_MSG_GET_STATS},
		resp: ch,
	}

	msg := <-ch
	stats := msg.data.(*controller.Stats)
	return stats, nil
}
