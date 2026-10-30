# Printrment

An app for watching and looking after a Bambu Lab printer (built and tested against an X2D),
for **Windows** (one exe), **Fedora Linux** x86-64 (an .rpm) and **Docker / TrueNAS** (a server
you open in a browser). No Bambu software, Python or other runtime needed.

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

## Docker and TrueNAS (server mode)

GitHub Actions builds an image on every push to `main`: `ghcr.io/clarkyau/printrment:latest`
(and `:<version>`). It contains only the app (no shell or OS), runs as user 568 (TrueNAS's
`apps` user) and keeps everything in `/data`.

In server mode Printrment has no window: you open `http://<server>:8787` in any browser on your
home network and sign in with the **admin password**. Everything else works as in the desktop app,
including the PIN-protected remote access viewer (port 8788). Alert sounds play in the browser tab.

### TrueNAS SCALE (24.10 or newer)

1. Create a dataset for it, e.g. `tank/apps/printrment`, and give the `apps` user (568) access.
2. Apps > Discover Apps > ⋮ > **Install via YAML**, paste [`docker-compose.yml`](docker-compose.yml),
   change the volume path and `TZ`, and deploy.
3. Open the app's **Logs**: the first start prints the admin password (also saved as
   `initial-admin-password.txt` in the dataset until you change it).
4. Browse to `http://<truenas>:8787`, sign in, add your printer, and change the password under
   Settings > App.

### Plain Docker

```sh
docker compose up -d            # with docker-compose.yml from this repo
docker logs printrment          # shows the first admin password
```

### Pulling the image

The repository is private, so the image is private too. Either log the server in to `ghcr.io`
with a GitHub token that has `read:packages`, or make just the package public (GitHub > your
profile > Packages > printrment > Package settings > Change visibility). A public image exposes the
compiled app, not the source code.

### Settings (environment variables)

| Variable | Default | What it does |
|---|---|---|
| `PRINTRMENT_ADMIN_PASSWORD` | (generated) | Fixes the admin password (10+ characters); it then can't be changed in the app |
| `PRINTRMENT_PORT` | `8787` | Port for the app |
| `PRINTRMENT_KEY_FILE` | `/data/key/secret.key` | Encryption key for saved access codes; point it at a separate mount to keep it apart from the data |
| `PRINTRMENT_SECURE_COOKIES` | off | Set to `1` when serving through an HTTPS reverse proxy |
| `TZ` | `UTC` | Time zone for times and file names |

### Notes

- **Networking:** host networking (as in the compose file) lets Printrment find printers and shows
  the right viewer address. With bridge networking, publish 8787 (and 8788 for the viewer) and add
  printers by IP address.
- **Home network only:** the app answers private-network addresses only. Don't publish port 8787
  to the internet or through a tunnel; use a VPN such as Tailscale for access from outside.
- **Security:** admin password stored as a PBKDF2-SHA256 hash; sign-in cookie is signed,
  HttpOnly, SameSite=Strict and expires after 30 days; changing the password signs everyone out;
  wrong passwords lock out after 5 tries; every API call needs Printrment's own header, so other
  websites can't drive it; desktop-only actions (quit, window, file manager) don't exist on the server.
- **One printer, two apps:** running the desktop app and the server at the same time works, but both
  will connect to the printer's camera.

## Building

Requires **Go 1.27.3** (or newer). Everything else is in this repo (`third_party/` holds the
exact library sources used, including a patched go-webview2).

`./build.sh` (from any OS) builds:

- `Printrment.exe` (Windows x86-64)
- `build/linux/printrment` (Linux x86-64, static)
- `build/printrment-<version>-1.x86_64.rpm` if [nfpm](https://nfpm.goreleaser.com) is installed
  (`pip install nfpm`); packaging settings are in `packaging/linux/`.

The Docker image is built from the `Dockerfile` (by GitHub Actions; `docker build .` works locally too).

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
| `server_mode.go` | Server / Docker mode: admin sign-in, browser sounds, file viewing, health check |
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
