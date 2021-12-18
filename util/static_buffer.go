package util

import (
	"errors"
	"io"
)

var (
	ErrNegativeRead   = errors.New("util.StaticBuffer: reader returned negative count from Read")
	ErrInvalidWrite   = errors.New("util.StaticBuffer: writer returned invalid count from Write")
	ErrNotEnoughSpace = errors.New("util.StaticBuffer: not enough space left for writting")
)

type StaticBuffer struct {
	Buf []byte // buffer
	W   int64  // start point of writting
	R   int64  // start point of reading
}

func NewStaticBuffer(cap int64) *StaticBuffer {
	return &StaticBuffer{make([]byte, cap), 0, 0}
}

func (sb *StaticBuffer) Read(p []byte) (int, error) {
	if len(p) == 0 {
		return 0, nil
	}

	dataLen := sb.W - sb.R
	if dataLen < 0 {
		panic("StaticBuffer content length negtive")
	}

	if dataLen == 0 {
		return 0, io.EOF
	}

	n := copy(p, sb.Buf[sb.R:sb.W])
	sb.R += int64(n)
	if sb.R == sb.W {
		sb.Reset()
	}

	return n, nil
}

func (sb *StaticBuffer) ReadFrom(r io.Reader) (int64, error) {
	var totalRead int64

	if sb.W == sb.Cap() {
		return 0, nil
	}

	for {
		n, err := r.Read(sb.Buf[sb.W:])
		if n < 0 {
			panic(ErrNegativeRead)
		}

		sb.W += int64(n)
		totalRead += int64(n)

		if err == io.EOF {
			return totalRead, err
		}

		if err != nil {
			return totalRead, err
		}

		if sb.W == sb.Cap() {
			return totalRead, nil
		}
	}
}

func (sb *StaticBuffer) Write(p []byte) (int, error) {
	left := int64(len(sb.Buf)) - sb.W

	if len(p) == 0 {
		return 0, nil
	}

	if left == 0 {
		return 0, ErrNotEnoughSpace
	}

	n := copy(sb.Buf[sb.W:], p)
	sb.W += int64(n)

	if n != len(p) {
		return n, ErrNotEnoughSpace
	}

	return n, nil
}

func (sb *StaticBuffer) WriteTo(w io.Writer) (int64, error) {
	l := int(sb.W - sb.R)
	n, err := w.Write(sb.Buf[sb.R:sb.W])
	if n > l {
		panic(ErrInvalidWrite)
	}

	sb.R += int64(n)

	if err != nil {
		return int64(n), err
	}

	if n != l {
		return int64(n), io.ErrShortWrite
	}

	// All the data's been written to writer
	sb.Reset()
	return int64(n), nil
}

func (sb *StaticBuffer) Cap() int64 {
	return int64(len(sb.Buf))
}

func (sb *StaticBuffer) Size() int64 {
	if sb.R < sb.W {
		panic("Buffer Read point exceeds the boundary")
	}
	return sb.W - sb.R
}

func (sb *StaticBuffer) Reset() {
	sb.W = 0
	sb.R = 0
}

func (sb *StaticBuffer) Close() error {
	return nil
}
