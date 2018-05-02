package record

import (
	"errors"

	"github.com/garryfan2013/goget/record/file"
)

const (
	LocalFileType = 0
)

type Handler interface {
	Open(path string) error
	WriteAt(data []byte, offset int, size int) (int, error)
	Close() error
}

type HandlerFactory struct {
	HandlerType int
	Create      func() (Handler, error)
}

var (
	Factories = []HandlerFactory{
		HandlerFactory{
			HandlerType: LocalFileType,
			Create: func() (Handler, error) {
				var h Handler
				var err error
				h, err = file.NewLocalFileWriter()
				return h, err
			}}}
)

func NewHandler(t int) (Handler, error) {
	if t != LocalFileType {
		return nil, errors.New("Illegal handler type")
	}
	return Factories[t].Create()
}
