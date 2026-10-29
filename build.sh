#!/bin/sh
# Builds Printrment for Windows (Printrment.exe) and Fedora/Linux x86-64
# (build/linux/printrment, plus an .rpm if nfpm is installed). Needs Go 1.27.3+.
set -e
cd "$(dirname "$0")"
VERSION=$(sed -n 's/^const version = "\(.*\)"/\1/p' server.go)
export GOFLAGS=-mod=mod GOPROXY=off CGO_ENABLED=0

GOOS=windows GOARCH=amd64 go build -trimpath -ldflags "-H windowsgui -s -w" -o Printrment.exe .
echo "Built Printrment.exe ($VERSION)"

mkdir -p build/linux/icons
GOOS=linux GOARCH=amd64 go build -trimpath -ldflags "-s -w" -o build/linux/printrment .
PRINTRMENT_ICON_DIR="$PWD/build/linux/icons" CGO_ENABLED=0 go test -run TestIconPNGs . >/dev/null
echo "Built build/linux/printrment ($VERSION)"

if command -v nfpm >/dev/null 2>&1; then
  PRINTRMENT_VERSION="$VERSION" nfpm package -f packaging/linux/nfpm.yaml -p rpm -t build/
  echo "Built $(ls build/printrment-*.rpm)"
else
  echo "nfpm not installed: skipped the .rpm (pip install nfpm)"
fi
