package client

import (
	"context"
	"errors"
	"io"

	"github.com/garryfan2013/goget/client/ftp"
	"github.com/garryfan2013/goget/client/http"
)

const (
	HttpProtocol = 0
	FtpProtocol  = 1
)

type Crawler interface {
	Open(url string) error
	GetFileSize(ctx context.Context) (int64, error)
	SetConfig(key string, value string)
	GetFileBlock(ctx context.Context, offset int64, size int64) (io.ReadCloser, error)
	Close()
}

type CrawlerFactory struct {
	Protocol int
	Create   func() interface{}
}

var (
	Factories []CrawlerFactory = []CrawlerFactory{
		CrawlerFactory{
			Protocol: HttpProtocol,
			Create:   http.NewHttpCrawler},

		CrawlerFactory{
			Protocol: FtpProtocol,
			Create:   ftp.NewFtpCrawler}}
)

func NewCrawler(proto int) (Crawler, error) {
	if proto < HttpProtocol {
		return nil, errors.New("Unsupported protocol")
	}

	if proto > FtpProtocol {
		return nil, errors.New("Unsupported protocol")
	}

	f := Factories[proto]
	var c Crawler = f.Create().(Crawler)
	return c, nil
}
