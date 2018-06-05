package source

import (
	"context"
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

func Get(scheme string) Creator {
	if c, ok := ctors[scheme]; ok {
		return c
	}
	return nil
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
