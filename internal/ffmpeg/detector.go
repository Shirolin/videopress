package ffmpeg

import (
	"os/exec"
	"sync"
)

var (
	cachedGPUEncoder string
	cacheMu          sync.RWMutex
)

// ResetGPUEncoderCache 重置 GPU 探测缓存。主要用于测试。
func ResetGPUEncoderCache() {
	cacheMu.Lock()
	cachedGPUEncoder = ""
	cacheMu.Unlock()
}

// DetectGPUEncoder 探测系统可用的硬件加速编码器。
// 探测顺序: h264_nvenc -> h264_qsv -> h264_amf.
// 如果都不支持，则返回 "libx264".
func DetectGPUEncoder(ffmpegPath string, runCmd func(name string, args []string) error) string {
	isTest := runCmd != nil

	// 如果非测试环境且已有缓存，直接使用缓存
	if !isTest {
		cacheMu.RLock()
		cached := cachedGPUEncoder
		cacheMu.RUnlock()
		if cached != "" {
			return cached
		}
	}

	if runCmd == nil {
		runCmd = func(name string, args []string) error {
			cmd := exec.Command(name, args...)
			prepareCmd(cmd)
			return cmd.Run()
		}
	}

	encoders := []string{"h264_nvenc", "h264_qsv", "h264_amf"}
	var detected string = "libx264"
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
			detected = enc
			break
		}
	}

	// 缓存探测结果
	if !isTest {
		cacheMu.Lock()
		cachedGPUEncoder = detected
		cacheMu.Unlock()
	}

	return detected
}
