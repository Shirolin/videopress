package main

import (
	"testing"

	"github.com/egoist/mygo/plugins/render"
	"github.com/egoist/mygo/ui"
)

func TestSliceRunesUTF8Safe(t *testing.T) {
	const sample = "videopress --preset standard 中文"
	if got := sliceRunes(sample, -0.5); got != "" {
		t.Fatalf("expected empty at negative progress, got %q", got)
	}
	if got := sliceRunes(sample, 1.5); got != sample {
		t.Fatalf("expected full string at progress > 1, got %q", got)
	}
	mid := sliceRunes(sample, 0.5)
	if len([]rune(mid)) != len([]rune(sample))/2 {
		t.Fatalf("unexpected rune count at 0.5: %d", len([]rune(mid)))
	}
}

func TestPart2KeyframesRenderWithoutPanic(t *testing.T) {
	assets, err := LoadAssets("")
	if err != nil {
		t.Fatalf("LoadAssets: %v", err)
	}
	comp := render.Composition{
		Width:      Width,
		Height:     Height,
		FPS:        FPS,
		Frames:     TotalFrames,
		Scale:      Scale,
		Dark:       true,
		Background: DarkPalette().Background,
		View: func(c *ui.Context, f render.Frame) {
			BuildPart2(c, f, assets)
		},
	}

	keyframes := []int{
		Scene4Start + 10,
		CueCropLock,
		Scene5Start + 25,
		CueCLISummary,
		Scene6Start + 36,
		CueFinaleLock,
		TotalFrames - 1,
	}
	for _, fr := range keyframes {
		img, err := render.Still(comp, fr, render.Options{Renderer: render.CPU})
		if err != nil {
			t.Fatalf("render still frame %d: %v", fr, err)
		}
		if img == nil || img.Bounds().Dx() != Width || img.Bounds().Dy() != Height {
			t.Fatalf("frame %d unexpected bounds: %v", fr, img)
		}
	}
}
