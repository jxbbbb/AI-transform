package ffmpeg

import "runtime"

var FFmpeg = "pkg/ffmpeg/windows/bin/ffmpeg.exe"

func init() {
	if runtime.GOOS == "windows" {
		FFmpeg = "pkg/ffmpeg/windows/bin/ffmpeg.exe"
	} else {
		FFmpeg = "pkg/ffmpeg/linux/ffmpeg"
	}
}
