package main

import (
	"os"
	"path/filepath"
	"strconv"
	"testing"
)

// Writes the app icon as PNGs for the Linux package when PRINTRMENT_ICON_DIR
// is set (used by build.sh); otherwise just checks the icon renders.
func TestIconPNGs(t *testing.T) {
	dir := os.Getenv("PRINTRMENT_ICON_DIR")
	for _, n := range []int{64, 128, 256} {
		b := iconPNG(n)
		if len(b) < 100 {
			t.Fatalf("icon %d: %d bytes", n, len(b))
		}
		if dir != "" {
			if err := os.WriteFile(filepath.Join(dir, "printrment-"+strconv.Itoa(n)+".png"), b, 0o644); err != nil {
				t.Fatal(err)
			}
		}
	}
}
