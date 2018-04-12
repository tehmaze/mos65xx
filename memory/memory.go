// Package memory implements access to 16-bit address space.
package memory

import (
	"fmt"
	"io/ioutil"
)

const zeroBlockSize = 128

// Blank memory always returns the same value
type Blank uint8

func (mem Blank) Fetch(_ uint16) uint8 { return uint8(mem) }
func (Blank) Store(_ uint16, _ uint8)  {}

func (mem Blank) String() string {
	return fmt.Sprintf("%#02x", uint8(mem))
}

// RAM is Rendom Access Memory.
type RAM []uint8

// New creates new RAM.
func New(size int) *RAM {
	mem := make(RAM, size)
	return &mem
}

// Fetch a byte at addr.
func (mem RAM) Fetch(addr uint16) uint8 {
	return mem[addr]
}

// Store a byte at addr.
func (mem *RAM) Store(addr uint16, value uint8) {
	(*mem)[addr] = value
}

// Reset RAM with the provided zero value overwriting the current memory.
func (mem *RAM) Reset(zero uint8) *RAM {
	b := make([]uint8, zeroBlockSize)
	for i := range b {
		b[i] = zero
	}
	for i := 0; i < len(*mem); i += zeroBlockSize {
		copy((*mem)[i:], b)
	}
	return mem
}

func (mem RAM) String() string {
	return fmt.Sprintf("%s RAM", sizeOf(len(mem)))
}

// ROM is Read-Only Memory.
type ROM []uint8

// Load a new ROM.
func Load(name string) (ROM, error) {
	b, err := ioutil.ReadFile(name)
	if err != nil {
		return nil, err
	}
	return ROM(b), nil
}

// Fetch a byte at addr.
func (mem ROM) Fetch(addr uint16) uint8 {
	return mem[addr]
}

// Store is a no-op.
func (ROM) Store(_ uint16, _ uint8) {}

func (mem ROM) String() string {
	return fmt.Sprintf("%s ROM", sizeOf(len(mem)))
}

func sizeOf(l int) string {
	switch {
	case l >= 8192:
		return fmt.Sprintf("%dkB", l>>10)
	default:
		return fmt.Sprintf("%dB", l)
	}
}

// Interface checks
var (
	_ Memory = (*RAM)(nil)
	_ Memory = (*ROM)(nil)
)
