package main

import (
	"context"
	"flag"
	"fmt"
	"image/png"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/ELY-Open-Source/MujicaUI/core"
	"github.com/egoist/mygo/plugins/render"
	"github.com/egoist/mygo/ui"
)

func main() {
	outPath := flag.String("out", filepath.Join("dist", "promo", "videopress-promo.mp4"), "output video path (.mp4/.mov/.webm) or PNG path when -still >= 0")
	stillFrame := flag.Int("still", -1, "render a single frame index (0..TotalFrames-1) to a PNG image instead of full video")
	rendererName := flag.String("renderer", "auto", "frame drawer: auto, gpu, or cpu")
	encoderName := flag.String("encoder", "system", "video encoder: system (Windows Media Foundation) or ffmpeg")
	ffmpegBin := flag.String("ffmpeg", "", "path to ffmpeg executable when -encoder=ffmpeg (defaults to PATH)")
	codecName := flag.String("codec", "h264", "video codec: h264 or hevc")
	noAudio := flag.Bool("no-audio", false, "render video without soundtrack")
	repoRoot := flag.String("root", "", "optional repository root path for locating docs/screenshots")
	flag.Parse()

	assets, err := LoadAssets(*repoRoot)
	if err != nil {
		log.Fatalf("load assets: %v", err)
	}

	var soundtrack *render.Sound
	if !*noAudio && *stillFrame < 0 {
		soundtrack = BuildSoundtrack()
	}

	comp := buildComposition(assets, soundtrack)
	opts, err := buildOptions(*rendererName, *encoderName, *codecName, *ffmpegBin)
	if err != nil {
		log.Fatal(err)
	}

	if *stillFrame >= 0 {
		if *stillFrame >= TotalFrames {
			log.Fatalf("invalid -still frame %d: must be in [0, %d]", *stillFrame, TotalFrames-1)
		}
		target := *outPath
		if strings.EqualFold(filepath.Ext(target), ".mp4") || target == "" {
			target = filepath.Join("dist", "promo", fmt.Sprintf("frame-%04d.png", *stillFrame))
		}
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			log.Fatalf("mkdir output dir: %v", err)
		}
		start := time.Now()
		img, err := render.Still(comp, *stillFrame, opts)
		if err != nil {
			log.Fatalf("render still frame %d: %v", *stillFrame, err)
		}
		f, err := os.Create(target)
		if err != nil {
			log.Fatalf("create %s: %v", target, err)
		}
		if err := png.Encode(f, img); err != nil {
			_ = f.Close()
			log.Fatalf("encode png %s: %v", target, err)
		}
		if err := f.Close(); err != nil {
			log.Fatalf("close %s: %v", target, err)
		}
		abs, _ := filepath.Abs(target)
		fmt.Printf("still frame %d -> %s (%v)\n", *stillFrame, abs, time.Since(start).Round(time.Millisecond))
		return
	}

	if err := os.MkdirAll(filepath.Dir(*outPath), 0o755); err != nil {
		log.Fatalf("mkdir output dir: %v", err)
	}

	opts.Progress = func(p render.Progress) {
		if p.Done%FPS == 0 || p.Done == p.Frames {
			rate := 0.0
			if secs := p.Elapsed.Seconds(); secs > 0 {
				rate = float64(p.Done) / secs
			}
			fmt.Printf("\r[%s] %d/%d frames (%.0f%%, %.1f fps)", p.Renderer, p.Done, p.Frames, 100*float64(p.Done)/float64(p.Frames), rate)
		}
	}

	start := time.Now()
	if err := render.Video(context.Background(), comp, *outPath, opts); err != nil {
		fmt.Println()
		log.Fatalf("render video: %v", err)
	}
	fmt.Println()
	abs, _ := filepath.Abs(*outPath)
	fmt.Printf("rendered %s (%dx%d @ %dfps, %d frames, %.1fs) in %v\n",
		abs, Width, Height, FPS, TotalFrames, TotalDurationSeconds(), time.Since(start).Round(time.Millisecond))
}

func buildComposition(assets *Assets, soundtrack *render.Sound) render.Composition {
	darkBg := DarkPalette().Background
	return render.Composition{
		Width:      Width,
		Height:     Height,
		FPS:        FPS,
		Frames:     TotalFrames,
		Scale:      Scale,
		Dark:       true,
		Background: darkBg,
		Locale:     "zh-Hans",
		View: func(c *ui.Context, f render.Frame) {
			core.Use(c, core.Settings{Mode: core.Dark, Locale: core.ZhCN})
			if soundtrack != nil {
				f.Audio(soundtrack)
			}
			BuildPart1(c, f, assets)
			BuildPart2(c, f, assets)
		},
	}
}

func buildOptions(rendererName, encoderName, codecName, ffmpegBin string) (render.Options, error) {
	var renderer render.Renderer
	switch strings.ToLower(rendererName) {
	case "", "auto":
		renderer = render.Auto
	case "gpu":
		renderer = render.GPU
	case "cpu":
		renderer = render.CPU
	default:
		return render.Options{}, fmt.Errorf("unknown -renderer %q (use auto, gpu, or cpu)", rendererName)
	}

	var encoder render.Encoder
	switch strings.ToLower(encoderName) {
	case "", "system":
		encoder = render.System
	case "ffmpeg":
		encoder = render.FFmpeg
	default:
		return render.Options{}, fmt.Errorf("unknown -encoder %q (use system or ffmpeg)", encoderName)
	}

	var codec render.Codec
	switch strings.ToLower(codecName) {
	case "", "h264":
		codec = render.H264
	case "hevc", "h265":
		codec = render.HEVC
	default:
		return render.Options{}, fmt.Errorf("unknown -codec %q (use h264 or hevc)", codecName)
	}

	return render.Options{
		Renderer:     renderer,
		Encoder:      encoder,
		Codec:        codec,
		FFmpeg:       ffmpegBin,
		AudioBitrate: 192_000,
	}, nil
}
