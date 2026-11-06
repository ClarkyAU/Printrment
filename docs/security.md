# Security

## Desktop app (Windows, Linux)

- Local UI and API need a per-run secret token; requests with the wrong Host header are refused.
- Printer certificates are remembered on first connect and any change is refused (with an alert).
- Access codes and the PIN are encrypted: Windows DPAPI, or on Linux AES-256-GCM with an
  owner-only key stored apart from the settings.
- Remote access only answers private/home-network addresses, needs a PIN (with lockout), and
  never exposes printer files, settings or (unless allowed) controls.
- FFmpeg is verified by SHA-256 before every run and limited to pipe/file inputs.
- Commands the printer refuses without Developer Mode are never sent.
- Every page is served with a Content-Security-Policy (own origin only, no framing, no plugins),
  `nosniff` and no-referrer headers.
- Replies from the printer are read with size limits (lines, headers, messages, downloads), so a
  misbehaving device on the network can't use up the app's memory.

## Server mode (Docker / TrueNAS)

- Answers home-network (private) addresses only. Don't publish it to the internet or through a
  tunnel; use a VPN such as Tailscale for access from outside.
- Admin password stored as a PBKDF2-SHA256 hash (600,000 rounds); a generated first password is
  shown once in the log and deleted from disk when changed.
- Sign-in cookie is signed, HttpOnly, SameSite=Strict and expires after 30 days; changing the
  password signs everyone out. Five wrong passwords lock that device out for 15 minutes.
- Every API call needs Printrment's own header, so other websites can't drive it.
- Plain HTTP by default (fine on a trusted home network); `PRINTRMENT_HTTPS=1` or your own
  certificate turns on HTTPS with Secure cookies. See [Docker and TrueNAS](docker-truenas.md).
- Desktop-only actions (quit, window, file manager) don't exist on the server.
- The image contains only the app (no shell or OS) and runs as an unprivileged user.

## Reporting a problem

Open an issue in this repository (it's private) or contact Clarky3D.
