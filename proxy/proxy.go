package proxy

import (
	"errors"
	_ "fmt"
)

const (
	ProxyLocal = "localProxy"
	ProxyRPC   = "rpcProxy"
	ProxyREST  = "restProxy"
)

var (
	builders = make(map[string]Builder)

	ErrJobNotExists = errors.New("Job not exists")
)

type Stats struct {
	Size int64
	Done int64
}

type JobInfo struct {
	Id   string
	Url  string
	Path string
}

type ProxyManager interface {
	Add(url string, path string, username string, passwd string, cnt int) (*JobInfo, error)

	Get(id string) (*JobInfo, error)

	GetAll() ([]*JobInfo, error)

	Progress(id string) (*Stats, error)

	Start(id string) error

	Stop(id string) error

	Delete(id string) error
}

type Builder interface {
	Build() (ProxyManager, error)
	Name() string
}

func Register(b Builder) {
	builders[b.Name()] = b
}

func GetProxyManager(name string) (ProxyManager, error) {
	b, exists := builders[name]
	if !exists {
		return nil, errors.New("Unsupported proxy name")
	}

	return b.Build()
}
