//go:build !windows

package main

import (
	"encoding/base64"
	"errors"
	"strings"
)

// Only used for testing on other systems: NOT encrypted, just labelled.
func seal(plain string) (string, error) {
	if plain == "" {
		return "", nil
	}
	return "plain:" + base64.StdEncoding.EncodeToString([]byte(plain)), nil
}

func unseal(s string) (string, error) {
	enc, ok := strings.CutPrefix(s, "plain:")
	if !ok {
		return "", errors.New("unknown format")
	}
	b, err := base64.StdEncoding.DecodeString(enc)
	return string(b), err
}
