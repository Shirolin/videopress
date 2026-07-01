package ffmpeg

import (
	"fmt"

	"videopress/internal/compress"
)

func BuildArgs(inputPath string, outputPath string, preset compress.Preset) []string {
	scaleFilter := fmt.Sprintf(
		"scale='if(gt(iw,ih),min(%d,iw),-2)':'if(gt(iw,ih),-2,min(%d,ih))'",
		preset.MaxDimension,
		preset.MaxDimension,
	)

	return []string{
		"-hide_banner",
		"-y",
		"-i", inputPath,
		"-vf", scaleFilter,
		"-c:v", "libx264",
		"-preset", preset.Preset,
		"-crf", fmt.Sprintf("%d", preset.CRF),
		"-c:a", "aac",
		"-b:a", preset.AudioBitrate,
		"-movflags", "+faststart",
		outputPath,
	}
}
