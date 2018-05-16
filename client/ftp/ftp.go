package ftp

import (
	"context"
	"fmt"
	"io"
	"net/url"

	"github.com/garryfan2013/goget/config"
	"github.com/jlaffaye/ftp"
)

const (
	AnonymousUser   = "anonymous"
	AnonymousPasswd = "anonymous"
)

type FtpCrawler struct {
	Url     *url.URL
	Configs map[string]string
}

func NewFtpCrawler() interface{} {
	return new(FtpCrawler)
}

func (fc *FtpCrawler) Open(u string) error {
	fc.Configs = make(map[string]string)
	purl, err := url.Parse(u)
	if err != nil {
		return err
	}

	fc.Url = purl
	/*
		In case of "ftp://[username[:passwd]]@ipaddr:port"
		Set the corresponding user and password to the configs
		The password of UserInfo is not mandatory, so must needs test
	*/
	if fc.Url.User != nil {
		fc.SetConfig(config.KeyUserName, fc.Url.User.Username())
		if passwd, set := fc.Url.User.Password(); set {
			fc.SetConfig(config.KeyPasswd, passwd)
		}
	}

	return nil
}

func (fc *FtpCrawler) SetConfig(key string, value string) {
	fc.Configs[key] = value
}

func (fc *FtpCrawler) GetFileSize(ctx context.Context) (int64, error) {
	conn, err := ftp.Dial(fc.Url.Host)
	if err != nil {
		return 0, err
	}

	defer func() {
		if err := conn.Quit(); err != nil {
			fmt.Println(err)
		}
	}()

	fmt.Printf("%s:%s\n", fc.Configs[config.KeyUserName], fc.Configs[config.KeyPasswd])
	if err = conn.Login(fc.Configs[config.KeyUserName], fc.Configs[config.KeyPasswd]); err != nil {
		return 0, err
	}

	defer func() {
		if err := conn.Logout(); err != nil {
			fmt.Println(err)
		}
	}()

	size, err := conn.FileSize(fc.Url.Path)
	if err != nil {
		return 0, err
	}

	return size, nil
}

func (fc *FtpCrawler) GetFileBlock(ctx context.Context, offset int64, size int64) (io.ReadCloser, error) {
	conn, err := ftp.Dial(fc.Url.Host)
	if err != nil {
		return nil, err
	}

	if err = conn.Login(fc.Configs[config.KeyUserName], fc.Configs[config.KeyPasswd]); err != nil {
		conn.Quit()
		return nil, err
	}

	resp, err := conn.RetrFrom(fc.Url.Path, uint64(offset))
	if err != nil {
		return nil, err
	}

	var rc io.ReadCloser = NewResponseReadCloser(resp, conn)
	return rc, nil
}

func (fc *FtpCrawler) Close() {

}

type ResponseReadCloser struct {
	R *ftp.Response
	C *ftp.ServerConn
}

func NewResponseReadCloser(r *ftp.Response, c *ftp.ServerConn) *ResponseReadCloser {
	return &ResponseReadCloser{r, c}
}

func (rrc *ResponseReadCloser) Read(buf []byte) (int, error) {
	return rrc.R.Read(buf)
}

func (rrc *ResponseReadCloser) Close() error {
	rrc.R.Close()
	rrc.C.Logout()
	rrc.C.Quit()

	return nil
}
