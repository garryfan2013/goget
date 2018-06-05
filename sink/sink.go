package sink

import (
	"io"
)

const (
	SchemeLocalFile = "LocalFile"
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
	Create() (StreamWriter, error)
	Scheme() string
}

type StreamWriter interface {
	Open(path string) error
	io.WriterAt
	Close() error
}
