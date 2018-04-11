package mos65xx

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"os"
	"runtime"
	"testing"

	"github.com/tehmaze/mos65xx/memory"
)

/*

Test harnass for our emulator. Most of these tests are stolen from neskell and
other emulators that I looked at. The tests included are:

 * Klaus2m5 Functional 6502 tests by Klaus Dormann
 * Blargg's tests by Shay Green <gblargg@gmail.com>
 * HMC 6502 tests by "Swift Elephant"
 * Neskell's tests by Tim C. Schröder <tim@blitzcode.net>

We also keep this text here so our test logger will align nicely :-)

So here is a nice ASCII diagram of a MOS Technology 6502:


         +------u------+
     VSS |1          40| RES
     RDY |2          39| Phi2
    Phi1 |3          38| S0
     IRQ |4          37| Phi0
      NC |5          36| NC
     NMI |6          35| NC
    SYNC |7          34| R/W
     VCC |8          33| D0
      A0 |9    6     32| D1
      A1 |10   5  M  31| D2
      A2 |11   0  O  30| D3
      A3 |12   2  S  29| D4
      A4 |13         28| D5
      A5 |14         27| D6
      A6 |15         26| D7
      A7 |16         26| A15
      A8 |17         26| A14
      A9 |18         26| A13
     A10 |19         26| A12
     A11 |20         26| A11
         +-------------+


Pinout differences:

  Pin | 6800            | 6501            | 6502
    ----+-----------------+-----------------+------------------
    2   | HLT             | RDY             | RDY
    3   | Phi1 (in)       | Phi1 (in)       | Phi1 (out)
    5   | Valid memory xs | Valid memory xs | Not connected
    7   | Bus available   | Bus available   | SYNC
    36  | Data bus enable | Data bus enable | Not connected
    37  | Phi2 (in)       | Phi2 (in)       | Phi0 (in)
    38  | Not connected   | Not connected   | Set overflow flag
    39  | Three-state ctl | Not connected   | Phi2 (out)










.. ahh finally our debugger log is at line 100+
*/

type testBinary struct {
	Model
	Name       string
	Offset, PC uint16
	S          uint8
	Stop, Pass *conds
	Done       func(CPU, *memory.RAM)
	Patch      map[uint16]uint8

	t       *testing.T
	cpu     CPU
	done    bool
	last    Instruction
	monitor Monitor
	remain  bool
}

func (test *testBinary) BeforeExecute(cpu CPU, in Instruction) bool {
	if test.monitor != nil {
		if !test.remain {
			test.t.Log("Cycles                Status Reg.    Instr. Operand Load      Stores")
			test.t.Log("Elapsed $PC  $A $X $Y $P:NVUBDIZC $S $I:Mne Data    $src→$val $val→$dst")
			test.t.Log("-------+----+--+--+--+-----------+--+------+-------+---------+---------")
			test.remain = true
		}
		test.monitor.BeforeExecute(cpu, in)
	}
	test.last = in
	if test.done = test.Stop.Cond(in); test.done {
		return false
	}
	return true
}

func (test *testBinary) Run(t *testing.T) {
	t.Helper()
	//t.Run(test.Name, func(t *testing.T) {
	test.t = t
	var (
		mem      = memory.New(test.Model.ExternalMemory)
		cpu      CPU
		bin, err = ioutil.ReadFile(test.Name)
	)
	if err != nil {
		t.Skip(err)
	}

	// Load binary at offset
	copy((*mem)[test.Offset:], bin)

	// Patch bytes
	for addr, v := range test.Patch {
		(*mem)[addr] = v
	}

	// References
	cpu = New(test.Model, mem)
	test.cpu = cpu
	if trace {
		test.monitor = InstructionPrinter(func(s string) { fmt.Println(s) })
	}

	// Attach monitor
	cpu.Attach(test)

	// PC
	if test.PC > 0x0000 {
		cpu.Registers().PC = test.PC
	}
	cpu.Registers().P = U | I
	if test.S != 0x00 {
		cpu.Registers().S = test.S
	} else {
		cpu.Registers().S = 0xff
	}

	// Run
	cycles := 0
	for !(test.done || cpu.Halted()) {
		cycles += cpu.Step()
	}

	pass := test.Pass.Cond(test.last)
	if testing.Verbose() || !pass {
		t.Logf("stop reason....: %s", test.Stop.Reason())
		t.Logf("final cycles...: %d", cycles)
		t.Logf("final CPU state: %+v", cpu.Registers())
		t.Log("zero page......: $00-$0F")
		t.Logf("0000 %s", padX((*mem)[:16]))
		t.Log("stack..........: $80-$FF")
		t.Logf("0180 %s", padX((*mem)[0x0180:0x0190]))
		t.Logf("0190 %s", padX((*mem)[0x0190:0x01a0]))
		t.Logf("01a0 %s", padX((*mem)[0x01a0:0x01b0]))
		t.Logf("01b0 %s", padX((*mem)[0x01b0:0x01c0]))
		t.Logf("01c0 %s", padX((*mem)[0x01c0:0x01d0]))
		t.Logf("01d0 %s", padX((*mem)[0x01d0:0x01e0]))
		t.Logf("01e0 %s", padX((*mem)[0x01e0:0x01f0]))
		t.Logf("01f0 %s", padX((*mem)[0x01f0:0x0200]))
	}

	if !pass {
		t.Log("failure:")
		test.Pass.Print(func(s string) { t.Log(s) })
		t.Fatal()
	}

	if testing.Verbose() {
		if test.Pass.Any {
			t.Log("success (any):")
		} else {
			t.Log("success (all):")
		}
		test.Pass.Print(func(s string) { t.Log(s) })
	}

	if test.Done != nil {
		test.Done(cpu, mem)
	}
	//})
}

var trace bool

func TestMain(m *testing.M) {
	trace = os.Getenv("TEST_TRACE") != ""
	log.Printf("testing on %d cores in %d threads", runtime.NumCPU(), runtime.GOMAXPROCS(-1))
	os.Exit(m.Run())
}

func TestLoadStore(t *testing.T) {
	test := &testBinary{
		Model:  MOS6502,
		Name:   "testdata/hmc-6502/load_store_test.bin",
		Offset: 0x0600,
		PC:     0x0600,
		Stop: &conds{Any: true, Conds: []cond{
			condOp(BRK),
			condCycles{1000, math.MaxInt16},
		}},
		Pass: &conds{Conds: []cond{
			condA(0x55),
			condX(0x2a),
			condY(0x73),
			condCycles{161, 161},
		}},
	}
	test.Run(t)
}

func TestAND_ORA_EOR(t *testing.T) {
	test := &testBinary{
		Model:  MOS6502,
		Name:   "testdata/hmc-6502/and_or_xor_test.bin",
		Offset: 0x0600,
		PC:     0x0600,
		Stop: &conds{Any: true, Conds: []cond{
			condOp(BRK),
			condCycles{1000, math.MaxInt16},
		}},
		Pass: &conds{Conds: []cond{
			condA(0xaa),
			condX(0x10),
			condY(0xf0),
			condCycles{332, 332},
		}},
	}
	test.Run(t)
}

func TestINC_DEC(t *testing.T) {
	test := &testBinary{
		Model:  MOS6502,
		Name:   "testdata/hmc-6502/inc_dec_test.bin",
		Offset: 0x0600,
		PC:     0x0600,
		Stop: &conds{Any: true, Conds: []cond{
			condOp(BRK),
			condCycles{1000, math.MaxInt16},
		}},
		Pass: &conds{Conds: []cond{
			condP(N | U | I),
			condCycles{149, 149},
		}},
	}
	test.Run(t)
}

func TestBitshift(t *testing.T) {
	test := &testBinary{
		Model:  MOS6502,
		Name:   "testdata/hmc-6502/bitshift_test.bin",
		Offset: 0x0600,
		PC:     0x0600,
		Stop: &conds{Any: true, Conds: []cond{
			condOp(BRK),
			condCycles{1000, math.MaxInt16},
		}},
		Pass: &conds{Conds: []cond{
			condByte{0x01dd, 0x6e},
			condA(0xdd),
			condX(0xdd),
			condCycles{253, 253},
		}},
	}
	test.Run(t)
}

func TestJMP_JSR_RET(t *testing.T) {
	test := &testBinary{
		Model:  MOS6502,
		Name:   "testdata/hmc-6502/jump_ret_test.bin",
		Offset: 0x0600,
		PC:     0x0600,
		Stop: &conds{Any: true, Conds: []cond{
			condOp(BRK),
			condCycles{1000, math.MaxInt16},
		}},
		Pass: &conds{Conds: []cond{
			condByte{0x0040, 0x42},
			condA(0x42),
			condX(0x33),
			condS(0xff),
			condPC(0x0626),
			condCycles{50, 50},
		}},
	}
	test.Run(t)
}

func TestTransfer(t *testing.T) {
	test := &testBinary{
		Model:  MOS6502,
		Name:   "testdata/hmc-6502/reg_transf_test.bin",
		Offset: 0x0600,
		PC:     0x0600,
		Stop: &conds{Any: true, Conds: []cond{
			condOp(BRK),
			condCycles{1000, math.MaxInt16},
		}},
		Pass: &conds{Conds: []cond{
			condByte{0x0040, 0x33},
			condA(0x33),
			condX(0x33),
			condY(0x33),
			condS(0x33),
			condCycles{37, 37},
		}},
	}
	test.Run(t)
}

func TestADD_SUB(t *testing.T) {
	test := &testBinary{
		Model:  MOS6502,
		Name:   "testdata/hmc-6502/add_sub_test.bin",
		Offset: 0x0600,
		PC:     0x0600,
		Stop: &conds{Any: true, Conds: []cond{
			condOp(BRK),
			condCycles{1000, math.MaxInt16},
		}},
		Pass: &conds{Conds: []cond{
			condByte{0x0030, 0xaa},
			condP(N | U | I | C),
			condA(0xaa),
			condX(0x34),
			condY(0x01),
			condCycles{205, 205},
		}},
	}
	test.Run(t)
}

func TestCMP_BEQ_BNE(t *testing.T) {
	test := &testBinary{
		Model:  MOS6502,
		Name:   "testdata/hmc-6502/cmp_beq_bne_test.bin",
		Offset: 0x0600,
		PC:     0x0600,
		Stop: &conds{Any: true, Conds: []cond{
			condOp(BRK),
			condCycles{1000, math.MaxInt16},
		}},
		Pass: &conds{Conds: []cond{
			condByte{0x0015, 0x7f},
			condA(0x7f),
			condY(0x7f),
			condCycles{152, 152},
		}},
	}
	test.Run(t)
}

func TestCPX_CPY_BIT(t *testing.T) {
	test := &testBinary{
		Model:  MOS6502,
		Name:   "testdata/hmc-6502/cpx_cpy_bit_test.bin",
		Offset: 0x0600,
		PC:     0x0600,
		Stop: &conds{Any: true, Conds: []cond{
			condOp(BRK),
			condCycles{1000, math.MaxInt16},
		}},
		Pass: &conds{Conds: []cond{
			condByte{0x0042, 0xa5},
			condA(0xa5),
			condP(V | U | I | Z | C),
			condCycles{85, 85},
		}},
	}
	test.Run(t)
}

func TestBranch(t *testing.T) {
	test := &testBinary{
		Model:  MOS6502,
		Name:   "testdata/hmc-6502/misc_branch_test.bin",
		Offset: 0x0600,
		PC:     0x0600,
		Stop: &conds{Any: true, Conds: []cond{
			condOp(BRK),
			condCycles{1000, math.MaxInt16},
		}},
		Pass: &conds{Conds: []cond{
			condByte{0x0080, 0x1f},
			condA(0x1f),
			condX(0x0d),
			condY(0x54),
			condP(V | U | I | C),
			condCycles{109, 109},
		}},
	}
	test.Run(t)
}

func TestBranchBackwards(t *testing.T) {
	test := &testBinary{
		Model:  MOS6502,
		Name:   "testdata/unit/branch_backwards_test.bin",
		Offset: 0x0600,
		PC:     0x0600,
		Stop: &conds{Any: true, Conds: []cond{
			condOp(BRK),
			condCycles{1000, math.MaxInt16},
		}},
		Pass: &conds{Conds: []cond{
			condX(0xff),
			condCycles{31, 31},
		}},
	}
	test.Run(t)
}

func TestBranchPagecross(t *testing.T) {
	test := &testBinary{
		Model:  MOS6502,
		Name:   "testdata/unit/branch_pagecross_test.bin",
		Offset: 0x02f9,
		PC:     0x02f9,
		Stop: &conds{Any: true, Conds: []cond{
			condOp(BRK),
			condCycles{1000, math.MaxInt16},
		}},
		Pass: &conds{Conds: []cond{
			condA(0xff),
			condCycles{14, 14},
		}},
	}
	test.Run(t)
}

func TestProcessorStatus(t *testing.T) {
	test := &testBinary{
		Model:  MOS6502,
		Name:   "testdata/hmc-6502/flag_test.bin",
		Offset: 0x0600,
		PC:     0x0600,
		Stop: &conds{Any: true, Conds: []cond{
			condOp(BRK),
			condCycles{1000, math.MaxInt16},
		}},
		Pass: &conds{Conds: []cond{
			condByte{0x0030, 0xce},
			condP(N | U | I),
			condCycles{29, 29},
		}},
	}
	test.Run(t)
}

func TestProcessorStatusSpecial(t *testing.T) {
	test := &testBinary{
		Model:  MOS6502,
		Name:   "testdata/hmc-6502/special_flag_test.bin",
		Offset: 0x0600,
		PC:     0x0600,
		Stop: &conds{Any: true, Conds: []cond{
			condOp(BRK),
			condCycles{1000, math.MaxInt16},
		}},
		Pass: &conds{Conds: []cond{
			condByte{0x0020, 0x3c},
			condByte{0x0021, 0x6c},
			condP(U),
			condS(0xff),
			condCycles{31, 31},
		}},
	}
	test.Run(t)
}

func TestStack(t *testing.T) {
	test := &testBinary{
		Model:  MOS6502,
		Name:   "testdata/hmc-6502/stack_test.bin",
		Offset: 0x0600,
		PC:     0x0600,
		Stop: &conds{Any: true, Conds: []cond{
			condOp(BRK),
			condCycles{1000, math.MaxInt16},
		}},
		Pass: &conds{Conds: []cond{
			condByte{0x0030, 0x29},
			condP(U | I),
			condS(0xff),
			condCycles{29, 29},
		}},
	}
	test.Run(t)
}

func TestADD_SUB_BCD(t *testing.T) {
	test := &testBinary{
		Model:  MOS6502,
		Name:   "testdata/unit/bcd_add_sub_test.bin",
		Offset: 0x0600,
		PC:     0x0600,
		Stop: &conds{Any: true, Conds: []cond{
			condOp(BRK),
			condCycles{1000, math.MaxInt16},
		}},
		Pass: &conds{Conds: append([]cond{
			condS(0xf6),
			condP(N | U | D | I),
			condCycles{73, 73},
		}, condStack(
			0x87, 0x91, 0x29, 0x27, 0x34, 0x73, 0x41, 0x46, 0x05,
		)...)},
	}
	test.Run(t)
}

func Test_ADD_SUB_ProcessorStatus(t *testing.T) {
	test := &testBinary{
		Model:  MOS6502,
		Name:   "testdata/unit/add_sub_cvzn_flag_test.bin",
		Offset: 0x0600,
		PC:     0x0600,
		Stop: &conds{Any: true, Conds: []cond{
			condOp(BRK),
			condCycles{1000, math.MaxInt16},
		}},
		Pass: &conds{Conds: []cond{
			condA(0x80),
			condS(0xf3),
			condCycles{108, 108},
			condByte{0x01ff, U | B | I},
			condByte{0x01fe, U | B | I | Z | C},
			condByte{0x01fd, N | V | U | B | I},
			condByte{0x01fc, V | U | B | I | C},
			condByte{0x01fb, U | B | I},
			condByte{0x01fa, U | B | I | Z | C},
			condByte{0x01f9, N | V | U | B | I},
			condByte{0x01f8, V | U | B | I | C},
			condByte{0x01f7, N | U | B | I},
			condByte{0x01f6, V | U | B | I | C},
			condByte{0x01f5, N | V | U | B | I},
			condByte{0x01f4, N | V | U | B | I},
		}},
	}
	test.Run(t)
}

func TestRTI(t *testing.T) {
	test := &testBinary{
		Model:  MOS6502,
		Name:   "testdata/hmc-6502/rti_test.bin",
		Offset: 0x0600,
		PC:     0x0600,
		Stop: &conds{Any: true, Conds: []cond{
			condOp(BRK),
			condCycles{1000, math.MaxInt16},
		}},
		Pass: &conds{Conds: []cond{
			condByte{0x0033, 0x42},
			condP(U | I | C),
			condS(0xff),
			condCycles{40, 40},
		}},
	}
	test.Run(t)
}

func TestBRK(t *testing.T) {
	test := &testBinary{
		Model:  MOS6502,
		Name:   "testdata/unit/brk_test.bin",
		Offset: 0x0600,
		PC:     0x0600,
		Stop: &conds{Any: true, Conds: []cond{
			condOp(NOP),
			condCycles{1000, math.MaxInt16},
		}},
		Pass: &conds{Conds: []cond{
			condByte{0x00ff, 0x44},
			condP(U | I),
			condS(0xff),
			condCycles{89, 89},
		}},
	}
	test.Run(t)
}

func TestHLT(t *testing.T) {
	test := &testBinary{
		Model:  MOS6502,
		Name:   "testdata/unit/kil_test.bin",
		Offset: 0x0600,
		PC:     0x0600,
		Stop: &conds{Any: true, Conds: []cond{
			condCycles{1000, math.MaxInt16},
		}},
		Pass: &conds{Conds: []cond{
			condPC(0x0600),
		}},
	}
	test.Run(t)
}

func TestNOP(t *testing.T) {
	test := &testBinary{
		Model:  MOS6502,
		Name:   "testdata/unit/nop_test.bin",
		Offset: 0x0600,
		PC:     0x0600,
		Stop: &conds{Any: true, Conds: []cond{
			condOp(BRK),
			condCycles{1000, math.MaxInt16},
		}},
		Pass: &conds{Conds: []cond{
			condPC(0x0639),
			condCycles{86, 86},
		}},
	}
	test.Run(t)
}

func TestLAX(t *testing.T) {
	test := &testBinary{
		Model:  MOS6502,
		Name:   "testdata/unit/lax_test.bin",
		Offset: 0x0600,
		PC:     0x0600,
		Stop: &conds{Any: true, Conds: []cond{
			condOp(BRK),
			condCycles{1000, math.MaxInt16},
		}},
		Pass: &conds{Conds: []cond{
			condS(0xf3),
			condP(N | U | I),
			condCycles{135, 135},
			// condStack{0xdb, 0xdb, 0x55, 0x55, 0xff, 0xff, 0x11, 0x11, 0xc3, 0xc3, 0x21, 0x21},
		}},
	}
	test.Run(t)
}

func TestSAX(t *testing.T) {
	test := &testBinary{
		Model:  MOS6502,
		Name:   "testdata/unit/sax_test.bin",
		Offset: 0x0600,
		PC:     0x0600,
		Stop: &conds{Any: true, Conds: []cond{
			condOp(BRK),
			condCycles{1000, math.MaxInt16},
		}},
		Pass: &conds{Conds: []cond{
			condByte{0x0000, 0x80},
			condByte{0x0001, 0x80},
			condByte{0x0002, 0x80},
			condByte{0x0003, 0x80},
			condP(U | I),
			condCycles{33, 33},
		}},
	}
	test.Run(t)
}

func TestIllegalRMW(t *testing.T) {
	test := &testBinary{
		Model:  MOS6502,
		Name:   "testdata/unit/illegal_rmw_test.bin",
		Offset: 0x0600,
		PC:     0x0600,
		Stop: &conds{Any: true, Conds: []cond{
			condOp(BRK),
			condCycles{3000, math.MaxInt16},
		}},
		Pass: &conds{Conds: append([]cond{
			condS(0x81),
			condCycles{1158, 1158},
		}, condStack(
			0x00, 0x00, 0x00, 0x34, 0x0f, 0x4c, 0xb5, 0xd5, 0x4e, 0x35, 0x7d, 0x7b, 0xb4, 0x8d, 0x04, 0x35,
			0x0d, 0x50, 0xb4, 0xa0, 0x76, 0xb5, 0x9b, 0x00, 0xb5, 0xbf, 0x32, 0xb5, 0xbb, 0x3a, 0x35, 0x3b,
			0xec, 0xb5, 0xfe, 0x22, 0xb5, 0xb3, 0x42, 0xb5, 0xf2, 0xba, 0xb5, 0xff, 0x80, 0xb5, 0x80, 0xcc,
			0x75, 0x66, 0xef, 0x35, 0x7b, 0x78, 0x35, 0x6a, 0x9d, 0xb4, 0xd7, 0x79, 0x35, 0x59, 0x11, 0x34,
			0x35, 0x01, 0x34, 0x01, 0x33, 0x35, 0x11, 0x3b, 0x35, 0x33, 0xed, 0xb5, 0xe4, 0x21, 0x37, 0x00,
			0x42, 0x35, 0x40, 0x43, 0x34, 0x01, 0x00, 0xb5, 0xfe, 0x9a, 0xb4, 0xfe, 0x35, 0xb4, 0x97, 0xf1,
			0xb4, 0xfe, 0x3b, 0xb4, 0xfe, 0xf3, 0xb4, 0xfc, 0x38, 0xb4, 0xff, 0xff, 0x37, 0xff, 0x98, 0x35,
			0x99, 0x33, 0x37, 0x33, 0xef, 0x35, 0xf0, 0x39, 0x35, 0x3a, 0xf1, 0xb4, 0xf0, 0x36, 0x35, 0x37,
		)...),
		},
	}
	test.Run(t)
}

func TestXAA(t *testing.T) {
	test := &testBinary{
		Model:  MOS6502,
		Name:   "testdata/unit/illegal_xb_test.bin",
		Offset: 0x0600,
		PC:     0x0600,
		Stop: &conds{Any: true, Conds: []cond{
			condOp(BRK),
			condCycles{1000, math.MaxInt16},
		}},
		Pass: &conds{Conds: []cond{
			condS(0xde),
			condCycles{267, 267},
			//condStack{},
		}},
	}
	test.Run(t)
}

func TestAHX_TAS_SHX_SHY(t *testing.T) {
	test := &testBinary{
		Model:  MOS6502,
		Name:   "testdata/unit/ahx_tas_shx_shy_test.bin",
		Offset: 0x0600,
		PC:     0x0600,
		Stop: &conds{Any: true, Conds: []cond{
			condOp(BRK),
			condCycles{1000, math.MaxInt16},
		}},
		Pass: &conds{Conds: []cond{
			condS(0xf2),
			condCycles{232, 232},
			//condStack{0x01, 0xc9, 0x01, 0x80, 0xc0, 0xe0, 0x01, 0x55, 0x80, 0x80, 0x01, 0x34, 0x10},
		}},
	}
	test.Run(t)
}

func TestNESTest(t *testing.T) {
	test := &testBinary{
		Model:  Ricoh2A03, // This test only works on CPU without BCD!
		Name:   "testdata/nestest/nestest.bin",
		Offset: 0xc000,
		PC:     0xc000,
		S:      0xfd,
		Stop: &conds{Any: true, Conds: []cond{
			condPC(0x0001),
			condOp(BRK),
			condCycles{30000, math.MaxInt16},
		}},
		Pass: &conds{Conds: []cond{
			condPC(0x0001),
			condByte{0x0002, 0x00},
			condByte{0x0003, 0x00},
			condA(0x00),
			condX(0xff),
			condY(0x15),
			condS(0xff),
			condP(U | I | Z | C),
			condCycles{26553, 26553},
		}},
	}
	test.Run(t)
}

func testBlargg(t *testing.T, name, value string, cycles int) {
	if testing.Short() {
		t.Skip("these tests take long to run")
	}

	t.Helper()
	t.Parallel()

	test := &testBinary{
		Model:  Ricoh2A03, // This test only works on CPU without BCD!
		Name:   name,
		Offset: 0x8000,
		//PC:     0xe583,
		Patch: map[uint16]uint8{
			0x6000: 0xff, // Test result 0xff
		},
		S: 0xff,
		Stop: &conds{Any: true, Conds: []cond{
			condTrap{},
			condByte{0x6000, 0x00},
			condCycles{cycles + 128, math.MaxInt32},
		}},
		Pass: &conds{Conds: append([]cond{
			condByte{0x6000, 0x00}, // Success
			condByte{0x6001, 0xDE},
			condByte{0x6002, 0xB0},
			condByte{0x6003, 0x61},
			condCycles{cycles, cycles},
		}, condString(0x6004, value)...)},
		Done: func(cpu CPU, mem *memory.RAM) {
			i := bytes.IndexByte((*mem)[0x6004:], 0x00)
			if i > 0 {
				t.Logf("test result: %q", (*mem)[0x6004:0x6004+i])
			}
		},
	}
	test.Run(t)
}

func TestBlargg01Basics(t *testing.T) {
	testBlargg(t, "testdata/instr_test-v4/rom_singles/01-basics.bin", "\n01-basics\n\nPassed\n", 330200)
}

func TestBlargg02Implied(t *testing.T) {
	testBlargg(t, "testdata/instr_test-v4/rom_singles/02-implied.bin", "\n02-implied\n\nPassed\n", 2687506)
}

func TestBlargg03Immediate(t *testing.T) {
	if os.Getenv("TEST_VERY_LONG") == "" {
		t.Skip("this test taskes *very* long, set TEST_VERY_LONG to run")
	}
	testBlargg(t, "testdata/instr_test-v4/rom_singles/03-immediate.bin", "\n03-immediate\n\nPassed\n", 2388550)
}

func TestBlargg04ZeroPage(t *testing.T) {
	testBlargg(t, "testdata/instr_test-v4/rom_singles/04-zero_page.bin", "\n04-zero_page\n\nPassed\n", 3273464)
}

func TestBlargg05ZeroPageXY(t *testing.T) {
	if os.Getenv("TEST_VERY_LONG") == "" {
		t.Skip("this test taskes *very* long, set TEST_VERY_LONG to run")
	}
	testBlargg(t, "testdata/instr_test-v4/rom_singles/05-zp_xy.bin", "\n05-zp_xy\n\nPassed\n", 7558100)
}

func TestBlargg06Absolute(t *testing.T) {
	testBlargg(t, "testdata/instr_test-v4/rom_singles/06-absolute.bin", "\n06-absolute\n\nPassed\n", 3093993)
}

func TestBlargg07AbsoluteXY(t *testing.T) {
	testBlargg(t, "testdata/instr_test-v4/rom_singles/07-abs_xy.bin", "\n07-abs_xy\n\nPassed\n", 10675054)
}

func TestBlargg08IndirectX(t *testing.T) {
	testBlargg(t, "testdata/instr_test-v4/rom_singles/08-ind_x.bin", "\n08-ind_x\n\nPassed\n", 4145448)
}

func TestBlargg09IndirectY(t *testing.T) {
	testBlargg(t, "testdata/instr_test-v4/rom_singles/09-ind_y.bin", "\n09-ind_y\n\nPassed\n", 3888212)
}

func TestBlargg10Branches(t *testing.T) {
	testBlargg(t, "testdata/instr_test-v4/rom_singles/10-branches.bin", "\n10-branches\n\nPassed\n", 1033363)
}

func TestBlargg11Stack(t *testing.T) {
	testBlargg(t, "testdata/instr_test-v4/rom_singles/11-stack.bin", "\n11-stack\n\nPassed\n", 4682762)
}

func TestBlargg12Jump(t *testing.T) {
	testBlargg(t, "testdata/instr_test-v4/rom_singles/12-jmp_jsr.bin", "\n12-jmp_jsr\n\nPassed\n", 322698)
}

func TestBlargg13RTS(t *testing.T) {
	testBlargg(t, "testdata/instr_test-v4/rom_singles/13-rts.bin", "\n13-rts\n\nPassed\n", 223119)
}

func TestBlargg14RTI(t *testing.T) {
	testBlargg(t, "testdata/instr_test-v4/rom_singles/14-rti.bin", "", 225084)
}

func TestBlargg15BRK(t *testing.T) {
	testBlargg(t, "testdata/instr_test-v4/rom_singles/15-brk.bin", "", 540320)
}

func TestBlargg16SPecial(t *testing.T) {
	testBlargg(t, "testdata/instr_test-v4/rom_singles/16-special.bin", "", 157793)
}

func TestFunctional6502Test(t *testing.T) {
	if testing.Short() {
		t.Skip("these tests take long to run")
	}
	if os.Getenv("TEST_VERY_LONG") == "" {
		t.Skip("this test taskes *very* long, set TEST_VERY_LONG to run")
	}
	t.Parallel()

	test := &testBinary{
		Model:  Ricoh2A03, // This test only works on CPU without BCD!
		Name:   "testdata/6502_functional_tests/6502_functional_test.bin",
		Offset: 0x0400,
		PC:     0x0400,
		Stop: &conds{Any: true, Conds: []cond{
			condTrap{},
		}},
		Pass: &conds{Conds: []cond{
			condPC(0x32e9),
			condCycles{92608051, 92608051},
		}},
	}
	test.Run(t)
}

func TestTrap(t *testing.T) {
	test := &testBinary{
		Model:  MOS6502,
		Name:   "testdata/trap.bin",
		Offset: 0x0000,
		PC:     0x0000,
		Stop: &conds{Any: true, Conds: []cond{
			condTrap{},
		}},
		Pass: &conds{Conds: []cond{
			condCycles{4, 4},
			condPC(0x0002),
		}},
	}
	test.Run(t)
}

func TestTrapIndirect(t *testing.T) {
	test := &testBinary{
		Model:  MOS6502,
		Name:   "testdata/trap_ind.bin",
		Offset: 0x0000,
		PC:     0x0000,
		Stop: &conds{Any: true, Conds: []cond{
			condTrap{},
		}},
		Pass: &conds{Conds: []cond{
			condCycles{5, 5},
			condByte{0x00f0, 0x04},
			condPC(0x0004),
		}},
	}
	test.Run(t)
}

func TestTrapJSR(t *testing.T) {
	test := &testBinary{
		Model:  MOS6502,
		Name:   "testdata/trap_jsr.bin",
		Offset: 0x0000,
		PC:     0x0000,
		Stop: &conds{Any: true, Conds: []cond{
			condTrap{},
		}},
		Pass: &conds{Conds: []cond{
			condCycles{4, 4},
			condPC(0x0002),
		}},
	}
	test.Run(t)
}
