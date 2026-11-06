package webviewloader

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
)

// extractLoader writes the embedded WebView2Loader.dll into a folder under
// base that only this user can write to, and returns its full path. The file
// is named after its own SHA-256 and rewritten if what's on disk differs, so
// it is always exactly the embedded copy. Loading it by full path (instead of
// by name) means nothing in the program's folder or PATH can stand in for it.
func extractLoader(base string, dll []byte) (string, error) {
	sum := sha256.Sum256(dll)
	dir := filepath.Join(base, "Printrment", "webview2", hex.EncodeToString(sum[:])[:16])
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return "", err
	}
	path := filepath.Join(dir, "WebView2Loader.dll")
	if cur, err := os.ReadFile(path); err == nil && bytes.Equal(cur, dll) {
		return path, nil
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, dll, 0o600); err != nil {
		return "", fmt.Errorf("can't unpack WebView2Loader.dll: %w", err)
	}
	if err := os.Rename(tmp, path); err != nil {
		os.Remove(tmp)
		return "", fmt.Errorf("can't unpack WebView2Loader.dll: %w", err)
	}
	return path, nil
}
