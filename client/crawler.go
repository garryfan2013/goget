package client

import (
	"errors"

	"github.com/garryfan2013/goget/client/ftp"
	"github.com/garryfan2013/goget/client/http"
)

const (
	HttpProtocol = 0
	FtpProtocol  = 1
)

type Crawler interface {
	Open(url string) error
	GetFileSize() (int, error)
	SetConfig(key string, value string)
	GetFileBlock(offset int, size int) ([]byte, error)
	Close()
}

type CrawlerFactory struct {
	Protocol int
	Create   func() (Crawler, error)
}

var (
	Factories []CrawlerFactory = []CrawlerFactory{
		CrawlerFactory{
			Protocol: HttpProtocol,
			Create: func() (Crawler, error) {
				var c Crawler
				var err error
				c, err = http.NewHttpCrawler()
				return c, err
			}},

		CrawlerFactory{
			Protocol: FtpProtocol,
			Create: func() (Crawler, error) {
				var c Crawler
				var err error
				c, err = ftp.NewFtpCrawler()
				return c, err
			}}}
)

func NewCrawler(proto int) (Crawler, error) {
	if proto < HttpProtocol {
		return nil, errors.New("Unsupported protocol")
	}

	if proto > FtpProtocol {
		return nil, errors.New("Unsupported protocol")
	}

	f := Factories[proto]
	return f.Create()
}
