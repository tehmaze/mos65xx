package memory

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNew(t *testing.T) {
	mem := New(8).Reset(0x2a)
	if v := mem.Fetch(1); v != 0x2a {
		t.Fatalf("expected 0x2a at 0x0001, got %#02x", v)
	}
	if v := mem.String(); v != "8B RAM" {
		t.Fatalf("expected %q, got %q", "8B RAM", v)
	}
}

func TestLoad(t *testing.T) {
	mem, err := Load(filepath.Join("testdata", "zero.rom"))
	if err != nil && os.IsNotExist(err) {
		t.Skip(err)
	} else if err != nil {
		t.Fatal(err)
	}

	if v := mem.Fetch(0x2a); v != 0x00 {
		t.Fatalf("expected 0x00 at 0x002a, got %#02x", v)
	}
	if v := mem.String(); v != "128B ROM" {
		t.Fatalf("expected %q, got %q", "128B ROM", v)
	}
}

func TestLoadError(t *testing.T) {
	mem, err := Load("/doesnotexistanywhere.rom")
	if !os.IsNotExist(err) {
		t.Fatalf("expected ROM to not exist, got %v", err)
	}
	if mem != nil {
		t.Fatalf("expected ROM to be nil, got %v", mem)
	}
}

func TestSizeOf(t *testing.T) {
	for _, test := range []struct {
		Size int
		Want string
	}{
		{0x0400, "1024B"},
		{0x0800, "2048B"},
		{0x1000, "4096B"},
		{0x2000, "8kB"},
		{0x4000, "16kB"},
	} {
		if v := sizeOf(test.Size); v != test.Want {
			t.Fatalf("expected %q for %#04x, got %q", test.Want, test.Size, v)
		} else {
			t.Logf("%#04x = %s", test.Size, v)
		}
	}
}
