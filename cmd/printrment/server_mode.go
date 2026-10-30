package main

// Server mode: Printrment running without a window, e.g. in Docker on a NAS.
//
// The desktop app relies on "only this PC can reach the page, and only our
// own window knows the per-run token". A server is reached over the network
// by a browser, so instead:
//   - the full app is served on PRINTRMENT_PORT (default 8787) to home-network
//     addresses only, behind an admin password (PBKDF2-hashed in the settings);
//   - sign-in uses a signed, expiring, HttpOnly, SameSite=Strict cookie, and
//     every /api call must also carry the X-Printrment header (other sites
//     can't add it), as remote access does;
//   - wrong passwords are rate-limited and locked out like PINs;
//   - the PIN-protected remote access viewer still runs on its own port.

import (
	"crypto/hmac"
	"crypto/pbkdf2"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"mime"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"
	_ "time/tzdata" // containers often have no time zone files; TZ still works
)

var serverMode = os.Getenv("PRINTRMENT_SERVER") == "1"

func serverPort() int {
	if n, err := strconv.Atoi(os.Getenv("PRINTRMENT_PORT")); err == nil && n >= 1024 && n <= 65535 {
		return n
	}
	return 8787
}

// envPassword: the admin password is set by PRINTRMENT_ADMIN_PASSWORD (then it
// can't be changed in the app), or chosen in the app after a first password
// is generated and shown in the log.
func envPassword() string { return os.Getenv("PRINTRMENT_ADMIN_PASSWORD") }

// ---------------------------------------------------------------- passwords

const pbkdf2Iter = 600000

func hashPassword(pw string) (string, error) {
	salt := make([]byte, 16)
	if _, err := io.ReadFull(rand.Reader, salt); err != nil {
		return "", err
	}
	k, err := pbkdf2.Key(sha256.New, pw, salt, pbkdf2Iter, 32)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("pbkdf2-sha256$%d$%s$%s", pbkdf2Iter, base64.RawStdEncoding.EncodeToString(salt), base64.RawStdEncoding.EncodeToString(k)), nil
}

func checkPassword(pw, stored string) bool {
	parts := strings.Split(stored, "$")
	if len(parts) != 4 || parts[0] != "pbkdf2-sha256" {
		return false
	}
	iter, err := strconv.Atoi(parts[1])
	if err != nil || iter < 100000 || iter > 10000000 {
		return false
	}
	salt, err1 := base64.RawStdEncoding.DecodeString(parts[2])
	want, err2 := base64.RawStdEncoding.DecodeString(parts[3])
	if err1 != nil || err2 != nil {
		return false
	}
	got, err := pbkdf2.Key(sha256.New, pw, salt, iter, len(want))
	return err == nil && hmac.Equal(got, want)
}

func validPassword(pw string) error {
	if len(pw) < 10 {
		return errors.New("use at least 10 characters")
	}
	if len(pw) > 200 {
		return errors.New("that's too long")
	}
	return nil
}

func initialPasswordFile() string {
	return filepath.Join(os.Getenv("PRINTRMENT_DATA"), "initial-admin-password.txt")
}

// setupAdminPassword runs at start-up in server mode.
func setupAdminPassword() error {
	if pw := envPassword(); pw != "" {
		if err := validPassword(pw); err != nil {
			return fmt.Errorf("PRINTRMENT_ADMIN_PASSWORD: %v", err)
		}
		if checkPassword(pw, cfg.get().AdminHash) {
			return nil
		}
		h, err := hashPassword(pw)
		if err != nil {
			return err
		}
		return cfg.update(func(s *Settings) { s.AdminHash = h })
	}
	if cfg.get().AdminHash != "" {
		return nil
	}
	pw := strings.ToLower(randHex(4) + "-" + randHex(4) + "-" + randHex(4))
	h, err := hashPassword(pw)
	if err != nil {
		return err
	}
	if err := cfg.update(func(s *Settings) { s.AdminHash = h }); err != nil {
		return err
	}
	if os.Getenv("PRINTRMENT_DATA") != "" {
		os.WriteFile(initialPasswordFile(), []byte(pw+"\n"), 0o600)
	}
	fmt.Println("==============================================================")
	fmt.Println(" Printrment admin password: " + pw)
	fmt.Println(" Sign in with it, then change it under Settings > App.")
	fmt.Println("==============================================================")
	return nil
}

// ---------------------------------------------------------------- sessions

const adminCookie = "pm_admin"
const sessionDays = 30

// The cookie is "<unix time>.<HMAC>", keyed by the stored password hash: it
// expires, and changing the password signs every browser out.
func sessionValue(t int64) string {
	mac := hmac.New(sha256.New, []byte("printrment-admin-session|"+cfg.get().AdminHash))
	mac.Write([]byte(strconv.FormatInt(t, 10)))
	return strconv.FormatInt(t, 10) + "." + hex.EncodeToString(mac.Sum(nil))
}

func sessionOK(r *http.Request) bool {
	c, err := r.Cookie(adminCookie)
	if err != nil || cfg.get().AdminHash == "" {
		return false
	}
	ts, _, ok := strings.Cut(c.Value, ".")
	t, err := strconv.ParseInt(ts, 10, 64)
	if !ok || err != nil || time.Since(time.Unix(t, 0)) > sessionDays*24*time.Hour || time.Unix(t, 0).After(time.Now().Add(time.Minute)) {
		return false
	}
	return hmac.Equal([]byte(c.Value), []byte(sessionValue(t)))
}

func secureCookies(r *http.Request) bool {
	return r.TLS != nil || os.Getenv("PRINTRMENT_SECURE_COOKIES") == "1"
}

func setSession(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{Name: adminCookie, Value: sessionValue(time.Now().Unix()), Path: "/",
		HttpOnly: true, Secure: secureCookies(r), SameSite: http.SameSiteStrictMode, MaxAge: sessionDays * 24 * 3600})
}

var adminLogins = &loginGuard{what: "admin passwords were entered", perIP: map[string][]time.Time{}}

// adminGuard protects the full app in server mode.
func adminGuard(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !lanClient(r.RemoteAddr) {
			http.Error(w, "only available on your local network", 403)
			return
		}
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("Referrer-Policy", "no-referrer")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		ip, _, _ := net.SplitHostPort(r.RemoteAddr)
		switch {
		case r.URL.Path == "/api/version" || strings.HasPrefix(r.URL.Path, "/fonts/"):
			next.ServeHTTP(w, r)
			return
		case r.URL.Path == "/login" && r.Method == http.MethodPost:
			if adminLogins.blocked(ip) {
				adminLoginPage(w, 429, "Too many wrong passwords. Try again in 15 minutes.")
				return
			}
			r.ParseForm()
			if checkPassword(r.PostForm.Get("password"), cfg.get().AdminHash) {
				adminLogins.ok(ip)
				setSession(w, r)
				http.Redirect(w, r, "/", http.StatusSeeOther)
				return
			}
			adminLogins.fail(ip)
			time.Sleep(1500 * time.Millisecond)
			adminLoginPage(w, 200, "Wrong password, try again.")
			return
		case r.URL.Path == "/logout" && r.Method == http.MethodPost:
			http.SetCookie(w, &http.Cookie{Name: adminCookie, Value: "", Path: "/", MaxAge: -1, HttpOnly: true, SameSite: http.SameSiteStrictMode})
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}
		if !sessionOK(r) {
			if r.URL.Path == "/" {
				adminLoginPage(w, 200, "")
			} else {
				http.Error(w, "sign in first", 401)
			}
			return
		}
		if strings.HasPrefix(r.URL.Path, "/api/") && r.Header.Get("X-Printrment") == "" {
			http.Error(w, "forbidden", 403)
			return
		}
		if desktopOnly[r.URL.Path] {
			http.NotFound(w, r)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// desktopOnly are things that act on the desktop (window, file manager, quit):
// meaningless - or harmful, like stopping the container - on a server.
var desktopOnly = map[string]bool{"/api/window": true, "/api/openremote": true, "/api/openfolder": true,
	"/api/show": true, "/api/closing": true, "/api/quit": true}

func adminLoginPage(w http.ResponseWriter, code int, msg string) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(code)
	io.WriteString(w, strings.Replace(adminLoginHTML, "{{ERR}}", msg, 1))
}

const adminLoginHTML = `<!doctype html><meta charset=utf-8><meta name=viewport content="width=device-width,initial-scale=1">
<title>Printrment</title>
<style>body{margin:0;min-height:100vh;display:flex;align-items:center;justify-content:center;background:#0d0f12;color:#d8dce2;font:16px system-ui,sans-serif}
form{background:#171a20;border:1px solid #2a303a;border-radius:12px;padding:24px;width:min(340px,90vw);display:flex;flex-direction:column;gap:12px}
input{background:#0d0f12;color:#d8dce2;border:1px solid #333a45;border-radius:8px;padding:12px;font:inherit}
button{background:#2f6fed;color:#fff;border:0;border-radius:8px;padding:12px;font:inherit}
p{margin:0;color:#8b93a1;font-size:14px}.e{color:#ff7a7a}</style>
<form method=post action="/login"><b>Printrment</b><p>Enter the admin password.</p>
<input type=password name=password autocomplete=current-password autofocus required><p class=e>{{ERR}}</p><button>Sign in</button></form>`

// ---------------------------------------------------------------- server-only handlers

func registerServerHandlers(mux *http.ServeMux) {
	// Change the admin password (only when it isn't set by the environment).
	mux.HandleFunc("/api/adminpassword", localOnly(func(w http.ResponseWriter, r *http.Request) {
		if !serverMode || envPassword() != "" || r.Method != http.MethodPost {
			http.NotFound(w, r)
			return
		}
		var in struct{ Current, New string }
		if readJSON(r, &in) != nil {
			http.Error(w, "bad request", 400)
			return
		}
		ip, _, _ := net.SplitHostPort(r.RemoteAddr)
		if adminLogins.blocked(ip) {
			writeJSON(w, map[string]any{"ok": false, "reason": "Too many wrong passwords. Try again in 15 minutes."})
			return
		}
		if !checkPassword(in.Current, cfg.get().AdminHash) {
			adminLogins.fail(ip)
			writeJSON(w, map[string]any{"ok": false, "reason": "The current password is wrong."})
			return
		}
		if err := validPassword(in.New); err != nil {
			writeJSON(w, map[string]any{"ok": false, "reason": "New password: " + err.Error() + "."})
			return
		}
		h, err := hashPassword(in.New)
		if err == nil {
			err = cfg.update(func(s *Settings) { s.AdminHash = h })
		}
		if err != nil {
			writeJSON(w, map[string]any{"ok": false, "reason": err.Error()})
			return
		}
		os.Remove(initialPasswordFile())
		setSession(w, r) // keep this browser signed in; all others are signed out
		activity("", "info", "Admin password changed")
		writeJSON(w, map[string]any{"ok": true})
	}))

	// Shows a saved picture or timelapse in the browser (server mode has no
	// file manager to open). Only files inside the snapshot / timelapse folders.
	mux.HandleFunc("/media", localOnly(func(w http.ResponseWriter, r *http.Request) {
		if !serverMode {
			http.NotFound(w, r)
			return
		}
		s := cfg.get()
		p := r.URL.Query().Get("path")
		if !(insideDir(p, s.SnapshotDir) || insideDir(p, s.TimelapseDir)) {
			http.Error(w, "not one of Printrment's files", 403)
			return
		}
		// no following links out of the folders either
		real, err := filepath.EvalSymlinks(p)
		sd, _ := filepath.EvalSymlinks(s.SnapshotDir)
		td, _ := filepath.EvalSymlinks(s.TimelapseDir)
		if err != nil || !(insideDir(real, sd) || insideDir(real, td)) {
			http.NotFound(w, r)
			return
		}
		f, err := os.Open(real)
		if err != nil {
			http.NotFound(w, r)
			return
		}
		defer f.Close()
		st, err := f.Stat()
		if err != nil || !st.Mode().IsRegular() {
			http.NotFound(w, r)
			return
		}
		ext := strings.ToLower(filepath.Ext(p))
		if ext != ".jpg" && ext != ".jpeg" && ext != ".mp4" {
			http.Error(w, "unsupported file", 403)
			return
		}
		w.Header().Set("Content-Type", mime.TypeByExtension(ext))
		w.Header().Set("Content-Disposition", mime.FormatMediaType("inline", map[string]string{"filename": filepath.Base(p)}))
		w.Header().Set("Content-Security-Policy", "default-src 'none'; img-src 'self'; media-src 'self'")
		http.ServeContent(w, r, filepath.Base(p), st.ModTime(), f)
	}))

	// Alert sounds, played by the browser in server mode.
	mux.HandleFunc("/sound", localOnly(func(w http.ResponseWriter, r *http.Request) {
		wav := soundWAV(r.URL.Query().Get("kind"))
		if wav == nil {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "audio/wav")
		w.Header().Set("Cache-Control", "max-age=86400")
		w.Write(wav)
	}))
}

// ---------------------------------------------------------------- start-up

// runServer is main() for server mode. It never returns.
func runServer() {
	port := serverPort()
	ln, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", port))
	if err != nil {
		fmt.Fprintln(os.Stderr, "Printrment: can't listen on port", port, "-", err)
		os.Exit(1)
	}
	cfg.load()
	for _, msg := range cfg.takeProblems() {
		alert("", "warn", msg)
		fmt.Fprintln(os.Stderr, "Printrment:", msg)
	}
	if err := setupAdminPassword(); err != nil {
		fmt.Fprintln(os.Stderr, "Printrment:", err)
		os.Exit(1)
	}
	mux := newMux()
	appMux = mux
	syncMonitors()
	if cfg.get().RemotePort == port {
		cfg.update(func(s *Settings) { s.RemotePort = port + 1 })
	}
	applySettings()
	go newServer(adminGuard(mux)).Serve(ln)
	appendLog(fmt.Sprintf("server mode: listening on port %d", port))
	fmt.Printf("Printrment %s running. Open http://<this machine's address>:%d in a browser on your home network.\n", version, port)

	// docker stop / Ctrl+C: finish timelapses and close printer connections properly
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGTERM, syscall.SIGINT)
	go func() { <-sig; quitApp() }()
	go func() {
		for {
			updateTray()
			time.Sleep(30 * time.Second)
		}
	}()
	select {}
}

// healthcheck is "printrment --healthcheck" for Docker's HEALTHCHECK.
func healthcheck() {
	c := http.Client{Timeout: 4 * time.Second}
	r, err := c.Get(fmt.Sprintf("http://127.0.0.1:%d/api/version", serverPort()))
	if err != nil {
		os.Exit(1)
	}
	r.Body.Close()
	if r.StatusCode != 200 {
		os.Exit(1)
	}
	os.Exit(0)
}
