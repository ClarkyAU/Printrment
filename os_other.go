//go:build !windows

package main

import (
	"os"
	"os/exec"
)

func hideWindow(cmd *exec.Cmd) {}

func openWindow(url string) {
	if os.Getenv("BROWSER") == "true" {
		return
	}
	exec.Command("xdg-open", url).Start()
}

func openInBrowser(url string) {
	if os.Getenv("BROWSER") == "true" {
		appendLog("open in browser: " + url)
		return
	}
	exec.Command("xdg-open", url).Start()
}

func openFolder(dir string)             {}
func showInFolder(path string)          {}
func setWindowMode(onTop, compact bool) {}
func flashWindow()                      {}
func setAutostart(on bool)              {}
func trayAvailable() bool               { return false }
func runTray(onOpen, onQuit func())     {}
func quitTray()                         {}
func hideAppWindow()                    {}
func setTrayTooltip(s string)           {}

// playSound writes the sound to a file when BAMBUCAM_SOUNDLOG is set (tests).
func playSound(kind string) {
	if d := os.Getenv("BAMBUCAM_SOUNDLOG"); d != "" {
		f, err := os.OpenFile(d, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
		if err == nil {
			f.WriteString(kind + "\n")
			f.Close()
		}
		os.WriteFile(d+"."+kind+".wav", soundWAV(kind), 0o644)
	}
}

var _ = iconICO
