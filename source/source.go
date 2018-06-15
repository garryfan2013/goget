package source

import (
	"context"
	"errors"
	"io"
)

const (
	SchemeHTTP  = "http"
	SchemeHTTPS = "https"
	SchemeFTP   = "ftp"
)

var (
	ctors = make(map[string]Creator)
)

func Register(c Creator) {
	ctors[c.Scheme()] = c
}

type Creator interface {
	Create() (StreamReader, error)
	Scheme() string
}

type StreamReader interface {
	Open(url string) error

	Size(ctx context.Context) (int64, error)

	SetConfig(key string, value string)

	Get(ctx context.Context, offset int64, size int64) (io.ReadCloser, error)

	Close()
}

func GetStreamReader(scheme string) (StreamReader, error) {
	ctor, exists := ctors[scheme]
	if !exists {
		return nil, errors.New("Unsupported source scheme")
	}

	return ctor.Create()
}
