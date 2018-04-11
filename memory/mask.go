package memory

// Masked memory access allows one to restrict and/or translate 16-bit memory
// to a smaller range.
type Masked struct {
	Memory

	// Mask is the memory mask. To limit memory access to 1k for example, use
	// a mask of 2^10-1 (0x3ff).
	Mask uint16
}

// Fetch a byte
func (m Masked) Fetch(addr uint16) uint8 {
	return m.Memory.Fetch(addr & m.Mask)
}

// Store a byte
func (m Masked) Store(addr uint16, value uint8) {
	m.Memory.Store(addr&m.Mask, value)
}
