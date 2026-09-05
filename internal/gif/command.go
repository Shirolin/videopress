package gif

import (
	"fmt"
)

// ScaleFilter 构造按宽度上限等比缩放的 scale 滤镜（偶数对齐）。
func scaleFilter(maxWidth int) string {
	if maxWidth <= 0 {
		return "scale='trunc(iw/2)*2':'trunc(ih/2)*2'"
	}
	return fmt.Sprintf(
		"scale='if(gt(iw,ih),-2,trunc(min(iw,%d)/2)*2)':'if(gt(iw,ih),trunc(min(ih,%d)/2)*2,-2)'",
		maxWidth, maxWidth,
	)
}

// BuildGIFArgs 构造 GIF 动图输出参数。
func BuildGIFArgs(inputPath string, outputPath string, t Tier) []string {
	scale := scaleFilter(t.MaxWidth)
	filter := fmt.Sprintf(
		"fps=%d,%s,split[s0][s1];[s0]palettegen=128[p];[s1][p]paletteuse=dither=bayer:diff_mode=rectangle",
		t.FPS, scale,
	)
	return []string{
		"-y", "-i", inputPath,
		"-t", t.MaxDuration,
		"-filter_complex", filter,
		"-loop", "0",
		outputPath,
	}
}

// BuildAPNGArgs 构造 APNG 动图输出参数（无损）。fps 为自适应计算后的目标帧率。
func BuildAPNGArgs(inputPath string, outputPath string, t Tier, fps int) []string {
	scale := scaleFilter(t.MaxWidth)
	filter := fmt.Sprintf("fps=%d,%s", fps, scale)
	return []string{
		"-y", "-i", inputPath,
		"-t", t.MaxDuration,
		"-vf", filter,
		"-plays", "0",
		"-pix_fmt", "rgba",
		"-compression_level", "6",
		"-preset", "none",
		outputPath,
	}
}

// BuildWebPArgs 构造动态 WebP 输出参数（有损，按档位质量）。
func BuildWebPArgs(inputPath string, outputPath string, t Tier) []string {
	scale := scaleFilter(t.MaxWidth)
	filter := fmt.Sprintf("fps=%d,%s", t.FPS, scale)
	// 档位越低，质量越低，压缩越狠。
	quality := 40
	switch t.Name {
	case "smooth":
		quality = 35
	case "balanced":
		quality = 50
	case "hd":
		quality = 65
	}
	return []string{
		"-y", "-i", inputPath,
		"-t", t.MaxDuration,
		"-vf", filter,
		"-loop", "0",
		"-c:v", "libwebp_anim",
		"-lossless", "0",
		"-quality", fmt.Sprintf("%d", quality),
		"-compression_level", "6",
		outputPath,
	}
}

// BuildArgs 根据格式分派到对应参数构造器。APNG 使用传入的 fps。
func BuildArgs(inputPath string, outputPath string, t Tier, format Format, apngFPS int) []string {
	switch format {
	case FormatGIF:
		return BuildGIFArgs(inputPath, outputPath, t)
	case FormatAPNG:
		if apngFPS < 1 {
			apngFPS = t.FPS
		}
		return BuildAPNGArgs(inputPath, outputPath, t, apngFPS)
	case FormatWebP:
		return BuildWebPArgs(inputPath, outputPath, t)
	default:
		return BuildGIFArgs(inputPath, outputPath, t)
	}
}