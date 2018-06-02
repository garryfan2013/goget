package sink

import (
	"fmt"
	"os"
)

type FileStreamWriter struct {
	fp   *os.File
	path string
}

func NewFileStreamWriter() interface{} {
	return new(FileStreamWriter)
}

func (fs *FileStreamWriter) Open(path string) error {
	fp, err := os.Create(path)
	if err != nil {
		return err
	}

	fs.fp = fp
	fs.path = path
	return nil
}

func (fs *FileStreamWriter) WriteAt(data []byte, offset int64) (int, error) {
	if fs.fp == nil {
		return 0, fmt.Errorf("File %s not ready, cant write to it", fs.path)
	}

	n, err := fs.fp.WriteAt(data, offset)
	if err != nil {
		return n, err
	}

	return n, nil
}

func (fs *FileStreamWriter) Close() error {
	return fs.fp.Close()
}
