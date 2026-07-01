package ffmpeg

import (
	"strings"
	"testing"

	"videopress/internal/compress"
)

func TestBuildArgsForStandardPreset(t *testing.T) {
	preset, err := compress.PresetByName("standard")
	if err != nil {
		t.Fatalf("PresetByName returned error: %v", err)
	}

	args := BuildArgs(`C:\videos\input.mov`, `C:\videos\compressed\output.mp4`, preset)
	joined := strings.Join(args, " ")

	required := []string{
		"-i C:\\videos\\input.mov",
		"-c:v libx264",
		"-preset faster",
		"-crf 27",
		"-c:a aac",
		"-b:a 96k",
		"-movflags +faststart",
		"C:\\videos\\compressed\\output.mp4",
	}
	for _, part := range required {
		if !strings.Contains(joined, part) {
			t.Fatalf("expected args to contain %q, got %s", part, joined)
		}
	}

	if !strings.Contains(joined, "scale='if(gt(iw,ih),min(1080,iw),-2)':'if(gt(iw,ih),-2,min(1080,ih))'") {
		t.Fatalf("expected scale filter for 1080 max dimension, got %s", joined)
	}
}
