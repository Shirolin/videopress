package ffmpeg

import (
	"bufio"
	"fmt"
	"io"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var durationRegex = regexp.MustCompile(`Duration:\s*(\d{2}):(\d{2}):(\d{2})\.(\d+)`)

// ParseDuration 从 ffmpeg 元数据输出中解析视频时长
func ParseDuration(metadata string) (time.Duration, error) {
	matches := durationRegex.FindStringSubmatch(metadata)
	if len(matches) != 5 {
		return 0, fmt.Errorf("could not find duration in metadata")
	}

	hours, _ := strconv.Atoi(matches[1])
	minutes, _ := strconv.Atoi(matches[2])
	seconds, _ := strconv.Atoi(matches[3])

	// 小数秒解析：例如 123 表示 0.123 秒
	fractionStr := "0." + matches[4]
	var fractionVal float64
	if val, err := strconv.ParseFloat(fractionStr, 64); err == nil {
		fractionVal = val
	}

	total := time.Duration(hours)*time.Hour +
		time.Duration(minutes)*time.Minute +
		time.Duration(seconds)*time.Second +
		time.Duration(fractionVal*float64(time.Second))

	return total, nil
}

// GetDuration 运行 ffmpeg -i 探测视频时长
func GetDuration(ffmpegPath string, inputPath string) (time.Duration, error) {
	cmd := exec.Command(ffmpegPath, "-i", inputPath)
	prepareCmd(cmd)
	output, _ := cmd.CombinedOutput()
	return ParseDuration(string(output))
}

// TrackProgress 读取 ffmpeg -progress 输出流并回调进度百分比
func TrackProgress(r io.Reader, duration time.Duration, onProgress func(percent float64)) {
	scanner := bufio.NewScanner(r)
	var outTimeMs int64
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "out_time_ms=") {
			valStr := strings.TrimPrefix(line, "out_time_ms=")
			if val, err := strconv.ParseInt(valStr, 10, 64); err == nil {
				outTimeMs = val
				if duration > 0 {
					percent := float64(outTimeMs) / 1000000.0 / duration.Seconds() * 100.0
					if percent > 100 {
						percent = 100
					}
					if percent < 0 {
						percent = 0
					}
					onProgress(percent)
				}
			}
		}
	}
}

// RenderProgressBar 在输出端渲染动态刷新进度条
func RenderProgressBar(w io.Writer, percent float64, prefix string) {
	width := 25
	completed := int(percent / 100.0 * float64(width))
	if completed > width {
		completed = width
	}
	if completed < 0 {
		completed = 0
	}
	remaining := width - completed

	displayPrefix := prefix
	// 针对中英文字符长度，简单的截断
	runes := []rune(prefix)
	if len(runes) > 18 {
		displayPrefix = string(runes[:15]) + "..."
	}

	bar := "\033[32m" + strings.Repeat("█", completed) + "\033[90m" + strings.Repeat("░", remaining) + "\033[0m"
	// \r 回到行首，输出对齐的进度条
	fmt.Fprintf(w, "\r%-20s %s \033[36m%5.1f%%\033[0m", displayPrefix, bar, percent)
}
