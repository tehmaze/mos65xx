package mos65xx

// Frequency scale
const (
	Hz  = 1
	KHz = 1000 * Hz
	MHz = 1000 * KHz
)

// Model of the MOS Technology 65xx (or compatible) CPU
type Model struct {
	Name           string
	Frequency      float64 // Typical clock frequency in Hz
	ExternalMemory int     // External addressable memory size
	InternalMemory int     // Internal RAM size
	HasBCD         bool    // Decimal mode support
	HasIRQ         bool    // IRQ support
	HasNMI         bool    // NMI support
	HasReady       bool    // RDY support
}

// Models
var (
	MOS6502 = Model{
		Name:           "MOS Technology 6502",
		Frequency:      1 * MHz,
		ExternalMemory: 0x10000,
		HasBCD:         true,
		HasIRQ:         true,
		HasNMI:         true,
	}

	MOS6503 = Model{
		Name:           "MOS Technology 6503",
		Frequency:      1 * MHz,
		ExternalMemory: 0x1000, // 4 kB
		HasBCD:         true,
		HasIRQ:         true,
		HasNMI:         true,
	}

	MOS6504 = Model{
		Name:           "MOS Technology 6504",
		Frequency:      1 * MHz,
		ExternalMemory: 0x2000, // 8 kB
		HasBCD:         true,
		HasIRQ:         true,
	}

	MOS6505 = Model{
		Name:           "MOS Technology 6505",
		Frequency:      1 * MHz,
		ExternalMemory: 0x1000, // 4 kB
		HasBCD:         true,
		HasIRQ:         true,
		HasReady:       true,
	}

	MOS6506 = Model{
		Name:           "MOS Technology 6506",
		Frequency:      1 * MHz,
		ExternalMemory: 0x1000, // 4 kB
		HasBCD:         true,
		HasIRQ:         true,
	}

	MOS6507 = Model{
		Name:           "MOS Technology 6507",
		Frequency:      1 * MHz,
		ExternalMemory: 0x2000, // 8 kB
	}

	MOS6510 = Model{
		Name:           "MOS Technology 6510",
		Frequency:      1.023 * MHz, // On NTSC, for PAL use 0.985 MHz
		ExternalMemory: 0x10000,
		HasBCD:         true,
		HasNMI:         true,
		HasReady:       true,
	}

	MOS6510T = Model{
		Name:           "MOS Technology 6510T",
		Frequency:      1.023 * MHz, // On NTSC, for PAL use 0.985 MHz
		ExternalMemory: 0x10000,
		HasBCD:         true,
	}

	MOS7501 = Model{
		Name:           "MOS Technology 7501",
		Frequency:      1.023 * MHz, // On NTSC, for PAL use 0.985 MHz
		ExternalMemory: 0x10000,
		HasBCD:         true,
		HasReady:       true,
	}

	MOS8501 = Model{
		Name:           "MOS Technology 8501",
		Frequency:      1.023 * MHz, // On NTSC, for PAL use 0.985 MHz
		ExternalMemory: 0x10000,
		HasBCD:         true,
		HasReady:       true,
	}

	MOS8502 = Model{
		Name:           "MOS Technology 8502",
		Frequency:      2 * MHz,
		ExternalMemory: 0x10000,
		HasBCD:         true,
		HasNMI:         true,
		HasReady:       true,
	}

	// Ricoh2A03 is the 8-bit microprocessor in the Nintendo Entertainment System (NTSC version)
	Ricoh2A03 = Model{
		Name:           "Ricoh 2A03",
		Frequency:      1 * MHz,
		ExternalMemory: 0x10000,
		HasIRQ:         true,
		HasNMI:         true,
	}

	// Ricoh2A07 is the 8-bit microprocessor in the Nintendo Entertainment System (PAL version)
	Ricoh2A07 = Model{
		Name:           "Ricoh 2A07",
		Frequency:      1 * MHz,
		ExternalMemory: 0x10000,
		HasIRQ:         true,
		HasNMI:         true,
	}
)
