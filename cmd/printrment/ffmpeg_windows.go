//go:build windows

package main

import (
	_ "embed"
	"io"
	"strings"
)

// FFmpeg 7.1 for Windows (GPL v3 build with x264), from imageio-ffmpeg 0.7.0.
//
// Embedded as-is, not compressed: a big block of compressed data inside an exe
// looks like a packed program to virus scanners. assets/ffmpeg.exe is made from
// the committed assets/ffmpeg.exe.gz by scripts/build.sh. (A string, so it sits
// in the read-only part of the program.)
//
//go:embed assets/ffmpeg.exe
var ffmpegData string

func ffmpegSource() (io.Reader, error) { return strings.NewReader(ffmpegData), nil }

// SHA-256 of the unpacked file (checked by TestFFmpegHash, so it can't drift).
const ffmpegSHA256 = "2ce797a0f88d7f067180338fb227f7b1928ea727bd9a4d7a1d022f7c52af71a3"

const ffmpegExt = ".exe"
