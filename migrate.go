package main

import (
	"io"
	"os"
	"path/filepath"
)

// The app used to be called Bambu Cam. These bring an existing install across
// so nothing has to be set up again: settings (the encryption they use is
// unchanged), pictures/videos folders and tidy-up of the old cache.

func legacyDir(base func() (string, error)) string {
	d, err := base()
	if err != nil {
		return ""
	}
	return filepath.Join(d, "BambuCam")
}

func legacyConfigDir() string { return legacyDir(os.UserConfigDir) }
func legacyLocalDir() string  { return legacyDir(os.UserCacheDir) }

func copyFile(from, to string) error {
	in, err := os.Open(from)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.OpenFile(to, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o600)
	if err != nil {
		return err
	}
	_, err = io.Copy(out, in)
	if cerr := out.Close(); err == nil {
		err = cerr
	}
	if err != nil {
		os.Remove(to)
	}
	return err
}

// migrateLegacyFiles runs before settings are loaded.
func migrateLegacyFiles() {
	if old := legacyConfigDir(); old != "" && old != dataDir() {
		if _, err := os.Stat(filepath.Join(dataDir(), "config.json")); err != nil {
			if copyFile(filepath.Join(old, "config.json"), filepath.Join(dataDir(), "config.json")) == nil {
				appendLog("moved settings over from the old Bambu Cam folder")
			}
		}
	}
	if old := legacyLocalDir(); old != "" && old != localDir() {
		go os.RemoveAll(old) // old video helper copy, browser cache and token
	}
}

// migrateLegacyPaths runs after settings are loaded: if the picture / video
// folders are still the old default ones, switch to the new names (moving the
// folder when that is safe).
func migrateLegacyPaths() {
	home, err := os.UserHomeDir()
	if err != nil {
		return
	}
	fix := func(sub string, get func(*Settings) *string) {
		oldP := filepath.Join(home, sub, "BambuCam")
		newP := filepath.Join(home, sub, "Printrment")
		cfg.update(func(s *Settings) {
			p := get(s)
			if *p != oldP {
				return
			}
			if _, err := os.Stat(oldP); err == nil {
				if _, err := os.Stat(newP); err != nil && os.Rename(oldP, newP) != nil {
					return // couldn't move it: keep using the old folder
				}
			}
			*p = newP
		})
	}
	fix("Pictures", func(s *Settings) *string { return &s.SnapshotDir })
	fix("Videos", func(s *Settings) *string { return &s.TimelapseDir })
}
