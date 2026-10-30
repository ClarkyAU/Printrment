# Fedora Linux

Download the `.rpm` and install it:

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
