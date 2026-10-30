#!/bin/sh
# Builds Printrment from the repository root:
#   build/Printrment.exe                         Windows x86-64
#   build/linux/printrment                       Linux x86-64 (static)
#   build/printrment-<version>-1.x86_64.rpm      Fedora package (needs nfpm)
# Needs Go 1.27.3 or newer. The Docker image is built from the Dockerfile.
set -e
cd "$(dirname "$0")/.."
VERSION=$(sed -n 's/^const version = "\(.*\)"/\1/p' cmd/printrment/server.go)
export GOFLAGS=-mod=mod GOPROXY=off CGO_ENABLED=0

mkdir -p build/linux/icons
GOOS=windows GOARCH=amd64 go build -trimpath -ldflags "-H windowsgui -s -w" -o build/Printrment.exe ./cmd/printrment
echo "Built build/Printrment.exe ($VERSION)"

GOOS=linux GOARCH=amd64 go build -trimpath -ldflags "-s -w" -o build/linux/printrment ./cmd/printrment
PRINTRMENT_ICON_DIR="$PWD/build/linux/icons" go test -run TestIconPNGs ./cmd/printrment >/dev/null
echo "Built build/linux/printrment ($VERSION)"

if command -v nfpm >/dev/null 2>&1; then
  PRINTRMENT_VERSION="$VERSION" nfpm package -f packaging/linux/nfpm.yaml -p rpm -t build/
  echo "Built $(ls build/printrment-*.rpm)"
else
  echo "nfpm not installed: skipped the .rpm (pip install nfpm)"
fi
