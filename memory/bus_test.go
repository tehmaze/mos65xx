package memory

import (
	"io"
	"testing"
)

func TestReaderAt(t *testing.T) {
	mem := New(0x400).Reset(0xff)
	rdr := ReaderAt{mem}

	if n, err := rdr.ReadAt(make([]byte, 8), 0x100); err != nil {
		t.Fatal(err)
	} else if n != 8 {
		t.Fatalf("expected to read 8 bytes, got %d", n)
	}

	if n, err := rdr.ReadAt(make([]byte, 16), 0x3f0); err != nil {
		t.Fatal(err)
	} else if n != 16 {
		t.Fatalf("expected to read 8 bytes, got %d", n)
	}

	if _, err := rdr.ReadAt(make([]byte, 1), 0xffff); err != io.EOF {
		t.Fatalf("expected EOF; got %v", err)
	}
	if _, err := rdr.ReadAt(make([]byte, 1), 0x1ffff); err != io.ErrShortBuffer {
		t.Fatalf("expected ErrShortBuffer; got %v", err)
	}
}
