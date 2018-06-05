package source

import (
	"context"
	"fmt"
	"io"
	"net/url"

	"github.com/jlaffaye/ftp"

	"github.com/garryfan2013/goget/config"
	"github.com/garryfan2013/goget/source"
)

const (
	AnonymousUser   = "anonymous"
	AnonymousPasswd = "anonymous"
)

func init() {
	source.Register(&FtpStreamReaderCreator{})
}

type FtpStreamReaderCreator struct{}

func (*FtpStreamReaderCreator) Create() (source.StreamReader, error) {
	sr := newFtpStreamReader().(source.StreamReader)
	return sr, nil
}

func (*FtpStreamReaderCreator) Scheme() string {
	return SchemeFTP
}

func newFtpStreamReader() interface{} {
	return new(FtpStreamReader)
}

type FtpStreamReader struct {
	url     *url.URL
	configs map[string]string
}

func (fs *FtpStreamReader) Open(u string) error {
	fs.configs = make(map[string]string)
	purl, err := url.Parse(u)
	if err != nil {
		return err
	}

	fs.url = purl
	/*
		In case of "ftp://[username[:passwd]]@ipaddr:port"
		Set the corresponding user and password to the configs
		The password of UserInfo is not mandatory, so must needs test
	*/
	if fs.url.User != nil {
		fs.SetConfig(config.KeyUserName, fs.url.User.Username())
		if passwd, set := fs.url.User.Password(); set {
			fs.SetConfig(config.KeyPasswd, passwd)
		}
	}

	return nil
}

func (fs *FtpStreamReader) SetConfig(key string, value string) {
	fs.configs[key] = value
}

func (fs *FtpStreamReader) Size(ctx context.Context) (int64, error) {
	conn, err := ftp.Dial(fs.url.Host)
	if err != nil {
		return 0, err
	}

	defer func() {
		if err := conn.Quit(); err != nil {
			fmt.Println(err)
		}
	}()

	fmt.Printf("%s:%s\n", fs.configs[config.KeyUserName], fs.configs[config.KeyPasswd])
	if err = conn.Login(fs.configs[config.KeyUserName], fs.configs[config.KeyPasswd]); err != nil {
		return 0, err
	}

	defer func() {
		if err := conn.Logout(); err != nil {
			fmt.Println(err)
		}
	}()

	size, err := conn.FileSize(fs.url.Path)
	if err != nil {
		return 0, err
	}

	return size, nil
}

func (fs *FtpStreamReader) Get(ctx context.Context, offset int64, size int64) (io.ReadCloser, error) {
	conn, err := ftp.Dial(fs.url.Host)
	if err != nil {
		return nil, err
	}

	if err = conn.Login(fs.configs[config.KeyUserName], fs.configs[config.KeyPasswd]); err != nil {
		conn.Quit()
		return nil, err
	}

	resp, err := conn.RetrFrom(fs.url.Path, uint64(offset))
	if err != nil {
		return nil, err
	}

	var rc io.ReadCloser = NewResponseReadCloser(resp, conn)
	return rc, nil
}

func (fs *FtpStreamReader) Close() {

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
