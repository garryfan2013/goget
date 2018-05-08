package util

import (
	"io"
)

type OffsetWriter struct {
	W      io.WriterAt //WriterAt interface
	Offset int64       //Write offset
}

func NewOffsetWriter(wa io.WriterAt, offset int64) *OffsetWriter {
	return &OffsetWriter{wa, offset}
}

func (ow *OffsetWriter) Write(p []byte) (int, error) {
	n, err := ow.W.WriteAt(p, ow.Offset)
	if n > 0 {
		ow.Offset += int64(n)
	}

	return n, err
}
