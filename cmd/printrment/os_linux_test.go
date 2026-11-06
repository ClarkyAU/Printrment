//go:build linux

package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestAutostartUsesAppImage(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(dir, "config"))
	img := filepath.Join(dir, "My Apps", "Printrment-0.32-x86_64.AppImage")
	os.MkdirAll(filepath.Dir(img), 0o755)
	os.WriteFile(img, []byte("x"), 0o755)

	t.Setenv("APPIMAGE", img)
	setAutostart(true)
	b, err := os.ReadFile(autostartFile())
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(b), "Exec=\""+img+"\" --tray") {
		t.Fatalf("autostart doesn't start the .AppImage file:\n%s", b)
	}

	// A relative or missing path isn't trusted: fall back to the running program.
	for _, bad := range []string{"relative.AppImage", filepath.Join(dir, "gone.AppImage"), dir} {
		t.Setenv("APPIMAGE", bad)
		got, err := startCommand()
		exe, _ := os.Executable()
		if err != nil || got != exe {
			t.Fatalf("APPIMAGE=%q gave %q, want %q", bad, got, exe)
		}
	}

	setAutostart(false)
	if _, err := os.Stat(autostartFile()); err == nil {
		t.Fatal("autostart file not removed")
	}
}
