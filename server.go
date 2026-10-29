package main

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"
)

const listenAddr = "127.0.0.1:8787"
const version = "0.30"

var (
	lastActivity   time.Time
	lastActivityMu sync.Mutex
)

func touch() {
	lastActivityMu.Lock()
	lastActivity = time.Now()
	lastActivityMu.Unlock()
}

func idleFor() time.Duration {
	lastActivityMu.Lock()
	defer lastActivityMu.Unlock()
	return time.Since(lastActivity)
}

type ctxKey int

const remoteKey ctxKey = 1

func isRemote(r *http.Request) bool { return r.Context().Value(remoteKey) == true }

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	json.NewEncoder(w).Encode(v)
}

// newServer sets time limits so a client that opens connections and sends
// requests very slowly can't tie the app up. (No write limit: the camera
// stream is one long response.)
func newServer(h http.Handler) *http.Server {
	return &http.Server{Handler: h, ReadHeaderTimeout: 10 * time.Second,
		IdleTimeout: 2 * time.Minute, MaxHeaderBytes: 64 << 10}
}

func readJSON(r *http.Request, v any) error {
	return json.NewDecoder(io.LimitReader(r.Body, 1<<20)).Decode(v)
}

// ------------------------------------------------------------ handlers

func newMux() *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Header().Set("Cache-Control", "no-store")
		io.WriteString(w, strings.Replace(page, "{{VERSION}}", version, -1))
	})

	mux.HandleFunc("/fonts/", serveFont)

	registerFileHandlers(mux)
	registerCheckHandler(mux)
	registerBackupHandlers(mux)

	mux.HandleFunc("/api/version", func(w http.ResponseWriter, r *http.Request) { io.WriteString(w, version) })

	mux.HandleFunc("/api/state", func(w http.ResponseWriter, r *http.Request) {
		since, _ := strconv.ParseInt(r.URL.Query().Get("since"), 10, 64)
		s := cfg.get()
		type pState struct {
			ID        string  `json:"id"`
			Name      string  `json:"name"`
			IP        string  `json:"ip,omitempty"`
			Cam       camInfo `json:"cam"`
			Status    Status  `json:"status"`
			Timelapse tlInfo  `json:"timelapse"`
		}
		ps := []pState{} // never null, even with no printers yet
		for _, p := range s.Printers {
			st := pState{ID: p.ID, Name: p.Name, Cam: getCam(p.ID).info(), Timelapse: timelapseInfo(p.ID)}
			if !isRemote(r) {
				st.IP = p.IP
			}
			if m := getMon(p.ID); m != nil {
				st.Status = m.status()
				if isRemote(r) && st.Status.Info != nil {
					// phones don't need the printer's serial numbers
					in := *st.Status.Info
					in.Serial = ""
					in.Modules = nil
					st.Status.Info = &in
				}
			} else {
				st.Status.Err = "not set up"
			}
			ps = append(ps, st)
		}
		evs := []Event{}
		var last int64
		if since < 0 {
			evs, last = recentEvents(30), 0
			_, last = eventsSince(1 << 62)
		} else {
			evs, last = eventsSince(since)
		}
		writeJSON(w, map[string]any{
			"version": version, "os": runtime.GOOS, "remote": isRemote(r), "active": s.Active,
			"printers": ps, "events": evs, "last_event": last,
			"remote_controls": s.RemoteControls, "theme": s.Theme,
		})
	})

	mux.HandleFunc("/stream", func(w http.ResponseWriter, r *http.Request) {
		id := r.URL.Query().Get("id")
		if _, ok := cfg.printer(id); !ok {
			http.NotFound(w, r)
			return
		}
		cam := getCam(id)
		cam.addClient()
		defer cam.removeClient()
		w.Header().Set("Content-Type", "multipart/x-mixed-replace; boundary=frame")
		w.Header().Set("Cache-Control", "no-store")
		fl, _ := w.(http.Flusher)
		last := -1
		for {
			if r.Context().Err() != nil {
				return
			}
			cam.mu.Lock()
			for cam.seq == last && r.Context().Err() == nil && !cam.closed {
				t := time.AfterFunc(time.Second, func() { cam.cond.Broadcast() })
				cam.cond.Wait()
				t.Stop()
			}
			if cam.closed {
				cam.mu.Unlock()
				return
			}
			last = cam.seq
			f := cam.frame
			cam.mu.Unlock()
			if f == nil {
				continue
			}
			if _, err := fmt.Fprintf(w, "--frame\r\nContent-Type: image/jpeg\r\nContent-Length: %d\r\n\r\n", len(f)); err != nil {
				return
			}
			if _, err := w.Write(f); err != nil {
				return
			}
			io.WriteString(w, "\r\n")
			if fl != nil {
				fl.Flush()
			}
		}
	})

	// download a snapshot (used by phones)
	mux.HandleFunc("/snapshot.jpg", func(w http.ResponseWriter, r *http.Request) {
		id := r.URL.Query().Get("id")
		if _, ok := cfg.printer(id); !ok {
			http.NotFound(w, r)
			return
		}
		f := getCam(id).waitFrame(20 * time.Second)
		if f == nil {
			http.Error(w, "no picture from the camera", 504)
			return
		}
		w.Header().Set("Content-Type", "image/jpeg")
		w.Header().Set("Content-Disposition", "attachment; filename=snapshot-"+time.Now().Format("20060102-150405")+".jpg")
		w.Write(f)
	})

	// save a snapshot into the snapshot folder (desktop)
	mux.HandleFunc("/api/snapshot", localOnly(func(w http.ResponseWriter, r *http.Request) {
		id := r.URL.Query().Get("id")
		p, ok := cfg.printer(id)
		if !ok {
			http.NotFound(w, r)
			return
		}
		f := getCam(id).waitFrame(20 * time.Second)
		if f == nil {
			reason := "No picture from the camera"
			if getCam(id).info().Paused {
				reason = "The camera is paused - resume it to take a snapshot"
			}
			writeJSON(w, map[string]any{"ok": false, "reason": reason})
			return
		}
		dir := cfg.get().SnapshotDir
		if err := os.MkdirAll(dir, 0o755); err != nil {
			writeJSON(w, map[string]any{"ok": false, "reason": err.Error()})
			return
		}
		name := unsafeName.ReplaceAllString(p.Name, "") + " " + time.Now().Format("2006-01-02 150405") + ".jpg"
		out := filepath.Join(dir, name)
		if err := os.WriteFile(out, f, 0o644); err != nil {
			writeJSON(w, map[string]any{"ok": false, "reason": err.Error()})
			return
		}
		writeJSON(w, map[string]any{"ok": true, "path": out})
	}))

	mux.HandleFunc("/api/campause", func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			ID     string
			Paused bool
		}
		if readJSON(r, &req) != nil {
			http.Error(w, "bad request", 400)
			return
		}
		if isRemote(r) && !cfg.get().RemoteControls {
			http.Error(w, "controls are turned off for other devices", 403)
			return
		}
		if _, ok := cfg.printer(req.ID); !ok {
			http.NotFound(w, r)
			return
		}
		getCam(req.ID).setPaused(req.Paused)
		writeJSON(w, map[string]any{"ok": true})
	})

	mux.HandleFunc("/api/retry", func(w http.ResponseWriter, r *http.Request) {
		getCam(r.URL.Query().Get("id")).retryNow()
		if m := getMon(r.URL.Query().Get("id")); m != nil {
			select {
			case m.wake <- struct{}{}:
			default:
			}
		}
		writeJSON(w, map[string]any{"ok": true})
	})

	mux.HandleFunc("/api/cmd", func(w http.ResponseWriter, r *http.Request) {
		if isRemote(r) && !cfg.get().RemoteControls {
			http.Error(w, "controls are turned off for other devices", 403)
			return
		}
		var req struct{ ID, Cmd, Value string }
		if readJSON(r, &req) != nil {
			http.Error(w, "bad request", 400)
			return
		}
		m := getMon(req.ID)
		if m == nil {
			writeJSON(w, map[string]any{"ok": false, "reason": "Not connected to this printer's status feed"})
			return
		}
		if req.Cmd == "retry" {
			m.forgetRefusals()
			writeJSON(w, map[string]any{"ok": true})
			return
		}
		mqttCmd := map[string]string{"pause": "pause", "resume": "resume", "stop": "stop", "speed": "print_speed", "light": "ledctrl"}[req.Cmd]
		if m.status().needsDevMode(mqttCmd) {
			writeJSON(w, map[string]any{"ok": false, "reason": devOnlyMsg})
			return
		}
		if mqttCmd != "" && m.status().refused(mqttCmd) {
			// it already refused this one: sending again only puts another error on its screen
			writeJSON(w, map[string]any{"ok": false, "reason": lockedMsg})
			return
		}
		var ok bool
		var reason string
		sent := time.Now()
		switch req.Cmd {
		case "pause", "resume", "stop":
			ok, reason = m.command("print", req.Cmd, map[string]any{"param": ""})
		case "speed":
			if n, err := strconv.Atoi(req.Value); err != nil || n < 1 || n > 4 {
				http.Error(w, "bad speed", 400)
				return
			}
			ok, reason = m.command("print", "print_speed", map[string]any{"param": req.Value})
		case "light":
			if req.Value != "on" && req.Value != "off" {
				http.Error(w, "bad value", 400)
				return
			}
			ok, reason = m.command("system", "ledctrl", map[string]any{"led_node": "chamber_light", "led_mode": req.Value,
				"led_on_time": 500, "led_off_time": 500, "loop_times": 0, "interval_time": 0})
		default:
			http.Error(w, "unknown command", 400)
			return
		}
		if ok {
			// A reply of "OK" isn't proof (a printer that refuses unsigned
			// commands can still say OK): wait until the change shows up.
			ok, reason = false, "Sent, but the printer hasn't changed yet. Check the printer's screen."
			for deadline := time.Now().Add(6 * time.Second); time.Now().Before(deadline); time.Sleep(250 * time.Millisecond) {
				if cmdTookEffect(req.Cmd, req.Value, m.status()) {
					ok, reason = true, ""
					break
				}
				if m.refusedSince(sent) {
					m.markRefused(mqttCmd)
					reason = devModeMsg
					break
				}
			}
		}

		writeJSON(w, map[string]any{"ok": ok, "reason": reason})
	})

	mux.HandleFunc("/api/timelapse", func(w http.ResponseWriter, r *http.Request) {
		if isRemote(r) && !cfg.get().RemoteControls {
			http.Error(w, "controls are turned off for other devices", 403)
			return
		}
		var req struct{ ID, Action string }
		if readJSON(r, &req) != nil {
			http.Error(w, "bad request", 400)
			return
		}
		var err error
		switch req.Action {
		case "start":
			file := "manual"
			if m := getMon(req.ID); m != nil && m.status().File != "" {
				file = m.status().File
			}
			err = timelapseStart(req.ID, file)
		case "stop":
			timelapseStop(req.ID, "manual")
		}
		if err != nil {
			writeJSON(w, map[string]any{"ok": false, "reason": err.Error()})
			return
		}
		writeJSON(w, map[string]any{"ok": true})
	})

	// ---- desktop-only below

	mux.HandleFunc("/api/settings", localOnly(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			var in struct {
				Settings
				NewPIN bool `json:"new_pin"`
			}
			if readJSON(r, &in) != nil {
				http.Error(w, "bad request", 400)
				return
			}
			err := cfg.update(func(s *Settings) {
				s.SnapshotDir = strings.TrimSpace(in.SnapshotDir)
				s.TimelapseDir = strings.TrimSpace(in.TimelapseDir)
				s.TimelapseAuto = in.TimelapseAuto
				s.TimelapseMode = in.TimelapseMode
				s.TimelapseInterval = in.TimelapseInterval
				s.SoundProblems = in.SoundProblems
				s.SoundFinished = in.SoundFinished
				if in.Theme != "" {
					s.Theme = in.Theme
				}
				s.NotifyFlash = in.NotifyFlash
				s.Tray = in.Tray
				s.Autostart = in.Autostart
				s.Remote = in.Remote
				if in.RemotePort >= 1024 && in.RemotePort <= 65535 && in.RemotePort != 8787 {
					s.RemotePort = in.RemotePort
				}
				s.RemoteControls = in.RemoteControls
				if in.NewPIN {
					s.RemotePIN = randPIN()
					s.Secret = randHex(16) // signs everyone out
				}
			})
			if err != nil {
				http.Error(w, err.Error(), 500)
				return
			}
			applySettings()
		}
		writeJSON(w, settingsView())
	}))

	mux.HandleFunc("/api/printer", localOnly(func(w http.ResponseWriter, r *http.Request) {
		var in struct {
			Printer
			TrustNew bool `json:"trust_new"` // forget the remembered certificates
		}
		if readJSON(r, &in) != nil {
			http.Error(w, "bad request", 400)
			return
		}
		p := cleanPrinter(in.Printer)
		p.Pins, p.CodeEnc = nil, ""
		if net.ParseIP(p.IP) == nil {
			writeJSON(w, map[string]any{"ok": false, "reason": "That isn't a valid IP address"})
			return
		}
		if strings.ContainsAny(p.Code, "\r\n\x00") {
			writeJSON(w, map[string]any{"ok": false, "reason": "The access code can't contain line breaks"})
			return
		}
		_, known := cfg.printer(p.ID)
		if p.Code == "" && !(known && p.ID != "") {
			writeJSON(w, map[string]any{"ok": false, "reason": "Enter the access code"})
			return
		}
		var old Printer
		var existed bool
		cfg.update(func(s *Settings) {
			for i := range s.Printers {
				if s.Printers[i].ID == p.ID && p.ID != "" {
					old, existed = s.Printers[i], true
					if p.Code == "" {
						p.Code = old.Code // the page never sees the saved code; empty means unchanged
					}
					if old.IP == p.IP && !in.TrustNew {
						p.Pins = old.Pins
					}
					p.FwSeen = old.FwSeen
					if old.IP != p.IP {
						p.Serial = "" // different printer: detect again
					}
					s.Printers[i] = p
					return
				}
			}
			p.ID = randHex(4)
			s.Printers = append(s.Printers, p)
			if len(s.Printers) == 1 {
				s.Active = p.ID
			}
		})
		if existed && in.TrustNew {
			appendLog("certificates for " + p.Name + " will be learned again")
		}
		if existed && (old.IP != p.IP || old.Code != p.Code || old.Mode != p.Mode || in.TrustNew) {
			getCam(p.ID).restart()
			if m := getMon(p.ID); m != nil {
				select {
				case m.wake <- struct{}{}:
				default:
				}
			}
		}
		syncMonitors()
		writeJSON(w, map[string]any{"ok": true, "id": p.ID})
	}))

	mux.HandleFunc("/api/printer/delete", localOnly(func(w http.ResponseWriter, r *http.Request) {
		var req struct{ ID string }
		if readJSON(r, &req) != nil {
			http.Error(w, "bad request", 400)
			return
		}
		timelapseStop(req.ID, "removed")
		cfg.update(func(s *Settings) {
			var keep []Printer
			for _, p := range s.Printers {
				if p.ID != req.ID {
					keep = append(keep, p)
				}
			}
			s.Printers = keep
		})
		dropCam(req.ID)
		syncMonitors()
		writeJSON(w, map[string]any{"ok": true})
	}))

	mux.HandleFunc("/api/active", localOnly(func(w http.ResponseWriter, r *http.Request) {
		var req struct{ ID string }
		if readJSON(r, &req) == nil {
			cfg.update(func(s *Settings) { s.Active = req.ID })
		}
		writeJSON(w, map[string]any{"ok": true})
	}))

	mux.HandleFunc("/api/scan", localOnly(func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 6*time.Second)
		defer cancel()
		writeJSON(w, scanPrinters(ctx))
	}))

	mux.HandleFunc("/api/raw", localOnly(func(w http.ResponseWriter, r *http.Request) {
		m := getMon(r.URL.Query().Get("id"))
		w.Header().Set("Content-Type", "application/json")
		if m == nil {
			io.WriteString(w, "{}")
			return
		}
		w.Write(m.rawJSON())
	}))

	mux.HandleFunc("/api/window", localOnly(func(w http.ResponseWriter, r *http.Request) {
		var req struct{ OnTop, Compact bool }
		if readJSON(r, &req) != nil {
			http.Error(w, "bad request", 400)
			return
		}
		setWindowMode(req.OnTop, req.Compact)
		writeJSON(w, map[string]any{"ok": true})
	}))

	// Opens the remote access page in this PC's normal browser, to try it out.
	// The address is fixed (this PC, the configured port): nothing from the
	// request goes into it.
	mux.HandleFunc("/api/openremote", localOnly(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.Error(w, "POST only", 405)
			return
		}
		if !remoteRunning() {
			writeJSON(w, map[string]any{"ok": false, "reason": "Turn remote access on first."})
			return
		}
		openInBrowser(fmt.Sprintf("http://127.0.0.1:%d/", cfg.get().RemotePort))
		writeJSON(w, map[string]any{"ok": true})
	}))

	mux.HandleFunc("/api/openfolder", localOnly(func(w http.ResponseWriter, r *http.Request) {
		var req struct{ Which, Path string }
		readJSON(r, &req)
		s := cfg.get()
		var dir string
		switch req.Which {
		case "snapshots":
			dir = s.SnapshotDir
		case "timelapses":
			dir = s.TimelapseDir
		case "file":
			// only existing files inside our own folders
			p := filepath.Clean(req.Path)
			if !(insideDir(p, s.SnapshotDir) || insideDir(p, s.TimelapseDir)) || strings.ContainsAny(p, "\"\x00") {
				http.Error(w, "not one of Printrment's files", 403)
				return
			}
			if st, err := os.Stat(p); err != nil || !st.Mode().IsRegular() {
				writeJSON(w, map[string]any{"ok": false, "reason": "That file isn't there any more"})
				return
			}
			showInFolder(p)
			writeJSON(w, map[string]any{"ok": true})
			return
		default:
			http.Error(w, "bad request", 400)
			return
		}
		os.MkdirAll(dir, 0o755)
		openFolder(dir)
		writeJSON(w, map[string]any{"ok": true})
	}))

	mux.HandleFunc("/api/testnotify", localOnly(func(w http.ResponseWriter, r *http.Request) {
		var req struct{ Kind string }
		readJSON(r, &req)
		s := cfg.get()
		switch req.Kind {
		case "error":
			alert(s.Active, "error", "Test: this is how a failed print or error sounds")
		case "warn":
			alert(s.Active, "warn", "Test: this is how a paused print or printer alert sounds")
		default:
			alert(s.Active, "success", "Test: this is how a finished print sounds")
		}
		writeJSON(w, map[string]any{"ok": true})
	}))

	mux.HandleFunc("/api/show", localOnly(func(w http.ResponseWriter, r *http.Request) {
		go openWindow(appURL())
		writeJSON(w, map[string]any{"ok": true})
	}))

	// The Edge fallback window can't tell us it was closed, so the page sends
	// this as it unloads. If no page comes back within a few seconds (a reload
	// comes straight back), the app shuts down - unless it lives in the tray.
	mux.HandleFunc("/api/closing", localOnly(func(w http.ResponseWriter, r *http.Request) {
		at := time.Now()
		go func() {
			time.Sleep(5 * time.Second)
			s := cfg.get()
			if idleFor() >= time.Since(at)-500*time.Millisecond && !(s.Tray && trayAvailable()) && !timelapseRecordingAny() {
				quitApp()
			}
		}()
		w.WriteHeader(204)
	}))

	mux.HandleFunc("/api/quit", localOnly(func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, map[string]any{"ok": true})
		go quitApp()
	}))

	return mux
}

func cmdTookEffect(cmd, value string, st Status) bool {
	switch cmd {
	case "pause":
		return st.State == "PAUSE"
	case "resume":
		return st.State == "RUNNING" || st.State == "PREPARE"
	case "stop":
		return !activeState(st.State)
	case "speed":
		return strconv.Itoa(st.Speed) == value
	case "light":
		return st.Light == value
	}
	return true
}

func localOnly(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if isRemote(r) {
			http.Error(w, "only available on the PC running Printrment", 403)
			return
		}
		h(w, r)
	}
}

func settingsView() map[string]any {
	s := cfg.get()
	// The page never gets the saved access codes back - it only needs to know
	// one is set. They stay inside this program.
	for i := range s.Printers {
		s.Printers[i].Code, s.Printers[i].Pins = "", nil
	}
	var urls []string
	for _, ip := range lanIPs() {
		urls = append(urls, fmt.Sprintf("http://%s:%d", ip, s.RemotePort))
	}
	return map[string]any{"settings": s, "remote_urls": urls, "remote_running": remoteRunning(), "remote_error": remoteError()}
}

func lanIPs() []string {
	var out []string
	addrs, _ := net.InterfaceAddrs()
	for _, a := range addrs {
		if ipn, ok := a.(*net.IPNet); ok {
			if ip4 := ipn.IP.To4(); ip4 != nil && !ip4.IsLoopback() && !ip4.IsLinkLocalUnicast() {
				out = append(out, ip4.String())
			}
		}
	}
	return out
}

// ------------------------------------------------------------ protection

// localGuard protects the desktop listener. It is only reachable from this PC,
// but other programs (and other Windows users) on this PC, and web pages open
// in a browser, can all send it requests. So:
//   - the Host must be ours (blocks DNS-rebinding tricks), and
//   - everything except the page itself, its fonts and the version number
//     needs this run's secret token, which only Printrment's own window gets.
func localGuard(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Host != listenAddr && r.Host != "localhost:8787" {
			http.Error(w, "forbidden", 403)
			return
		}
		open := r.URL.Path == "/" || r.URL.Path == "/api/version" || strings.HasPrefix(r.URL.Path, "/fonts/")
		if !open && !tokenOK(r) {
			http.Error(w, "forbidden", 403)
			return
		}
		if r.URL.Path != "/api/closing" {
			touch()
		}
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("Referrer-Policy", "no-referrer")
		next.ServeHTTP(w, r)
	})
}

func pinCookie(s Settings) string {
	mac := hmac.New(sha256.New, []byte(s.Secret))
	mac.Write([]byte(s.RemotePIN))
	return hex.EncodeToString(mac.Sum(nil))
}

// remoteGuard protects the remote access listener with the PIN.
func remoteGuard(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		s := cfg.get()
		if !s.Remote {
			http.Error(w, "remote access is turned off", 403)
			return
		}
		// Home network / VPN only, even if this PC is reachable from the internet.
		if !lanClient(r.RemoteAddr) {
			http.Error(w, "only available on your local network", 403)
			return
		}
		ip, _, _ := net.SplitHostPort(r.RemoteAddr)
		r = r.WithContext(context.WithValue(r.Context(), remoteKey, true))
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("Referrer-Policy", "no-referrer")
		if r.URL.Path == "/login" && r.Method == http.MethodPost {
			if logins.blocked(ip) {
				w.Header().Set("Content-Type", "text/html; charset=utf-8")
				w.WriteHeader(429)
				io.WriteString(w, strings.Replace(loginPage, "{{ERR}}", "Too many wrong PINs. Try again in 15 minutes.", 1))
				return
			}
			r.ParseForm()
			if hmac.Equal([]byte(r.PostForm.Get("pin")), []byte(s.RemotePIN)) {
				logins.ok(ip)
				http.SetCookie(w, &http.Cookie{Name: "bc_auth", Value: pinCookie(s), Path: "/", HttpOnly: true,
					SameSite: http.SameSiteStrictMode, MaxAge: 60 * 60 * 24 * 90})
				http.Redirect(w, r, "/", http.StatusSeeOther)
				return
			}
			logins.fail(ip)
			time.Sleep(1500 * time.Millisecond)
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			io.WriteString(w, strings.Replace(loginPage, "{{ERR}}", "Wrong PIN, try again.", 1))
			return
		}
		if strings.HasPrefix(r.URL.Path, "/fonts/") {
			next.ServeHTTP(w, r)
			return
		}
		c, err := r.Cookie("bc_auth")
		if err != nil || !hmac.Equal([]byte(c.Value), []byte(pinCookie(s))) {
			if strings.HasPrefix(r.URL.Path, "/api/") || r.URL.Path == "/stream" || r.URL.Path == "/snapshot.jpg" {
				http.Error(w, "sign in first", 401)
				return
			}
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			io.WriteString(w, strings.Replace(loginPage, "{{ERR}}", "", 1))
			return
		}
		// Cross-site protection: the sign-in cookie is SameSite=Strict, and every
		// /api call must also carry our header, which other sites can't add.
		if strings.HasPrefix(r.URL.Path, "/api/") && r.Header.Get("X-Printrment") == "" {
			http.Error(w, "forbidden", 403)
			return
		}
		next.ServeHTTP(w, r)
	})
}

var (
	remoteMu   sync.Mutex
	remoteSrv  *http.Server
	remotePort int
	remoteErr  string
)

func remoteRunning() bool {
	remoteMu.Lock()
	defer remoteMu.Unlock()
	return remoteSrv != nil
}

func remoteError() string {
	remoteMu.Lock()
	defer remoteMu.Unlock()
	return remoteErr
}

func applyRemote(mux http.Handler) {
	s := cfg.get()
	remoteMu.Lock()
	defer remoteMu.Unlock()
	if remoteSrv != nil && (!s.Remote || remotePort != s.RemotePort) {
		remoteSrv.Close()
		remoteSrv = nil
	}
	if s.Remote && remoteSrv == nil {
		ln, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", s.RemotePort))
		if err != nil {
			remoteErr = "Couldn't open port " + strconv.Itoa(s.RemotePort) + ": " + err.Error()
			return
		}
		remoteErr = ""
		remotePort = s.RemotePort
		remoteSrv = newServer(remoteGuard(mux))
		go remoteSrv.Serve(ln)
	}
}

const loginPage = `<!doctype html><meta charset=utf-8><meta name=viewport content="width=device-width,initial-scale=1">
<title>Printrment</title>
<style>body{margin:0;min-height:100vh;display:flex;align-items:center;justify-content:center;background:#0d0f12;color:#d8dce2;font:16px system-ui,sans-serif}
form{background:#171a20;border:1px solid #2a303a;border-radius:12px;padding:24px;width:min(320px,90vw);display:flex;flex-direction:column;gap:12px}
input{background:#0d0f12;color:#d8dce2;border:1px solid #333a45;border-radius:8px;padding:12px;font:inherit;letter-spacing:4px;text-align:center}
button{background:#2f6fed;color:#fff;border:0;border-radius:8px;padding:12px;font:inherit}
p{margin:0;color:#8b93a1;font-size:14px}.e{color:#ff7a7a}</style>
<form method=post action="/login"><b>Printrment</b><p>Enter the PIN shown on your PC in Printrment, under Settings &gt; Remote access.</p>
<input name=pin inputmode=numeric autocomplete=one-time-code maxlength=6 autofocus><p class=e>{{ERR}}</p><button>Open camera</button></form>`
