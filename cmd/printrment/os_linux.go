//go:build linux

package main

import (
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"fyne.io/systray"
	"github.com/godbus/dbus/v5"
)

func hideWindow(cmd *exec.Cmd) {}

// ---------------------------------------------------------------- window

// openWindow shows the app in its own browser window (Chrome / Chromium / Edge
// / Brave "app" mode: no tabs or address bar), or as a tab in the default
// browser if none of those is installed. The secret token travels in the
// address fragment, which browsers never send over the network.
func openWindow(url string) {
	if os.Getenv("BROWSER") == "true" { // tests
		return
	}
	args := []string{"--app=" + url, "--window-size=1280,800"}
	for _, b := range []string{"google-chrome", "google-chrome-stable", "chromium-browser", "chromium", "microsoft-edge", "microsoft-edge-stable", "brave-browser"} {
		if p, err := exec.LookPath(b); err == nil {
			if start(exec.Command(p, args...)) {
				return
			}
		}
	}
	// Flatpak installs (common on Fedora) aren't on the PATH
	if fp, err := exec.LookPath("flatpak"); err == nil {
		for _, id := range []string{"com.google.Chrome", "org.chromium.Chromium", "com.microsoft.Edge", "com.brave.Browser"} {
			if exec.Command(fp, "info", id).Run() == nil && start(exec.Command(fp, append([]string{"run", id}, args...)...)) {
				return
			}
		}
	}
	openInBrowser(url) // e.g. Firefox: opens as a tab
}

func start(cmd *exec.Cmd) bool {
	if cmd.Start() != nil {
		return false
	}
	go cmd.Wait()
	return true
}

func openInBrowser(url string) {
	if os.Getenv("BROWSER") == "true" {
		appendLog("open in browser: " + url)
		return
	}
	cmd := exec.Command("xdg-open", url)
	if cmd.Start() == nil {
		go cmd.Wait()
	}
}

func openFolder(dir string) {
	cmd := exec.Command("xdg-open", dir)
	if cmd.Start() == nil {
		go cmd.Wait()
	}
}

// showInFolder opens the file manager with the file selected (the standard
// FileManager1 interface: Files, Dolphin, Nemo, Caja...), or just its folder.
func showInFolder(path string) {
	if conn, err := dbus.SessionBus(); err == nil {
		u := (&url.URL{Scheme: "file", Path: path}).String()
		obj := conn.Object("org.freedesktop.FileManager1", "/org/freedesktop/FileManager1")
		if obj.Call("org.freedesktop.FileManager1.ShowItems", 0, []string{u}, "").Err == nil {
			return
		}
	}
	openFolder(filepath.Dir(path))
}

// These control the Windows app window; a browser window can't be pinned on
// top or flashed from outside, so they do nothing here.
func setWindowMode(onTop, compact bool) {}
func flashWindow()                      {}
func hideAppWindow()                    {}

// ---------------------------------------------------------------- autostart

func autostartFile() string {
	d, err := os.UserConfigDir()
	if err != nil {
		return ""
	}
	return filepath.Join(d, "autostart", "printrment.desktop")
}

// desktopQuote quotes a path for a .desktop Exec line.
func desktopQuote(s string) string {
	r := strings.NewReplacer(`\`, `\\\\`, `"`, `\\"`, "`", "\\\\`", "$", `\\$`)
	return `"` + r.Replace(s) + `"`
}

func setAutostart(on bool) {
	p := autostartFile()
	if p == "" {
		return
	}
	if !on {
		os.Remove(p)
		return
	}
	exe, err := os.Executable()
	if err != nil || strings.ContainsAny(exe, "\n\r") {
		return
	}
	os.MkdirAll(filepath.Dir(p), 0o700)
	entry := fmt.Sprintf("[Desktop Entry]\nType=Application\nName=Printrment\nComment=Bambu Lab printer monitor\nExec=%s --tray\nTerminal=false\nX-GNOME-Autostart-enabled=true\n", desktopQuote(exe))
	os.WriteFile(p, []byte(entry), 0o600)
}

// ---------------------------------------------------------------- tray

var (
	trayMu    sync.Mutex
	trayReady bool
)

// trayHost reports whether the desktop shows tray icons (KDE, Cinnamon, XFCE,
// GNOME with the AppIndicator extension...). Without one, the app behaves as
// if the tray is off: it quits when its window has been closed for a while.
func trayHost() bool {
	conn, err := dbus.SessionBus()
	if err != nil {
		return false
	}
	var has bool
	err = conn.BusObject().Call("org.freedesktop.DBus.NameHasOwner", 0, "org.kde.StatusNotifierWatcher").Store(&has)
	return err == nil && has
}

func trayAvailable() bool {
	trayMu.Lock()
	defer trayMu.Unlock()
	return trayReady
}

func runTray(onOpen, onQuit func()) {
	if !trayHost() {
		return
	}
	systray.Run(func() {
		systray.SetIcon(iconPNG(64))
		systray.SetTitle("Printrment")
		systray.SetTooltip("Printrment")
		open := systray.AddMenuItem("Open Printrment", "Show the window")
		systray.AddSeparator()
		quit := systray.AddMenuItem("Quit", "Close Printrment completely")
		trayMu.Lock()
		trayReady = true
		trayMu.Unlock()
		go func() {
			for {
				select {
				case <-open.ClickedCh:
					onOpen()
				case <-quit.ClickedCh:
					go onQuit()
				}
			}
		}()
	}, func() {})
}

func quitTray() {
	if trayAvailable() {
		systray.Quit()
	}
}

func setTrayTooltip(s string) {
	if trayAvailable() {
		systray.SetTooltip(s)
	}
}

// ---------------------------------------------------------------- sound

var soundMu sync.Mutex

// playSound plays through PipeWire / PulseAudio / ALSA, whichever is there.
func playSound(kind string) {
	wav := soundWAV(kind)
	if wav == nil {
		return
	}
	if d := os.Getenv("PRINTRMENT_SOUNDLOG"); d != "" { // tests
		if f, err := os.OpenFile(d, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644); err == nil {
			f.WriteString(kind + "\n")
			f.Close()
		}
		return
	}
	soundMu.Lock()
	defer soundMu.Unlock()
	p := filepath.Join(localDir(), "sound-"+kind+".wav")
	if err := os.WriteFile(p, wav, 0o600); err != nil {
		return
	}
	for _, player := range []string{"pw-play", "paplay", "aplay"} {
		if bin, err := exec.LookPath(player); err == nil {
			args := []string{p}
			if player == "aplay" {
				args = []string{"-q", p}
			}
			cmd := exec.Command(bin, args...)
			if cmd.Start() == nil {
				done := make(chan struct{})
				go func() { cmd.Wait(); close(done) }()
				select { // don't let a stuck player pile up
				case <-done:
				case <-time.After(10 * time.Second):
					cmd.Process.Kill()
				}
				return
			}
		}
	}
}

var _ = iconICO
