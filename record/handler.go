package record

import (
	"errors"
	"io"

	"github.com/garryfan2013/goget/record/file"
)

const (
	LocalFileType = 0
)

type Handler interface {
	Open(path string) error
	io.WriterAt
	Close() error
}

type HandlerFactory struct {
	HandlerType int
	Create      func() interface{}
}

var (
	Factories = []HandlerFactory{
		HandlerFactory{
			HandlerType: LocalFileType,
			Create:      file.NewLocalFileWriter}}
)

func NewHandler(t int) (Handler, error) {
	if t != LocalFileType {
		return nil, errors.New("Illegal handler type")
	}
	var i Handler = Factories[t].Create().(Handler)
	return i, nil
}
