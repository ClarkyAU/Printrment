//go:build windows

package main

import (
	"encoding/base64"
	"errors"
	"strings"
	"unsafe"

	"golang.org/x/sys/windows"
)

// Secrets on disk are encrypted with Windows DPAPI: only this Windows user on
// this PC can decrypt them, so a copied settings file is useless elsewhere.

var sealEntropy = []byte("BambuCam settings v1")

func blob(b []byte) *windows.DataBlob {
	if len(b) == 0 {
		return &windows.DataBlob{}
	}
	return &windows.DataBlob{Size: uint32(len(b)), Data: &b[0]}
}

func takeBlob(out windows.DataBlob) []byte {
	defer windows.LocalFree(windows.Handle(unsafe.Pointer(out.Data)))
	return append([]byte(nil), unsafe.Slice(out.Data, out.Size)...)
}

func seal(plain string) (string, error) {
	if plain == "" {
		return "", nil
	}
	var out windows.DataBlob
	if err := windows.CryptProtectData(blob([]byte(plain)), nil, blob(sealEntropy), 0, nil, windows.CRYPTPROTECT_UI_FORBIDDEN, &out); err != nil {
		return "", err
	}
	return "dpapi:" + base64.StdEncoding.EncodeToString(takeBlob(out)), nil
}

func unseal(s string) (string, error) {
	enc, ok := strings.CutPrefix(s, "dpapi:")
	if !ok {
		return "", errors.New("unknown format")
	}
	raw, err := base64.StdEncoding.DecodeString(enc)
	if err != nil || len(raw) == 0 {
		return "", errors.New("damaged value")
	}
	var out windows.DataBlob
	if err := windows.CryptUnprotectData(blob(raw), nil, blob(sealEntropy), 0, nil, windows.CRYPTPROTECT_UI_FORBIDDEN, &out); err != nil {
		return "", err
	}
	return string(takeBlob(out)), nil
}
