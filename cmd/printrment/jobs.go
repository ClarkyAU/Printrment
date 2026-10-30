package main

import (
	"fmt"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// JobSummary is the little report shown after a print ends.
type JobSummary struct {
	File     string `json:"file"`
	Result   string `json:"result"` // finished | failed
	Ended    int64  `json:"ended"`  // unix seconds
	Seconds  int    `json:"seconds,omitempty"`
	Layers   int    `json:"layers,omitempty"`
	Snapshot string `json:"snapshot,omitempty"` // path of the picture taken when it ended
}

func fmtDuration(sec int) string {
	if sec <= 0 {
		return ""
	}
	h, m := sec/3600, sec%3600/60
	if h > 0 {
		return fmt.Sprintf("%dh %02dm", h, m)
	}
	if m == 0 {
		return fmt.Sprintf("%ds", sec)
	}
	return fmt.Sprintf("%dm", m)
}

// shot saves the camera's latest picture next to your snapshots, if the
// camera is running (window open or a timelapse recording). Returns the path.
func (m *Monitor) shot(tag, label string) string {
	p, ok := cfg.printer(m.id)
	if !ok {
		return ""
	}
	f := getCam(m.id).latest(15 * time.Second)
	if f == nil {
		return ""
	}
	dir := cfg.get().SnapshotDir
	if os.MkdirAll(dir, 0o755) != nil {
		return ""
	}
	name := strings.TrimSpace(unsafeName.ReplaceAllString(p.Name+" - "+label, "")) + " (" + tag + ") " + time.Now().Format("2006-01-02 150405") + ".jpg"
	out := filepath.Join(dir, name)
	if os.WriteFile(out, f, 0o644) != nil {
		return ""
	}
	return out
}

// problemShot: a picture for when something goes wrong (spaghetti pause,
// failure, error). At most one a minute per printer.
func (m *Monitor) problemShot(label string) string {
	if time.Since(m.lastShot) < time.Minute {
		return ""
	}
	path := m.shot("problem", strings.TrimSuffix(strings.TrimSuffix(label, ".3mf"), ".gcode"))
	if path != "" {
		m.lastShot = time.Now()
	}
	return path
}

func shotNote(path string) string {
	if path == "" {
		return ""
	}
	return " - picture saved"
}

// jobStarted / jobEnded keep the numbers for the summary.
func (m *Monitor) jobStarted() {
	m.jobStart = time.Now()
	m.mu.Lock()
	m.st.LastJob = nil
	m.mu.Unlock()
}

func (m *Monitor) jobEnded(cur Status, result, file, fallbackShot string) (sum *JobSummary) {
	secs := 0
	if cur.StartedAt > 0 {
		if d := int(time.Now().Unix() - cur.StartedAt); d > 0 && d < 14*24*3600 {
			secs = d
		}
	}
	if secs == 0 && !m.jobStart.IsZero() {
		secs = int(time.Since(m.jobStart).Seconds())
	}
	layers := cur.TotalLayers
	if result != "finished" || layers == 0 {
		layers = cur.Layer // how far it got
	}
	sum = &JobSummary{File: file, Result: result, Ended: time.Now().Unix(), Seconds: secs, Layers: layers}
	label := strings.TrimSuffix(strings.TrimSuffix(file, ".3mf"), ".gcode")
	tag := result
	sum.Snapshot = fallbackShot
	if sum.Snapshot == "" {
		sum.Snapshot = m.shot(tag, label)
	}
	m.jobStart = time.Time{}
	m.mu.Lock()
	m.st.LastJob = sum
	m.mu.Unlock()
	return sum
}

func (s *JobSummary) text() string {
	var parts []string
	if d := fmtDuration(s.Seconds); d != "" {
		parts = append(parts, d)
	}
	if s.Layers > 0 {
		parts = append(parts, strconv.Itoa(s.Layers)+" layers")
	}
	if s.Snapshot != "" {
		parts = append(parts, "picture saved")
	}
	if len(parts) == 0 {
		return ""
	}
	return " (" + strings.Join(parts, ", ") + ")"
}

// ---------------------------------------------------------------- firmware updates

// checkFirmware raises one alert when the printer reports a firmware update
// it hasn't told you about yet. What it last told you is remembered, so a
// restart doesn't repeat it.
func (m *Monitor) checkFirmware(cur Status) {
	if cur.Info == nil || len(cur.Info.Updates) == 0 {
		return
	}
	key := strings.Join(cur.Info.Updates, "; ")
	p, ok := cfg.printer(m.id)
	if !ok || p.FwSeen == key {
		return
	}
	cfg.update(func(s *Settings) {
		for i := range s.Printers {
			if s.Printers[i].ID == m.id {
				s.Printers[i].FwSeen = key
			}
		}
	})
	alert(m.id, "info", "Firmware update available: "+key+". Install it from the printer's screen or Bambu Handy.")
}

// ---------------------------------------------------------------- settings backup

type backupFile struct {
	App      string         `json:"app"`
	Version  string         `json:"version"`
	Exported string         `json:"exported"`
	Note     string         `json:"note"`
	Settings map[string]any `json:"settings"`
	Printers []backupPrn    `json:"printers"`
}

type backupPrn struct {
	Name   string `json:"name"`
	IP     string `json:"ip"`
	Serial string `json:"serial,omitempty"`
	Mode   string `json:"mode,omitempty"`
}

func registerBackupHandlers(mux *http.ServeMux) {
	mux.HandleFunc("/api/export", localOnly(func(w http.ResponseWriter, r *http.Request) {
		s := cfg.get()
		b := backupFile{App: "Printrment", Version: version, Exported: time.Now().Format(time.RFC3339),
			Note: "Access codes, the remote access PIN and printer certificates are not included. Enter the access codes again after importing.",
			Settings: map[string]any{"theme": s.Theme, "timelapse_auto": s.TimelapseAuto, "timelapse_mode": s.TimelapseMode,
				"timelapse_interval": s.TimelapseInterval, "sound_problems": s.SoundProblems, "sound_finished": s.SoundFinished,
				"notify_flash": s.NotifyFlash, "tray": s.Tray, "autostart": s.Autostart, "remote": s.Remote,
				"remote_port": s.RemotePort, "remote_controls": s.RemoteControls}}
		for _, p := range s.Printers {
			b.Printers = append(b.Printers, backupPrn{p.Name, p.IP, p.Serial, p.Mode})
		}
		w.Header().Set("Content-Disposition", "attachment; filename=printrment-settings.json")
		writeJSON(w, b)
	}))

	mux.HandleFunc("/api/import", localOnly(func(w http.ResponseWriter, r *http.Request) {
		var b backupFile
		if readJSON(r, &b) != nil || (b.App != "Printrment" && b.App != "BambuCam") {
			writeJSON(w, map[string]any{"ok": false, "reason": "That isn't a Printrment settings file."})
			return
		}
		if len(b.Printers) > 20 {
			writeJSON(w, map[string]any{"ok": false, "reason": "Too many printers in that file."})
			return
		}
		str := func(k string) (string, bool) { v, ok := b.Settings[k].(string); return v, ok }
		boo := func(k string) (bool, bool) { v, ok := b.Settings[k].(bool); return v, ok }
		num := func(k string) (int, bool) { v, ok := b.Settings[k].(float64); return int(v), ok }
		added, updated, needCode := 0, 0, []string{}
		cfg.update(func(s *Settings) {
			if v, ok := str("theme"); ok && (v == "default" || v == "webstore" || v == "makerworld") {
				s.Theme = v
			}
			if v, ok := str("timelapse_mode"); ok && (v == "layer" || v == "interval") {
				s.TimelapseMode = v
			}
			if v, ok := num("timelapse_interval"); ok && v >= 2 && v <= 3600 {
				s.TimelapseInterval = v
			}
			for k, dst := range map[string]*bool{"timelapse_auto": &s.TimelapseAuto, "sound_problems": &s.SoundProblems,
				"sound_finished": &s.SoundFinished, "notify_flash": &s.NotifyFlash, "tray": &s.Tray,
				"autostart": &s.Autostart, "remote": &s.Remote, "remote_controls": &s.RemoteControls} {
				if v, ok := boo(k); ok {
					*dst = v
				}
			}
			if v, ok := num("remote_port"); ok && v >= 1024 && v <= 65535 && v != 8787 {
				s.RemotePort = v
			}
			for _, bp := range b.Printers {
				ip := strings.TrimSpace(bp.IP)
				if net.ParseIP(ip) == nil {
					continue
				}
				name := strings.TrimSpace(bp.Name)
				if len(name) > 40 {
					name = name[:40]
				}
				mode := "auto"
				if bp.Mode == "jpeg" {
					mode = "jpeg"
				}
				found := false
				for i := range s.Printers {
					if s.Printers[i].IP == ip {
						found = true
						updated++
						if name != "" {
							s.Printers[i].Name = name
						}
						s.Printers[i].Mode = mode
						break
					}
				}
				if !found {
					added++
					needCode = append(needCode, name)
					s.Printers = append(s.Printers, Printer{ID: randHex(4), Name: name, IP: ip,
						Serial: strings.TrimSpace(bp.Serial), Mode: mode})
				}
			}
		})
		syncMonitors()
		applySettings()
		writeJSON(w, map[string]any{"ok": true, "added": added, "updated": updated, "need_code": needCode})
	}))
}
