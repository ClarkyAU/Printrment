//go:build linux

package main

import _ "embed"

// FFmpeg 7.0.2 for Linux x86-64 (static GPL v3 build with x264, by John Van
// Sickle), from imageio-ffmpeg 0.7.0.
//
//go:embed ffmpeg-linux.gz
var ffmpegGz []byte

// SHA-256 of the unpacked file (checked by TestFFmpegHash, so it can't drift).
const ffmpegSHA256 = "e7e7fb30477f717e6f55f9180a70386c62677ef8a4d4d1a5d948f4098aa3eb99"

const ffmpegExt = ""
