package file

import (
	"fmt"
	"os"
)

type LocalFileWriter struct {
	fp       *os.File
	FilePath string
}

func NewLocalFileWriter() interface{} {
	return new(LocalFileWriter)
}

func (w *LocalFileWriter) Open(path string) error {
	fp, err := os.Create(path)
	if err != nil {
		return err
	}

	w.fp = fp
	w.FilePath = path
	return nil
}

func (w *LocalFileWriter) WriteAt(data []byte, offset int64) (int, error) {
	if w.fp == nil {
		return 0, fmt.Errorf("File %s not ready, cant write to it", w.FilePath)
	}

	n, err := w.fp.WriteAt(data, offset)
	if err != nil {
		return n, err
	}

	return n, nil
}

func (w *LocalFileWriter) Close() error {
	return w.fp.Close()
}
