package ffmpeg

import (
	"os/exec"
)

// DetectGPUEncoder 探测系统可用的硬件加速编码器。
// 探测顺序: h264_nvenc -> h264_qsv -> h264_amf.
// 如果都不支持，则返回 "libx264".
func DetectGPUEncoder(ffmpegPath string, runCmd func(name string, args []string) error) string {
	if runCmd == nil {
		runCmd = func(name string, args []string) error {
			cmd := exec.Command(name, args...)
			prepareCmd(cmd)
			return cmd.Run()
		}
	}

	encoders := []string{"h264_nvenc", "h264_qsv", "h264_amf"}
	for _, enc := range encoders {
		// 运行一个超轻量级的测试命令，仅生成 1 帧视频进行编码
		args := []string{
			"-f", "lavfi",
			"-i", "color=c=black:s=16x16",
			"-vframes", "1",
			"-c:v", enc,
			"-f", "null",
			"-",
		}
		if err := runCmd(ffmpegPath, args); err == nil {
			return enc
		}
	}

	return "libx264"
}
