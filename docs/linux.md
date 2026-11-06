# Linux

Printrment for Linux is a single **AppImage** file that runs on any x86-64 desktop distribution
(Fedora, Ubuntu, Debian, Arch, openSUSE...). Nothing is installed and no root is needed.

1. Download `Printrment-<version>-x86_64.AppImage` from [Releases](../../releases).
2. Make it executable and run it:

   ```sh
   chmod +x Printrment-*-x86_64.AppImage
   ./Printrment-*-x86_64.AppImage
   ```

   (Most file managers also do this from the file's Properties > Permissions > "Allow executing".)

Notes:

- **FUSE:** AppImages normally need FUSE 2. Fedora has it (`fuse-libs`); on Ubuntu 22.04+ install
  `libfuse2` (`libfuse2t64` on 24.04). Without it, run
  `./Printrment-*.AppImage --appimage-extract-and-run`, which works anywhere.
- **Window:** opens as an app window if Chrome, Chromium, Edge or Brave is installed (including
  Flatpak); otherwise as a tab in your default browser (Firefox).
- **Tray icon:** KDE shows it. GNOME needs the AppIndicator extension (on Fedora:
  `sudo dnf install gnome-shell-extension-appindicator`, then sign out and in).
- **Start when I sign in** (Settings > App) starts the `.AppImage` file from where it is now, so
  keep it in one place (for example `~/Applications`). If you move it, switch the setting off and on.
- **Remote access:** if other devices can't connect, your firewall may be blocking the port. On
  Fedora and similar: `sudo firewall-cmd --permanent --add-port=8788/tcp && sudo firewall-cmd --reload`
- **Data:** settings in `~/.config/Printrment`, the encryption key for saved access codes in
  `~/.local/share/Printrment/secret.key` (owner-only), FFmpeg and cache in `~/.cache/Printrment`.
- **Timelapses** are off until you turn on "Record every print automatically" (they use disk space).
  Videos go to `~/Videos/Printrment`, pictures to `~/Pictures/Printrment`.
- **Updating:** download the new AppImage and replace the old file. **Removing:** delete the file
  (your settings stay in your home folder).
