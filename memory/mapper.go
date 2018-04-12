package memory

import (
	"fmt"
	"sort"
	"strings"
)

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
func (m *Mapper) Map(addr, stop uint16, memory Memory) {
	m.mapped = append(m.mapped, memoryRange{
		Memory: memory,
		addr:   addr,
		stop:   stop,
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

// Reset the mappings
func (m *Mapper) Reset() *Mapper {
	m.mapped = m.mapped[:0]
	return m
}

func (m Mapper) String() string {
	return fmt.Sprintf("Mapper{%s}", m.mapped)
}

type memoryRange struct {
	Memory
	addr, stop uint16
}

func (r memoryRange) String() string {
	return fmt.Sprintf("$%04X-$%04X: %v", r.addr, r.stop, r.Memory)
}

// memoryRanges are 0-n (non-contiguous) memory ranges
type memoryRanges []memoryRange

func (r memoryRanges) Len() int {
	return len(r)
}

func (r memoryRanges) Less(i, j int) bool {
	if r[j].addr >= r[i].addr && r[j].stop <= r[i].stop {
		// If j is contained in i, return false; the smaller area takes precdence
		return false
	}
	return r[i].addr < r[j].addr
}

func (r memoryRanges) Swap(i, j int) {
	r[i], r[j] = r[j], r[i]
}

func (r memoryRanges) Sort() {
	sort.Stable(r)
}

func (r memoryRanges) String() string {
	s := make([]string, len(r))
	for i, m := range r {
		s[i] = m.String()
	}
	return strings.Join(s, ", ")
}

func (r memoryRanges) Bank(addr uint16) Memory {
	l := len(r)
	if i := sort.Search(l, func(i int) bool {
		return addr <= r[i].stop
	}); i < l {
		if it := r[i]; addr >= it.addr && addr <= it.stop {
			return it
		}
	}
	return nil
}
