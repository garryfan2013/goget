package source

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/garryfan2013/goget/source"
)

const (
	HeaderKeyContentLength = "Content-Length"
	HeaderKeyRange         = "Range"
)

func init() {
	source.Register(&HttpStreamReaderCreator{})
	source.Register(&HttpsStreamReaderCreator{})
}

type HttpStreamReaderCreator struct{}

func (*HttpStreamReaderCreator) Create() (source.StreamReader, error) {
	sr := newHttpStreamReader().(source.StreamReader)
	return sr, nil
}

func (*HttpStreamReaderCreator) Scheme() string {
	return source.SchemeHTTP
}

type HttpsStreamReaderCreator struct{}

func (*HttpsStreamReaderCreator) Create() (source.StreamReader, error) {
	sr := newHttpStreamReader().(source.StreamReader)
	return sr, nil
}

func (*HttpsStreamReaderCreator) Scheme() string {
	return source.SchemeHTTPS
}

type HttpStreamReader struct {
	url  string
	clnt *http.Client
}

func newHttpStreamReader() interface{} {
	return new(HttpStreamReader)
}

func (hs *HttpStreamReader) Open(url string) error {
	hs.clnt = &http.Client{}
	hs.url = url
	return nil
}

func (hs *HttpStreamReader) SetConfig(key string, value string) {

}

func (hs *HttpStreamReader) Size(ctx context.Context) (int64, error) {
	req, err := http.NewRequest(http.MethodHead, hs.url, nil)
	if err != nil {
		return 0, err
	}

	req = req.WithContext(ctx)
	resp, err := hs.clnt.Do(req)
	if err != nil {
		return 0, err
	}

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("Get Header for %s failed: StatusCode = %d", hs.url, resp.StatusCode)
	}

	str := resp.Header.Get(HeaderKeyContentLength)
	fmt.Printf("Content-Length: %s\n", str)
	len, err := strconv.Atoi(str)
	if err != nil {
		return 0, err
	}

	return int64(len), nil
}

func (hs *HttpStreamReader) Get(ctx context.Context, offset int64, size int64) (io.ReadCloser, error) {
	req, err := http.NewRequest(http.MethodGet, hs.url, nil)
	if err != nil {
		return nil, err
	}

	req = req.WithContext(ctx)
	req.Header.Add(HeaderKeyRange, fmt.Sprintf("bytes=%d-%d", offset, offset+size-1))
	resp, err := hs.clnt.Do(req)
	if err != nil {
		return nil, err
	}

	return resp.Body, err
}

func (hs *HttpStreamReader) Close() {

}
