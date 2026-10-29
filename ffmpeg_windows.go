//go:build windows

package main

import _ "embed"

// FFmpeg 7.1 for Windows (GPL v3 build with x264), from imageio-ffmpeg 0.7.0.
//
//go:embed ffmpeg.exe.gz
var ffmpegGz []byte

// SHA-256 of the unpacked file (checked by TestFFmpegHash, so it can't drift).
const ffmpegSHA256 = "2ce797a0f88d7f067180338fb227f7b1928ea727bd9a4d7a1d022f7c52af71a3"

const ffmpegExt = ".exe"
