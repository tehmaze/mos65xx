package mos65xx

import (
	"github.com/tehmaze/mos65xx/memory"
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

// FetchWord is a helper to fetch a 16-bit word from memory
func FetchWord(mem memory.Memory, addr uint16) uint16 {
	var (
		lo = uint16(mem.Fetch(addr))
		hi = uint16(mem.Fetch(addr+1)) << 8
	)
	return lo | hi
}

// FetchWordBug is a helper to fetch a 16-bit word from memory
func FetchWordBug(mem memory.Memory, addr uint16) uint16 {
	var (
		lo = uint16(mem.Fetch(addr))
		hi = uint16(mem.Fetch(addr&0xff00)) | uint16(uint8(addr+1))<<8
	)
	return lo | hi
}

// StoreWord is a helper to store a 16-bit word on a bus
func StoreWord(mem memory.Memory, addr, value uint16) {
	mem.Store(addr+0, uint8(value))
	mem.Store(addr+1, uint8(value>>8))
}
