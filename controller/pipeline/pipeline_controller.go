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
	CTRL_MSG_GET_STATS = 0
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
	Source source.StreamReader
	Tasks  []ReaderTask
	P      int
	Notify chan<- int64
}

type ReaderTask struct {
	Offset int64
	Size   int64
}

func NewUrlRequestHandler(s source.StreamReader, n int, notify chan<- int64) *UrlRequestHandler {
	return &UrlRequestHandler{
		Source: s,
		Tasks:  make([]ReaderTask, n),
		P:      0,
		Notify: notify}
}

func (h *UrlRequestHandler) Handle(ctx context.Context, arg interface{}) (interface{}, error) {
	_, ok := arg.(string)
	if !ok {
		fmt.Printf("Unexpected data type %t\n", arg)
		return nil, errors.New("Unexpected data type")
	}

	total, err := h.Source.Size(ctx)
	if err != nil {
		return nil, err
	}

	// Inform the controller of the whole file size
	h.Notify <- total

	cnt := len(h.Tasks)
	size := total / int64(cnt)

	for i := 0; i < cnt; i++ {
		h.Tasks[i].Offset = int64(i) * size
		h.Tasks[i].Size = size
		if i == cnt-1 {
			h.Tasks[i].Size = h.Tasks[i].Size + total%int64(cnt)
		}
	}

	return h, nil
}

func (h *UrlRequestHandler) Next(ctx context.Context) (interface{}, error) {
	if h.P < len(h.Tasks) {
		var tp *ReaderTask = &h.Tasks[h.P]
		h.P += 1
		return tp, nil
	}

	return nil, util.ErrIteratorEOF
}

// Reader-AsyncExecutor
// Input(ReaderTask): the specific offset and size of a target file block
// Output([]byte): stream of block data
type ReaderTaskHandler struct {
	Source     source.StreamReader // Source crawler
	RC         io.ReadCloser       // io interface obtained from crawler
	BufferPool *sync.Pool          // This BufferPool will live as long as the ReaderTaskHandler
	Offset     int64               // Block offset in the whole file
	Size       int64               // Block size
	Left       int64               // Block data remained to read
}

// It's the writer's duty to put the buffer back to the pool
type WriterTask struct {
	Buf        *util.StaticBuffer
	BufferPool *sync.Pool
	Offset     int64
}

func NewReaderTaskHandler(s source.StreamReader, bs int64) *ReaderTaskHandler {
	return &ReaderTaskHandler{
		Source: s,
		BufferPool: &sync.Pool{
			New: func() interface{} {
				t := new(WriterTask)
				t.Buf = util.NewStaticBuffer(bs)
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

	h.Offset = t.Offset
	h.Size = t.Size
	h.Left = t.Size

	rc, err := h.Source.Get(ctx, h.Offset, h.Size)
	if err != nil {
		return nil, err
	}

	h.RC = rc
	return h, nil
}

// Implement the iterator interface
func (h *ReaderTaskHandler) Next(ctx context.Context) (interface{}, error) {
	task := h.BufferPool.Get().(*WriterTask)

	task.Buf.Reset()
	task.Offset = h.Offset + h.Size - h.Left
	task.BufferPool = h.BufferPool

	n, err := task.Buf.ReadFrom(h.RC)

	if n > 0 {
		h.Left -= int64(n)
	}

	if err == io.EOF {
		if h.Left != 0 {
			panic("ReaderHandler.Next reach EOF, but got not enough data")
		}
		return task, util.ErrIteratorEOF
	}

	if err != nil {
		h.BufferPool.Put(task)
		return task, err
	}

	if h.Left < 0 {
		panic("ReaderTaskHandler.Next: left negative!")
	}

	if h.Left == 0 {
		return task, util.ErrIteratorEOF
	}

	return task, nil
}

// Implement the Closer interface
func (h *ReaderTaskHandler) Close() error {
	h.RC.Close()
	return nil
}

// FileWriter-AsyncExecutor
// Input([]byte): the actual data block
// Output(bool): status
type WriterTaskHandler struct {
	Sink sink.StreamWriter
}

func NewWriterTaskHandler(s sink.StreamWriter) *WriterTaskHandler {
	return &WriterTaskHandler{Sink: s}
}

func (h *WriterTaskHandler) Handle(ctx context.Context, arg interface{}) (interface{}, error) {
	t, ok := arg.(*WriterTask)
	if !ok {
		fmt.Printf("Unexpected data type %t\n", arg)
		return nil, errors.New("Unexpected data type")
	}

	defer t.BufferPool.Put(t)

	wa := h.Sink.(io.WriterAt)
	w := util.NewOffsetWriter(wa, t.Offset)
	n, err := t.Buf.WriteTo(w)
	if err != nil {
		return nil, err
	}

	return n, nil
}

// Sync request-response module
type Message struct {
	cmd  int
	data interface{}
}

type Roundtrip struct {
	M Message
	R chan *Message
}

// PiplelineController constructor
type PipelineController struct {
	Configs  map[string]string
	Source   source.StreamReader
	Sink     sink.StreamWriter
	CtrlChan chan *Roundtrip
	Cancel   context.CancelFunc
}

func newPipelineController() interface{} {
	return new(PipelineController)
}

func (pc *PipelineController) Open(c source.StreamReader, h sink.StreamWriter) error {
	pc.Configs = make(map[string]string)
	pc.Source = c
	pc.Sink = h
	pc.CtrlChan = make(chan *Roundtrip)

	return nil
}

func (pc *PipelineController) SetConfig(key string, value string) {
	pc.Configs[key] = value
}

var cancel context.CancelFunc

func (pc *PipelineController) Start() error {
	if pc.Source == nil {
		return errors.New("No source set yet")
	}

	if pc.Sink == nil {
		return errors.New("No sink set yet")
	}

	url, exists := pc.Configs[config.KeyRemoteUrl]
	if exists == false {
		return errors.New("Source url not set!")
	}

	if err := pc.Source.Open(url); err != nil {
		return err
	}

	if user, exists := pc.Configs[config.KeyUserName]; exists {
		pc.Source.SetConfig(config.KeyUserName, user)
	}

	if passwd, exists := pc.Configs[config.KeyPasswd]; exists {
		pc.Source.SetConfig(config.KeyPasswd, passwd)
	}

	path, exists := pc.Configs[config.KeyLocalPath]
	if exists == false {
		return errors.New("Sink path not set!")
	}

	taskCountStr, exists := pc.Configs[config.KeyTaskCount]
	if exists == false {
		return errors.New("TaskCount not set!")
	}
	taskCount, err := strconv.Atoi(taskCountStr)
	if err != nil {
		return err
	}

	if err := pc.Sink.Open(path); err != nil {
		return err
	}

	ctx := context.Background()
	ctx, pc.Cancel = context.WithCancel(ctx)

	// Prepare the start channel and notify channel
	startCh := make(chan interface{})
	notifyCh := make(chan int64)

	urlHandler := NewUrlRequestHandler(pc.Source, taskCount, notifyCh)
	urlEx := util.NewAsyncExecutor(urlHandler)

	crawlerExes := make([]util.Executor, taskCount)
	for i := 0; i < taskCount; i++ {
		crawlerHandler := NewReaderTaskHandler(pc.Source, PoolBufferAllocSize)
		crawlerExes[i] = util.NewAsyncExecutor(crawlerHandler)
	}

	fanOutEx := util.NewFanOutAsyncExecutor()
	fanInEx := util.NewFanInAsyncExecutor()

	writerHandler := NewWriterTaskHandler(pc.Sink)
	writerEx := util.NewAsyncExecutor(writerHandler)

	// Start the url executor
	taskCh, urlErrCh := urlEx.Run(ctx, startCh)

	// Start the fanOut executor
	fanOutChs, fanOutErrChs := fanOutEx.Run(ctx, taskCh, crawlerExes)

	// Start the fanIn executor
	fanInCh, fanInErrCh := fanInEx.Run(ctx, fanOutChs, fanOutErrChs)

	// Start the writer executor
	counterCh, writerErrCh := writerEx.Run(ctx, fanInCh)

	var written, total int64

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
			case total = <-notifyCh:
				/*if written == total {
					fmt.Printf("Progress: Done\n")
					return
				}*/

			// Stats updated
			case ret := <-counterCh:
				cnt, ok := ret.(int)
				if !ok {
					panic("counterCh ret cant type switch to int")
				}

				written += int64(cnt)
				/*if written == total {
					return
				}*/

			// Ctrl channel
			case msg := <-pc.CtrlChan:
				switch msg.M.cmd {
				case CTRL_MSG_GET_STATS:
					msg.R <- &Message{
						cmd: msg.M.cmd,
						data: &controller.Stats{
							Size: total,
							Done: written,
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
	if pc.Cancel != nil {
		pc.Cancel()
	}

	return nil
}

func (mc *PipelineController) Close() {
	mc.Sink.Close()
	mc.Source.Close()
}

func (mc *PipelineController) Progress() (*controller.Stats, error) {
	ch := make(chan *Message, 1)
	mc.CtrlChan <- &Roundtrip{
		M: Message{cmd: CTRL_MSG_GET_STATS},
		R: ch,
	}

	msg := <-ch
	stats := msg.data.(*controller.Stats)
	return stats, nil
}
