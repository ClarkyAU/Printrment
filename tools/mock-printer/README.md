# Mock printer

Services that behave like a Bambu Lab printer, for testing Printrment without one:

- `mqtt/`: status feed on :8883 (env: `FUN` Developer Mode bits, `VERIFYFAIL`, `SCENARIO=idle`,
  `TRAYNOW`, `EXTPETG`, ...)
- `ftp/`: storage on :990, serving `root/`
- `camera/rtsp_mock.py`: camera on :322 (needs an H.264 file `t.h264`)

Each needs a self-signed certificate in its folder (`c.pem` / `k.pem`), not committed:

```sh
openssl req -x509 -newkey rsa:2048 -nodes -days 3650 -subj /CN=mock -keyout k.pem -out c.pem
```

Run the app against them with environment `PRINTRMENT_FFMPEG=$(which ffmpeg)` (system FFmpeg) and add
a printer at `127.0.0.1` with any 8-digit code.

Build the Go mocks with `go build` inside `mqtt/` and `ftp/`.
