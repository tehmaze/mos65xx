package mos65xx

import (
	"fmt"
	"io"
	"math"

	"github.com/tehmaze/mos65xx/memory"
)

// fast CPU variant is not timing accurate, but optimized for execution speed
type fast struct {
	reg     *Registers
	bus     memory.Memory // External memory
	ram     *memory.RAM   // Internal memory
	ramSize int
	ramMask uint16

	// https://hashrocket.com/blog/posts/switch-vs-map-which-is-the-better-way-to-branch-in-go
	//ops     map[Mnemonic]func(uint16)
	ops     [mnemonics]func(uint16)
	monitor Monitor

	interrupt   Interrupt
	cycles      int
	halted      bool
	addressMode AddressMode

	hasBCD   bool
	hasNMI   bool
	hasIRQ   bool
	hasReady bool
	notReady bool
}

// New creates a new CPU for the specified model
func New(model Model, mem memory.Memory) CPU {
	cpu := &fast{
		reg:      new(Registers),
		bus:      mem,
		ramSize:  model.InternalMemory,
		ramMask:  uint16(model.InternalMemory - 1),
		hasBCD:   model.HasBCD,
		hasNMI:   model.HasNMI,
		hasIRQ:   model.HasIRQ,
		hasReady: model.HasReady,
	}

	if cpu.ramSize > 0 {
		cpu.ram = memory.New(int(cpu.ramSize)).Reset(0xff)
	}

	cpu.ops = [mnemonics]func(uint16){
		cpu.adc,
		cpu.and,
		cpu.asl,
		cpu.bcc,
		cpu.bcs,
		cpu.beq,
		cpu.bit,
		cpu.bmi,
		cpu.bne,
		cpu.bpl,
		cpu.brk,
		cpu.bvc,
		cpu.bvs,
		cpu.clc,
		cpu.cld,
		cpu.cli,
		cpu.clv,
		cpu.cmp,
		cpu.cpx,
		cpu.cpy,
		cpu.dec,
		cpu.dex,
		cpu.dey,
		cpu.eor,
		cpu.inc,
		cpu.inx,
		cpu.iny,
		cpu.jmp,
		cpu.jsr,
		cpu.lda,
		cpu.ldx,
		cpu.ldy,
		cpu.lsr,
		cpu.nop,
		cpu.ora,
		cpu.pha,
		cpu.php,
		cpu.pla,
		cpu.plp,
		cpu.rol,
		cpu.ror,
		cpu.rti,
		cpu.rts,
		cpu.sbc,
		cpu.sec,
		cpu.sed,
		cpu.sei,
		cpu.sta,
		cpu.stx,
		cpu.sty,
		cpu.tax,
		cpu.tay,
		cpu.tsx,
		cpu.txa,
		cpu.txs,
		cpu.tya,
		cpu.hlt,
		cpu.lax,
		cpu.sax,
		cpu.dcp,
		cpu.isc,
		cpu.rla,
		cpu.rra,
		cpu.slo,
		cpu.sre,
		cpu.anc,
		cpu.alr,
		cpu.arr,
		cpu.xaa,
		cpu.ahx,
		cpu.tas,
		cpu.shx,
		cpu.shy,
		cpu.las,
		cpu.axs,
	}

	cpu.Reset()

	return cpu
}

// Fetch a byte from RAM or the address bus
func (cpu *fast) Fetch(addr uint16) uint8 {
	if cpu.ramSize > 0 && int(addr) < cpu.ramSize {
		return cpu.ram.Fetch(addr)
	}
	return cpu.bus.Fetch(addr)
}

// Store a byte in RAM or the address bus
func (cpu *fast) Store(addr uint16, value uint8) {
	if cpu.ramSize > 0 && int(addr) < cpu.ramSize {
		cpu.ram.Store(addr, value)
	} else {
		cpu.bus.Store(addr, value)
	}
}

// ReadAt reads a portion of the memory
func (cpu *fast) ReadAt(p []byte, offs int64) (n int, err error) {
	if offs < 0 || offs > math.MaxUint16 {
		err = io.EOF
	} else {
		l := len(p)
		if cpu.ramSize > 0 && int(offs) < cpu.ramSize {
			n = copy(p, (*cpu.ram)[offs:])
			l -= n
		}
		if l > 0 {
			// Read remainder from external RAM
			var (
				r, ok = cpu.bus.(io.ReaderAt)
				m     int
			)
			if !ok {
				r = memory.ReaderAt{Memory: cpu.bus}
			}
			if m, err = r.ReadAt(p[n:], offs+int64(n)); err == nil {
				n += m
			}
		}
	}
	return
}

// Push a byte onto the stack
func (cpu *fast) Push(value uint8) {
	cpu.Store(0x0100|uint16(cpu.reg.S), value)
	cpu.reg.S--
}

// PushWord pushes a word onto the stack
func (cpu *fast) PushWord(value uint16) {
	cpu.Push(uint8(value >> 8))
	cpu.Push(uint8(value))
}

// Pull a byte from the stack
func (cpu *fast) Pull() uint8 {
	cpu.reg.S++
	return cpu.Fetch(0x0100 | uint16(cpu.reg.S))
}

// PullWord pulls a word from the stack
func (cpu *fast) PullWord() uint16 {
	var (
		lo = uint16(cpu.Pull())
		hi = uint16(cpu.Pull())
	)
	return (hi << 8) | lo
}

// Registers returns a pointer to the CPU registers
func (cpu *fast) Registers() *Registers {
	return cpu.reg
}

// IRQ requests an interrupt
func (cpu *fast) IRQ() {
	if !cpu.hasIRQ {
		return
	}
	cpu.interrupt = IRQ
}

// NMI requests an non-maskable interrupt
func (cpu *fast) NMI() {
	if !cpu.hasNMI {
		return
	}
	cpu.interrupt = NMI
}

// Reset requests a cold reset
func (cpu *fast) Reset() {
	cpu.reg.PC = FetchWord(cpu, ResetVector)
	cpu.reg.S = 0xfd
	cpu.reg.P = 0x34
	cpu.interrupt = None
	cpu.halted = false
	cpu.notReady = false
}

// Ready
func (cpu *fast) Ready(on bool) {
	if !cpu.hasReady {
		return
	}
	cpu.notReady = !on
}

// Run until halted
func (cpu *fast) Run() int {
	cpu.cycles = 0
	cpu.halted = false
	for !cpu.halted {
		cpu.Step()
	}
	return cpu.cycles
}

// Step one instruction
func (cpu *fast) Step() int {
	// RDY line
	if cpu.notReady {
		return 0
	}

	cpu.handleInterrupts()

	var (
		start  = cpu.cycles
		opcode = cpu.nextOpcode()
	)

	if cpu.monitor != nil {
		raw := make([]byte, opcode.Size)
		cpu.ReadAt(raw, int64(cpu.reg.PC))

		if !cpu.monitor.BeforeExecute(cpu, Instruction{
			CPU:         cpu,
			Cycles:      cpu.cycles,
			Mnemonic:    opcode.Mnemonic,
			Registers:   *cpu.reg,
			AddressMode: opcode.Mode,
			Raw:         raw,
		}) {
			return 0
		}
	}

	cpu.addressMode = opcode.Mode

	pageCrossed, addr := cpu.resolveAddr()
	if pageCrossed {
		cpu.cycles += opcode.PageCrossCycles
	}

	cpu.reg.PC += uint16(opcode.Size)
	cpu.ops[opcode.Mnemonic](addr)
	cpu.cycles += opcode.Cycles

	return cpu.cycles - start
}

func (cpu *fast) Halted() bool { return cpu.halted }

// Attach a monitor
func (cpu *fast) Attach(m Monitor) { cpu.monitor = m }

// Operations

func (cpu *fast) handleInterrupts() {
	switch cpu.interrupt {
	case NMI:
		cpu.nmi()
	case IRQ:
		cpu.irq()
	}
	cpu.interrupt = None
}

func (cpu *fast) nextOpcode() opcode {
	return opcodes[cpu.Fetch(cpu.reg.PC)]
}

func differentPage(a, b uint16) bool {
	return (a & 0xff00) != (b & 0xff00)
}

func (cpu *fast) resolveAddr() (pageCrossed bool, addr uint16) {
	switch cpu.addressMode {
	case Implied:
		return
	case Accumulator:
		return
	case Immediate:
		addr = cpu.reg.PC + 1
		return
	case ZeroPage:
		addr = uint16(cpu.Fetch(cpu.reg.PC + 1))
		return
	case ZeroPageX:
		addr = uint16(cpu.Fetch(cpu.reg.PC+1) + cpu.reg.X)
		return
	case ZeroPageY:
		addr = uint16(cpu.Fetch(cpu.reg.PC+1) + cpu.reg.Y)
		return
	case Relative:
		off := uint16(cpu.Fetch(cpu.reg.PC + 1))
		addr = cpu.reg.PC + off + 2
		if off&0x80 == 0x80 {
			addr -= 0x0100
		}
		return
	case Absolute:
		addr = FetchWord(cpu, cpu.reg.PC+1)
		return
	case AbsoluteX:
		src := FetchWord(cpu, cpu.reg.PC+1)
		addr = src + uint16(cpu.reg.X)
		pageCrossed = differentPage(src, addr)
		return
	case AbsoluteY:
		src := FetchWord(cpu, cpu.reg.PC+1)
		addr = src + uint16(cpu.reg.Y)
		pageCrossed = differentPage(src, addr)
		return
	case Indirect:
		addr = FetchWord(cpu, FetchWord(cpu, cpu.reg.PC+1))
		return
	case IndexedIndirect:
		addr = uint16(cpu.Fetch(cpu.reg.PC+1) + cpu.reg.X)
		var (
			lo = uint16(cpu.Fetch((addr)))
			hi = uint16(cpu.Fetch((addr + 1) & 0x00ff))
		)
		addr = (hi << 8) | lo
		return
	case IndirectIndexed:
		addr = uint16(cpu.Fetch(cpu.reg.PC + 1))
		var (
			lo = uint16(cpu.Fetch((addr)))
			hi = uint16(cpu.Fetch((addr + 1) & 0x00ff))
		)
		addr = (hi << 8) | lo
		pageCrossed = differentPage(addr, addr+uint16(cpu.reg.Y))
		addr += uint16(cpu.reg.Y)
		return
	default:
		panic(fmt.Sprintf("resolveAddr() called for mode %s", cpu.addressMode))
	}
}

// Load/store

func (cpu *fast) lda(addr uint16) {
	cpu.reg.A = cpu.Fetch(addr)
	cpu.reg.setZN(cpu.reg.A)
}

func (cpu *fast) ldx(addr uint16) {
	cpu.reg.X = cpu.Fetch(addr)
	cpu.reg.setZN(cpu.reg.X)
}

func (cpu *fast) ldy(addr uint16) {
	cpu.reg.Y = cpu.Fetch(addr)
	cpu.reg.setZN(cpu.reg.Y)
}

func (cpu *fast) sta(addr uint16) {
	cpu.Store(addr, cpu.reg.A)
}

func (cpu *fast) stx(addr uint16) {
	cpu.Store(addr, cpu.reg.X)
}

func (cpu *fast) sty(addr uint16) {
	cpu.Store(addr, cpu.reg.Y)
}

// Transfer

func (cpu *fast) tax(_ uint16) {
	cpu.reg.X = cpu.reg.A
	cpu.reg.setZN(cpu.reg.X)
}

func (cpu *fast) tay(_ uint16) {
	cpu.reg.Y = cpu.reg.A
	cpu.reg.setZN(cpu.reg.Y)
}

func (cpu *fast) tsx(_ uint16) {
	cpu.reg.X = cpu.reg.S
	cpu.reg.setZN(cpu.reg.X)
}

func (cpu *fast) txa(_ uint16) {
	cpu.reg.A = cpu.reg.X
	cpu.reg.setZN(cpu.reg.A)
}

func (cpu *fast) txs(_ uint16) {
	cpu.reg.S = cpu.reg.X
}

func (cpu *fast) tya(_ uint16) {
	cpu.reg.A = cpu.reg.Y
	cpu.reg.setZN(cpu.reg.A)
}

// Increment/decrement register

func (cpu *fast) dec(addr uint16) {
	v := cpu.Fetch(addr) - 1
	cpu.Store(addr, v)
	cpu.reg.setZN(v)
}

func (cpu *fast) dex(_ uint16) {
	cpu.reg.X--
	cpu.reg.setZN(cpu.reg.X)
}

func (cpu *fast) dey(_ uint16) {
	cpu.reg.Y--
	cpu.reg.setZN(cpu.reg.Y)
}

func (cpu *fast) inc(addr uint16) {
	v := cpu.Fetch(addr) + 1
	cpu.Store(addr, v)
	cpu.reg.setZN(v)
}

func (cpu *fast) inx(_ uint16) {
	cpu.reg.X++
	cpu.reg.setZN(cpu.reg.X)
}

func (cpu *fast) iny(_ uint16) {
	cpu.reg.Y++
	cpu.reg.setZN(cpu.reg.Y)
}

// Compare

func (cpu *fast) cmp(addr uint16) {
	cpu.reg.cmp(cpu.reg.A, cpu.Fetch(addr))
}

func (cpu *fast) cpx(addr uint16) {
	cpu.reg.cmp(cpu.reg.X, cpu.Fetch(addr))
}

func (cpu *fast) cpy(addr uint16) {
	cpu.reg.cmp(cpu.reg.Y, cpu.Fetch(addr))
}

// Processor status register

func (cpu *fast) clc(_ uint16) { cpu.reg.P &= (^C) }
func (cpu *fast) cld(_ uint16) { cpu.reg.P &= (^D) }
func (cpu *fast) cli(_ uint16) { cpu.reg.P &= (^I) }
func (cpu *fast) clv(_ uint16) { cpu.reg.P &= (^V) }
func (cpu *fast) sec(_ uint16) { cpu.reg.P |= C }
func (cpu *fast) sed(_ uint16) { cpu.reg.P |= D }
func (cpu *fast) sei(_ uint16) { cpu.reg.P |= I }

// Logical operations

func (cpu *fast) and(addr uint16) {
	cpu.reg.A &= cpu.Fetch(addr)
	cpu.reg.setZN(cpu.reg.A)
}

func (cpu *fast) eor(addr uint16) {
	cpu.reg.A ^= cpu.Fetch(addr)
	cpu.reg.setZN(cpu.reg.A)
}

func (cpu *fast) ora(addr uint16) {
	cpu.reg.A |= cpu.Fetch(addr)
	cpu.reg.setZN(cpu.reg.A)
}

func (cpu *fast) bit(addr uint16) {
	v := cpu.Fetch(addr)
	cpu.reg.P = setFlag(cpu.reg.P, V, v&0x40 == 0x40)
	cpu.reg.P = setFlag(cpu.reg.P, N, v&0x80 == 0x80)
	cpu.reg.P = setFlag(cpu.reg.P, Z, v&cpu.reg.A == 0)
}

func (cpu *fast) asl(addr uint16) {
	switch cpu.addressMode {
	case Accumulator:
		v := cpu.reg.A
		cpu.reg.P = setFlag(cpu.reg.P, C, (v>>7)&1 == 1)
		cpu.reg.A = v << 1
		cpu.reg.setZN(cpu.reg.A)
	default:
		v := cpu.Fetch(addr)
		cpu.reg.P = setFlag(cpu.reg.P, C, (v>>7)&1 == 1)
		v <<= 1
		cpu.Store(addr, v)
		cpu.reg.setZN(v)
	}
}

func (cpu *fast) lsr(addr uint16) {
	switch cpu.addressMode {
	case Accumulator:
		v := cpu.reg.A
		cpu.reg.P = setFlag(cpu.reg.P, C, v&1 == 1)
		cpu.reg.A = v >> 1
		cpu.reg.setZN(cpu.reg.A)
	default:
		v := cpu.Fetch(addr)
		cpu.reg.P = setFlag(cpu.reg.P, C, v&1 == 1)
		v >>= 1
		cpu.Store(addr, v)
		cpu.reg.setZN(v)
	}
}

func (cpu *fast) rol(addr uint16) {
	var v, carry uint8
	if cpu.reg.P&C == C {
		carry = 1
	}
	switch cpu.addressMode {
	case Accumulator:
		v = cpu.reg.A
	default:
		v = cpu.Fetch(addr)
	}
	cpu.reg.P = setFlag(cpu.reg.P, C, (v>>7) == 1)
	v = (v << 1) | carry
	cpu.reg.setZN(v)
	switch cpu.addressMode {
	case Accumulator:
		cpu.reg.A = v
	default:
		cpu.Store(addr, v)
	}
}

func (cpu *fast) ror(addr uint16) {
	var v, carry uint8
	if cpu.reg.P&C == C {
		carry = 1 << 7
	}
	switch cpu.addressMode {
	case Accumulator:
		v = cpu.reg.A
	default:
		v = cpu.Fetch(addr)
	}
	cpu.reg.P = setFlag(cpu.reg.P, C, v&1 == 1)
	v = (v >> 1) | carry
	cpu.reg.setZN(v)
	switch cpu.addressMode {
	case Accumulator:
		cpu.reg.A = v
	default:
		cpu.Store(addr, v)
	}
}

/// Arithmetic

func overflow(a, b, r uint8) bool  { return (a^r)&(b^r)&0x80 == 0x80 }
func underflow(a, b, r uint8) bool { return (a^b)&0x80 == 0x80 && (a^r)&0x80 == 0x80 }

func (cpu *fast) adc(addr uint16) {
	var n, v, z, c bool
	cpu.reg.A, n, v, z, c = adc(
		cpu.reg.A, cpu.Fetch(addr),
		cpu.reg.P&C == C,               // carry
		cpu.reg.P&D == D && cpu.hasBCD, // bcd
	)
	cpu.reg.P = setFlag(cpu.reg.P, N, n)
	cpu.reg.P = setFlag(cpu.reg.P, V, v)
	cpu.reg.P = setFlag(cpu.reg.P, Z, z)
	cpu.reg.P = setFlag(cpu.reg.P, C, c)
}

func (cpu *fast) sbc(addr uint16) {
	var n, v, z, c bool
	cpu.reg.A, n, v, z, c = sbc(
		cpu.reg.A, cpu.Fetch(addr),
		cpu.reg.P&C == C,               // carry
		cpu.reg.P&D == D && cpu.hasBCD, // bcd
	)
	cpu.reg.P = setFlag(cpu.reg.P, N, n)
	cpu.reg.P = setFlag(cpu.reg.P, V, v)
	cpu.reg.P = setFlag(cpu.reg.P, Z, z)
	cpu.reg.P = setFlag(cpu.reg.P, C, c)
}

// Branching

func (cpu *fast) branch(pc uint16) {
	// Branch taken: add cycle
	cpu.cycles++

	if differentPage(cpu.reg.PC, pc) {
		// Page cross; add cycle
		cpu.cycles++
	}

	cpu.reg.PC = pc
}

func (cpu *fast) bcc(addr uint16) {
	if cpu.reg.P&C == 0 {
		cpu.branch(addr)
	}
}

func (cpu *fast) bcs(addr uint16) {
	if cpu.reg.P&C == C {
		cpu.branch(addr)
	}
}

func (cpu *fast) bne(addr uint16) {
	if cpu.reg.P&Z == 0 {
		cpu.branch(addr)
	}
}

func (cpu *fast) beq(addr uint16) {
	if cpu.reg.P&Z == Z {
		cpu.branch(addr)
	}
}

func (cpu *fast) bpl(addr uint16) {
	if cpu.reg.P&N == 0 {
		cpu.branch(addr)
	}
}

func (cpu *fast) bmi(addr uint16) {
	if cpu.reg.P&N == N {
		cpu.branch(addr)
	}
}

func (cpu *fast) bvc(addr uint16) {
	if cpu.reg.P&V == 0 {
		cpu.branch(addr)
	}
}

func (cpu *fast) bvs(addr uint16) {
	if cpu.reg.P&V == V {
		cpu.branch(addr)
	}
}

// Jumps and interrupts

func (cpu *fast) jmp(addr uint16) {
	cpu.reg.PC = addr
}

func (cpu *fast) jsr(addr uint16) {
	cpu.PushWord(cpu.reg.PC - 1)
	cpu.reg.PC = addr
}

func (cpu *fast) rti(_ uint16) {
	cpu.reg.P = (cpu.Pull() & 0xef) | 0x20
	cpu.reg.PC = cpu.PullWord()
}

func (cpu *fast) rts(_ uint16) {
	cpu.reg.PC = cpu.PullWord() + 1
}

func (cpu *fast) brk(addr uint16) {
	cpu.PushWord(cpu.reg.PC + 1)
	cpu.Push(cpu.reg.P | 0x10) // php
	cpu.reg.P |= I             // sei
	cpu.reg.PC = FetchWord(cpu, IRQVector)
}

func (cpu *fast) nmi() {
	cpu.PushWord(cpu.reg.PC)
	cpu.Push(cpu.reg.P)
	cpu.reg.P |= I
	cpu.reg.PC = FetchWord(cpu, NMIVector)
	cpu.cycles += 7
}

func (cpu *fast) irq() {
	cpu.PushWord(cpu.reg.PC)
	cpu.Push(cpu.reg.P)
	cpu.reg.P |= I
	cpu.reg.PC = FetchWord(cpu, IRQVector)
	cpu.cycles += 7
}

// Push/Pull values

func (cpu *fast) pha(_ uint16) {
	cpu.Push(cpu.reg.A)
}

func (cpu *fast) php(_ uint16) {
	cpu.Push(cpu.reg.P | B)
}

func (cpu *fast) pla(_ uint16) {
	cpu.reg.A = cpu.Pull()
	cpu.reg.setZN(cpu.reg.A)
}

func (cpu *fast) plp(_ uint16) {
	cpu.reg.P = (cpu.Pull() & 0xef) | 0x20
}

// Misc.

func (cpu *fast) nop(_ uint16) {}

func (cpu *fast) hlt(_ uint16) {
	cpu.reg.PC--
	cpu.halted = true
}

// Undocumented

func (cpu *fast) alr(addr uint16) {
	cpu.and(addr)
	cpu.addressMode = Accumulator
	cpu.lsr(addr)
}

func (cpu *fast) anc(addr uint16) {
	cpu.reg.A &= cpu.Fetch(addr)
	cpu.reg.setZN(cpu.reg.A)
	cpu.reg.P = setFlag(cpu.reg.P, C, cpu.reg.P&N == N)
}

func (cpu *fast) arr(addr uint16) {
	cpu.and(addr)
	cpu.addressMode = Accumulator
	cpu.ror(addr)
	var (
		b5 = (cpu.reg.A>>5)&1 == 1
		b6 = (cpu.reg.A>>6)&1 == 1
	)
	cpu.reg.P = setFlag(cpu.reg.P, C, b6)
	cpu.reg.P = setFlag(cpu.reg.P, V, b5 != b6) // XOR
}

func (cpu *fast) axs(addr uint16) {
	var (
		a = cpu.reg.A & cpu.reg.X
		v = cpu.Fetch(addr)
	)
	cpu.reg.X = a - v
	cpu.reg.P = setFlag(cpu.reg.P, C, a >= v)
	cpu.reg.setZN(cpu.reg.X)
}

func (cpu *fast) lax(addr uint16) {
	cpu.reg.A = cpu.Fetch(addr)
	cpu.reg.X = cpu.reg.A
	cpu.reg.setZN(cpu.reg.A)
}

func (cpu *fast) las(addr uint16) {
	v := cpu.Fetch(addr) & cpu.reg.S
	cpu.reg.S = v
	cpu.reg.X = v
	cpu.reg.A = v
	cpu.reg.setZN(cpu.reg.A)
}

func (cpu *fast) sax(addr uint16) {
	cpu.Store(addr, cpu.reg.A&cpu.reg.X)
}

func (cpu *fast) dcp(addr uint16) {
	cpu.dec(addr)
	cpu.cmp(addr)
}

func (cpu *fast) isc(addr uint16) {
	cpu.inc(addr)
	cpu.sbc(addr)
}

func (cpu *fast) rla(addr uint16) {
	cpu.rol(addr)
	cpu.and(addr)
}

func (cpu *fast) slo(addr uint16) {
	cpu.asl(addr)
	cpu.ora(addr)
}

func (cpu *fast) sre(addr uint16) {
	cpu.lsr(addr)
	cpu.eor(addr)
}

func (cpu *fast) rra(addr uint16) {
	cpu.ror(addr)
	cpu.adc(addr)
}

func (cpu *fast) ahx(addr uint16) {
	cpu.Store(addr, (uint8(addr>>8)+1)&cpu.reg.A&cpu.reg.X)
}

func (cpu *fast) shx(addr uint16) {
	cpu.Store(addr, (uint8(addr>>8)+1)&cpu.reg.X)
}

func (cpu *fast) shy(addr uint16) {
	cpu.Store(addr, (uint8(addr>>8)+1)&cpu.reg.Y)
}

func (cpu *fast) tas(addr uint16) {
	cpu.reg.S = cpu.reg.A & cpu.reg.X
	var (
		v = (uint16(cpu.reg.S) & ((addr >> 8) + 1)) & 0xff
		t = (addr - uint16(cpu.reg.Y)) & 0xff
	)
	if uint16(cpu.reg.Y)+t <= 0xff {
		cpu.Store(addr, uint8(v))
	} else {
		cpu.Store(addr, cpu.Fetch(addr))
	}
}

func (cpu *fast) xaa(addr uint16) {
	cpu.reg.A = cpu.reg.X & cpu.Fetch(addr)
	cpu.reg.setZN(cpu.reg.A)
}

// Interface checks
var (
	_ CPU           = (*fast)(nil)
	_ memory.Memory = (*fast)(nil)
)
