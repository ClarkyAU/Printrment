package main

import (
	"crypto/sha256"
	"crypto/tls"
	"encoding/hex"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// ---------------------------------------------------------------- local API token
//
// The app's local web server is reachable by every program on this PC (and
// every web page open in a browser). So every request needs a secret token
// that only Printrment's own window knows. It changes every time the app
// starts. The window gets it directly; a second copy of the app reads it from
// a file only this Windows user can open.

var sessionToken = randHex(24)

// localDir is for large, machine-specific files (video helper, browser cache):
// %LOCALAPPDATA%\Printrment, so they don't travel with a roaming profile.
func localDir() string {
	d, err := os.UserCacheDir()
	if err != nil {
		return dataDir()
	}
	p := filepath.Join(d, "Printrment")
	os.MkdirAll(p, 0o700)
	return p
}

// appURL opens the app; the token rides in the fragment, which browsers never
// send over the network.
func appURL() string { return "http://" + listenAddr + "/#t=" + sessionToken }

func tokenPath() string { return filepath.Join(localDir(), "session.token") }

func writeSessionToken() { os.WriteFile(tokenPath(), []byte(sessionToken), 0o600) }

func runningToken() string {
	b, _ := os.ReadFile(tokenPath())
	if len(b) == 0 { // a copy from before the app was renamed keeps its token elsewhere
		b, _ = os.ReadFile(filepath.Join(legacyLocalDir(), "session.token"))
	}
	return strings.TrimSpace(string(b))
}

func tokenOK(r *http.Request) bool {
	t := r.Header.Get("X-Printrment")
	if t == "" {
		t = r.URL.Query().Get("t") // <img>/<video> and beacons can't send headers
	}
	return t != "" && hmacEqual(t, sessionToken)
}

func hmacEqual(a, b string) bool {
	if len(a) != len(b) {
		return false
	}
	var v byte
	for i := 0; i < len(a); i++ {
		v |= a[i] ^ b[i]
	}
	return v == 0
}

// ---------------------------------------------------------------- printer certificates
//
// Bambu printers use their own self-signed certificates, so normal web-style
// checking can't work. Instead Printrment remembers each printer's certificate
// the first time it connects (per port) and refuses to talk to anything that
// presents a different one later - someone pretending to be the printer can't
// collect the access code or show a fake camera.

type certChangedErr struct{ port string }

func (e *certChangedErr) Error() string {
	return "the printer's security certificate on port " + e.port + " has changed"
}

func certMessage(p Printer) string {
	return "The printer's security certificate has changed, so Printrment stopped connecting. If you reset, replaced or updated the printer, open Settings > Printers > Edit " + p.Name + " and choose \"Trust the printer's new certificate\". Otherwise something on your network may be pretending to be the printer."
}

var (
	pinMu      sync.Mutex
	certWarned = map[string]bool{} // printer|port, so the alert shows once per run
)

func printerTLS(p Printer, port string) *tls.Config {
	return &tls.Config{
		// Chain checks are replaced by the pinning below (self-signed certs).
		InsecureSkipVerify: true,
		MinVersion:         tls.VersionTLS12,
		VerifyConnection: func(cs tls.ConnectionState) error {
			if len(cs.PeerCertificates) == 0 {
				return errors.New("the printer sent no certificate")
			}
			sum := sha256.Sum256(cs.PeerCertificates[0].Raw)
			fp := hex.EncodeToString(sum[:])
			pinMu.Lock()
			defer pinMu.Unlock()
			cur, ok := cfg.printer(p.ID)
			if !ok {
				return errors.New("printer removed")
			}
			if want := cur.Pins[port]; want != "" {
				if !hmacEqual(want, fp) {
					appendLog(fmt.Sprintf("certificate mismatch for %s port %s: have %s..., got %s...", cur.Name, port, want[:12], fp[:12]))
					if k := cur.ID + "|" + port; !certWarned[k] {
						certWarned[k] = true
						go alert(cur.ID, "error", certMessage(cur))
					}
					return &certChangedErr{port}
				}
				return nil
			}
			cfg.update(func(s *Settings) {
				for i := range s.Printers {
					if s.Printers[i].ID == p.ID {
						pins := map[string]string{}
						for k, v := range s.Printers[i].Pins {
							pins[k] = v
						}
						pins[port] = fp
						s.Printers[i].Pins = pins // copy-on-write: readers keep their own map
					}
				}
			})
			delete(certWarned, cur.ID+"|"+port)
			appendLog("remembered " + cur.Name + "'s certificate for port " + port + " (" + tls.VersionName(cs.Version) + ")")
			return nil
		},
	}
}

// ---------------------------------------------------------------- remote access

// lanClient is true for addresses on a private network (home Wi-Fi, VPNs like
// Tailscale), so remote access is never open to the internet even if this PC
// has a public address or a forwarded port.
func lanClient(remoteAddr string) bool {
	host, _, err := net.SplitHostPort(remoteAddr)
	if err != nil {
		return false
	}
	ip := net.ParseIP(host)
	if ip == nil {
		return false
	}
	if ip.IsLoopback() || ip.IsPrivate() || ip.IsLinkLocalUnicast() {
		return true
	}
	_, cgnat, _ := net.ParseCIDR("100.64.0.0/10") // Tailscale and similar
	return cgnat.Contains(ip)
}

// PIN guessing is limited: 5 wrong PINs locks that device out for 15 minutes,
// and 20 wrong PINs from anywhere locks all sign-ins for 15 minutes.
type loginGuard struct {
	mu      sync.Mutex
	perIP   map[string][]time.Time
	all     []time.Time
	lockAll time.Time
}

var logins = &loginGuard{perIP: map[string][]time.Time{}}

func recent(ts []time.Time, d time.Duration) []time.Time {
	out := ts[:0]
	for _, t := range ts {
		if time.Since(t) < d {
			out = append(out, t)
		}
	}
	return out
}

func (g *loginGuard) blocked(ip string) bool {
	g.mu.Lock()
	defer g.mu.Unlock()
	if time.Now().Before(g.lockAll) {
		return true
	}
	g.perIP[ip] = recent(g.perIP[ip], 15*time.Minute)
	return len(g.perIP[ip]) >= 5
}

func (g *loginGuard) fail(ip string) {
	g.mu.Lock()
	g.perIP[ip] = append(recent(g.perIP[ip], 15*time.Minute), time.Now())
	g.all = append(recent(g.all, 15*time.Minute), time.Now())
	n, total := len(g.perIP[ip]), len(g.all)
	if total >= 20 {
		g.lockAll = time.Now().Add(15 * time.Minute)
	}
	g.mu.Unlock()
	if n == 5 || total == 20 {
		alert("", "warn", "Several wrong PINs were entered for remote access from "+ip+" - sign-ins are paused for 15 minutes")
	}
}

func (g *loginGuard) ok(ip string) {
	g.mu.Lock()
	delete(g.perIP, ip)
	g.mu.Unlock()
}

// ---------------------------------------------------------------- printer storage paths

// remotePath checks a path on the printer's storage. Control characters are
// refused outright: a line break would otherwise smuggle extra commands
// (like deleting files) to the printer.
func remotePath(p string) (string, error) {
	for _, c := range p {
		if c < 0x20 || c == 0x7f {
			return "", errors.New("that file name contains characters Printrment can't use")
		}
	}
	return cleanRemotePath(p), nil
}

// insideDir reports whether file is inside dir (no prefix tricks like
// "BambuCam2", no "..").
func insideDir(file, dir string) bool {
	if dir == "" {
		return false
	}
	rel, err := filepath.Rel(filepath.Clean(dir), filepath.Clean(file))
	if err != nil {
		return false
	}
	return rel != "." && rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator)) && !filepath.IsAbs(rel)
}

// ---------------------------------------------------------------- misc

// tailBuf keeps the last few KB a helper program wrote to its error output.
// It is safe to read while the program is still writing.
type tailBuf struct {
	mu sync.Mutex
	b  []byte
}

func (t *tailBuf) Write(p []byte) (int, error) {
	t.mu.Lock()
	t.b = append(t.b, p...)
	if len(t.b) > 8<<10 {
		t.b = append([]byte(nil), t.b[len(t.b)-4<<10:]...)
	}
	t.mu.Unlock()
	return len(p), nil
}

func (t *tailBuf) String() string {
	t.mu.Lock()
	defer t.mu.Unlock()
	return string(t.b)
}
