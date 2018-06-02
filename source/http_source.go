package source

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strconv"
)

const (
	HeaderKeyContentLength = "Content-Length"
	HeaderKeyRange         = "Range"
)

type HttpStreamReader struct {
	url  string
	clnt *http.Client
}

func NewHttpStreamReader() interface{} {
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
