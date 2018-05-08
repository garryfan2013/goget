package http

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
)

const (
	HeaderKeyContentLength = "Content-Length"
	HeaderRange            = "Range"
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

func (h *HttpCrawler) GetFileSize() (int, error) {
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

	return len, nil
}

func (h *HttpCrawler) GetFileBlock(offset int, size int) ([]byte, error) {
	req, err := http.NewRequest(http.MethodGet, h.FileUrl, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Add(HeaderRange, fmt.Sprintf("bytes=%d-%d", offset, offset+size-1))
	resp, err := h.c.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	return body, err
}

func (h *HttpCrawler) Close() {

}
