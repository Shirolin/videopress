package main

import (
	"fmt"
	"image/png"
	"os"
	"path/filepath"
	"testing"

	"github.com/ELY-Open-Source/MujicaUI/core"
	"github.com/egoist/mygo/plugins/render"
	"github.com/egoist/mygo/ui"
)

func TestBuildPart1RenderKeyframes(t *testing.T) {
	assets, err := LoadAssets("")
	if err != nil {
		t.Fatalf("LoadAssets: %v", err)
	}

	comp := render.Composition{
		Width:      Width,
		Height:     Height,
		FPS:        FPS,
		Frames:     Scene3Start + Scene3Duration,
		Scale:      Scale,
		Dark:       true,
		Background: DarkPalette().Background,
		Locale:     "zh-Hans",
		View: func(c *ui.Context, f render.Frame) {
			core.Use(c, core.Settings{Mode: core.Dark, Locale: core.ZhCN})
			BuildPart1(c, f, assets)
		},
	}

	dumpDir := os.Getenv("PROMO_DUMP_DIR")
	if dumpDir != "" {
		if err := os.MkdirAll(dumpDir, 0o755); err != nil {
			t.Fatalf("mkdir dump dir: %v", err)
		}
	}

	keyframes := []struct {
		name  string
		frame int
	}{
		{"act1-reveal", 96},
		{"act2-workspace", 165},
		{"act2-batch-swap", 210},
		{"act3-workers-mid", 300},
		{"act3-workers-done", 355},
	}

	for _, kf := range keyframes {
		t.Run(kf.name, func(t *testing.T) {
			img, err := render.Still(comp, kf.frame, render.Options{Renderer: render.CPU})
			if err != nil {
				t.Fatalf("render.Still(frame=%d): %v", kf.frame, err)
			}
			if b := img.Bounds(); b.Dx() != Width || b.Dy() != Height {
				t.Fatalf("unexpected image bounds %v, want %dx%d", b, Width, Height)
			}
			if dumpDir != "" {
				out := filepath.Join(dumpDir, fmt.Sprintf("%s-f%04d.png", kf.name, kf.frame))
				f, err := os.Create(out)
				if err != nil {
					t.Fatalf("create %s: %v", out, err)
				}
				if err := png.Encode(f, img); err != nil {
					_ = f.Close()
					t.Fatalf("encode %s: %v", out, err)
				}
				if err := f.Close(); err != nil {
					t.Fatalf("close %s: %v", out, err)
				}
			}
		})
	}
}
