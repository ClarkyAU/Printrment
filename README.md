# Printrment

An app for watching and looking after a Bambu Lab printer (built and tested against an X2D),
for **Windows** (one exe) and **Fedora Linux** x86-64 (an .rpm). No Bambu software, Python or
other runtime needed.

By Clarky3D, with Claude (Anthropic's AI assistant). Version is in `server.go` (`const version`).

## Features

- Live camera (printer's RTSPS stream, decoded by the bundled FFmpeg)
- Print status, progress, stage, temperatures, fans, printer info, firmware update alerts
- AMS / AMS HT / external spools, with the feeding slot worked out per nozzle on dual-nozzle
  printers (same rules as the Bambu Monitor Stream Dock plugin)
- Printer file browser with plate previews and print start (shown only when the printer accepts it)
- Controls: pause / resume / stop / speed / chamber light (anything the printer won't accept
  without Developer Mode is hidden)
- Alerts with sounds, finish and failure photos, job summaries, timelapses, snapshots
- Remote access: a web version for other devices on the home network, PIN protected
- Three themes, compact and full-screen modes, tray icon, multiple printers, settings backup

## Installing on Fedora

```sh
sudo dnf install ./printrment-<version>-1.x86_64.rpm
```

Then start **Printrment** from the app menu (or run `printrment`). Notes:

- **Window:** opens as an app window if Chrome, Chromium, Edge or Brave is installed (including
  Flatpak); otherwise as a tab in your default browser (Firefox).
- **Tray icon:** KDE shows it. Fedora's GNOME needs the AppIndicator extension:
  `sudo dnf install gnome-shell-extension-appindicator`, then sign out and in.
- **Remote access:** Fedora Workstation's firewall normally allows it already. If not:
  `sudo firewall-cmd --permanent --add-port=8788/tcp && sudo firewall-cmd --reload`
- **Data:** settings in `~/.config/Printrment`, the encryption key for saved access codes in
  `~/.local/share/Printrment/secret.key` (owner-only), FFmpeg and cache in `~/.cache/Printrment`.
- **Remove:** `sudo dnf remove printrment` (your settings stay in your home folder).

## Building

Requires **Go 1.27.3** (or newer). Everything else is in this repo (`third_party/` holds the
exact library sources used, including a patched go-webview2).

`./build.sh` (from any OS) builds:

- `Printrment.exe` (Windows x86-64)
- `build/linux/printrment` (Linux x86-64, static)
- `build/printrment-<version>-1.x86_64.rpm` if [nfpm](https://nfpm.goreleaser.com) is installed
  (`pip install nfpm`); packaging settings are in `packaging/linux/`.

Tests (run on Linux or Windows):

```sh
GOFLAGS=-mod=mod GOPROXY=off go test ./...
```

`go.mod` keeps the `go 1.24` language level on purpose, so TLS behaviour towards the printer
stays as tested. Raising it tightens some TLS defaults; test against a real printer first.

## Layout

| File | What it does |
|---|---|
| `main.go` | Start-up, single-instance hand-over, FFmpeg unpack + hash check, tray, quit |
| `server.go` | Local web server, API, remote access server and its PIN login |
| `page.go` | The whole UI (one HTML/CSS/JS string) |
| `security.go` | Session token, printer certificate pinning, PIN lockout, path checks |
| `seal_windows.go` | Encrypts access codes / PIN with Windows DPAPI |
| `seal_linux.go` | Same on Linux: AES-256-GCM with an owner-only key file kept apart from the settings |
| `config.go`, `migrate.go` | Settings storage; migration from the old "Bambu Cam" name |
| `mqtt.go` | Printer status/commands over MQTT-TLS; AMS parsing; Developer Mode rules |
| `printerinfo.go`, `check.go` | Printer info card; Settings > Printers > Check |
| `rtsp.go`, `camera.go` | Camera: RTSPS client (port 322) and the older JPEG camera (port 6000) |
| `ftp.go`, `files.go` | Printer storage over FTPS; file browser and print start |
| `timelapse.go`, `jobs.go` | Timelapses, job summaries, photos, settings export/import |
| `events.go`, `sound.go` | Alerts and sounds |
| `window_windows.go`, `os_windows.go` | WebView2 window, tray, autostart, Windows helpers |
| `os_linux.go` | Linux: browser app window, tray (StatusNotifier), autostart, sounds, file manager |
| `ffmpeg_windows.go`, `ffmpeg_linux.go` | Which bundled FFmpeg each build carries, and its pinned SHA-256 |
| `ffmpeg.exe.gz`, `ffmpeg-linux.gz` | Bundled FFmpeg for each system (see Licences) |
| `fonts/` | Exo 2, Space Mono, Silkscreen (SIL OFL) |
| `devtools/` | Mock printer services for testing (see below) |

## Testing without a printer

`devtools/` has mock services that behave like the printer:

- `mockmqtt`: status feed on :8883 (env: `FUN` Developer Mode bits, `VERIFYFAIL`, `SCENARIO=idle`,
  `TRAYNOW`, `EXTPETG`, ...)
- `mockftp`: storage on :990, serving `root/`
- `rtspmock/rtsp_mock.py`: camera on :322 (needs an H.264 file `t.h264`)

Each needs a self-signed certificate in its folder (`c.pem` / `k.pem`), not committed:

```sh
openssl req -x509 -newkey rsa:2048 -nodes -days 3650 -subj /CN=mock -keyout k.pem -out c.pem
```

Run the app against them with environment `BAMBUCAM_FFMPEG=$(which ffmpeg)` (system FFmpeg) and add
a printer at `127.0.0.1` with any 8-digit code.

## Security model (summary)

- Local UI and API need a per-run secret token; requests with the wrong Host header are refused.
- Printer certificates are remembered on first connect and any change is refused (with an alert).
- Access codes and the PIN are encrypted: Windows DPAPI, or on Linux AES-256-GCM with an
  owner-only key stored apart from the settings.
- Remote access only answers private/home-network addresses, needs a PIN (with lockout), and
  never exposes printer files, settings or (unless allowed) controls.
- FFmpeg is verified by SHA-256 before every run and limited to pipe/file inputs.
- Commands the printer refuses without Developer Mode are never sent.

## Licences

Printrment's own code: © Clarky3D. All rights reserved (private repository).

Third-party components keep their own licences (see each folder in `third_party/`):
go-webview2 (MIT), go-winloader (ISC), fyne.io/systray (Apache 2.0), godbus/dbus (BSD 2-clause),
golang.org/x/sys (BSD 3-clause). Fonts: SIL Open Font Licence 1.1.

`ffmpeg.exe.gz` is FFmpeg 7.1 and `ffmpeg-linux.gz` is FFmpeg 7.0.2 (static build by John
Van Sickle), both GPL v3 builds including x264 (GPL v2+), taken from the `imageio-ffmpeg` 0.7.0
package. If you distribute Printrment, you must also offer the
corresponding FFmpeg/x264 source (ffmpeg.org, videolan.org).
