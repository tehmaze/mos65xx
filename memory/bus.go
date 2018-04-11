package memory

import (
	"io"
	"math"
)

// Memory implements a 16-bit address bus.
type Memory interface {
	// Fetch a byte
	Fetch(addr uint16) (value uint8)

	// Store a byte
	Store(addr uint16, value uint8)
}

// ReaderAt implements io.ReaderAt on Memory.
type ReaderAt struct {
	Memory
}

// ReadAt reads len(p) bytes into p starting at offset off in the underlying
// input source. It returns the number of bytes read (0 <= n <= len(p)) and any
// error encountered.
func (bus ReaderAt) ReadAt(p []byte, off int64) (n int, err error) {
	if off > math.MaxUint16 {
		return 0, io.ErrShortBuffer
	}

	var (
		size = int64(len(p))
		addr int64
	)
	if off+size > math.MaxUint16 {
		size = math.MaxUint16 - off
		err = io.EOF
	}
	for ; addr < size; addr++ {
		p[addr] = bus.Fetch(uint16(addr + off))
	}
	return int(size), err
}
