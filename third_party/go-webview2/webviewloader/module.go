package webviewloader

import (
	"fmt"
	"os"
	"sync"
	"unsafe"

	"golang.org/x/sys/windows"
)

// Printrment change: the embedded WebView2Loader.dll is unpacked to a private
// folder and loaded from there by its full path, like any normal program does.
// (The original code could also map the DLL into memory by hand with a custom
// PE loader, and first tried loading "WebView2Loader" by name, which searches
// the program's folder and PATH. Both are gone.)

var (
	loadOnce sync.Once
	loadErr  error
	dll      *windows.LazyDLL
)

func load() error {
	loadOnce.Do(func() {
		base, err := os.UserCacheDir() // %LOCALAPPDATA%
		if err != nil {
			base = os.TempDir()
		}
		path, err := extractLoader(base, WebView2Loader)
		if err != nil {
			loadErr = err
			return
		}
		dll = windows.NewLazyDLL(path)
		if err := dll.Load(); err != nil {
			loadErr = fmt.Errorf("unable to load WebView2Loader.dll: %w", err)
		}
	})
	return loadErr
}

func proc(name string) (*windows.LazyProc, error) {
	if err := load(); err != nil {
		return nil, err
	}
	p := dll.NewProc(name)
	if err := p.Find(); err != nil {
		return nil, err
	}
	return p, nil
}

// CompareBrowserVersions will compare the 2 given versions and return:
//
//	-1 = v1 < v2
//	 0 = v1 == v2
//	 1 = v1 > v2
func CompareBrowserVersions(v1 string, v2 string) (int, error) {
	_v1, err := windows.UTF16PtrFromString(v1)
	if err != nil {
		return 0, err
	}
	_v2, err := windows.UTF16PtrFromString(v2)
	if err != nil {
		return 0, err
	}
	p, err := proc("CompareBrowserVersions")
	if err != nil {
		return 0, err
	}
	var result int
	_, _, err = p.Call(uintptr(unsafe.Pointer(_v1)), uintptr(unsafe.Pointer(_v2)), uintptr(unsafe.Pointer(&result)))
	if err != windows.ERROR_SUCCESS {
		return result, err
	}
	return result, nil
}

// GetInstalledVersion returns the installed version of the webview2 runtime.
// If there is no version installed, a blank string is returned.
func GetInstalledVersion() (string, error) {
	// public STDAPI GetAvailableCoreWebView2BrowserVersionString(PCWSTR browserExecutableFolder, LPWSTR * versionInfo)
	p, err := proc("GetAvailableCoreWebView2BrowserVersionString")
	if err != nil {
		return "", err
	}
	var result *uint16
	hr, _, _ := p.Call(uintptr(unsafe.Pointer(nil)), uintptr(unsafe.Pointer(&result)))
	defer windows.CoTaskMemFree(unsafe.Pointer(result)) // Safe even if result is nil
	if hr != uintptr(windows.S_OK) {
		if hr&0xFFFF == uintptr(windows.ERROR_FILE_NOT_FOUND) {
			return "", nil // the runtime isn't installed
		}
		return "", fmt.Errorf("GetAvailableCoreWebView2BrowserVersionString returned HRESULT 0x%X", hr)
	}
	return windows.UTF16PtrToString(result), nil
}

// CreateCoreWebView2EnvironmentWithOptions loads WebView2Loader and calls the
// CreateCoreWebView2EnvironmentWithOptions routine.
func CreateCoreWebView2EnvironmentWithOptions(browserExecutableFolder, userDataFolder *uint16, environmentOptions uintptr, environmentCompletedHandle uintptr) (uintptr, error) {
	p, err := proc("CreateCoreWebView2EnvironmentWithOptions")
	if err != nil {
		return 0, err
	}
	res, _, _ := p.Call(
		uintptr(unsafe.Pointer(browserExecutableFolder)),
		uintptr(unsafe.Pointer(userDataFolder)),
		environmentOptions,
		environmentCompletedHandle,
	)
	return res, nil
}
