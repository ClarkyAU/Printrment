# Development

## Building

Requires **Go 1.27.3** (or newer). Everything else is in this repo (`third_party/` holds the
exact library sources used, including a patched go-webview2).

`scripts/build.sh` (run from anywhere) builds:

- `build/Printrment.exe` (Windows x86-64)
- `build/linux/printrment` (Linux x86-64, static)
- `build/printrment-<version>-1.x86_64.rpm` if [nfpm](https://nfpm.goreleaser.com) is installed
  (`pip install nfpm`); packaging settings are in `packaging/linux/`.

The Docker image is built from the `Dockerfile` (by GitHub Actions; `docker build .` works locally too).

Tests (run on Linux or Windows):

```sh
GOFLAGS=-mod=mod GOPROXY=off go test ./cmd/printrment
```

`go.mod` keeps the `go 1.24` language level on purpose, so TLS behaviour towards the printer
stays as tested. Raising it tightens some TLS defaults; test against a real printer first.

## Releasing

1. Set the new version in `cmd/printrment/server.go` (`const version`) and push to `main`.
   GitHub Actions publishes the Docker image (`latest` and the version).
2. Tag it and push the tag:

   ```sh
   git tag v0.32 && git push origin v0.32
   ```

   GitHub Actions tests, builds `Printrment.exe` and the Fedora `.rpm`, and publishes them as a
   Release with their SHA-256 checksums.

## Source layout

All the app code is one Go package in `cmd/printrment/`:

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
| `assets/` | Bundled FFmpeg for each system (`ffmpeg.exe.gz`, `ffmpeg-linux.gz`) and the fonts |

## Testing without a printer

`tools/mock-printer/` has services that behave like a Bambu printer, so the whole app can be
tested without one. See [its README](../tools/mock-printer/README.md).
