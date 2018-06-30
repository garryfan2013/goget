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

var (
	wg sync.WaitGroup
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

type workerStats struct {
	offset int64
	size   int64
	done   int64
}

// ReaderTaskGenerator-AsyncExecutor
// Input(string): download url
// Output(ReaderTaskIterator): Iterator for reader task
type UrlRequestHandler struct {
	src    source.StreamReader
	sm     controller.WorkerStatsManager // Provided by up level component for workers' stats related operation
	count  int
	tasks  []*ReaderTask
	pos    int
	notify chan<- *Roundtrip
}

type UrlRequestTask struct {
	url string
}

type ReaderTask struct {
	offset int64
	size   int64
}

func NewUrlRequestHandler(s source.StreamReader, n int, notify chan<- *Roundtrip, sm controller.WorkerStatsManager) *UrlRequestHandler {
	return &UrlRequestHandler{
		src:    s,
		sm:     sm,
		count:  n,
		tasks:  make([]*ReaderTask, 0, n),
		pos:    0,
		notify: notify}
}

func (h *UrlRequestHandler) prepareTasksLayout(total int64) {
	cnt := h.count
	size := total / int64(cnt)
	ext := total % int64(cnt)

	for i := 0; i < cnt; i++ {
		blockSize := size
		if i == cnt-1 {
			blockSize += ext
		}

		h.tasks = append(h.tasks, &ReaderTask{
			offset: int64(i) * size,
			size:   blockSize,
		})
	}
}

func (h *UrlRequestHandler) restoreTasksLayout(stats []*controller.WorkerStats) {
	for i, v := range stats {
		if v.Done == v.Size {
			continue
		}

		h.tasks = append(h.tasks, &ReaderTask{
			offset: stats[i].Offset + stats[i].Done,
			size:   stats[i].Size - stats[i].Done,
		})
	}
}

func (h *UrlRequestHandler) Init() error {
	wg.Add(1)
	return nil
}

func (h *UrlRequestHandler) Finish() error {
	fmt.Printf("UrlRequestHandler done\n")
	wg.Done()
	return nil
}

func (h *UrlRequestHandler) Handle(ctx context.Context, arg interface{}) (interface{}, error) {
	_, ok := arg.(string)
	if !ok {
		fmt.Printf("Unexpected data type %t\n", arg)
		return nil, errors.New("Unexpected data type")
	}

	if h.sm == nil {
		panic("UrlRequestHandler has a nil StatsManager")
	}

	total, err := h.src.Size(ctx)
	if err != nil {
		return nil, err
	}

	retrievedStats, retrievedTotal, retrievedDone := h.sm.Retrieve()
	fmt.Printf("Retrieve: total=%d done=%d\n", retrievedTotal, retrievedDone)
	for id, val := range retrievedStats {
		fmt.Printf("Retrive: worker[%d] offset=%d size=%d done=%d\n", id, val.Offset, val.Size, val.Done)
	}
	if retrievedStats != nil {
		if len(retrievedStats) != h.count {
			panic("UrlRequestHandler count not equal retrieved stats count")
		}
	}

	if retrievedTotal > 0 {
		if retrievedTotal != total {
			return nil, errors.New("File length changed(remote size not equal the retrieved size)")
		}
	}

	var stats []*controller.WorkerStats
	var done int64
	if retrievedStats == nil {
		// This is a probably fresh task
		h.prepareTasksLayout(total)
		stats = make([]*controller.WorkerStats, len(h.tasks))
		for i, v := range h.tasks {
			stats[i] = &controller.WorkerStats{
				Offset: v.offset,
				Size:   v.size,
				Done:   0,
			}
		}
		done = 0
	} else {
		// For restored task
		h.restoreTasksLayout(retrievedStats)
		stats = retrievedStats
		done = retrievedDone
	}

	fmt.Printf("Tasks layout:\n")
	for i, v := range h.tasks {
		fmt.Printf("task[%d]: offset=%d size=%d\n", i, v.offset, v.size)
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
				done:  done,
			},
		},
		resp: resp,
	}
	<-resp

	return h, nil
}

func (h *UrlRequestHandler) Next(ctx context.Context) (interface{}, error) {
	if h.pos < len(h.tasks) {
		tp := h.tasks[h.pos]
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

func (h *ReaderTaskHandler) Init() error {
	wg.Add(1)
	return nil
}

func (h *ReaderTaskHandler) Finish() error {
	fmt.Printf("ReaderTaskHandler %d-%d done\n", h.offset, h.offset+h.size)
	wg.Done()
	return nil
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
		if h.left > 0 {
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

func (h *WriterTaskHandler) Init() error {
	wg.Add(1)
	return nil
}

func (h *WriterTaskHandler) Finish() error {
	fmt.Printf("WriterTaskHandler done\n")
	wg.Done()
	return nil
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

// FanInAsyncExecutor
// Input([]chan interface{}, []chan error): the output channel of all readers
// Output(chan interface{}, chan error): single channel for output
type FanInHandler struct {
}

func (*FanInHandler) Init() error {
	wg.Add(1)
	return nil
}

func (*FanInHandler) Finish() error {
	fmt.Printf("FanInHandler done\n")
	wg.Done()
	return nil
}

func (*FanInHandler) Handle(ctx context.Context, d interface{}) (interface{}, error) {
	return nil, nil
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
	configs map[string]string             // config params map
	src     source.StreamReader           // source
	snk     sink.StreamWriter             // sink
	ctrl    chan *Roundtrip               // The ctrl channel for pipeline
	cancel  context.CancelFunc            // Cancel function
	sm      controller.WorkerStatsManager // Provided by up level component for workers' stat related operation
	nf      controller.Notifier           // Provided by up level component for controller to deliver specific events
	stats   []*controller.WorkerStats     // This is the stats slice for all workers
	rate    int64                         // The download speed for whole job
	total   int64                         // This indicates the total lengh of the file stream
	done    int64                         // This indicates the finished bytes
}

type streamInfo struct {
	total int64
	done  int64
	stats []*controller.WorkerStats
}

func newPipelineController() interface{} {
	return new(PipelineController)
}

func (pc *PipelineController) Open(c source.StreamReader, h sink.StreamWriter, sm controller.WorkerStatsManager, nf controller.Notifier) error {
	pc.configs = make(map[string]string)
	pc.src = c
	pc.snk = h
	pc.ctrl = make(chan *Roundtrip)
	pc.sm = sm
	pc.nf = nf

	return nil
}

func (pc *PipelineController) SetConfig(key string, value string) {
	pc.configs[key] = value
}

func (pc *PipelineController) handleRoudTripMessage(rt *Roundtrip) {
	switch rt.msg.cmd {
	// Deal with the Progress request from upper level component
	case CTRL_MSG_GET_STATS:
		rt.resp <- &Message{
			cmd: rt.msg.cmd,
			data: &controller.Stats{
				Rate: pc.rate,
				Size: pc.total,
				Done: pc.done,
			},
		}
		break

	default:
		fmt.Printf("Recv unkown round trip message: cmd = %d\n", rt.msg.cmd)
	}
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

	defer func() {
		if err != nil {
			pc.cancel()
			wg.Wait()
		}
	}()

	// Prepare the start channel and notify channel
	startCh := make(chan interface{})
	notifyCh := make(chan *Roundtrip)

	urlHandler := NewUrlRequestHandler(pc.src, taskCount, notifyCh, pc.sm)
	urlEx := util.NewAsyncExecutor(urlHandler)

	crawlerExes := make([]util.Executor, taskCount)
	for i := 0; i < taskCount; i++ {
		crawlerHandler := NewReaderTaskHandler(pc.src, PoolBufferAllocSize)
		crawlerExes[i] = util.NewAsyncExecutor(crawlerHandler)
	}

	fanOutEx := util.NewFanOutAsyncExecutor()
	fanInEx := util.NewFanInAsyncExecutor(&FanInHandler{})

	writerHandler := NewWriterTaskHandler(pc.snk)
	writerEx := util.NewAsyncExecutor(writerHandler)

	// Start the url executor
	taskCh, urlErrCh, err := urlEx.Run(ctx, startCh)
	if err != nil {
		return err
	}

	// Start the fanOut executor
	fanOutChs, fanOutErrChs, err := fanOutEx.Run(ctx, taskCh, crawlerExes)
	if err != nil {
		return err
	}

	// Start the fanIn executor
	fanInCh, fanInErrCh, err := fanInEx.Run(ctx, fanOutChs, fanOutErrChs)
	if err != nil {
		return err
	}

	// Start the writer executor
	counterCh, writerErrCh, err := writerEx.Run(ctx, fanInCh)
	if err != nil {
		return err
	}

	var sinfo *streamInfo
	wg.Add(1)
	go func() {
		defer wg.Done()
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
					pc.done = sinfo.done

					if rt.resp != nil {
						rt.resp <- &Message{
							cmd:  rt.msg.cmd,
							data: nil,
						}
					}

					/*
						Probably this is a newly started job
						neccessary to update the info to the upper level component
					*/
					if pc.done == 0 {
						err := pc.sm.Update(pc.stats)
						if err != nil {
							fmt.Println(err)
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
					if pc.stats[i].Offset+pc.stats[i].Size >= wi.offset+wi.size {
						if pc.stats[i].Offset <= wi.offset {
							pc.stats[i].Done += wi.size
							pc.done += wi.size
						}
					}
				}

				err := pc.sm.Update(pc.stats)
				if err != nil {
					fmt.Println(err)
				}

				// Here the whole job's completed, deal with the rest request and quit
				if pc.done == pc.total {
					err := pc.nf.Notify(controller.NotifyEventDone)
					if err != nil {
						fmt.Println(err)
					}

					/*
						After the NotifyEventDone delivered to the upper component, There
						shouldnt be any ctrl messages sent to this controller.
						But the controller still need to get rid of the left ctrl messages
						that were sent before the NotifyEventDone delivered, in case the
						upper level component might be blocked
					*/
				loop:
					for {
						select {
						case rt := <-pc.ctrl:
							pc.handleRoudTripMessage(rt)
							break
						default:
							break loop
						}
					}

					/*
						Since this is one of the go routines need to be waitted
						for completion, so have to start a new go routine to do the wait
					*/
					go func() {
						pc.Stop()
						pc.Close()
					}()
				}
			// Ctrl channel
			case rt := <-pc.ctrl:
				pc.handleRoudTripMessage(rt)

			// Cancel channel
			case <-ctx.Done():
				fmt.Printf("Controller done\n")
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
		wg.Wait()
	}

	pc.snk.Close()
	pc.src.Close()
	return nil
}

func (pc *PipelineController) Close() {
	pc.cancel = nil
	pc.src = nil
	pc.snk = nil
	pc.sm = nil
	pc.configs = nil
	pc.ctrl = nil
	pc.done = 0
	pc.total = 0
	pc.stats = nil
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
