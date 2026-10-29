package main

import (
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	_ "embed"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

//go:embed ffmpeg.exe.gz
var ffmpegGz []byte

// ---------------------------------------------------------------- ffmpeg

var ffmpegOnce sync.Once
var ffmpegPath string
var ffmpegErr error

// ffmpegSHA256 is the SHA-256 of the unpacked ffmpeg.exe (checked by
// TestFFmpegHash, so it can't drift from the embedded file).
const ffmpegSHA256 = "2ce797a0f88d7f067180338fb227f7b1928ea727bd9a4d7a1d022f7c52af71a3"

func fileSHA256(p string) string {
	f, err := os.Open(p)
	if err != nil {
		return ""
	}
	defer f.Close()
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return ""
	}
	return hex.EncodeToString(h.Sum(nil))
}

// getFFmpeg unpacks the video helper once and, every time the app starts,
// checks it is byte-for-byte the copy built into Printrment before running it.
// A damaged or swapped file is replaced.
func getFFmpeg() (string, error) {
	if p := os.Getenv("BAMBUCAM_FFMPEG"); p != "" {
		return p, nil
	}
	ffmpegOnce.Do(func() {
		p := filepath.Join(localDir(), "ffmpeg-"+ffmpegSHA256[:16]+".exe")
		if fileSHA256(p) == ffmpegSHA256 {
			ffmpegPath = p
			return
		}
		zr, err := gzip.NewReader(bytes.NewReader(ffmpegGz))
		if err != nil {
			ffmpegErr = err
			return
		}
		tmp := p + "." + randHex(4) + ".part"
		f, err := os.OpenFile(tmp, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o700)
		if err != nil {
			ffmpegErr = err
			return
		}
		h := sha256.New()
		_, err = io.Copy(io.MultiWriter(f, h), zr)
		if cerr := f.Close(); err == nil {
			err = cerr
		}
		if err == nil && hex.EncodeToString(h.Sum(nil)) != ffmpegSHA256 {
			err = errors.New("the built-in copy is damaged - download Printrment again")
		}
		if err != nil {
			os.Remove(tmp)
			ffmpegErr = err
			return
		}
		os.Remove(p)
		if err := os.Rename(tmp, p); err != nil {
			os.Remove(tmp)
			ffmpegErr = err
			return
		}
		ffmpegPath = p
		// tidy up copies left by older versions (they lived in the roaming profile)
		for _, dir := range []string{dataDir(), localDir()} {
			old, _ := filepath.Glob(filepath.Join(dir, "ffmpeg-*.exe"))
			for _, o := range old {
				if o != p {
					os.Remove(o)
				}
			}
		}
	})
	return ffmpegPath, ffmpegErr
}

var logMu sync.Mutex

func appendLog(s string) {
	logMu.Lock()
	defer logMu.Unlock()
	path := filepath.Join(dataDir(), "log.txt")
	if st, err := os.Stat(path); err == nil && st.Size() > 2<<20 {
		os.Rename(path, path+".old")
	}
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o600)
	if err != nil {
		return
	}
	defer f.Close()
	fmt.Fprintf(f, "%s  %s\n", time.Now().Format("2006-01-02 15:04:05"), s)
}

// ---------------------------------------------------------------- app

var appMux http.Handler

// takeOver: if another copy is running, either reuse it (same version) or ask
// it to quit (older version). Returns true if this process should exit.
func takeOver(showWindow bool) bool {
	conn, err := net.DialTimeout("tcp", listenAddr, 400*time.Millisecond)
	if err != nil {
		return false
	}
	conn.Close()
	cl := http.Client{Timeout: 3 * time.Second}
	same := false
	if r, err := cl.Get("http://" + listenAddr + "/api/version"); err == nil {
		b, _ := io.ReadAll(r.Body)
		r.Body.Close()
		same = strings.TrimSpace(string(b)) == version
	}
	if same {
		if showWindow {
			// ask the running copy to show its window
			req, _ := http.NewRequest("POST", "http://"+listenAddr+"/api/show", nil)
			req.Header.Set("X-Printrment", runningToken())
			if r, err := cl.Do(req); err == nil {
				r.Body.Close()
			}
		}
		return true
	}
	// Versions 21+ take this run's token; 13-20 a POST with "1"; older ones a
	// plain GET.
	for _, t := range []string{runningToken(), "1"} {
		if t == "" {
			continue
		}
		req, _ := http.NewRequest("POST", "http://"+listenAddr+"/api/quit", nil)
		req.Header.Set("X-Printrment", t)
		req.Header.Set("X-BambuCam", t) // copies from before the rename
		if r, err := cl.Do(req); err == nil {
			r.Body.Close()
		}
	}
	if r, err := cl.Get("http://" + listenAddr + "/api/quit"); err == nil {
		r.Body.Close()
	}
	time.Sleep(1500 * time.Millisecond)
	return false
}

func main() {
	hidden := false
	for _, a := range os.Args[1:] {
		if a == "--tray" {
			hidden = true
		}
	}
	if takeOver(!hidden) {
		return
	}
	var ln net.Listener
	var err error
	for i := 0; i < 20; i++ {
		if ln, err = net.Listen("tcp", listenAddr); err == nil {
			break
		}
		time.Sleep(500 * time.Millisecond)
	}
	if err != nil {
		os.Exit(1)
	}

	migrateLegacyFiles()
	writeSessionToken()
	// big machine-specific files moved to the local (non-roaming) profile in v21
	go os.RemoveAll(filepath.Join(dataDir(), "WebView2"))
	go os.RemoveAll(filepath.Join(dataDir(), "timelapse-work"))
	cfg.load()
	migrateLegacyPaths()
	for _, msg := range cfg.takeProblems() {
		alert("", "warn", msg)
	}
	touch()
	mux := newMux()
	appMux = mux
	syncMonitors()
	applySettings()

	go newServer(localGuard(mux)).Serve(ln)

	// Exit when the window is closed, unless we live in the tray or a
	// timelapse is being recorded.
	go func() {
		for {
			time.Sleep(10 * time.Second)
			s := cfg.get()
			if s.Tray && trayAvailable() {
				continue
			}
			busy := timelapseRecordingAny()
			for _, c := range allCams() {
				if c.active() {
					busy = true
				}
			}
			if !busy && idleFor() > 3*time.Minute {
				quitApp()
			}
		}
	}()

	go func() {
		for {
			updateTray()
			time.Sleep(5 * time.Second)
		}
	}()

	if !hidden {
		go func() {
			time.Sleep(300 * time.Millisecond)
			openWindow(appURL())
		}()
	}
	runTray(func() { openWindow(appURL()) }, quitApp) // blocks (or returns at once without a tray)
	select {}
}

func applySettings() {
	s := cfg.get()
	setAutostart(s.Autostart)
	if appMux != nil {
		applyRemote(appMux)
	}
}

var quitOnce sync.Once

func quitApp() {
	quitOnce.Do(func() {
		appendLog("quitting")
		hideAppWindow()
		// close camera connections politely (TEARDOWN) and the status feeds
		for _, c := range allCams() {
			c.shutdown()
		}
		stopMonitors()
		timelapseStopAll()
		os.Remove(tokenPath())
		time.Sleep(600 * time.Millisecond)
		quitTray()
		os.Exit(0)
	})
}

func fmtMinutes(m float64) string {
	if m <= 0 {
		return ""
	}
	h, mm := int(m)/60, int(m)%60
	if h > 0 {
		return fmt.Sprintf("%dh %02dm", h, mm)
	}
	return fmt.Sprintf("%dm", mm)
}

var stateNames = map[string]string{
	"IDLE": "Idle", "PREPARE": "Preparing", "RUNNING": "Printing", "PAUSE": "Paused",
	"FINISH": "Finished", "FAILED": "Failed", "SLICING": "Slicing",
}

func updateTray() {
	s := cfg.get()
	lines := []string{"Printrment"}
	for _, p := range s.Printers {
		line := p.Name + ": "
		m := getMon(p.ID)
		if m == nil {
			line += "not set up"
		} else if st := m.status(); !st.Connected {
			line += "offline"
		} else {
			name := stateNames[st.State]
			if name == "" {
				name = st.State
			}
			line += name
			if st.State == "RUNNING" || st.State == "PAUSE" {
				line += fmt.Sprintf(" %d%%", int(st.Percent))
				if r := fmtMinutes(st.Remaining); r != "" {
					line += ", " + r + " left"
				}
			}
		}
		lines = append(lines, line)
	}
	t := strings.Join(lines, "\n")
	if len(t) > 120 {
		t = t[:120]
	}
	setTrayTooltip(t)
}
