package util

import (
	"fmt"
	"io"
	"testing"
)

func TestWriteRead(t *testing.T) {

	bsize := int64(1024 * 2)
	sb := NewStaticBuffer(bsize)

	// test Cap()
	if sb.Cap() != bsize {
		t.Errorf("Wrong cap %d, expected %d\n", sb.Cap(), bsize)
	}

	// test normal write
	var p []byte = []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	_, err := sb.Write(p)
	if err != nil {
		t.Error(err)
	}

	if sb.W != int64(len(p)) {
		t.Errorf("Write failed ,%d != %d\n", sb.W, len(p))
	}

	fmt.Println(p)
	// Output: [1 2 3 4 5 6 7 8 9 10]

	//test normal read
	p1 := make([]byte, 8)
	n, err := sb.Read(p1)
	if err != nil {
		t.Error(err)
	}

	if n != len(p1) {
		t.Error("Read return wrong bytes")
	}

	n, err = sb.Read(p1)
	if err != nil {
		t.Error(err)
	}

	if n != len(p)-len(p1) {
		t.Error("Read return bytes not amount to expected")
	}

	// Read empty buffer
	n, err = sb.Read(p1)
	if err != io.EOF {
		t.Error("Should get EOF")
	}

	// test write more than cap
	sb.Reset()
	p2 := make([]byte, sb.Cap()+512)
	n, err = sb.Write(p2)
	if int64(n) != sb.Cap() {
		t.Errorf("write bytes %d not amount to expected %d\n", n, sb.Cap())
	}

	if err == nil {
		t.Error("Should be ErrNotEnoughSpace")
	}

	if err != ErrNotEnoughSpace {
		t.Error(err)
	}
}
