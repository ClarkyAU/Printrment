module github.com/ClarkyAU/printrment

go 1.24

require (
	fyne.io/systray v1.11.0
	github.com/godbus/dbus/v5 v5.1.0
	github.com/jchv/go-webview2 v0.0.0-00010101000000-000000000000
	golang.org/x/sys v0.30.0
)

require github.com/jchv/go-winloader v0.0.0-20250406163304-c1995be93bd1 // indirect

replace (
	fyne.io/systray => ./third_party/systray
	github.com/godbus/dbus/v5 => ./third_party/dbus
	github.com/jchv/go-webview2 => ./third_party/go-webview2
	github.com/jchv/go-winloader => ./third_party/go-winloader
	golang.org/x/sys => ./third_party/sys
)
