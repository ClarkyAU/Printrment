# Licences

Printrment's own code: © Clarky3D. All rights reserved (private repository).

Third-party components keep their own licences (see each folder in `third_party/`):
go-webview2 (MIT; with Microsoft's WebView2Loader, BSD 3-clause), fyne.io/systray (Apache 2.0), godbus/dbus (BSD 2-clause),
golang.org/x/sys (BSD 3-clause). Fonts: SIL Open Font Licence 1.1.

`cmd/printrment/assets/ffmpeg.exe.gz` is FFmpeg 7.1 and `ffmpeg-linux.gz` is FFmpeg 7.0.2 (static build by John
Van Sickle), both GPL v3 builds including x264 (GPL v2+), taken from the `imageio-ffmpeg` 0.7.0
package. If you distribute Printrment, you must also offer the
corresponding FFmpeg/x264 source.

Source code for these builds:

- FFmpeg 7.0.2: <https://ffmpeg.org/releases/ffmpeg-7.0.2.tar.xz>; FFmpeg 7.1: <https://ffmpeg.org/releases/ffmpeg-7.1.tar.xz>
- x264: <https://code.videolan.org/videolan/x264>
- The Linux build's own build scripts and sources: <https://johnvansickle.com/ffmpeg/> (source tarballs are listed there)

`ffmpeg -version` on the bundled program shows the exact build configuration. Printrment only runs FFmpeg
on data piped to it (the printer's video and its own pictures), never on network addresses or other files.

**Known limitation:** these builds are older than the newest FFmpeg release. Newer prebuilt builds
are much larger and, on Linux, need system libraries the container image doesn't have, so they
aren't used. Updating FFmpeg means building a small static copy; that is on the to-do list.
