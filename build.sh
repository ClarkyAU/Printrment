#!/bin/sh
# Builds Printrment.exe for Windows (64-bit). Needs Go 1.27.3+.
set -e
cd "$(dirname "$0")"
GOOS=windows GOARCH=amd64 CGO_ENABLED=0 GOFLAGS=-mod=mod GOPROXY=off \
  go build -trimpath -ldflags "-H windowsgui -s -w" -o Printrment.exe .
echo "Built Printrment.exe"
