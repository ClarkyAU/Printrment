package main

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// Printer is one saved printer.
type Printer struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	IP     string `json:"ip"`
	Code   string `json:"code"`
	Serial string `json:"serial"` // auto-detected from the printer's certificate
	Mode   string `json:"mode"`   // auto (RTSP) | jpeg

	// Pins remembers the printer's certificate fingerprint per port, so a
	// different device can't pretend to be this printer later.
	Pins map[string]string `json:"pins,omitempty"`

	// FwSeen is the firmware update Printrment last told you about.
	FwSeen string `json:"fw_seen,omitempty"`

	// CodeEnc is the access code as stored on disk: encrypted with Windows
	// DPAPI so it only opens for this Windows user on this PC. Code is never
	// written to disk.
	CodeEnc string `json:"code_enc,omitempty"`
}

// Settings is everything the app remembers.
type Settings struct {
	Printers []Printer `json:"printers"`
	Active   string    `json:"active"`

	SnapshotDir  string `json:"snapshot_dir"`
	TimelapseDir string `json:"timelapse_dir"`

	TimelapseAuto     bool   `json:"timelapse_auto"`     // record every print automatically
	TimelapseMode     string `json:"timelapse_mode"`     // layer | interval
	TimelapseInterval int    `json:"timelapse_interval"` // seconds, for interval mode

	NotifySound   bool `json:"notify_sound"`   // (v11, replaced by the two below)
	SoundProblems bool `json:"sound_problems"` // failed, paused by the printer, error codes
	SoundFinished bool `json:"sound_finished"` // print finished
	NotifyFlash   bool `json:"notify_flash"`   // flash the taskbar button

	Theme string `json:"theme"` // default | webstore (Clarky3D) | makerworld

	ConfigVersion int `json:"config_version"`

	Tray      bool `json:"tray"`      // keep running in the tray when the window closes
	Autostart bool `json:"autostart"` // start with Windows (in the tray)

	Remote         bool   `json:"remote"` // allow viewing from phones on the same network
	RemotePort     int    `json:"remote_port"`
	RemotePIN      string `json:"remote_pin,omitempty"`
	RemoteControls bool   `json:"remote_controls"`  // allow pause/stop etc. from the phone
	Secret         string `json:"secret,omitempty"` // signs the phone login cookie

	// encrypted copies of RemotePIN / Secret, as stored on disk
	RemotePINEnc string `json:"remote_pin_enc,omitempty"`
	SecretEnc    string `json:"secret_enc,omitempty"`
}

type store struct {
	mu       sync.Mutex
	s        Settings
	problems []string // shown once the app is up
}

var cfg = &store{}

func dataDir() string {
	d, err := os.UserConfigDir()
	if err != nil {
		d = os.TempDir()
	}
	p := filepath.Join(d, "Printrment")
	os.MkdirAll(p, 0o700)
	return p
}

func randHex(n int) string {
	b := make([]byte, n)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func randPIN() string {
	b := make([]byte, 4)
	rand.Read(b)
	n := (uint32(b[0])<<24 | uint32(b[1])<<16 | uint32(b[2])<<8 | uint32(b[3])) % 1000000
	return fmt.Sprintf("%06d", n)
}

func (st *store) applyDefaults() {
	s := &st.s
	home, _ := os.UserHomeDir()
	if s.SnapshotDir == "" {
		s.SnapshotDir = filepath.Join(home, "Pictures", "Printrment")
	}
	if s.TimelapseDir == "" {
		s.TimelapseDir = filepath.Join(home, "Videos", "Printrment")
	}
	if s.TimelapseMode != "interval" {
		s.TimelapseMode = "layer"
	}
	if s.TimelapseInterval < 2 {
		s.TimelapseInterval = 10
	}
	if s.ConfigVersion < 12 {
		s.SoundProblems = true
		s.SoundFinished = true
		s.ConfigVersion = 12
	}
	if s.ConfigVersion < 13 {
		s.Tray = false // closing the window now quits the app unless chosen otherwise
		s.ConfigVersion = 13
	}
	if s.Theme != "makerworld" && s.Theme != "webstore" {
		s.Theme = "default"
	}
	if s.RemotePort == 0 {
		s.RemotePort = 8788
	}
	if s.RemotePIN == "" {
		s.RemotePIN = randPIN()
	}
	if s.Secret == "" {
		s.Secret = randHex(16)
	}
	for i := range s.Printers {
		p := &s.Printers[i]
		if p.ID == "" {
			p.ID = randHex(4)
		}
		if p.Mode != "jpeg" {
			p.Mode = "auto"
		}
		if p.Name == "" {
			p.Name = "Printer " + fmt.Sprint(i+1)
		}
	}
	if s.Printers == nil {
		s.Printers = []Printer{}
	}
	if len(s.Printers) > 0 && st.indexLocked(s.Active) < 0 {
		s.Active = s.Printers[0].ID
	}
}

func (st *store) load() {
	st.mu.Lock()
	defer st.mu.Unlock()
	path := filepath.Join(dataDir(), "config.json")
	b, err := os.ReadFile(path)
	fresh := err != nil
	if err == nil {
		// Older versions stored a single printer as {ip, code, mode}.
		var old struct {
			IP, Code, Mode string
			Printers       []Printer `json:"printers"`
		}
		if jerr := json.Unmarshal(b, &old); jerr != nil {
			// Don't silently replace a damaged settings file with an empty one:
			// keep it so nothing is lost, and say so.
			keep := path + ".damaged-" + time.Now().Format("20060102-150405")
			os.Rename(path, keep)
			appendLog("settings file couldn't be read (" + jerr.Error() + "), kept as " + keep)
			st.problems = append(st.problems, "Your settings file was damaged, so Printrment started fresh. The old file was kept as "+filepath.Base(keep)+" in "+dataDir()+".")
			fresh = true
		} else if old.Printers == nil && old.IP != "" {
			st.s = Settings{Printers: []Printer{{Name: "My printer", IP: old.IP, Code: old.Code, Mode: old.Mode}}}
			fresh = true
		} else {
			json.Unmarshal(b, &st.s)
		}
	}
	st.unsealLocked()
	if fresh {
		st.s.TimelapseAuto = true
		st.s.NotifyFlash = true
	}
	st.applyDefaults()
	if err := st.saveLocked(); err != nil {
		appendLog("couldn't save settings: " + err.Error())
	}
}

// unsealLocked decrypts the secrets that were read from disk. Plain values
// left by older versions are kept and get encrypted on the next save.
func (st *store) unsealLocked() {
	s := &st.s
	for i := range s.Printers {
		p := &s.Printers[i]
		if p.CodeEnc != "" {
			code, err := unseal(p.CodeEnc)
			if err != nil {
				appendLog("couldn't decrypt the access code for " + p.Name + ": " + err.Error())
				st.problems = append(st.problems, "Printrment couldn't read the saved access code for "+p.Name+" (settings copied from another PC or Windows account?). Enter it again in Settings > Printers.")
			}
			p.Code, p.CodeEnc = code, ""
		}
	}
	if s.RemotePINEnc != "" {
		s.RemotePIN, _ = unseal(s.RemotePINEnc) // a new one is made if it can't be read
		s.RemotePINEnc = ""
	}
	if s.SecretEnc != "" {
		s.Secret, _ = unseal(s.SecretEnc)
		s.SecretEnc = ""
	}
}

func (st *store) saveLocked() error {
	d := st.s
	d.Printers = make([]Printer, len(st.s.Printers))
	for i, p := range st.s.Printers {
		if p.Code != "" {
			enc, err := seal(p.Code)
			if err != nil {
				return fmt.Errorf("couldn't encrypt the access code: %v", err)
			}
			p.CodeEnc = enc
		}
		p.Code = ""
		d.Printers[i] = p
	}
	var err error
	if d.RemotePINEnc, err = seal(d.RemotePIN); err != nil {
		return err
	}
	if d.SecretEnc, err = seal(d.Secret); err != nil {
		return err
	}
	d.RemotePIN, d.Secret = "", ""
	b, _ := json.MarshalIndent(d, "", "  ")
	tmp := filepath.Join(dataDir(), "config.json.tmp")
	if err := os.WriteFile(tmp, b, 0o600); err != nil {
		return err
	}
	return os.Rename(tmp, filepath.Join(dataDir(), "config.json"))
}

// takeProblems returns (once) anything that went wrong while loading.
func (st *store) takeProblems() []string {
	st.mu.Lock()
	defer st.mu.Unlock()
	p := st.problems
	st.problems = nil
	return p
}

func (st *store) get() Settings {
	st.mu.Lock()
	defer st.mu.Unlock()
	s := st.s
	s.Printers = append([]Printer{}, st.s.Printers...) // never null in JSON
	return s
}

func (st *store) indexLocked(id string) int {
	for i, p := range st.s.Printers {
		if p.ID == id {
			return i
		}
	}
	return -1
}

func (st *store) printer(id string) (Printer, bool) {
	st.mu.Lock()
	defer st.mu.Unlock()
	if i := st.indexLocked(id); i >= 0 {
		return st.s.Printers[i], true
	}
	return Printer{}, false
}

// update runs f on the settings and saves them.
func (st *store) update(f func(s *Settings)) error {
	st.mu.Lock()
	defer st.mu.Unlock()
	f(&st.s)
	st.applyDefaults()
	return st.saveLocked()
}

func (st *store) setSerial(id, serial string) {
	st.update(func(s *Settings) {
		if i := st.indexLocked(id); i >= 0 {
			s.Printers[i].Serial = serial
		}
	})
}

func cleanPrinter(p Printer) Printer {
	p.IP = strings.TrimSpace(p.IP)
	p.Code = strings.TrimSpace(p.Code)
	p.Name = strings.TrimSpace(p.Name)
	p.Serial = strings.TrimSpace(p.Serial)
	if len(p.Name) > 40 {
		p.Name = p.Name[:40]
	}
	return p
}
