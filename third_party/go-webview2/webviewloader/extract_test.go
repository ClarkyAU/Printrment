package webviewloader

import (
	"os"
	"path/filepath"
	"testing"
)

func TestExtractLoader(t *testing.T) {
	base := t.TempDir()
	dll := []byte("pretend dll bytes")
	p, err := extractLoader(base, dll)
	if err != nil {
		t.Fatal(err)
	}
	if !filepath.IsAbs(p) {
		t.Fatalf("not an absolute path: %s", p)
	}
	if b, _ := os.ReadFile(p); string(b) != string(dll) {
		t.Fatal("wrong contents")
	}
	// tampered copy is replaced
	os.WriteFile(p, []byte("evil"), 0o600)
	p2, err := extractLoader(base, dll)
	if err != nil || p2 != p {
		t.Fatal(err, p2)
	}
	if b, _ := os.ReadFile(p); string(b) != string(dll) {
		t.Fatal("tampered file was trusted")
	}
}
