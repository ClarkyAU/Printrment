# Printrment

Watch and look after your Bambu Lab printer without Bambu's software: live camera, status, AMS,
files, alerts and timelapses. Built and tested with an X2D.

By Clarky3D, with Claude (Anthropic's AI assistant).

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
- Runs as a desktop app on Windows and Fedora, or as a server in Docker / TrueNAS

## Get it

| Platform | How |
|---|---|
| **Windows** | Download `Printrment.exe` from [Releases](../../releases) and run it, no installer needed |
| **Fedora Linux** | Download the `.rpm` from [Releases](../../releases), then `sudo dnf install ./printrment-*.rpm`: see [Fedora notes](docs/fedora.md) |
| **Docker / TrueNAS** | `ghcr.io/clarkyau/printrment:latest`: see [Docker and TrueNAS](docs/docker-truenas.md) |

## Documentation

- [Docker and TrueNAS](docs/docker-truenas.md): server mode, setup, settings
- [Fedora Linux](docs/fedora.md): install notes
- [Security](docs/security.md): how access, secrets and the printer connection are protected
- [Development](docs/development.md): building, source layout, testing with the mock printer
- [Licences](docs/licences.md): third-party components

## Repository

```
cmd/printrment/     the app (one Go package) and its bundled assets
deploy/             docker-compose.yml for Docker / TrueNAS
docs/               documentation
packaging/linux/    Fedora .rpm packaging
scripts/build.sh    builds the Windows exe, the Linux app and the .rpm
third_party/        exact sources of the libraries used
tools/mock-printer/ fake printer services for testing
Dockerfile          the server image (built by GitHub Actions)
```

© Clarky3D. All rights reserved. Third-party parts keep their own licences ([details](docs/licences.md)).
