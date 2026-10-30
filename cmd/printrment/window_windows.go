//go:build windows

package main

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"unsafe"

	webview2 "github.com/jchv/go-webview2"
	"github.com/jchv/go-webview2/webviewloader"
	"golang.org/x/sys/windows"
)

// Printrment's own window: a WebView2 view inside a frameless window. The page
// draws its own title bar (themed), and closing it shuts the app down
// properly. If WebView2 isn't installed, the Edge app window is used instead.

var (
	ole32               = windows.NewLazySystemDLL("ole32.dll")
	procCoInitializeEx  = ole32.NewProc("CoInitializeEx")
	dwmapi              = windows.NewLazySystemDLL("dwmapi.dll")
	procDwmExtendFrame  = dwmapi.NewProc("DwmExtendFrameIntoClientArea")
	procSetWindowLongPt = user32.NewProc("SetWindowLongPtrW")
	procCallWindowProc  = user32.NewProc("CallWindowProcW")
	procReleaseCapture  = user32.NewProc("ReleaseCapture")
	procPostMessage     = user32.NewProc("PostMessageW")
	procSendMessage     = user32.NewProc("SendMessageW")
	procIsZoomed        = user32.NewProc("IsZoomed")
	procIsIconic        = user32.NewProc("IsIconic")
	procShowWindowAsync = user32.NewProc("ShowWindowAsync")
	procGetDpiForWindow = user32.NewProc("GetDpiForWindow")
	procGetDpiForSystem = user32.NewProc("GetDpiForSystem")
	procGetSysMetricDpi = user32.NewProc("GetSystemMetricsForDpi")
	procSetDpiAwareCtx  = user32.NewProc("SetProcessDpiAwarenessContext")
	procCreateIconRes   = user32.NewProc("CreateIconFromResourceEx")
	procDestroyWindow   = user32.NewProc("DestroyWindow")
)

const (
	wmNCCalcSize    = 0x0083
	wmNCLButtonDown = 0x00A1
	wmSetIcon       = 0x0080
	htCaption       = 2
	swHide          = 0
	swMaximize      = 3
	swMinimize      = 6
	swShow          = 5
	swRestore       = 9
)

func init() {
	// Sharp text on high-DPI screens (must happen before any window exists).
	procSetDpiAwareCtx.Call(^uintptr(3)) // DPI_AWARENESS_CONTEXT_PER_MONITOR_AWARE_V2 (-4)
}

var (
	nativeMu      sync.Mutex
	nativeHWND    uintptr
	nativeOpening bool
	oldWndProc    uintptr
	framelessCB   = windows.NewCallback(framelessProc)
)

func nativeWindow() uintptr {
	nativeMu.Lock()
	defer nativeMu.Unlock()
	return nativeHWND
}

func webview2Installed() bool {
	v, err := webviewloader.GetInstalledVersion()
	return err == nil && v != ""
}

func systemDPI() int {
	if procGetDpiForSystem.Find() == nil {
		if d, _, _ := procGetDpiForSystem.Call(); d > 0 {
			return int(d)
		}
	}
	return 96
}

func windowDPI(h uintptr) int {
	if procGetDpiForWindow.Find() == nil {
		if d, _, _ := procGetDpiForWindow.Call(h); d > 0 {
			return int(d)
		}
	}
	return systemDPI()
}

func sysMetric(idx, dpi int) int32 {
	if procGetSysMetricDpi.Find() == nil {
		v, _, _ := procGetSysMetricDpi.Call(uintptr(idx), uintptr(dpi))
		return int32(v)
	}
	v, _, _ := procGetSystemMetrics.Call(uintptr(idx))
	return int32(v)
}

// openWindow shows the app window (or brings it to the front).
func openWindow(url string) {
	nativeMu.Lock()
	if h := nativeHWND; h != 0 {
		nativeMu.Unlock()
		focusWindow(h)
		return
	}
	if nativeOpening {
		nativeMu.Unlock()
		return
	}
	nativeOpening = true
	nativeMu.Unlock()
	if !webview2Installed() {
		nativeMu.Lock()
		nativeOpening = false
		nativeMu.Unlock()
		appendLog("WebView2 not installed - using an Edge window")
		openBrowserWindow(url)
		return
	}
	go runNativeWindow(url)
}

func focusWindow(h uintptr) {
	if r, _, _ := procIsIconic.Call(h); r != 0 {
		procShowWindowAsync.Call(h, swRestore)
	} else {
		procShowWindowAsync.Call(h, swShow)
	}
	procSetForegroundWindow.Call(h)
}

var themeBG = map[string]string{"default": "FF0B0D10", "webstore": "FF121212", "makerworld": "FFF4F5F7"}

func runNativeWindow(url string) {
	runtime.LockOSThread() // a window lives on the thread that made it
	procCoInitializeEx.Call(0, 2)

	// no white flash while the page loads
	os.Setenv("WEBVIEW2_DEFAULT_BACKGROUND_COLOR", themeBG[cfg.get().Theme])
	dpi := systemDPI()
	sc := func(v int) int { return v * dpi / 96 }

	w := webview2.NewWithOptions(webview2.WebViewOptions{
		DataPath:  filepath.Join(localDir(), "WebView2"),
		AutoFocus: true,
		WindowOptions: webview2.WindowOptions{
			Title: "Printrment", Width: uint(sc(1280)), Height: uint(sc(800)), Center: true,
		},
	})
	if w == nil {
		nativeMu.Lock()
		nativeOpening = false
		nativeMu.Unlock()
		appendLog("couldn't create the WebView2 window - using an Edge window")
		openBrowserWindow(url)
		return
	}
	hwnd := uintptr(w.Window())
	w.SetSize(sc(640), sc(400), webview2.HintMin)
	setWindowIcons(hwnd)
	makeFrameless(hwnd)

	w.Bind("bcWin", func(action string) map[string]any {
		return windowAction(hwnd, action)
	})
	// The token is handed straight to our own page, never over the network.
	// Only set on Printrment's own address, in case anything else ever loads.
	w.Init(`if(location.origin==="http://` + listenAddr + `"){window.bcNative=true;window.bcToken="` + sessionToken + `";}`)
	if i := strings.IndexByte(url, '#'); i >= 0 {
		url = url[:i]
	}
	w.Navigate(url)

	nativeMu.Lock()
	nativeHWND = hwnd
	nativeOpening = false
	nativeMu.Unlock()

	w.Run() // until the window is closed

	nativeMu.Lock()
	nativeHWND = 0
	nativeMu.Unlock()
	if !(cfg.get().Tray && trayAvailable()) {
		go quitApp()
	}
}

// windowAction is called from the page's own title bar.
func windowAction(hwnd uintptr, a string) map[string]any {
	hit := map[string]uintptr{"l": 10, "r": 11, "t": 12, "tl": 13, "tr": 14, "b": 15, "bl": 16, "br": 17}
	switch a {
	case "drag":
		procReleaseCapture.Call()
		procPostMessage.Call(hwnd, wmNCLButtonDown, htCaption, 0)
	case "min":
		procShowWindowAsync.Call(hwnd, swMinimize)
	case "max":
		if z, _, _ := procIsZoomed.Call(hwnd); z != 0 {
			procShowWindowAsync.Call(hwnd, swRestore)
		} else {
			procShowWindowAsync.Call(hwnd, swMaximize)
		}
	case "fs_on":
		enterFullscreen(hwnd)
	case "fs_off":
		exitFullscreen(hwnd)
	case "close":
		if cfg.get().Tray && trayAvailable() {
			procDestroyWindow.Call(hwnd) // hide to tray: just close the window
		} else {
			procShowWindowAsync.Call(hwnd, swHide) // disappear at once, then shut down cleanly
			go quitApp()
		}
	default:
		if ht, ok := hit[a]; ok {
			procReleaseCapture.Call()
			procPostMessage.Call(hwnd, wmNCLButtonDown, ht, 0)
		}
	}
	z, _, _ := procIsZoomed.Call(hwnd)
	nativeMu.Lock()
	fs := fsOn
	nativeMu.Unlock()
	return map[string]any{"max": z != 0, "fs": fs}
}

// ---- true fullscreen: the window covers the whole monitor, taskbar included

var (
	procMonitorFromWin = user32.NewProc("MonitorFromWindow")
	procGetMonitorInfo = user32.NewProc("GetMonitorInfoW")
	fsOn               bool
	fsSaved            rect
	fsWasMax           bool
)

type monitorInfo struct {
	Size    uint32
	Monitor rect
	Work    rect
	Flags   uint32
}

func enterFullscreen(hwnd uintptr) {
	nativeMu.Lock()
	if fsOn {
		nativeMu.Unlock()
		return
	}
	nativeMu.Unlock()
	z, _, _ := procIsZoomed.Call(hwnd)
	wasMax := z != 0
	if wasMax {
		procShowWindow.Call(hwnd, swRestore)
	}
	var r rect
	procGetWindowRect.Call(hwnd, uintptr(unsafe.Pointer(&r)))
	// Coming from compact view: leave it in the same step (so the two don't
	// fight), and come back to the normal-size window afterwards.
	winMu.Lock()
	if savedRect != nil {
		r = *savedRect
		savedRect = nil
	}
	winMu.Unlock()
	procSetWindowPos.Call(hwnd, ^uintptr(1), 0, 0, 0, 0, 0x1|0x2) // not topmost
	mon, _, _ := procMonitorFromWin.Call(hwnd, 2)                 // nearest monitor
	mi := monitorInfo{}
	mi.Size = uint32(unsafe.Sizeof(mi))
	if ok, _, _ := procGetMonitorInfo.Call(mon, uintptr(unsafe.Pointer(&mi))); ok == 0 {
		return
	}
	nativeMu.Lock()
	fsOn, fsSaved, fsWasMax = true, r, wasMax
	nativeMu.Unlock()
	m := mi.Monitor
	const swpShow, swpFrameChanged, swpNoOwnerZ = 0x40, 0x20, 0x200
	procSetWindowPos.Call(hwnd, 0, uintptr(m.L), uintptr(m.T), uintptr(m.R-m.L), uintptr(m.B-m.T), swpShow|swpFrameChanged|swpNoOwnerZ)
}

func exitFullscreen(hwnd uintptr) {
	nativeMu.Lock()
	if !fsOn {
		nativeMu.Unlock()
		return
	}
	fsOn = false
	s, wasMax := fsSaved, fsWasMax
	nativeMu.Unlock()
	const swpShow, swpFrameChanged, swpNoOwnerZ = 0x40, 0x20, 0x200
	procSetWindowPos.Call(hwnd, 0, uintptr(s.L), uintptr(s.T), uintptr(s.R-s.L), uintptr(s.B-s.T), swpShow|swpFrameChanged|swpNoOwnerZ)
	if wasMax {
		procShowWindow.Call(hwnd, swMaximize)
	}
}

type margins struct{ L, R, T, B int32 }

func makeFrameless(hwnd uintptr) {
	nativeMu.Lock()
	oldWndProc, _, _ = procSetWindowLongPt.Call(hwnd, ^uintptr(3), framelessCB) // GWLP_WNDPROC (-4)
	nativeMu.Unlock()
	// keep Windows' drop shadow and rounded corners
	m := margins{0, 0, 1, 0}
	procDwmExtendFrame.Call(hwnd, uintptr(unsafe.Pointer(&m)))
	const swpNoSize, swpNoMove, swpNoZOrder, swpFrameChanged = 0x1, 0x2, 0x4, 0x20
	procSetWindowPos.Call(hwnd, 0, 0, 0, 0, 0, swpNoSize|swpNoMove|swpNoZOrder|swpFrameChanged)
}

// framelessProc removes the standard title bar and border: the whole window
// is the page. Everything else goes to the normal window procedure, so
// snapping, minimise/maximise animations and the shadow still work.
func framelessProc(hwnd, msg, wp, lp uintptr) uintptr {
	if msg == wmNCCalcSize && wp != 0 {
		if z, _, _ := procIsZoomed.Call(hwnd); z != 0 {
			// a maximised window overhangs the screen by its frame; pull it in
			dpi := windowDPI(hwnd)
			pad := sysMetric(92, dpi) // SM_CXPADDEDBORDER
			fx, fy := sysMetric(32, dpi)+pad, sysMetric(33, dpi)+pad
			r := *(**rect)(unsafe.Pointer(&lp)) // lParam points at the window rectangle
			r.L += fx
			r.T += fy
			r.R -= fx
			r.B -= fy
		}
		return 0
	}
	nativeMu.Lock()
	old := oldWndProc
	nativeMu.Unlock()
	r, _, _ := procCallWindowProc.Call(old, hwnd, msg, wp, lp)
	return r
}

func setWindowIcons(hwnd uintptr) {
	for _, ic := range []struct{ size, which int }{{16, 0}, {64, 1}} {
		png := iconPNG(ic.size)
		h, _, _ := procCreateIconRes.Call(uintptr(unsafe.Pointer(&png[0])), uintptr(len(png)), 1, 0x00030000,
			uintptr(ic.size), uintptr(ic.size), 0)
		if h != 0 {
			procSendMessage.Call(hwnd, wmSetIcon, uintptr(ic.which), h)
		}
	}
}

// hideAppWindow is used on quit so the window vanishes immediately.
func hideAppWindow() {
	if h := nativeWindow(); h != 0 {
		procShowWindowAsync.Call(h, swHide)
	}
}
