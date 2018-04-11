package memory

import "testing"

func TestMapper(t *testing.T) {
	m := NewMapper()
	r := New(0x2000).Reset(0xaa)

	// Test Map (and sorting)
	m.Map(0x2000, 0x2000, Masked{New(0x2000).Reset(0x55), 0x1fff})
	m.Map(0x0000, 0x2000, r)
	m.Map(0x4000, 0x2000, Masked{make(ROM, 0x2000), 0x1fff})

	// Test RAM
	if v := m.Fetch(0x1234); v != 0xaa {
		t.Fatalf("expected 0xaa at 0x1234, got %#02x", v)
	}
	m.Store(0x1234, 0xff)
	if v := m.Fetch(0x1234); v != 0xff {
		t.Fatalf("expected 0xff at 0x1234, got %#02x", v)
	}

	// Test masked RAM
	if v := m.Fetch(0x2345); v != 0x55 {
		t.Fatalf("expected 0x55 at 0x2345, got %#02x", v)
	}
	m.Store(0x2345, 0xff)
	if v := m.Fetch(0x2345); v != 0xff {
		t.Fatalf("expected 0xff at 0x2345, got %#02x", v)
	}

	// Test ROM
	if v := m.Fetch(0x4567); v != 0x00 {
		t.Fatalf("expected 0x00 at 0x4567, got %#02x", v)
	}
	m.Store(0x4567, 0xff)
	if v := m.Fetch(0x4567); v != 0x00 {
		t.Fatalf("expected 0x00 at 0x4567, got %#02x", v)
	}

	// Test zero value
	if v := m.Fetch(0xffff); v != 0xff {
		t.Fatalf("expected 0xff at 0xffff, got %#02x", v)
	}

	// Test unmap
	if !m.Unmap(r) {
		t.Fatal("unmap failed")
	}
	if m.Unmap(r) {
		t.Fatal("unmap should have returned false for unmapped memory")
	}
}
