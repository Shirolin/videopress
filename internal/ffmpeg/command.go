package ffmpeg

import (
	"fmt"

	"videopress/internal/compress"
)

func BuildArgs(inputPath string, outputPath string, preset compress.Preset, hwEncoder string, copyAudio bool) []string {
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

	if hwEncoder == "" || hwEncoder == "libx264" {
		args = append(args,
			"-c:v", "libx264",
			"-preset", preset.Preset,
			"-crf", fmt.Sprintf("%d", preset.CRF),
		)
	} else {
		args = append(args, "-c:v", hwEncoder)
		switch hwEncoder {
		case "h264_nvenc":
			args = append(args, "-rc", "constqp", "-qp", fmt.Sprintf("%d", preset.CRF))
		case "h264_qsv":
			args = append(args, "-global_quality", fmt.Sprintf("%d", preset.CRF))
		case "h264_amf":
			args = append(args, "-rc", "cqp", "-qp_i", fmt.Sprintf("%d", preset.CRF), "-qp_p", fmt.Sprintf("%d", preset.CRF))
		default:
			// 兜底硬解参数，如果是不认识的硬解，不传特定参数以防报错，由 ffmpeg 使用默认值
		}
	}

	if copyAudio {
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
