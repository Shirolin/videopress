package ffmpeg

import (
	"fmt"
	"strings"

	"videopress/internal/compress"
)

func BuildArgs(inputPath string, outputPath string, preset compress.Preset, hwEncoder string, copyAudio bool, maxFPS int, audioMode string, crfOverride int) []string {
	var scaleFilter string
	if preset.MaxDimension > 0 {
		scaleFilter = fmt.Sprintf(
			"scale='if(gt(iw,ih),trunc(min(iw*%.2f,%d)/2)*2,-2)':'if(gt(iw,ih),-2,trunc(min(ih*%.2f,%d)/2)*2)'",
			preset.ScaleFactor, preset.MaxDimension,
			preset.ScaleFactor, preset.MaxDimension,
		)
	} else {
		scaleFilter = fmt.Sprintf(
			"scale='if(gt(iw,ih),trunc(iw*%.2f/2)*2,-2)':'if(gt(iw,ih),-2,trunc(ih*%.2f/2)*2)'",
			preset.ScaleFactor,
			preset.ScaleFactor,
		)
	}

	args := []string{
		"-hide_banner",
		"-y",
		"-i", inputPath,
		"-vf", scaleFilter,
	}

	if maxFPS > 0 {
		args = append(args, "-r", fmt.Sprintf("%d", maxFPS))
	}

	crf := preset.CRF
	if crfOverride > 0 {
		crf = crfOverride
	}

	isSoft := hwEncoder == "" || hwEncoder == "libx264" || hwEncoder == "libx265" || hwEncoder == "libsvtav1"
	if isSoft {
		codec := hwEncoder
		if codec == "" {
			codec = "libx264"
		}
		args = append(args, "-c:v", codec)
		if codec == "libx264" || codec == "libx265" {
			args = append(args,
				"-preset", preset.Preset,
				"-crf", fmt.Sprintf("%d", crf),
			)
		} else if codec == "libsvtav1" {
			svtPreset := "6"
			switch preset.Preset {
			case "veryfast":
				svtPreset = "8"
			case "faster":
				svtPreset = "6"
			case "medium":
				svtPreset = "4"
			}
			args = append(args,
				"-preset", svtPreset,
				"-crf", fmt.Sprintf("%d", crf),
			)
		}
	} else {
		args = append(args, "-c:v", hwEncoder)
		if strings.HasSuffix(hwEncoder, "_nvenc") {
			args = append(args, "-rc", "constqp", "-qp", fmt.Sprintf("%d", crf))
		} else if strings.HasSuffix(hwEncoder, "_qsv") {
			args = append(args, "-global_quality", fmt.Sprintf("%d", crf))
		} else if strings.HasSuffix(hwEncoder, "_amf") {
			args = append(args, "-rc", "cqp", "-qp_i", fmt.Sprintf("%d", crf), "-qp_p", fmt.Sprintf("%d", crf))
		}
	}

	useCopy := audioMode == "copy" || (audioMode == "" && copyAudio)
	if audioMode == "mute" {
		args = append(args, "-an")
	} else if useCopy {
		args = append(args, "-c:a", "copy")
	} else {
		args = append(args,
			"-c:a", "aac",
			"-b:a", preset.AudioBitrate,
		)
	}

	args = append(args,
		"-movflags", "+faststart",
		outputPath,
	)

	return args
}
