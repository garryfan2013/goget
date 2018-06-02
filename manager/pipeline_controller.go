package manager

import (
	"context"
	"errors"
	"fmt"
	"io"
	"sync"

	"github.com/garryfan2013/goget/config"
	"github.com/garryfan2013/goget/sink"
	"github.com/garryfan2013/goget/source"
	"github.com/garryfan2013/goget/util"
	"github.com/garryfan2013/goget/util/pipeline"
)

const (
	PoolBufferAllocSize = 128 * 1024
)

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

	return nil, pipeline.ErrIteratorEOF
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
		return task, pipeline.ErrIteratorEOF
	}

	if err != nil {
		h.BufferPool.Put(task)
		return task, err
	}

	if h.Left < 0 {
		panic("ReaderTaskHandler.Next: left negative!")
	}

	if h.Left == 0 {
		return task, pipeline.ErrIteratorEOF
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

// PiplelineController constructor
type PipelineController struct {
	Configs map[string]string
	Source  source.StreamReader
	Sink    sink.StreamWriter
}

func NewPipelineController() interface{} {
	return new(PipelineController)
}

func (pc *PipelineController) Open(c source.StreamReader, h sink.StreamWriter) error {
	pc.Configs = make(map[string]string)
	pc.Source = c
	pc.Sink = h
	return nil
}

func (pc *PipelineController) SetConfig(key string, value string) {
	pc.Configs[key] = value
}

var wg sync.WaitGroup
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

	if err := pc.Sink.Open(path); err != nil {
		return err
	}

	ctx := context.Background()
	ctx, cancel = context.WithCancel(ctx)

	// Prepare the start channel and notify channel
	startCh := make(chan interface{})
	notifyCh := make(chan int64)

	urlHandler := NewUrlRequestHandler(pc.Source, DefaultTaskCount, notifyCh)
	urlEx := pipeline.NewAsyncExecutor(urlHandler)

	crawlerExes := make([]pipeline.Executor, DefaultTaskCount)
	for i := 0; i < DefaultTaskCount; i++ {
		crawlerHandler := NewReaderTaskHandler(pc.Source, PoolBufferAllocSize)
		crawlerExes[i] = pipeline.NewAsyncExecutor(crawlerHandler)
	}

	fanOutEx := pipeline.NewFanOutAsyncExecutor()
	fanInEx := pipeline.NewFanInAsyncExecutor()

	writerHandler := NewWriterTaskHandler(pc.Sink)
	writerEx := pipeline.NewAsyncExecutor(writerHandler)

	// Start the url executor
	taskCh, urlErrCh := urlEx.Run(ctx, startCh)

	// Start the fanOut executor
	fanOutChs, fanOutErrChs := fanOutEx.Run(ctx, taskCh, crawlerExes)

	// Start the fanIn executor
	fanInCh, fanInErrCh := fanInEx.Run(ctx, fanOutChs, fanOutErrChs)

	// Start the writer executor
	counterCh, writerErrCh := writerEx.Run(ctx, fanInCh)

	var written, total int64
	wg.Add(1)
	go func() {
		for {
			select {
			case e := <-urlErrCh:
				fmt.Printf("urlErrch: %s\n", e.Error())
			case e := <-fanInErrCh:
				fmt.Printf("fanInErrCh: %s\n", e.Error())
			case e := <-writerErrCh:
				fmt.Printf("writerErrCh: %s\n", e.Error())
			case total = <-notifyCh:
				fmt.Printf("Notify file length: %d\n", total)
				if written == total {
					fmt.Printf("Progress: Done\n")
					wg.Done()
				}
			case ret := <-counterCh:
				cnt, ok := ret.(int)
				if !ok {
					panic("counterCh ret cant type switch to int")
				}

				written += int64(cnt)
				fmt.Printf("Progress: %d bytes written\n", written)
				if written == total {
					fmt.Printf("Progress: Done\n")
					wg.Done()
				}
			case <-ctx.Done():
				fmt.Printf("Controller canceled\n")
				return
			}
		}
	}()

	// Start the pipeline
	startCh <- url
	wg.Wait()
	cancel()
	return nil
}

func (pc *PipelineController) Stop() error {
	fmt.Println("Now calling stop...")
	if cancel != nil {
		cancel()
	}

	return nil
}

func (mc *PipelineController) Close() {
	mc.Sink.Close()
	mc.Source.Close()
}
