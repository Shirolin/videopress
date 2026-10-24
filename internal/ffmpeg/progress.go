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

	"videopress/internal/util"
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

// VideoInfo 汇总一次 ffmpeg -i 探测得到的时长与分辨率。
type VideoInfo struct {
	Duration time.Duration
	Width    int
	Height   int
}

// Probe 运行 ffmpeg -i 探测视频时长与分辨率（单进程，避免逐项各起一次 ffmpeg）。
func Probe(ffmpegPath string, inputPath string) (VideoInfo, error) {
	cmd := exec.Command(ffmpegPath, "-i", inputPath)
	util.HideConsoleWindow(cmd)
	output, _ := cmd.CombinedOutput()
	metadata := string(output)
	duration, durationErr := ParseDuration(metadata)
	w, h, dimErr := ParseDimensions(metadata)
	if durationErr != nil && dimErr != nil {
		return VideoInfo{}, fmt.Errorf("could not parse video metadata: %w", durationErr)
	}
	return VideoInfo{Duration: duration, Width: w, Height: h}, nil
}

var dimensionsRegex = regexp.MustCompile(`, (\d{2,5})x(\d{2,5})[, ]`)

// parseDimensionsFromLine 在单行元数据中查找 "…, WxH, …" 尺寸片段。
func parseDimensionsFromLine(line string) (int, int, bool) {
	matches := dimensionsRegex.FindStringSubmatch(line)
	if matches == nil {
		return 0, 0, false
	}
	w, _ := strconv.Atoi(matches[1])
	h, _ := strconv.Atoi(matches[2])
	if w <= 0 || h <= 0 {
		return 0, 0, false
	}
	return w, h, true
}

// ParseDimensions 从 ffmpeg 元数据输出中解析视频分辨率。
// 优先在普通（非 attached_pic 封面）视频流行内匹配，避免取到封面流的尺寸。
func ParseDimensions(metadata string) (int, int, error) {
	for _, line := range strings.Split(metadata, "\n") {
		if !strings.Contains(line, "Video:") || strings.Contains(line, "attached_pic") {
			continue
		}
		if w, h, ok := parseDimensionsFromLine(line); ok {
			return w, h, nil
		}
	}
	matches := dimensionsRegex.FindStringSubmatch(metadata)
	if matches == nil {
		return 0, 0, fmt.Errorf("could not find dimensions in metadata")
	}
	w, _ := strconv.Atoi(matches[1])
	h, _ := strconv.Atoi(matches[2])
	if w <= 0 || h <= 0 {
		return 0, 0, fmt.Errorf("invalid dimensions in metadata: %dx%d", w, h)
	}
	return w, h, nil
}

// EvenFloor 向下取整到偶数，最小为 2，避免 yuv420p 非偶数尺寸问题（crop/scale 均使用）。
func EvenFloor(v int) int {
	if v < 2 {
		return 2
	}
	return v / 2 * 2
}

// ComputeCenterCrop 计算居中比例裁剪的 crop=W:H:X:Y：
// 在 iw×ih 内取 RatioW:RatioH 的最大内接矩形，宽高同乘 scale 后仍居中，宽高落到偶数。
func ComputeCenterCrop(iw, ih, ratioW, ratioH int, scale float64) (int, int, int, int) {
	var rectW, rectH float64
	if float64(iw)*float64(ratioH) > float64(ih)*float64(ratioW) {
		rectW = float64(ih) * float64(ratioW) / float64(ratioH)
		rectH = float64(ih)
	} else {
		rectW = float64(iw)
		rectH = float64(iw) * float64(ratioH) / float64(ratioW)
	}
	cw := EvenFloor(int(rectW*float64(scale) + 0.5))
	ch := EvenFloor(int(rectH*float64(scale) + 0.5))
	if cw > iw {
		cw = EvenFloor(iw)
	}
	if ch > ih {
		ch = EvenFloor(ih)
	}
	x := (iw - cw) / 2
	y := (ih - ch) / 2
	return cw, ch, x, y
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
