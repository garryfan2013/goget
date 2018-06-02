package sink

import (
	"errors"
	"io"
)

const (
	LocalFileType = 0
)

type StreamWriter interface {
	Open(path string) error
	io.WriterAt
	Close() error
}

type StreamWriterFactory struct {
	Type   int
	Create func() interface{}
}

var (
	Factories = []StreamWriterFactory{
		StreamWriterFactory{
			Type:   LocalFileType,
			Create: NewFileStreamWriter}}
)

func NewStreamWriter(t int) (StreamWriter, error) {
	if t != LocalFileType {
		return nil, errors.New("Illegal StreamWriter type")
	}
	var sw StreamWriter = Factories[t].Create().(StreamWriter)
	return sw, nil
}
