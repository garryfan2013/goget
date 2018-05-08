package manage

import (
	"errors"

	"github.com/garryfan2013/goget/client"
	"github.com/garryfan2013/goget/manage/multi_task"
	"github.com/garryfan2013/goget/record"
)

const (
	MultiTaskType = 0
)

type Controller interface {
	Open(c client.Crawler, h record.Handler) error
	SetConfig(key string, value string)
	Start() error
	Close()
}

type ControllerFactory struct {
	ControllerType int
	Create         func() interface{}
}

var (
	Factories = []ControllerFactory{
		ControllerFactory{
			ControllerType: MultiTaskType,
			Create:         multi_task.NewMultiTaskController}}
)

func NewController(ct int) (Controller, error) {
	for _, ci := range Factories {
		if ci.ControllerType == ct {
			var c Controller = ci.Create().(Controller)
			return c, nil
		}
	}

	return nil, errors.New("Cannot find requested controller type")
}
