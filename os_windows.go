//go:build windows

package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"syscall"
	"unsafe"

	"fyne.io/systray"
	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/registry"
)

func init() { runtime.LockOSThread() } // the tray needs the main thread

func hideWindow(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true, CreationFlags: 0x08000000}
}

// openBrowserWindow shows the app in an app-style Edge window (no tabs or
// address bar), falling back to Chrome, then the default browser. Only used
// when WebView2 isn't available.
func openBrowserWindow(url string) {
	if h := findAppWindow(); h != 0 {
		procShowWindow.Call(uintptr(h), 9) // SW_RESTORE
		procSetForegroundWindow.Call(uintptr(h))
		return
	}
	candidates := []string{
		filepath.Join(os.Getenv("ProgramFiles(x86)"), `Microsoft\Edge\Application\msedge.exe`),
		filepath.Join(os.Getenv("ProgramFiles"), `Microsoft\Edge\Application\msedge.exe`),
		filepath.Join(os.Getenv("ProgramFiles"), `Google\Chrome\Application\chrome.exe`),
		filepath.Join(os.Getenv("ProgramFiles(x86)"), `Google\Chrome\Application\chrome.exe`),
		filepath.Join(os.Getenv("LocalAppData"), `Google\Chrome\Application\chrome.exe`),
	}
	for _, p := range candidates {
		if _, err := os.Stat(p); err == nil {
			cmd := exec.Command(p, "--app="+url, "--window-size=1280,800")
			if cmd.Start() == nil {
				return
			}
		}
	}
	exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
}

// openInBrowser opens a page in the user's normal browser (not the app window).
func openInBrowser(url string) { exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start() }

func openFolder(dir string) { exec.Command("explorer", dir).Start() }

func showInFolder(path string) {
	cmd := exec.Command("explorer")
	cmd.SysProcAttr = &syscall.SysProcAttr{CmdLine: `explorer /select,"` + path + `"`}
	cmd.Start()
}

// ---------------------------------------------------------------- window

var (
	user32                  = windows.NewLazySystemDLL("user32.dll")
	procEnumWindows         = user32.NewProc("EnumWindows")
	procGetWindowTextW      = user32.NewProc("GetWindowTextW")
	procIsWindowVisible     = user32.NewProc("IsWindowVisible")
	procSetWindowPos        = user32.NewProc("SetWindowPos")
	procGetWindowRect       = user32.NewProc("GetWindowRect")
	procFlashWindowEx       = user32.NewProc("FlashWindowEx")
	procGetForegroundWindow = user32.NewProc("GetForegroundWindow")
	procShowWindow          = user32.NewProc("ShowWindow")
	procGetSystemMetrics    = user32.NewProc("GetSystemMetrics")
	procSetForegroundWindow = user32.NewProc("SetForegroundWindow")
)

type rect struct{ L, T, R, B int32 }

func findAppWindow() windows.HWND {
	var found windows.HWND
	cb := windows.NewCallback(func(h windows.HWND, _ uintptr) uintptr {
		if v, _, _ := procIsWindowVisible.Call(uintptr(h)); v == 0 {
			return 1
		}
		buf := make([]uint16, 256)
		procGetWindowTextW.Call(uintptr(h), uintptr(unsafe.Pointer(&buf[0])), 256)
		t := windows.UTF16ToString(buf)
		// The app window's title is exactly the page title; ordinary browser
		// windows add the browser's name, so they are skipped.
		if strings.HasPrefix(t, "Printrment") && !strings.Contains(t, "Edge") && !strings.Contains(t, "Chrome") {
			found = h
			return 0
		}
		return 1
	})
	procEnumWindows.Call(cb, 0)
	return found
}

var (
	winMu     sync.Mutex
	savedRect *rect
)

func appWindow() uintptr {
	if h := nativeWindow(); h != 0 {
		return h
	}
	return uintptr(findAppWindow())
}

func setWindowMode(onTop, compact bool) {
	h := appWindow()
	if h == 0 {
		return
	}
	winMu.Lock()
	defer winMu.Unlock()
	const (
		swpNoMove = 0x0002
		swpNoSize = 0x0001
		swpShow   = 0x0040
	)
	insertAfter := ^uintptr(1) // HWND_NOTOPMOST (-2)
	if onTop {
		insertAfter = ^uintptr(0) // HWND_TOPMOST (-1)
	}
	var r rect
	procGetWindowRect.Call(h, uintptr(unsafe.Pointer(&r)))
	switch {
	case compact && savedRect == nil:
		saved := r
		savedRect = &saved
		// shrink into the same top-right corner it was in
		dpi := int32(windowDPI(h))
		w, hgt := 520*dpi/96, 340*dpi/96
		procSetWindowPos.Call(h, insertAfter, uintptr(r.R-w), uintptr(r.T), uintptr(w), uintptr(hgt), swpShow)
	case !compact && savedRect != nil:
		s := *savedRect
		savedRect = nil
		procSetWindowPos.Call(h, insertAfter, uintptr(s.L), uintptr(s.T), uintptr(s.R-s.L), uintptr(s.B-s.T), swpShow)
	default:
		procSetWindowPos.Call(h, insertAfter, 0, 0, 0, 0, swpNoMove|swpNoSize|swpShow)
	}
}

type flashInfo struct {
	Size    uint32
	Hwnd    windows.HWND
	Flags   uint32
	Count   uint32
	Timeout uint32
}

func flashWindow() {
	h := windows.HWND(appWindow())
	if h == 0 {
		return
	}
	if fg, _, _ := procGetForegroundWindow.Call(); fg == uintptr(h) {
		return
	}
	fi := flashInfo{Hwnd: h, Flags: 0x2 | 0xC, Count: 0} // tray button, until focused
	fi.Size = uint32(unsafe.Sizeof(fi))
	procFlashWindowEx.Call(uintptr(unsafe.Pointer(&fi)))
}

// ---------------------------------------------------------------- autostart

func setAutostart(on bool) {
	k, err := registry.OpenKey(registry.CURRENT_USER, `Software\Microsoft\Windows\CurrentVersion\Run`, registry.SET_VALUE|registry.QUERY_VALUE)
	if err != nil {
		return
	}
	defer k.Close()
	k.DeleteValue("BambuCam") // the entry from before the rename
	if on {
		exe, err := os.Executable()
		if err != nil {
			return
		}
		k.SetStringValue("Printrment", `"`+exe+`" --tray`)
	} else {
		k.DeleteValue("Printrment")
	}
}

// ---------------------------------------------------------------- tray

var (
	trayMu    sync.Mutex
	trayReady bool
)

func trayAvailable() bool {
	trayMu.Lock()
	defer trayMu.Unlock()
	return trayReady
}

func runTray(onOpen, onQuit func()) {
	systray.Run(func() {
		systray.SetIcon(iconICO())
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

var (
	winmm         = windows.NewLazySystemDLL("winmm.dll")
	procPlaySound = winmm.NewProc("PlaySoundW")
	soundMu       sync.Mutex
	soundPlaying  []byte // kept alive while Windows plays it
)

func playSound(kind string) {
	wav := soundWAV(kind)
	if wav == nil {
		return
	}
	soundMu.Lock()
	defer soundMu.Unlock()
	soundPlaying = wav
	const sndAsync, sndNoDefault, sndMemory = 0x0001, 0x0002, 0x0004
	procPlaySound.Call(uintptr(unsafe.Pointer(&soundPlaying[0])), 0, sndAsync|sndNoDefault|sndMemory)
}
