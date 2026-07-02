package ffmpeg

import (
	"testing"
)

func TestParseFFmpegError(t *testing.T) {
	tests := []struct {
		stderr string
		expect string
	}{
		{
			stderr: "[libx264 @ 0000021c17ee1240] width not divisible by 2 (1001x500)\nError initializing output stream 0:0 -- Error while opening encoder for output stream #0:0 - maybe incorrect parameters such as bit_rate, rate, width or height",
			expect: "视频分辨率的宽或高无法被 2 整除，编码器拒绝工作",
		},
		{
			stderr: "Unknown encoder 'h264_nvenc_typo'",
			expect: "指定的视频/音频编码器在当前 FFmpeg 中未找到，请更新您的 FFmpeg",
		},
		{
			stderr: "C:\\videos\\nonexistent.mp4: No such file or directory",
			expect: "找不到指定的文件或目录",
		},
		{
			stderr: "Some random line\n[some_module] Error: Invalid argument passed\nAnother line",
			expect: "[some_module] Error: Invalid argument passed",
		},
		{
			stderr: "Plain normal warning lines\nNo keywords here",
			expect: "FFmpeg 执行异常",
		},
	}

	for _, tt := range tests {
		result := ParseFFmpegError(tt.stderr)
		if result != tt.expect {
			t.Errorf("for stderr %q, expected %q, got %q", tt.stderr, tt.expect, result)
		}
	}
}
