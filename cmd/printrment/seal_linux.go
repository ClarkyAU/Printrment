//go:build linux

package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

// Linux has no DPAPI, so secrets on disk (access codes, the remote access PIN)
// are encrypted with AES-256-GCM using a random key kept in its own file,
// readable only by this user and stored away from the settings
// (~/.local/share/Printrment/secret.key, settings live in ~/.config). A
// copied or backed-up settings file is useless without that key.

var sealAAD = []byte("Printrment settings v1")

var (
	keyMu  sync.Mutex
	keyVal []byte
)

func keyPath() string {
	// Server / Docker: PRINTRMENT_KEY_FILE lets the key live on a separate
	// mount (e.g. a secret) from the data volume holding the settings.
	if p := os.Getenv("PRINTRMENT_KEY_FILE"); p != "" {
		return p
	}
	if d := os.Getenv("PRINTRMENT_DATA"); d != "" {
		return filepath.Join(d, "key", "secret.key")
	}
	d := os.Getenv("XDG_DATA_HOME")
	if d == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return ""
		}
		d = filepath.Join(home, ".local", "share")
	}
	return filepath.Join(d, "Printrment", "secret.key")
}

// sealKey loads the key, creating it the first time.
func sealKey() ([]byte, error) {
	keyMu.Lock()
	defer keyMu.Unlock()
	if keyVal != nil {
		return keyVal, nil
	}
	p := keyPath()
	if p == "" {
		return nil, errors.New("no home folder for the encryption key")
	}
	for try := 0; try < 2; try++ {
		if b, err := os.ReadFile(p); err == nil {
			if len(b) != 32 {
				return nil, errors.New("the encryption key file is damaged")
			}
			if st, err := os.Stat(p); err == nil && st.Mode().Perm()&0o077 != 0 {
				os.Chmod(p, 0o600) // someone loosened it: tighten again
			}
			keyVal = b
			return b, nil
		} else if !os.IsNotExist(err) {
			return nil, err
		}
		k, err := createKey(p)
		if os.IsExist(err) {
			continue // another copy made it at the same moment: read that one
		}
		if err != nil {
			return nil, err
		}
		keyVal = k
		return k, nil
	}
	return nil, errors.New("couldn't create the encryption key")
}

func createKey(p string) ([]byte, error) {
	if err := os.MkdirAll(filepath.Dir(p), 0o700); err != nil {
		return nil, err
	}
	k := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, k); err != nil {
		return nil, err
	}
	f, err := os.OpenFile(p, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o600)
	if err != nil {
		return nil, err
	}
	_, err = f.Write(k)
	if cerr := f.Close(); err == nil {
		err = cerr
	}
	if err != nil {
		os.Remove(p)
		return nil, err
	}
	return k, nil
}

func gcm() (cipher.AEAD, error) {
	k, err := sealKey()
	if err != nil {
		return nil, err
	}
	b, err := aes.NewCipher(k)
	if err != nil {
		return nil, err
	}
	return cipher.NewGCM(b)
}

func seal(plain string) (string, error) {
	if plain == "" {
		return "", nil
	}
	g, err := gcm()
	if err != nil {
		return "", err
	}
	nonce := make([]byte, g.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}
	out := g.Seal(nonce, nonce, []byte(plain), sealAAD)
	return "aesgcm:" + base64.StdEncoding.EncodeToString(out), nil
}

func unseal(s string) (string, error) {
	// "plain:" was written by early test builds; read it so it gets
	// re-saved encrypted.
	if enc, ok := strings.CutPrefix(s, "plain:"); ok {
		b, err := base64.StdEncoding.DecodeString(enc)
		return string(b), err
	}
	enc, ok := strings.CutPrefix(s, "aesgcm:")
	if !ok {
		return "", errors.New("unknown format")
	}
	b, err := base64.StdEncoding.DecodeString(enc)
	if err != nil {
		return "", err
	}
	g, err := gcm()
	if err != nil {
		return "", err
	}
	if len(b) < g.NonceSize() {
		return "", errors.New("damaged")
	}
	plain, err := g.Open(nil, b[:g.NonceSize()], b[g.NonceSize():], sealAAD)
	if err != nil {
		return "", errors.New("can't be decrypted with this computer's key")
	}
	return string(plain), nil
}
