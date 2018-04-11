package mos65xx

// Mnemonic is an instruction
type Mnemonic uint8

// mnemonics
const (
	ADC Mnemonic = iota
	AND
	ASL
	BCC
	BCS
	BEQ
	BIT
	BMI
	BNE
	BPL
	BRK
	BVC
	BVS
	CLC
	CLD
	CLI
	CLV
	CMP
	CPX
	CPY
	DEC
	DEX
	DEY
	EOR
	INC
	INX
	INY
	JMP
	JSR
	LDA
	LDX
	LDY
	LSR
	NOP
	ORA
	PHA
	PHP
	PLA
	PLP
	ROL
	ROR
	RTI
	RTS
	SBC
	SEC
	SED
	SEI
	STA
	STX
	STY
	TAX
	TAY
	TSX
	TXA
	TXS
	TYA
	HLT
	LAX
	SAX
	DCP
	ISC
	RLA
	RRA
	SLO
	SRE
	ANC
	ALR
	ARR
	XAA
	AHX
	TAS
	SHX
	SHY
	LAS
	AXS
	mnemonics // For counting
)

var mnemonicName = [mnemonics]string{
	"ADC", "AND", "ASL", "BCC", "BCS", "BEQ", "BIT", "BMI", "BNE", "BPL",
	"BRK", "BVC", "BVS", "CLC", "CLD", "CLI", "CLV", "CMP", "CPX", "CPY",
	"DEC", "DEX", "DEY", "EOR", "INC", "INX", "INY", "JMP", "JSR", "LDA",
	"LDX", "LDY", "LSR", "NOP", "ORA", "PHA", "PHP", "PLA", "PLP", "ROL",
	"ROR", "RTI", "RTS", "SBC", "SEC", "SED", "SEI", "STA", "STX", "STY",
	"TAX", "TAY", "TSX", "TXA", "TXS", "TYA", "HLT", "LAX", "SAX", "DCP",
	"ISC", "RLA", "RRA", "SLO", "SRE", "ANC", "ALR", "ARR", "XAA", "AHX",
	"TAS", "SHX", "SHY", "LAS", "AXS",
}

func (m Mnemonic) String() string {
	return mnemonicName[m]
}

// opcode is a CPU operation code
type opcode struct {
	Mnemonic
	Size            int
	Cycles          int
	PageCrossCycles int
	Mode            AddressMode
}

// opcodes
var opcodes = [0x100]opcode{
	{BRK, 1, 7, 0, Implied},         // 0x00
	{ORA, 2, 6, 0, IndexedIndirect}, // 0x01
	{HLT, 1, 0, 0, Implied},         // 0x02
	{SLO, 2, 8, 0, IndexedIndirect}, // 0x03
	{NOP, 2, 3, 0, ZeroPage},        // 0x04
	{ORA, 2, 3, 0, ZeroPage},        // 0x05
	{ASL, 2, 5, 0, ZeroPage},        // 0x06
	{SLO, 2, 5, 0, ZeroPage},        // 0x07
	{PHP, 1, 3, 0, Implied},         // 0x08
	{ORA, 2, 2, 0, Immediate},       // 0x09
	{ASL, 1, 2, 0, Accumulator},     // 0x0a
	{ANC, 2, 2, 0, Immediate},       // 0x0b
	{NOP, 3, 4, 0, Absolute},        // 0x0c
	{ORA, 3, 4, 0, Absolute},        // 0x0d
	{ASL, 3, 6, 0, Absolute},        // 0x0e
	{SLO, 3, 6, 0, Absolute},        // 0x0f
	{BPL, 2, 2, 0, Relative},        // 0x10
	{ORA, 2, 5, 1, IndirectIndexed}, // 0x11
	{HLT, 1, 0, 0, Implied},         // 0x12
	{SLO, 2, 8, 0, IndirectIndexed}, // 0x13
	{NOP, 2, 4, 0, ZeroPageX},       // 0x14
	{ORA, 2, 4, 0, ZeroPageX},       // 0x15
	{ASL, 2, 6, 0, ZeroPageX},       // 0x16
	{SLO, 2, 6, 0, ZeroPageX},       // 0x17
	{CLC, 1, 2, 0, Implied},         // 0x18
	{ORA, 3, 4, 1, AbsoluteY},       // 0x19
	{NOP, 1, 2, 0, Implied},         // 0x1a
	{SLO, 3, 7, 0, AbsoluteY},       // 0x1b
	{NOP, 3, 4, 1, AbsoluteX},       // 0x1c
	{ORA, 3, 4, 1, AbsoluteX},       // 0x1d
	{ASL, 3, 7, 0, AbsoluteX},       // 0x1e
	{SLO, 3, 7, 0, AbsoluteX},       // 0x1f
	{JSR, 3, 6, 0, Absolute},        // 0x20
	{AND, 2, 6, 0, IndexedIndirect}, // 0x21
	{HLT, 1, 0, 0, Implied},         // 0x22
	{RLA, 2, 8, 0, IndexedIndirect}, // 0x23
	{BIT, 2, 3, 0, ZeroPage},        // 0x24
	{AND, 2, 3, 0, ZeroPage},        // 0x25
	{ROL, 2, 5, 0, ZeroPage},        // 0x26
	{RLA, 2, 5, 0, ZeroPage},        // 0x27
	{PLP, 1, 4, 0, Implied},         // 0x28
	{AND, 2, 2, 0, Immediate},       // 0x29
	{ROL, 1, 2, 0, Accumulator},     // 0x2a
	{ANC, 2, 2, 0, Immediate},       // 0x2b
	{BIT, 3, 4, 0, Absolute},        // 0x2c
	{AND, 3, 4, 0, Absolute},        // 0x2d
	{ROL, 3, 6, 0, Absolute},        // 0x2e
	{RLA, 3, 6, 0, Absolute},        // 0x2f
	{BMI, 2, 2, 0, Relative},        // 0x30
	{AND, 2, 5, 1, IndirectIndexed}, // 0x31
	{HLT, 1, 0, 0, Implied},         // 0x32
	{RLA, 2, 8, 0, IndirectIndexed}, // 0x33
	{NOP, 2, 4, 0, ZeroPageX},       // 0x34
	{AND, 2, 4, 0, ZeroPageX},       // 0x35
	{ROL, 2, 6, 0, ZeroPageX},       // 0x36
	{RLA, 2, 6, 0, ZeroPageX},       // 0x37
	{SEC, 1, 2, 0, Implied},         // 0x38
	{AND, 3, 4, 1, AbsoluteY},       // 0x39
	{NOP, 1, 2, 0, Implied},         // 0x3a
	{RLA, 3, 7, 0, AbsoluteY},       // 0x3b
	{NOP, 3, 4, 1, AbsoluteX},       // 0x3c
	{AND, 3, 4, 1, AbsoluteX},       // 0x3d
	{ROL, 3, 7, 0, AbsoluteX},       // 0x3e
	{RLA, 3, 7, 0, AbsoluteX},       // 0x3f
	{RTI, 1, 6, 0, Implied},         // 0x40
	{EOR, 2, 6, 0, IndexedIndirect}, // 0x41
	{HLT, 1, 0, 0, Implied},         // 0x42
	{SRE, 2, 8, 0, IndexedIndirect}, // 0x43
	{NOP, 2, 3, 0, ZeroPage},        // 0x44
	{EOR, 2, 3, 0, ZeroPage},        // 0x45
	{LSR, 2, 5, 0, ZeroPage},        // 0x46
	{SRE, 2, 5, 0, ZeroPage},        // 0x47
	{PHA, 1, 3, 0, Implied},         // 0x48
	{EOR, 2, 2, 0, Immediate},       // 0x49
	{LSR, 1, 2, 0, Accumulator},     // 0x4a
	{ALR, 2, 2, 0, Immediate},       // 0x4b
	{JMP, 3, 3, 0, Absolute},        // 0x4c
	{EOR, 3, 4, 0, Absolute},        // 0x4d
	{LSR, 3, 6, 0, Absolute},        // 0x4e
	{SRE, 3, 6, 0, Absolute},        // 0x4f
	{BVC, 2, 2, 0, Relative},        // 0x50
	{EOR, 2, 5, 1, IndirectIndexed}, // 0x51
	{HLT, 1, 0, 0, Implied},         // 0x52
	{SRE, 2, 8, 0, IndirectIndexed}, // 0x53
	{NOP, 2, 4, 0, ZeroPageX},       // 0x54
	{EOR, 2, 4, 0, ZeroPageX},       // 0x55
	{LSR, 2, 6, 0, ZeroPageX},       // 0x56
	{SRE, 2, 6, 0, ZeroPageX},       // 0x57
	{CLI, 1, 2, 0, Implied},         // 0x58
	{EOR, 3, 4, 1, AbsoluteY},       // 0x59
	{NOP, 1, 2, 0, Implied},         // 0x5a
	{SRE, 3, 7, 0, AbsoluteY},       // 0x5b
	{NOP, 3, 4, 1, AbsoluteX},       // 0x5c
	{EOR, 3, 4, 1, AbsoluteX},       // 0x5d
	{LSR, 3, 7, 0, AbsoluteX},       // 0x5e
	{SRE, 3, 7, 0, AbsoluteX},       // 0x5f
	{RTS, 1, 6, 0, Implied},         // 0x60
	{ADC, 2, 6, 0, IndexedIndirect}, // 0x61
	{HLT, 1, 0, 0, Implied},         // 0x62
	{RRA, 2, 8, 0, IndexedIndirect}, // 0x63
	{NOP, 2, 3, 0, ZeroPage},        // 0x64
	{ADC, 2, 3, 0, ZeroPage},        // 0x65
	{ROR, 2, 5, 0, ZeroPage},        // 0x66
	{RRA, 2, 5, 0, ZeroPage},        // 0x67
	{PLA, 1, 4, 0, Implied},         // 0x68
	{ADC, 2, 2, 0, Immediate},       // 0x69
	{ROR, 1, 2, 0, Accumulator},     // 0x6a
	{ARR, 2, 2, 0, Immediate},       // 0x6b
	{JMP, 3, 5, 0, Indirect},        // 0x6c
	{ADC, 3, 4, 0, Absolute},        // 0x6d
	{ROR, 3, 6, 0, Absolute},        // 0x6e
	{RRA, 3, 6, 0, Absolute},        // 0x6f
	{BVS, 2, 2, 0, Relative},        // 0x70
	{ADC, 2, 5, 1, IndirectIndexed}, // 0x71
	{HLT, 1, 0, 0, Implied},         // 0x72
	{RRA, 2, 8, 0, IndirectIndexed}, // 0x73
	{NOP, 2, 4, 0, ZeroPageX},       // 0x74
	{ADC, 2, 4, 0, ZeroPageX},       // 0x75
	{ROR, 2, 6, 0, ZeroPageX},       // 0x76
	{RRA, 2, 6, 0, ZeroPageX},       // 0x77
	{SEI, 1, 2, 0, Implied},         // 0x78
	{ADC, 3, 4, 1, AbsoluteY},       // 0x79
	{NOP, 1, 2, 0, Implied},         // 0x7a
	{RRA, 3, 7, 0, AbsoluteY},       // 0x7b
	{NOP, 3, 4, 1, AbsoluteX},       // 0x7c
	{ADC, 3, 4, 1, AbsoluteX},       // 0x7d
	{ROR, 3, 7, 0, AbsoluteX},       // 0x7e
	{RRA, 3, 7, 0, AbsoluteX},       // 0x7f
	{NOP, 2, 2, 0, Immediate},       // 0x80
	{STA, 2, 6, 0, IndexedIndirect}, // 0x81
	{NOP, 2, 2, 0, Immediate},       // 0x82
	{SAX, 2, 6, 0, IndexedIndirect}, // 0x83
	{STY, 2, 3, 0, ZeroPage},        // 0x84
	{STA, 2, 3, 0, ZeroPage},        // 0x85
	{STX, 2, 3, 0, ZeroPage},        // 0x86
	{SAX, 2, 3, 0, ZeroPage},        // 0x87
	{DEY, 1, 2, 0, Implied},         // 0x88
	{NOP, 2, 2, 0, Immediate},       // 0x89
	{TXA, 1, 2, 0, Implied},         // 0x8a
	{XAA, 2, 2, 0, Immediate},       // 0x8b
	{STY, 3, 4, 0, Absolute},        // 0x8c
	{STA, 3, 4, 0, Absolute},        // 0x8d
	{STX, 3, 4, 0, Absolute},        // 0x8e
	{SAX, 3, 4, 0, Absolute},        // 0x8f
	{BCC, 2, 2, 0, Relative},        // 0x90
	{STA, 2, 6, 0, IndirectIndexed}, // 0x91
	{HLT, 1, 0, 0, Implied},         // 0x92
	{AHX, 2, 6, 0, IndirectIndexed}, // 0x93
	{STY, 2, 4, 0, ZeroPageX},       // 0x94
	{STA, 2, 4, 0, ZeroPageX},       // 0x95
	{STX, 2, 4, 0, ZeroPageY},       // 0x96
	{SAX, 2, 4, 0, ZeroPageY},       // 0x97
	{TYA, 1, 2, 0, Implied},         // 0x98
	{STA, 3, 5, 0, AbsoluteY},       // 0x99
	{TXS, 1, 2, 0, Implied},         // 0x9a
	{TAS, 3, 5, 0, AbsoluteY},       // 0x9b
	{SHY, 3, 5, 0, AbsoluteX},       // 0x9c
	{STA, 3, 5, 0, AbsoluteX},       // 0x9d
	{SHX, 3, 5, 0, AbsoluteY},       // 0x9e
	{AHX, 3, 5, 0, AbsoluteY},       // 0x9f
	{LDY, 2, 2, 0, Immediate},       // 0xa0
	{LDA, 2, 6, 0, IndexedIndirect}, // 0xa1
	{LDX, 2, 2, 0, Immediate},       // 0xa2
	{LAX, 2, 6, 0, IndexedIndirect}, // 0xa3
	{LDY, 2, 3, 0, ZeroPage},        // 0xa4
	{LDA, 2, 3, 0, ZeroPage},        // 0xa5
	{LDX, 2, 3, 0, ZeroPage},        // 0xa6
	{LAX, 2, 3, 0, ZeroPage},        // 0xa7
	{TAY, 1, 2, 0, Implied},         // 0xa8
	{LDA, 2, 2, 0, Immediate},       // 0xa9
	{TAX, 1, 2, 0, Implied},         // 0xaa
	{LAX, 2, 2, 0, Immediate},       // 0xab
	{LDY, 3, 4, 0, Absolute},        // 0xac
	{LDA, 3, 4, 0, Absolute},        // 0xad
	{LDX, 3, 4, 0, Absolute},        // 0xae
	{LAX, 3, 4, 0, Absolute},        // 0xaf
	{BCS, 2, 2, 0, Relative},        // 0xb0
	{LDA, 2, 5, 1, IndirectIndexed}, // 0xb1
	{HLT, 1, 0, 0, Implied},         // 0xb2
	{LAX, 2, 5, 1, IndirectIndexed}, // 0xb3
	{LDY, 2, 4, 0, ZeroPageX},       // 0xb4
	{LDA, 2, 4, 0, ZeroPageX},       // 0xb5
	{LDX, 2, 4, 0, ZeroPageY},       // 0xb6
	{LAX, 2, 4, 0, ZeroPageY},       // 0xb7
	{CLV, 1, 2, 0, Implied},         // 0xb8
	{LDA, 3, 4, 1, AbsoluteY},       // 0xb9
	{TSX, 1, 2, 0, Implied},         // 0xba
	{LAS, 3, 4, 1, AbsoluteY},       // 0xbb
	{LDY, 3, 4, 1, AbsoluteX},       // 0xbc
	{LDA, 3, 4, 1, AbsoluteX},       // 0xbd
	{LDX, 3, 4, 1, AbsoluteY},       // 0xbe
	{LAX, 3, 4, 1, AbsoluteY},       // 0xbf
	{CPY, 2, 2, 0, Immediate},       // 0xc0
	{CMP, 2, 6, 0, IndexedIndirect}, // 0xc1
	{NOP, 2, 2, 0, Immediate},       // 0xc2
	{DCP, 2, 8, 0, IndexedIndirect}, // 0xc3
	{CPY, 2, 3, 0, ZeroPage},        // 0xc4
	{CMP, 2, 3, 0, ZeroPage},        // 0xc5
	{DEC, 2, 5, 0, ZeroPage},        // 0xc6
	{DCP, 2, 5, 0, ZeroPage},        // 0xc7
	{INY, 1, 2, 0, Implied},         // 0xc8
	{CMP, 2, 2, 0, Immediate},       // 0xc9
	{DEX, 1, 2, 0, Implied},         // 0xca
	{AXS, 2, 2, 0, Immediate},       // 0xcb
	{CPY, 3, 4, 0, Absolute},        // 0xcc
	{CMP, 3, 4, 0, Absolute},        // 0xcd
	{DEC, 3, 6, 0, Absolute},        // 0xce
	{DCP, 3, 6, 0, Absolute},        // 0xcf
	{BNE, 2, 2, 0, Relative},        // 0xd0
	{CMP, 2, 5, 1, IndirectIndexed}, // 0xd1
	{HLT, 1, 0, 0, Implied},         // 0xd2
	{DCP, 2, 8, 0, IndirectIndexed}, // 0xd3
	{NOP, 2, 4, 0, ZeroPageX},       // 0xd4
	{CMP, 2, 4, 0, ZeroPageX},       // 0xd5
	{DEC, 2, 6, 0, ZeroPageX},       // 0xd6
	{DCP, 2, 6, 0, ZeroPageX},       // 0xd7
	{CLD, 1, 2, 0, Implied},         // 0xd8
	{CMP, 3, 4, 1, AbsoluteY},       // 0xd9
	{NOP, 1, 2, 0, Implied},         // 0xda
	{DCP, 3, 7, 0, AbsoluteY},       // 0xdb
	{NOP, 3, 4, 1, AbsoluteX},       // 0xdc
	{CMP, 3, 4, 1, AbsoluteX},       // 0xdd
	{DEC, 3, 7, 0, AbsoluteX},       // 0xde
	{DCP, 3, 7, 0, AbsoluteX},       // 0xdf
	{CPX, 2, 2, 0, Immediate},       // 0xe0
	{SBC, 2, 6, 0, IndexedIndirect}, // 0xe1
	{NOP, 2, 2, 0, Immediate},       // 0xe2
	{ISC, 2, 8, 0, IndexedIndirect}, // 0xe3
	{CPX, 2, 3, 0, ZeroPage},        // 0xe4
	{SBC, 2, 3, 0, ZeroPage},        // 0xe5
	{INC, 2, 5, 0, ZeroPage},        // 0xe6
	{ISC, 2, 5, 0, ZeroPage},        // 0xe7
	{INX, 1, 2, 0, Implied},         // 0xe8
	{SBC, 2, 2, 0, Immediate},       // 0xe9
	{NOP, 1, 2, 0, Implied},         // 0xea
	{SBC, 2, 2, 0, Immediate},       // 0xeb
	{CPX, 3, 4, 0, Absolute},        // 0xec
	{SBC, 3, 4, 0, Absolute},        // 0xed
	{INC, 3, 6, 0, Absolute},        // 0xee
	{ISC, 3, 6, 0, Absolute},        // 0xef
	{BEQ, 2, 2, 1, Relative},        // 0xf0
	{SBC, 2, 5, 1, IndirectIndexed}, // 0xf1
	{HLT, 1, 0, 0, Implied},         // 0xf2
	{ISC, 2, 8, 0, IndirectIndexed}, // 0xf3
	{NOP, 2, 4, 0, ZeroPageX},       // 0xf4
	{SBC, 2, 4, 0, ZeroPageX},       // 0xf5
	{INC, 2, 6, 0, ZeroPageX},       // 0xf6
	{ISC, 2, 6, 0, ZeroPageX},       // 0xf7
	{SED, 1, 2, 0, Implied},         // 0xf8
	{SBC, 3, 4, 1, AbsoluteY},       // 0xf9
	{NOP, 1, 2, 0, Implied},         // 0xfa
	{ISC, 3, 7, 0, AbsoluteY},       // 0xfb
	{NOP, 3, 4, 1, AbsoluteX},       // 0xfc
	{SBC, 3, 4, 1, AbsoluteX},       // 0xfd
	{INC, 3, 7, 0, AbsoluteX},       // 0xfe
	{ISC, 3, 7, 0, AbsoluteX},       // 0xff
}
