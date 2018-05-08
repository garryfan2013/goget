package http

import (
	"fmt"
	"io"
	"net/http"
	"strconv"
)

const (
	HeaderKeyContentLength = "Content-Length"
	HeaderKeyRange         = "Range"
)

type HttpCrawler struct {
	FileUrl string
	c       *http.Client
}

func NewHttpCrawler() interface{} {
	return new(HttpCrawler)
}

func (h *HttpCrawler) Open(url string) error {
	h.c = &http.Client{}
	h.FileUrl = url
	return nil
}

func (h *HttpCrawler) SetConfig(key string, value string) {

}

func (h *HttpCrawler) GetFileSize() (int64, error) {
	resp, err := h.c.Head(h.FileUrl)
	if err != nil {
		return 0, err
	}

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("Get Header for %s failed: StatusCode = %d", h.FileUrl, resp.StatusCode)
	}

	str := resp.Header.Get(HeaderKeyContentLength)
	fmt.Printf("Content-Length: %s\n", str)
	len, err := strconv.Atoi(str)
	if err != nil {
		return 0, err
	}

	return int64(len), nil
}

func (h *HttpCrawler) GetFileBlock(offset int64, size int64) (io.ReadCloser, error) {
	req, err := http.NewRequest(http.MethodGet, h.FileUrl, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Add(HeaderKeyRange, fmt.Sprintf("bytes=%d-%d", offset, offset+size-1))
	resp, err := h.c.Do(req)
	if err != nil {
		return nil, err
	}

	return resp.Body, err
}

func (h *HttpCrawler) Close() {

}
