package controller

import (
	"errors"

	"github.com/garryfan2013/goget/sink"
	"github.com/garryfan2013/goget/source"
)

const (
	SchemePipeline = "pipeline"
)

var (
	ctors = make(map[string]Creator)
)

func Register(c Creator) {
	ctors[c.Scheme()] = c
}

type Creator interface {
	Create() (ProgressController, error)
	Scheme() string
}

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

func GetProgressController(scheme string) (ProgressController, error) {
	ctor, exists := ctors[scheme]
	if !exists {
		return nil, errors.New("Unsupported controller scheme")
	}

	return ctor.Create()
}
