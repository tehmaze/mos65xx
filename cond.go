package mos65xx

import (
	"fmt"
)

const (
	condPass  = "    \x1b[1;32m✓\x1b[0m "
	condFail  = "    \x1b[1;31m✗\x1b[0m "
	condEqual = "\x1b[1;37m=\x1b[0m"
	condRange = "\x1b[1;37m∈\x1b[0m"
)

// cond is a condition checker for our test harness
type cond interface {
	Cond(Instruction) bool
	String() string
}

// conds are multiple conditions combined
type conds struct {
	Any   bool   // Any condition that returns true will make this pass
	Conds []cond // conditions

	res      []bool
	met, not []cond
}

func (t *conds) Cond(in Instruction) bool {
	// Each test restes the conditions
	t.res = make([]bool, len(t.Conds))
	t.met = make([]cond, 0, len(t.Conds))
	t.not = make([]cond, 0, len(t.Conds))

	// Check conditions
	for i, c := range t.Conds {
		if t.res[i] = c.Cond(in); t.res[i] {
			t.met = append(t.met, c)
		} else {
			t.not = append(t.not, c)
		}
	}

	if t.Any {
		// Any condition has to be met
		return len(t.met) > 0
	}

	// All conditions have to be met
	return len(t.met) == len(t.Conds)
}

// Reason (only valid for Any mode)
func (t *conds) Reason() string {
	if len(t.met) > 0 {
		return t.met[0].String()
	}
	return ""
}

func (t *conds) Print(f func(string)) {
	if len(t.met) == 0 {
		f("  no conditions met")
	} else {
		f("  conditions met:")
		for _, c := range t.met {
			f(condPass + c.String())
		}
	}

	if len(t.not) == 0 {
		f("  no conditions unmet")
	} else {
		f("  conditions unmet:")
		for _, c := range t.not {
			f(condFail + c.String())
		}
	}
}

// Register value conditions
type (
	condPC uint16
	condP  uint8
	condS  uint8
	condA  uint8
	condX  uint8
	condY  uint8
)

func (t condPC) Cond(in Instruction) bool { return in.CPU.Registers().PC == uint16(t) }
func (t condPC) String() string           { return fmt.Sprintf("PC     %s $%04X", condEqual, uint16(t)) }
func (t condP) Cond(in Instruction) bool  { return in.CPU.Registers().P == uint8(t) }
func (t condP) String() string            { return fmt.Sprintf("P      %s $%02X", condEqual, uint8(t)) }
func (t condS) Cond(in Instruction) bool  { return in.CPU.Registers().S == uint8(t) }
func (t condS) String() string            { return fmt.Sprintf("S      %s $%02X", condEqual, uint8(t)) }
func (t condA) Cond(in Instruction) bool  { return in.CPU.Registers().A == uint8(t) }
func (t condA) String() string            { return fmt.Sprintf("A      %s $%02X", condEqual, uint8(t)) }
func (t condX) Cond(in Instruction) bool  { return in.CPU.Registers().X == uint8(t) }
func (t condX) String() string            { return fmt.Sprintf("X      %s $%02X", condEqual, uint8(t)) }
func (t condY) Cond(in Instruction) bool  { return in.CPU.Registers().Y == uint8(t) }
func (t condY) String() string            { return fmt.Sprintf("Y      %s $%02X", condEqual, uint8(t)) }

// condCycles are conditional cycle boundaries
type condCycles [2]int

func (t condCycles) Cond(in Instruction) bool {
	if t[0] == t[1] || t[1] < t[0] {
		return in.Cycles >= t[0]
	}
	return in.Cycles >= t[0] && in.Cycles <= t[1]
}
func (t condCycles) String() string {
	if t[0] == t[1] || t[1] < t[0] {
		return fmt.Sprintf("cycles %s %d", condEqual, t[0])
	}
	return fmt.Sprintf("cycles  %s [%d, %d]", condRange, t[0], t[1])
}

// condOp is a condition for hitting a mnemonic
type condOp Mnemonic

func (t condOp) Cond(in Instruction) bool {
	return in.Mnemonic == Mnemonic(t)
}

func (t condOp) String() string {
	return fmt.Sprintf("opcode %s $%02X (%s)", condEqual, uint8(t), Mnemonic(t))
}

// condByte is a condition for the value of byte at Addr
type condByte struct {
	Addr  uint16 // Address of the byte
	Value uint8  // Value for comparison
}

func (t condByte) Cond(in Instruction) bool {
	return in.CPU.Fetch(t.Addr) == t.Value
}

func (t condByte) String() string {
	return fmt.Sprintf("$%04X  %s $%02X", t.Addr, condEqual, t.Value)
}

// contTrap is a condition for looping jumps
type condTrap struct{}

func (t condTrap) Cond(in Instruction) bool {
	switch in.Mnemonic {
	case JMP, JSR:
		addr := in.Addr()
		if in.AddressMode == Indirect {
			addr = FetchWord(in.CPU, addr)
		}
		return in.Registers.PC == addr
	default:
		return false
	}
}

func (t condTrap) String() string {
	return "PC trapped"
}

// condStack compiles a slice of condByte to match data in the stack
func condStack(stack ...uint8) []cond {
	var (
		l = uint16(len(stack))
		c = make([]cond, l)
	)
	for i, b := range stack {
		c[i] = condByte{0x0200 - l + uint16(i), b}
	}
	return c
}

func condString(addr uint16, value string) []cond {
	var (
		l = uint16(len(value))
		c = make([]cond, l)
	)
	for i, b := range []byte(value) {
		c[i] = condByte{addr + uint16(i), b}
	}
	return c
}
