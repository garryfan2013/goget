package manager

import (
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

func Get(scheme string) Creator {
	if c, ok := ctors[scheme]; ok {
		return c
	}
	return nil
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
