package multi_task

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/garryfan2013/goget/client"
	"github.com/garryfan2013/goget/config"
	"github.com/garryfan2013/goget/record"
)

type TaskInfo struct {
	Offset int
	Size   int
}

type MultiTaskController struct {
	Configs map[string]string
	Tasks   []TaskInfo
	Source  client.Crawler
	Sink    record.Handler
	Notify  chan int
}

const (
	DefaultTaskCount = 5
)

func NewMultiTaskController() (*MultiTaskController, error) {
	return new(MultiTaskController), nil
}

func (mc *MultiTaskController) Open(c client.Crawler, h record.Handler) error {
	mc.Configs = make(map[string]string)
	mc.Source = c
	mc.Sink = h
	mc.Notify = make(chan int)
	return nil
}

func (mc *MultiTaskController) SetConfig(key string, value string) {
	mc.Configs[key] = value
}

func RunTask(task *TaskInfo, crawler client.Crawler, handler record.Handler, notify chan int) {
	fmt.Printf("RunTask: offset = %d size = %d\n", task.Offset, task.Size)

	defer func(n chan<- int) {
		n <- 1
	}(notify)

	blockData, err := crawler.GetFileBlock(task.Offset, task.Size)
	if err != nil {
		fmt.Printf("GetFileBlock failed\n")
		return
	}

	if len(blockData) != task.Size {
		fmt.Printf("The number of bytes read(%d) not amount to the expecting(%d)\n", len(blockData), task.Size)
	}

	n, err := handler.WriteAt(blockData, task.Offset, task.Size)
	if n != task.Size {
		fmt.Printf("The number of bytes written(%d) not amount to the expecting(%d)\n", n, task.Size)
	}

	if err != nil {
		fmt.Printf("Write file failed\n")
		return
	}
}

func (mc *MultiTaskController) Start() error {
	if mc.Source == nil {
		return errors.New("No source set yet")
	}

	if mc.Sink == nil {
		return errors.New("No sink set yet")
	}

	url, exists := mc.Configs[config.KeyRemoteUrl]
	if exists == false {
		return errors.New("Source url not set!")
	}

	if err := mc.Source.Open(url); err != nil {
		return err
	}

	defer mc.Source.Close()

	path, exists := mc.Configs[config.KeyLocalPath]
	if exists == false {
		return errors.New("Sink path not set!")
	}

	if err := mc.Sink.Open(path); err != nil {
		return err
	}

	defer mc.Sink.Close()

	totalSize, err := mc.Source.GetFileSize()
	if err != nil {
		return err
	}

	blockCnt := DefaultTaskCount
	if v, exists := mc.Configs[config.KeyTaskCount]; exists {
		var err error
		fmt.Printf("%s\n", v)
		blockCnt, err = strconv.Atoi(v)
		if err != nil {
			return err
		}
	}

	mc.Tasks = make([]TaskInfo, blockCnt, blockCnt+1)
	blockSize := totalSize / blockCnt

	for i := 0; i < blockCnt; i++ {
		mc.Tasks[i].Offset = i * blockSize
		mc.Tasks[i].Size = blockSize
		go RunTask(&mc.Tasks[i], mc.Source, mc.Sink, mc.Notify)
	}

	if blockSize*blockCnt < totalSize {
		mc.Tasks = append(mc.Tasks, TaskInfo{Offset: blockSize * blockCnt, Size: totalSize % blockCnt})
		blockCnt = len(mc.Tasks)
		go RunTask(&mc.Tasks[blockCnt-1], mc.Source, mc.Sink, mc.Notify)
	}

	for i := 0; i < blockCnt; i++ {
		<-mc.Notify
	}

	return nil
}

func (mc *MultiTaskController) Close() {
	mc.Sink.Close()
	mc.Source.Close()
}
