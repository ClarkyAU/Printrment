package main

import (
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"os"
	"sync"
	"testing"
)

func TestFFmpegHash(t *testing.T) {
	zr, err := gzip.NewReader(bytes.NewReader(ffmpegGz))
	if err != nil {
		t.Fatal(err)
	}
	h := sha256.New()
	if _, err := io.Copy(h, zr); err != nil {
		t.Fatal(err)
	}
	if got := hex.EncodeToString(h.Sum(nil)); got != ffmpegSHA256 {
		t.Fatalf("ffmpegSHA256 is stale: embedded file hashes to %s", got)
	}
}

func TestFFmpegTamperedCopyIsReplaced(t *testing.T) {
	t.Setenv("BAMBUCAM_FFMPEG", "")
	t.Setenv("HOME", t.TempDir())
	t.Setenv("XDG_CACHE_HOME", "")
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	reset := func() { ffmpegOnce = sync.Once{}; ffmpegPath, ffmpegErr = "", nil }
	reset()
	p, err := getFFmpeg()
	if err != nil || fileSHA256(p) != ffmpegSHA256 {
		t.Fatalf("first unpack: %v", err)
	}
	// same size, one byte changed: the old size-only check would have run it
	f, _ := os.OpenFile(p, os.O_WRONLY, 0)
	f.WriteAt([]byte{'X'}, 4096)
	f.Close()
	reset()
	p2, err := getFFmpeg()
	if err != nil || p2 != p || fileSHA256(p2) != ffmpegSHA256 {
		t.Fatalf("tampered copy not replaced: %v", err)
	}
}
