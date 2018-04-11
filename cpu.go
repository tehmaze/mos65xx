package mos65xx

import "fmt"
import "github.com/tehmaze/mos65xx/memory"

// CPU represents a MOS Technology 65xx Central Processing Unit
type CPU interface {
	// Memory as observed by the CPU
	memory.Memory

	// Registers returns a pointer to the CPU registers
	Registers() *Registers

	// IRQ requests an interrupt
	IRQ()

	// NMI requests an non-maskable interrupt
	NMI()

	// Reset requests a cold reset
	Reset()

	// Ready
	Ready(bool)

	// Step fetches and executes the next instruction, returning the total
	// number of cycles spent on performing the operation.
	Step() int

	// Run until the CPU receives a HLT instruction, returning the total
	// number of cycles spent.
	Run() int

	// Halted returns true if the CPU received a HLT instruction
	Halted() bool

	// Attach a monitor
	Attach(Monitor)
}

/*
// TODO: Not implemented
// Accurate is a CPU implementation optimized for cycle accuracy
type Accurate interface {
	CPU

	// Tick a single cycle
	Tick()
}
*/

// Registers are the CPU registers
type Registers struct {
	PC uint16 // Program counter
	S  uint8  // Stack pointer
	P  uint8  // Processor status register
	A  uint8  // Accumulator register
	X  uint8  // X index register
	Y  uint8  // Y index register
}

// setFlag sets a process status register flag
func setFlag(mask, flag uint8, set bool) uint8 {
	if set {
		return mask | flag
	}
	return mask & ^flag
}

func (reg *Registers) setZ(value uint8) {
	reg.P = setFlag(reg.P, Z, value == 0x00)
}

func (reg *Registers) setV(value uint8) {
	reg.P = setFlag(reg.P, V, value&0x40 != 0x00)
}

func (reg *Registers) setN(value uint8) {
	reg.P = setFlag(reg.P, N, value&0x80 != 0x00)
}

// setZN sets the Z and N flags based on the value
func (reg *Registers) setZN(value uint8) {
	reg.P = setFlag(reg.P, Z, value == 0x00)
	reg.P = setFlag(reg.P, N, value&0x80 != 0x00)
}

func (reg *Registers) setBorrow(value uint16) {
	reg.P = setFlag(reg.P, C, value < 0x100)
}

// cmp compares two values and updates the Z, N and C flags accordingly
func (reg *Registers) cmp(a, b uint8) {
	/*
		log.Printf("cpu: CMP %02x <> %02X z:%t n:%t c:%t",
			a, b, (a-b) == 0, (a-b)&0x80 != 0, a >= b)
	*/

	reg.P = setFlag(reg.P, C, a >= b)
	reg.P = setFlag(reg.P, Z, a == b)
	reg.P = setFlag(reg.P, N, (a-b)&0x80 == 0x80)
}

func (reg *Registers) String() string {
	p := []rune("········")
	for i, c := range []rune("NVUBDIZC") {
		if reg.P&(1<<(7-uint(i))) != 0 {
			p[i] = c
		}
	}

	return fmt.Sprintf("PC:%04X A:%02X X:%02X Y:%02X S:%02X P:%02X(%s)",
		reg.PC, reg.A, reg.X, reg.Y, reg.S, reg.P, string(p))
}

// Processor status register flags
const (
	C uint8 = 1 << iota // Carry flag, 1 = true
	Z                   // Zero, 1 = Result zero
	I                   // IRQ disable, 1 = disable
	D                   // Decimal mode, 1 = true
	B                   // BRK command
	U                   // Unused
	V                   // Overflow, 1 = true
	N                   // Negative, 1 = true
)

// Logical shift right for the processor state register flags
const (
	clsr = iota
	zlsr
	ilsr
	dlsr
	blsr
	_
	vlsr
	nlsr
)

// Interrupt type
type Interrupt uint8

// Interrupt types
const (
	None Interrupt = iota //
	NMI                   // Non-Maskable interrupt
	IRQ                   // Interrupt request
)
