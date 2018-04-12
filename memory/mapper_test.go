package memory

import "testing"

func TestMapper(t *testing.T) {
	m := NewMapper()
	r := New(0x2000).Reset(0xaa)

	// Test Map (and sorting)
	m.Map(0x2000, 0x3fff, Masked{New(0x2000).Reset(0x55), 0x1fff})
	m.Map(0x0000, 0x1fff, r)
	m.Map(0x4000, 0x5fff, Masked{make(ROM, 0x2000), 0x1fff})
	m.Map(0x0000, 0x7fff, Blank(0x2a))

	// Test RAM
	if v := m.Fetch(0x1234); v != 0xaa {
		t.Fatalf("expected 0xaa at 0x1234, got %#02x", v)
	} else {
		t.Logf("0x1234 = %#02x", v)
	}
	m.Store(0x1234, 0xff)
	if v := m.Fetch(0x1234); v != 0xff {
		t.Fatalf("expected 0xff at 0x1234, got %#02x", v)
	} else {
		t.Logf("0x1234 = %#02x", v)
	}

	// Test masked RAM
	if v := m.Fetch(0x2345); v != 0x55 {
		t.Fatalf("expected 0x55 at 0x2345, got %#02x", v)
	} else {
		t.Logf("0x2345 = %#02x", v)
	}
	m.Store(0x2345, 0xff)
	if v := m.Fetch(0x2345); v != 0xff {
		t.Fatalf("expected 0xff at 0x2345, got %#02x", v)
	} else {
		t.Logf("0x2345 = %#02x", v)
	}

	// Test ROM
	if v := m.Fetch(0x4567); v != 0x00 {
		t.Fatalf("expected 0x00 at 0x4567, got %#02x", v)
	} else {
		t.Logf("0x4567 = %#02x", v)
	}
	m.Store(0x4567, 0xff)
	if v := m.Fetch(0x4567); v != 0x00 {
		t.Fatalf("expected 0x00 at 0x4567, got %#02x", v)
	} else {
		t.Logf("0x4567 = %#02x", v)
	}

	// Test blank value
	if v := m.Fetch(0x7fff); v != 0x2a {
		t.Fatalf("expected 0x2a at 0x7fff, got %#02x: %v", v, m)
	} else {
		t.Logf("0x7fff = %#02x", v)
	}

	// Test zero value
	if v := m.Fetch(0xffff); v != 0xff {
		t.Fatalf("expected 0xff at 0xffff, got %#02x", v)
	} else {
		t.Logf("0xffff = %#02x", v)
	}

	// Test string
	want := "Mapper{$0000-$1FFF: 8kB RAM, $2000-$3FFF: {8kB RAM 8191}, $4000-$5FFF: {8kB ROM 8191}, $0000-$7FFF: 0x2a}"
	if v := m.String(); v != want {
		t.Fatalf("expected %q, got %q", want, v)
	}

	// Test unmap
	if !m.Unmap(r) {
		t.Fatal("unmap failed")
	}
	if m.Unmap(r) {
		t.Fatal("unmap should have returned false for unmapped memory")
	}

	// Test reset
	if v := m.Reset().Fetch(0x1234); v != 0xff {
		t.Fatalf("expected 0xff at 0x1234, got %#02x", v)
	} else {
		t.Logf("0x1234 = %#02x", v)
	}
}
