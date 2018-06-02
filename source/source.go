package source

import (
	"context"
	"errors"
	"io"
)

const (
	HttpProtocol = 0
	FtpProtocol  = 1
)

type StreamReader interface {
	Open(url string) error
	Size(ctx context.Context) (int64, error)
	SetConfig(key string, value string)
	Get(ctx context.Context, offset int64, size int64) (io.ReadCloser, error)
	Close()
}

type StreamReaderFactory struct {
	Protocol int
	Create   func() interface{}
}

var (
	Factories []StreamReaderFactory = []StreamReaderFactory{
		StreamReaderFactory{
			Protocol: HttpProtocol,
			Create:   NewHttpStreamReader},

		StreamReaderFactory{
			Protocol: FtpProtocol,
			Create:   NewFtpStreamReader}}
)

func NewStreamReader(proto int) (StreamReader, error) {
	if proto < HttpProtocol {
		return nil, errors.New("Unsupported protocol")
	}

	if proto > FtpProtocol {
		return nil, errors.New("Unsupported protocol")
	}

	f := Factories[proto]
	var s StreamReader = f.Create().(StreamReader)
	return s, nil
}
