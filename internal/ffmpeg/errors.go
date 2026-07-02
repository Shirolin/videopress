package ffmpeg

import (
	"strings"
)

// ParseFFmpegError 从 stderr 输出中提取友好的核心错误提示
func ParseFFmpegError(stderr string) string {
	if strings.Contains(stderr, "width not divisible by 2") || strings.Contains(stderr, "height not divisible by 2") {
		return "视频分辨率的宽或高无法被 2 整除，编码器拒绝工作"
	}
	if strings.Contains(stderr, "Unknown encoder") {
		return "指定的视频/音频编码器在当前 FFmpeg 中未找到，请更新您的 FFmpeg"
	}
	if strings.Contains(stderr, "Unrecognized option") {
		return "未被识别的命令行选项，可能是 FFmpeg 版本过低"
	}
	if strings.Contains(stderr, "Error opening input") {
		return "打开输入文件失败，源视频可能已损坏或格式不受支持"
	}
	if strings.Contains(stderr, "Permission denied") {
		return "访问被拒绝，请检查文件是否被占用或缺少读取权限"
	}
	if strings.Contains(stderr, "No such file") {
		return "找不到指定的文件或目录"
	}
	if strings.Contains(stderr, "No space left on device") {
		return "磁盘空间不足，无法保存压缩输出"
	}

	// 找不到特征词，返回最后一行非空错误信息
	lines := strings.Split(stderr, "\n")
	for i := len(lines) - 1; i >= 0; i-- {
		line := strings.TrimSpace(lines[i])
		if line != "" && (strings.Contains(line, "Error") || strings.Contains(line, "Failed") || strings.Contains(line, "Invalid") || strings.Contains(line, "error")) {
			return line
		}
	}
	return "FFmpeg 执行异常"
}
