package mos65xx

import (
	"io"
)

// Vectors
const (
	NMIVector   = 0xfffa
	ResetVector = 0xfffc
	IRQVector   = 0xfffe
)

// zeros is empty memory
var zeros = make([]byte, 256)

func init() {
	for i := range zeros {
		zeros[i] = 0xff
	}
}

// AddressMode determines how the CPU will fetch the address
type AddressMode uint8

// Address modes
const (
	Implied AddressMode = iota
	Accumulator
	Immediate
	ZeroPage
	ZeroPageX
	ZeroPageY
	Relative
	Absolute
	AbsoluteX
	AbsoluteY
	Indirect
	IndexedIndirect
	IndirectIndexed
)

var (
	addressModeName = map[AddressMode]string{
		Implied:         "implied",
		Accumulator:     "accumulator",
		Immediate:       "immediate",
		ZeroPage:        "zero-page",
		ZeroPageX:       "zero-page indexed X",
		ZeroPageY:       "zero-page indexed Y",
		Relative:        "relative",
		Absolute:        "absolute",
		AbsoluteX:       "absolute indexed X",
		AbsoluteY:       "absolute indexed Y",
		Indirect:        "indirect",
		IndexedIndirect: "indexed indirect",
		IndirectIndexed: "indirect indexed",
	}
	addressModeCycles = map[AddressMode]int{
		Implied:         2,
		Accumulator:     0,
		Immediate:       2,
		ZeroPage:        3,
		ZeroPageX:       4,
		ZeroPageY:       4,
		Relative:        2, // +1 on branch, +1 if branch to different page
		Absolute:        4,
		AbsoluteX:       4, // +1 on page cross
		AbsoluteY:       4, // +1 on page cross
		Indirect:        0,
		IndexedIndirect: 6,
		IndirectIndexed: 5, // +1 on page cross
	}
)

// Cycles to fetch the operand address
func (mode AddressMode) Cycles() int {
	return addressModeCycles[mode]
}

// FetchPenalty is the cycle penalty for doing page cross or branch on a fetch operation.
func (mode AddressMode) FetchPenalty() int {
	return 0 // TODO
}

// StorePenalty is the cycle penalty for doing page cross on a store operation.
func (mode AddressMode) StorePenalty() int {
	switch mode {
	case AbsoluteX, AbsoluteY, IndirectIndexed:
		return 1
	default:
		return 0
	}
}

func (mode AddressMode) String() string {
	if s, ok := addressModeName[mode]; ok {
		return s
	}
	return "Invalid"
}

type AddressBus interface {
	// Fetch data from the bus
	Fetch(addr uint16) (value uint8)

	// Store data on the bus
	Store(addr uint16, value uint8)

	// Read a portion of memory from the bus
	io.ReaderAt
}

// addressBusMasked is a helper for a CPU with an address bus that supports
// less than 16-bit addresses
type addressBusMasked struct {
	bus  AddressBus
	mask uint16
}

func (b addressBusMasked) Fetch(addr uint16) (value uint8) { return b.bus.Fetch(addr & b.mask) }
func (b addressBusMasked) Store(addr uint16, value uint8)  { b.bus.Store(addr&b.mask, value) }

// RAM is Random-Access Memory
type RAM []byte

func NewRAM(size uint32) *RAM {
	b := make(RAM, size)
	return &b
}

func (b RAM) ReadAt(p []byte, off int64) (n int, err error) {
	if off > int64(len(b)) {
		err = io.EOF
	} else {
		n = copy(p, b[off:])
	}
	return
}

// Fetch data from RAM
func (b RAM) Fetch(addr uint16) (value uint8) {
	return b[addr]
}

// Store data in RAM
func (b *RAM) Store(addr uint16, value uint8) {
	(*b)[addr] = value
}

// Reset RAM
func (b *RAM) Reset() {
	for i, l := 0, len(*b); i < l; i += len(zeros) {
		copy((*b)[i:], zeros)
	}
}

// FetchWord is a helper to fetch a 16-bit word from a bus
func FetchWord(bus AddressBus, addr uint16) uint16 {
	var (
		lo = uint16(bus.Fetch(addr))
		hi = uint16(bus.Fetch(addr+1)) << 8
	)
	return lo | hi
}

// FetchWordBug is a helper to fetch a 16-bit word from a bus
func FetchWordBug(bus AddressBus, addr uint16) uint16 {
	var (
		lo = uint16(bus.Fetch(addr))
		hi = uint16(bus.Fetch(addr&0xff00)) | uint16(uint8(addr+1))<<8
	)
	return lo | hi
}

// StoreWord is a helper to store a 16-bit word on a bus
func StoreWord(bus AddressBus, addr, value uint16) {
	bus.Store(addr+0, uint8(value))
	bus.Store(addr+1, uint8(value>>8))
}
