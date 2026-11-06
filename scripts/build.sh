#!/bin/sh
# Builds Printrment from the repository root:
#   build/Printrment.exe                         Windows x86-64
#   build/linux/printrment                       Linux x86-64 (static)
#   build/Printrment-<version>-x86_64.AppImage   Linux, any distribution
# Needs Go 1.27.3 or newer. The Docker image is built from the Dockerfile.
set -e
cd "$(dirname "$0")/.."
VERSION=$(sed -n 's/^const version = "\(.*\)"/\1/p' cmd/printrment/server.go)
export GOFLAGS=-mod=mod GOPROXY=off CGO_ENABLED=0

mkdir -p build/linux/icons
PRINTRMENT_ICON_DIR="$PWD/build/linux/icons" go test -run TestIconPNGs ./cmd/printrment >/dev/null
# icon, version information and manifest for the Windows exe (see scripts/mkrsrc.py)
python3 scripts/mkrsrc.py "$VERSION" build/linux/icons cmd/printrment/rsrc_windows_amd64.syso >/dev/null
GOOS=windows GOARCH=amd64 go build -trimpath -ldflags "-H windowsgui -s -w" -o build/Printrment.exe ./cmd/printrment
echo "Built build/Printrment.exe ($VERSION)"

GOOS=linux GOARCH=amd64 go build -trimpath -ldflags "-s -w" -o build/linux/printrment ./cmd/printrment
echo "Built build/linux/printrment ($VERSION)"

# --- AppImage ---------------------------------------------------------------
# One file that runs on any x86-64 Linux desktop. The two tools that make it are
# downloaded once and checked against these SHA-256 values before they're used.
APPIMAGETOOL_URL=https://github.com/AppImage/appimagetool/releases/download/1.9.1/appimagetool-x86_64.AppImage
APPIMAGETOOL_SHA=ed4ce84f0d9caff66f50bcca6ff6f35aae54ce8135408b3fa33abfc3cb384eb0
RUNTIME_URL=https://github.com/AppImage/type2-runtime/releases/download/continuous/runtime-x86_64
RUNTIME_SHA=617aa77314d57c8b8e4fae9c3f0b35af7b30d5529cb69e98f64f92510ed98a46

fetch() { # url sha file
  if [ ! -f "$3" ] || [ "$(sha256sum "$3" | cut -d' ' -f1)" != "$2" ]; then
    curl -fsSL -o "$3" "$1"
  fi
  [ "$(sha256sum "$3" | cut -d' ' -f1)" = "$2" ] || { echo "Checksum mismatch for $3: not using it"; rm -f "$3"; exit 1; }
}
mkdir -p build/tools
fetch "$APPIMAGETOOL_URL" "$APPIMAGETOOL_SHA" build/tools/appimagetool
fetch "$RUNTIME_URL" "$RUNTIME_SHA" build/tools/runtime
chmod +x build/tools/appimagetool

APPDIR=build/Printrment.AppDir
rm -rf "$APPDIR"
mkdir -p "$APPDIR/usr/bin" "$APPDIR/usr/share/icons/hicolor/256x256/apps"
cp build/linux/printrment "$APPDIR/usr/bin/printrment"
cp packaging/linux/printrment.desktop "$APPDIR/printrment.desktop"
cp build/linux/icons/printrment-256.png "$APPDIR/printrment.png"
cp build/linux/icons/printrment-256.png "$APPDIR/usr/share/icons/hicolor/256x256/apps/printrment.png"
ln -s usr/bin/printrment "$APPDIR/AppRun"
rm -f build/Printrment-*-x86_64.AppImage
# --appimage-extract-and-run: appimagetool is itself an AppImage, and this way
# the build doesn't need FUSE.
ARCH=x86_64 build/tools/appimagetool --appimage-extract-and-run --no-appstream \
  --runtime-file build/tools/runtime "$APPDIR" "build/Printrment-$VERSION-x86_64.AppImage" >/dev/null
echo "Built build/Printrment-$VERSION-x86_64.AppImage"
