package file

import (
	"fmt"
	"os"

	"github.com/garryfan2013/goget/sink"
)

func init() {
	sink.Register(&FileStreamWriterCreator{})
}

type FileStreamWriterCreator struct{}

func (*FileStreamWriterCreator) Create() (sink.StreamWriter, error) {
	sw := newFileStreamWriter().(sink.StreamWriter)
	return sw, nil
}

func (*FileStreamWriterCreator) Scheme() string {
	return sink.SchemeLocalFile
}

type FileStreamWriter struct {
	fp   *os.File
	path string
}

func newFileStreamWriter() interface{} {
	return new(FileStreamWriter)
}

func (fs *FileStreamWriter) Open(path string) error {
	fp, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE, 0755)
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
	if fs.fp != nil {
		return fs.fp.Close()
	}
	return nil
}
