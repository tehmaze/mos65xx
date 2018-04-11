package memory

import (
	"os"
	"path/filepath"
	"testing"
)

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
