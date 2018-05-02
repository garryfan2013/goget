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
	Create         func() (Controller, error)
}

var (
	Factories = []ControllerFactory{
		ControllerFactory{
			ControllerType: MultiTaskType,
			Create: func() (Controller, error) {
				var c Controller
				var err error
				c, err = multi_task.NewMultiTaskController()

				return c, err
			}}}
)

func NewController(ct int) (Controller, error) {
	for _, ci := range Factories {
		if ci.ControllerType == ct {
			return ci.Create()
		}
	}

	return nil, errors.New("Cannot find requested controller type")
}
