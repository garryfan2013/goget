package ftp

import (
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

func (fc *FtpCrawler) GetFileSize() (int, error) {
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

	return int(size), nil
}

func (fc *FtpCrawler) GetFileBlock(offset int, size int) ([]byte, error) {
	conn, err := ftp.Dial(fc.Url.Host)
	if err != nil {
		return nil, err
	}

	defer conn.Quit()

	if err = conn.Login(fc.Configs[config.KeyUserName], fc.Configs[config.KeyPasswd]); err != nil {
		return nil, err
	}

	defer conn.Logout()

	resp, err := conn.RetrFrom(fc.Url.Path, uint64(offset))
	if err != nil {
		return nil, err
	}

	defer resp.Close()

	buf := make([]byte, size)
	var bytesRead, total = 0, size
	for bytesRead < total {
		var rb []byte = buf[bytesRead:]
		n, err := resp.Read(rb)
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}

		bytesRead = bytesRead + n
	}

	return buf, nil
}

func (fc *FtpCrawler) Close() {

}
