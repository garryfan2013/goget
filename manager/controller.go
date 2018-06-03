package manager

import (
	"errors"

	"github.com/garryfan2013/goget/sink"
	"github.com/garryfan2013/goget/source"
)

const (
	MultiTaskType = 0
)

type Controller interface {
	Open(src source.StreamReader, sink sink.StreamWriter) error
	SetConfig(key string, value string)
	Start() error
	Stop() error
	Close()
}

type Stats struct {
	Size int64
	Done int64
}

type Progresser interface {
	Progress() (*Stats, error)
}

type ProgressController interface {
	Controller
	Progresser
}

type ControllerFactory struct {
	ControllerType int
	Create         func() interface{}
}

var (
	Factories = []ControllerFactory{
		ControllerFactory{
			ControllerType: MultiTaskType,
			Create:         NewPipelineController}}
)

func NewController(ct int) (ProgressController, error) {
	for _, ci := range Factories {
		if ci.ControllerType == ct {
			var c ProgressController = ci.Create().(ProgressController)
			return c, nil
		}
	}

	return nil, errors.New("Cannot find requested controller type")
}
