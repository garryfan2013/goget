package store

import (
	"errors"
)

const (
	StoreBoltDB = "BoltDB"
)

var (
	builders = make(map[string]Builder)
)

type Storer interface {
	Open() error

	CreateBucket(bucket string) error

	Put(bucket string, key string, value []byte) error

	Get(bucket string, key string) ([]byte, error)

	GetSwap(bucket string, key string, fn func([]byte) ([]byte, error)) error

	Delete(bucket string, key string) error

	ForEach(bucket string, handle func(k string, v []byte) error) error

	Close() error
}

type Builder interface {
	Build() (Storer, error)

	Name() string
}

func Register(b Builder) {
	builders[b.Name()] = b
}

func GetStorer(name string) (Storer, error) {
	b, exists := builders[name]
	if !exists {
		return nil, errors.New("Unsupported storer name")
	}

	return b.Build()
}
