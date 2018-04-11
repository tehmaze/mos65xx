package memory

import "sort"

// Mapper for bank switched memory access.
type Mapper struct {
	// Zero value for unmapped areas.
	Zero uint8

	// mapper memory ranges
	mapped memoryRanges
}

// NewMapper creates a new mapper with 0xff as the zero value.
func NewMapper() *Mapper {
	return &Mapper{Zero: 0xff}
}

// Fetch a byte
func (m Mapper) Fetch(addr uint16) uint8 {
	if memory := m.mapped.Bank(addr); memory != nil {
		return memory.Fetch(addr)
	}
	return m.Zero
}

// Store a byte
func (m Mapper) Store(addr uint16, value uint8) {
	if memory := m.mapped.Bank(addr); memory != nil {
		memory.Store(addr, value)
	}
}

// Map memory starting at addr; the memory implementation is expected to do
// the address translation for the specified addr.
func (m *Mapper) Map(addr, size uint16, memory Memory) {
	m.mapped = append(m.mapped, memoryRange{
		Memory: memory,
		addr:   addr,
		stop:   addr + size,
	})
	m.mapped.Sort()
}

// Unmap a memory area; returns true if the memory was found. Returns at the
// first hit.
func (m *Mapper) Unmap(memory Memory) (found bool) {
	for i, r := range m.mapped {
		if found = r.Memory == memory; found {
			m.mapped = append(m.mapped[:i], m.mapped[i+1:]...)
			return
		}
	}
	return
}

type memoryRange struct {
	Memory
	addr, stop uint16
}

// memoryRanges are 0-n (non-contiguous) memory ranges
type memoryRanges []memoryRange

func (r memoryRanges) Len() int           { return len(r) }
func (r memoryRanges) Less(i, j int) bool { return r[i].addr < r[j].addr }
func (r memoryRanges) Swap(i, j int)      { r[i], r[j] = r[j], r[i] }
func (r memoryRanges) Sort()              { sort.Stable(r) }
func (r memoryRanges) Bank(addr uint16) Memory {
	l := len(r)
	if i := sort.Search(l, func(i int) bool {
		return addr < r[i].stop
	}); i < l {
		if it := r[i]; addr >= it.addr && addr < it.stop {
			return it
		}
	}
	return nil
}
