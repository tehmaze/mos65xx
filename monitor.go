package mos65xx

import (
	"bytes"
	"fmt"
	"html/template"
	"strings"
)

// Instruction formats
const (
	// FormatDefault resembles neskell's output format
	FormatDefault = `{{printf "%07d %04X %02X %02X %02X %02X:%s %02X %02X:%s %-7s %-9s %s" .C .PC .A .X .Y .P .PS .S .I .M .Operand .Fetch .Store}}`

	// FormatNintendulator resembles nintendulator's output format
	FormatNintendulator = `{{printf "%04X %-9s %s %-27s A:%02X X:%02X Y:%02X P:%02X SP:%02x" .PC .RawX .Mnemonic .Operand .A .X .Y .P .S}}`
)

var (
	// InstructionFormat is the default instruction format
	InstructionFormat = FormatDefault
)

// Instruction describes an instruction that's about to be executed
type Instruction struct {
	// CPU this instruction is executed on
	CPU CPU

	// Cycles elapsed
	Cycles int

	// Mnemonic is the current operation
	Mnemonic

	// Registers state for instruction
	Registers

	// AddressMode is the addressing mode for this instruction
	AddressMode

	// Raw opcode and address bytes
	Raw []byte
}

// Addr is the operand address for the current instruction.
func (in Instruction) Addr() (addr uint16) {
	switch in.AddressMode {
	case Immediate:
		addr = in.Registers.PC + 1
	case ZeroPage:
		addr = uint16(in.CPU.Fetch(in.Registers.PC + 1))
	case ZeroPageX:
		addr = uint16(in.CPU.Fetch(in.Registers.PC+1) + in.Registers.X)
	case ZeroPageY:
		addr = uint16(in.CPU.Fetch(in.Registers.PC+1) + in.Registers.Y)
	case Relative:
		off := uint16(in.CPU.Fetch(in.Registers.PC + 1))
		addr = in.Registers.PC + off + 2
		if off&0x80 == 0x80 {
			addr -= 0x0100
		}
	case Absolute:
		addr = FetchWord(in.CPU, in.Registers.PC+1)
	case AbsoluteX:
		addr = FetchWord(in.CPU, in.Registers.PC+1) + uint16(in.Registers.X)
	case AbsoluteY:
		addr = FetchWord(in.CPU, in.Registers.PC+1) + uint16(in.Registers.Y)
	case Indirect:
		addr = FetchWord(in.CPU, in.Registers.PC+1)
	case IndexedIndirect:
		addr = uint16(in.CPU.Fetch(in.Registers.PC+1) + in.Registers.X)
		var (
			lo = uint16(in.CPU.Fetch((addr)))
			hi = uint16(in.CPU.Fetch((addr + 1) & 0x00ff))
		)
		addr = (hi << 8) | lo
	case IndirectIndexed:
		addr = uint16(in.CPU.Fetch(in.Registers.PC + 1))
		var (
			lo = uint16(in.CPU.Fetch((addr)))
			hi = uint16(in.CPU.Fetch((addr + 1) & 0x00ff))
		)
		addr = (hi << 8) | lo
		addr += uint16(in.Registers.Y)
	default:
	}
	return
}

// fetches renders the fetch operations
func (in Instruction) fetches(cpu CPU) (out string) {
	out = "-"
	switch in.Mnemonic {
	case LDA, LDX, LDY, BIT, AND, EOR, ORA, ASL, LSR, ROL, ROR, ADC, SBC, INC, DEC, CMP, CPX, CPY:
		switch in.AddressMode {
		case Accumulator, Implied, Immediate:
		default:
			addr := in.Addr()
			out = fmt.Sprintf("%04X→%02X", addr, in.CPU.Fetch(addr))
		}
	case JMP:
		switch in.AddressMode {
		case Indirect:
			addr := in.Addr()
			out = fmt.Sprintf("%04X→%04X", addr, FetchWord(in.CPU, addr))
		case IndirectIndexed, IndexedIndirect:
			addr := in.Addr()
			out = fmt.Sprintf("%04X→%02X", addr, in.CPU.Fetch(addr))
		}
	}
	return
}

// stores renders the store operations
// TODO: this reimplements a lot of the instruction executions; can we do it
//       more optimal?
func (in Instruction) stores(cpu CPU) (out string) {
	var s []string
	switch in.Mnemonic {
	case LDA, LDX, LDY:
		var (
			p = in.Registers.P
			v = in.CPU.Fetch(in.Addr())
			r rune
		)
		if v == 0 {
			p |= Z
		} else {
			p &= ^Z
		}
		if v&0x80 == 0x80 {
			p |= N
		} else {
			p &= ^N
		}
		switch in.Mnemonic {
		case LDA:
			r = 'A'
		case LDX:
			r = 'X'
		case LDY:
			r = 'Y'
		}
		s = append(s, fmt.Sprintf("%02X→SR", p))
		s = append(s, fmt.Sprintf("%02X→%c", v, r))
	case STA, STX, STY:
		var (
			a = in.Addr()
			v uint8
		)
		switch in.Mnemonic {
		case STA:
			v = in.Registers.A
		case STX:
			v = in.Registers.X
		case STY:
			v = in.Registers.Y
		}
		s = append(s, fmt.Sprintf("%02X→%04X", v, a))
	case TAS, TAY, TAX, TSX, TXA, TXS, TYA:
		var (
			a     uint8
			p     = in.Registers.P
			r     string
			skipP bool
		)
		switch in.Mnemonic {
		case TAS:
			a = in.Registers.A
			r = "SP"
		case TAX:
			a = in.Registers.A
			r = "X"
		case TAY:
			a = in.Registers.A
			r = "Y"
		case TSX:
			a = in.Registers.S
			r = "X"
		case TXA:
			a = in.Registers.X
			r = "A"
		case TXS:
			a = in.Registers.X
			r = "SP"
			skipP = true
		case TYA:
			a = in.Registers.Y
			r = "A"
		}
		if !skipP {
			p = setFlag(p, Z, a == 0)
			p = setFlag(p, N, a&0x80 == 0x80)
			s = append(s, fmt.Sprintf("%02X→SR", p))
		}
		s = append(s, fmt.Sprintf("%02X→%s", a, r))
	case BIT:
		var (
			v = in.CPU.Fetch(in.Addr())
			p = in.Registers.P
		)
		if v&0x40 == 0x40 {
			p |= V
		} else {
			p &= ^V
		}
		if v&0x80 == 0x80 {
			p |= N
		} else {
			p &= ^N
		}
		if v&in.Registers.A == 0 {
			p |= Z
		} else {
			p &= ^Z
		}
		s = append(s, fmt.Sprintf("%02X→SR", p))
	case JSR:
		s = append(s, fmt.Sprintf("%02X→%04X", (in.Registers.PC+2)>>8, 0x0100|uint16(in.Registers.S)))
		s = append(s, fmt.Sprintf("%02X→%04X", (in.Registers.PC+2)&0xff, 0x0100|uint16(in.Registers.S-1)))
		s = append(s, fmt.Sprintf("%02X→SP", in.Registers.S-2))
		s = append(s, fmt.Sprintf("%04X→PC", in.Addr()))
	case RTI:
		s = append(s, fmt.Sprintf("%02X→SP", in.Registers.S+1))
		s = append(s, fmt.Sprintf("%02X→SR", (in.CPU.Fetch(0x0100|uint16(in.Registers.S+1))&0xef)|0x20))
		s = append(s, fmt.Sprintf("%04X→PC", FetchWord(in.CPU, 0x0100|uint16(in.Registers.S+2))+1))
		s = append(s, fmt.Sprintf("%02X→SP", in.Registers.S+3))

	case RTS:
		s = append(s, fmt.Sprintf("%02X→SP", in.Registers.S+2))
		s = append(s, fmt.Sprintf("%04X→PC", FetchWord(in.CPU, 0x0100|uint16(in.Registers.S+1))+1))
	case CLC:
		s = append(s, fmt.Sprintf("%02X→SR", in.Registers.P & ^C))
	case CLD:
		s = append(s, fmt.Sprintf("%02X→SR", in.Registers.P & ^D))
	case CLI:
		s = append(s, fmt.Sprintf("%02X→SR", in.Registers.P&^I))
	case CLV:
		s = append(s, fmt.Sprintf("%02X→SR", in.Registers.P&^V))
	case SEC:
		s = append(s, fmt.Sprintf("%02X→SR", in.Registers.P|C))
	case SED:
		s = append(s, fmt.Sprintf("%02X→SR", in.Registers.P|D))
	case SEI:
		s = append(s, fmt.Sprintf("%02X→SR", in.Registers.P|I))
	case PHA:
		s = append(s, fmt.Sprintf("%02X→%04X", in.Registers.A, 0x0100|uint16(in.Registers.S)))
		s = append(s, fmt.Sprintf("%02X→SP", in.Registers.S-1))
	case PHP:
		s = append(s, fmt.Sprintf("%02X→%04X", in.Registers.P|B, 0x0100|uint16(in.Registers.S)))
		s = append(s, fmt.Sprintf("%02X→SP", in.Registers.S-1))
	case AND:
		var (
			p = in.Registers.P
			v = in.CPU.Fetch(in.Addr())
			a = in.Registers.A & v
		)
		p = setFlag(p, N, a&0x80 == 0x80)
		p = setFlag(p, Z, a == 0)
		s = append(s, fmt.Sprintf("%02X→SR", p))
		s = append(s, fmt.Sprintf("%02X→A", a))
	case EOR:
		var (
			p = in.Registers.P
			v = in.CPU.Fetch(in.Addr())
			a = in.Registers.A ^ v
		)
		p = setFlag(p, N, a&0x80 == 0x80)
		p = setFlag(p, Z, a == 0)
		s = append(s, fmt.Sprintf("%02X→SR", p))
		s = append(s, fmt.Sprintf("%02X→A", a))
	case ORA:
		var (
			p = in.Registers.P
			v = in.CPU.Fetch(in.Addr())
			a = in.Registers.A | v
		)
		p = setFlag(p, N, a&0x80 == 0x80)
		p = setFlag(p, Z, a == 0)
		s = append(s, fmt.Sprintf("%02X→SR", p))
		s = append(s, fmt.Sprintf("%02X→A", a))
	case ASL:
		var (
			p = in.Registers.P
			v uint8
			r string
			t uint16
		)
		switch in.AddressMode {
		case Accumulator:
			v = in.Registers.A
			r = "A"
		default:
			t = in.Addr()
			v = in.CPU.Fetch(t)
			r = fmt.Sprintf("%04X", t)
		}
		p = setFlag(p, C, v&0x80 == 0x80)
		v <<= 1
		p = setFlag(p, N, v&0x80 == 0x80)
		p = setFlag(p, Z, v == 0)
		s = append(s, fmt.Sprintf("%02X→SR", p))
		s = append(s, fmt.Sprintf("%02X→%s", v, r))
	case LSR:
		var (
			p = in.Registers.P
			v uint8
			r string
			t uint16
		)
		switch in.AddressMode {
		case Accumulator:
			v = in.Registers.A
			r = "A"
		default:
			t = in.Addr()
			v = in.CPU.Fetch(t)
			r = fmt.Sprintf("%04X", t)
		}
		p = setFlag(p, C, v&1 == 1)
		v >>= 1
		p = setFlag(p, N, v&0x80 == 0x80)
		p = setFlag(p, Z, v == 0)
		s = append(s, fmt.Sprintf("%02X→SR", p))
		s = append(s, fmt.Sprintf("%02X→%s", v, r))
	case ROL:
		var (
			p = in.Registers.P
			v uint8
			r string
			t uint16
		)
		switch in.AddressMode {
		case Accumulator:
			v = in.Registers.A
			r = "A"
		default:
			t = in.Addr()
			v = in.CPU.Fetch(t)
			r = fmt.Sprintf("%04X", t)
		}
		p = setFlag(p, C, v&0x80 == 0x80)
		v = (v << 1) | (in.Registers.P & C)
		p = setFlag(p, N, v&0x80 == 0x80)
		p = setFlag(p, Z, v == 0)
		s = append(s, fmt.Sprintf("%02X→SR", p))
		s = append(s, fmt.Sprintf("%02X→%s", v, r))
	case ROR:
		var (
			p = in.Registers.P
			v uint8
			r string
			t uint16
		)
		switch in.AddressMode {
		case Accumulator:
			v = in.Registers.A
			r = "A"
		default:
			t = in.Addr()
			v = in.CPU.Fetch(t)
			r = fmt.Sprintf("%04X", t)
		}
		p = setFlag(p, C, v&0x01 == 0x01)
		v = (v >> 1) | (in.Registers.P&C)<<7
		p = setFlag(p, N, v&0x80 == 0x80)
		p = setFlag(p, Z, v == 0)
		s = append(s, fmt.Sprintf("%02X→SR", p))
		s = append(s, fmt.Sprintf("%02X→%s", v, r))
	case DEC, DEX, DEY, INC, INX, INY:
		var (
			v uint8
			r string
			p = in.Registers.P
			t uint16
		)
		switch in.Mnemonic {
		case DEC:
			if in.AddressMode == Accumulator {
				v = in.Registers.A - 1
				r = "A"
			} else {
				t = in.Addr()
				v = in.CPU.Fetch(t) - 1
				r = fmt.Sprintf("%04X", t)
			}
		case DEX:
			r = "X"
			v = in.Registers.X - 1
		case DEY:
			r = "Y"
			v = in.Registers.Y - 1
		case INC:
			if in.AddressMode == Accumulator {
				v = in.Registers.A + 1
				r = "A"
			} else {
				t = in.Addr()
				v = in.CPU.Fetch(t) + 1
				r = fmt.Sprintf("%04X", t)
			}
		case INX:
			r = "X"
			v = in.Registers.X + 1
		case INY:
			r = "Y"
			v = in.Registers.Y + 1
		}
		p = setFlag(p, N, v&0x80 == 0x80)
		p = setFlag(p, Z, v == 0)
		s = append(s, fmt.Sprintf("%02X→SR", p))
		s = append(s, fmt.Sprintf("%02X→%s", v, r))
	case PLA:
		var (
			p = in.Registers.P
			v = in.CPU.Fetch(0x0100 | uint16(in.Registers.S+1))
		)
		if v&0x80 == 0x80 {
			p |= N
		} else {
			p &= ^N
		}
		if v == 0 {
			p |= Z
		} else {
			p &= ^Z
		}
		s = append(s, fmt.Sprintf("%02X→SP", in.Registers.S+1))
		s = append(s, fmt.Sprintf("%02X→SR", p)) // Actually p, bug in neskell
		s = append(s, fmt.Sprintf("%02X→A", v))
	case PLP:
		var (
			p = (in.CPU.Fetch(0x0100|uint16(in.Registers.S+1)) & 0xef) | 0x20
		)
		s = append(s, fmt.Sprintf("%02X→SP", in.Registers.S+1))
		s = append(s, fmt.Sprintf("%02X→SR", p))
	case CMP, CPX, CPY:
		var (
			a uint8
			b = in.CPU.Fetch(in.Addr())
			p = in.Registers.P
		)
		switch in.Mnemonic {
		case CMP:
			a = in.Registers.A
		case CPX:
			a = in.Registers.X
		case CPY:
			a = in.Registers.Y
		}
		p = setFlag(p, C, a >= b)
		p = setFlag(p, Z, a == b)
		p = setFlag(p, N, (a-b)&0x80 == 0x80)
		s = append(s, fmt.Sprintf("%02X→SR", p))
	}

	switch in.Mnemonic {
	case JMP, JSR, RTS, RTI:
	default:
		// Not accurate; need page boundary cycles too
		//s = append(s, fmt.Sprintf("%04X→PC", in.Registers.PC+uint16(opcodes[in.Raw[0]].Cycles)))
	}

	return strings.Join(s, " ")
}

// Operand formats the instruction's mnemonic arguments
func (in Instruction) operand(cpu CPU) (out string) {
	switch in.AddressMode {
	case Accumulator:
		out = "A"
	case Immediate:
		out = fmt.Sprintf("#$%02X", in.CPU.Fetch(in.Registers.PC+1))
	case Absolute:
		out = fmt.Sprintf("$%04X", FetchWord(in.CPU, in.Registers.PC+1))
	case AbsoluteX:
		out = fmt.Sprintf("$%04X,X", FetchWord(in.CPU, in.Registers.PC+1))
	case AbsoluteY:
		out = fmt.Sprintf("$%04X,Y", FetchWord(in.CPU, in.Registers.PC+1))
	case Relative:
		out = fmt.Sprintf("$%02X", in.CPU.Fetch(in.Registers.PC+1))
	case Indirect:
		var (
			lo   = uint16(in.CPU.Fetch(in.Registers.PC + 1))
			hi   = uint16(in.CPU.Fetch(in.Registers.PC + 2))
			addr = (hi << 8) | lo
		)
		out = fmt.Sprintf("($%04X)", addr)
	case IndexedIndirect:
		out = fmt.Sprintf("($%02X,X)", in.CPU.Fetch(in.Registers.PC+1))
	case IndirectIndexed:
		out = fmt.Sprintf("($%02X),Y", in.CPU.Fetch(in.Registers.PC+1))
	case ZeroPage:
		out = fmt.Sprintf("$%02X", in.CPU.Fetch(in.Registers.PC+1))
	case ZeroPageX:
		out = fmt.Sprintf("$%02X,X", in.CPU.Fetch(in.Registers.PC+1))
	case ZeroPageY:
		out = fmt.Sprintf("$%02X,Y", in.CPU.Fetch(in.Registers.PC+1))
	}
	return
}

// Format returns a formatted string based on the format template for the
// referenced CPU.
func (in Instruction) Format(format string, cpu CPU) string {
	var (
		t = template.Must(template.New("instruction").Parse(format))
		b = new(bytes.Buffer)
		d = map[string]interface{}{
			"B":       in.CPU,
			"Mode":    in.AddressMode,
			"C":       in.Cycles,
			"M":       in.Mnemonic,
			"R":       in.Registers,
			"PC":      in.Registers.PC,
			"P":       in.Registers.P,
			"PS":      fmtP(in.Registers.P),
			"S":       in.Registers.S,
			"A":       in.Registers.A,
			"X":       in.Registers.X,
			"Y":       in.Registers.Y,
			"Raw":     in.Raw,
			"I":       in.Raw[0],
			"RawX":    padX(in.Raw),
			"Operand": in.operand(cpu),
			"Fetch":   in.fetches(cpu),
			"Store":   in.stores(cpu),
		}
	)
	if err := t.Execute(b, d); err != nil {
		return ""
	}
	return b.String()
}

func fmtP(p uint8) (s string) {
	var o = []rune("········")
	for i, c := range []rune("NVUBDIZC") {
		if p&(1<<uint(7-i)) != 0 {
			o[i] = c
		}
	}
	return string(o)
}

func padX(b []byte) (s string) {
	for i, c := range b {
		if i > 0 {
			s += " "
		}
		s += fmt.Sprintf("%02X", c)
	}
	return
}

// Monitor for the CPU monitors instruction executions
type Monitor interface {
	// BeforeExecute gets called before instruction execution, returning false
	// will stop execution and halt the CPU.
	BeforeExecute(CPU, Instruction) bool
}

// InstructionPrinter will output a formatted string before execution.
type InstructionPrinter func(string)

// BeforeExecute triggers the printer function.
func (m InstructionPrinter) BeforeExecute(cpu CPU, in Instruction) bool {
	m(in.Format(InstructionFormat, cpu))
	return true
}
