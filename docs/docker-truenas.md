# Docker and TrueNAS

GitHub Actions builds an image on every push to `main`: `ghcr.io/clarkyau/printrment:latest`
(and `:<version>`). It contains only the app (no shell or OS), runs as user 568 (TrueNAS's
`apps` user) and keeps everything in `/data`.

In server mode Printrment has no window: you open `http://<server>:8787` in any browser on your
home network and sign in with the **admin password**. Everything else works as in the desktop app,
including the PIN-protected remote access viewer (port 8788). Alert sounds play in the browser tab.

### TrueNAS SCALE (24.10 or newer)

1. Create a dataset for it, e.g. `tank/apps/printrment`, and give the `apps` user (568) access.
2. Apps > Discover Apps > ⋮ > **Install via YAML**, paste [`deploy/docker-compose.yml`](../deploy/docker-compose.yml),
   change the volume path and `TZ`, and deploy.
3. Open the app's **Logs**: the first start prints the admin password (also saved as
   `initial-admin-password.txt` in the dataset until you change it).
4. Browse to `http://<truenas>:8787`, sign in, add your printer, and change the password under
   Settings > App.

### Plain Docker

```sh
docker compose -f deploy/docker-compose.yml up -d
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
