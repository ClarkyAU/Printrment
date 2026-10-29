//go:build linux

package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSealLinux(t *testing.T) {
	d := t.TempDir()
	t.Setenv("XDG_DATA_HOME", d)
	keyMu.Lock()
	keyVal = nil
	keyMu.Unlock()

	enc, err := seal("12345678")
	if err != nil || !strings.HasPrefix(enc, "aesgcm:") || strings.Contains(enc, "12345678") {
		t.Fatalf("seal: %q %v", enc, err)
	}
	if again, _ := seal("12345678"); again == enc {
		t.Fatal("same text sealed twice gave the same output (nonce reused)")
	}
	if got, err := unseal(enc); err != nil || got != "12345678" {
		t.Fatalf("unseal: %q %v", got, err)
	}
	st, err := os.Stat(filepath.Join(d, "Printrment", "secret.key"))
	if err != nil || st.Mode().Perm() != 0o600 {
		t.Fatalf("key file: %v %v", st, err)
	}
	// tampering is detected
	b := []byte(enc)
	b[len(b)-3] ^= 1
	if _, err := unseal(string(b)); err == nil {
		t.Fatal("tampered value was accepted")
	}
	// a different key (another computer / user) can't read it
	keyMu.Lock()
	keyVal = nil
	keyMu.Unlock()
	t.Setenv("XDG_DATA_HOME", t.TempDir())
	if _, err := unseal(enc); err == nil {
		t.Fatal("value opened with a different key")
	}
	// old test-build values are still readable so they get re-saved encrypted
	if got, _ := unseal("plain:MTIzNDU2Nzg="); got != "12345678" {
		t.Fatalf("plain migration: %q", got)
	}
	keyMu.Lock()
	keyVal = nil
	keyMu.Unlock()
}
