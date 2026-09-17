package engine

import (
	"context"
	"errors"
	"os"
	"strings"
	"testing"
	"time"
)

func TestEngineRunSuccess(t *testing.T) {
	var runArgs [][]string
	deps := Dependencies{
		ExecutableDir: `C:\tools`,
		ResolveBinary: func(dir string) (string, error) {
			return `C:\ffmpeg\ffmpeg.exe`, nil
		},
		RunCommand: func(name string, args []string) error {
			runArgs = append(runArgs, args)
			return nil
		},
		GetDuration: func(ffmpegPath string, inputPath string) (time.Duration, error) {
			return 10 * time.Second, nil
		},
		DetectGPUEncoder: func(ffmpegPath string, codec string, runCmd func(string, []string) error) string {
			return "libx264"
		},
		MkdirAll: func(path string, perm os.FileMode) error {
			return nil
		},
		PathExists: func(path string) bool {
			return false
		},
		InputAccessible: func(path string) bool {
			return true
		},
	}

	eng := NewCompressEngine(deps)

	reports, err := eng.Run(context.Background(), JobRequest{
		Files:       []string{`C:\videos\test.mp4`},
		Preset:      "standard",
		Concurrency: 1,
	}, nil)

	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	if len(reports) != 1 {
		t.Fatalf("expected 1 report, got %d", len(reports))
	}

	if reports[0].Status != "success" {
		t.Fatalf("expected status success, got %s", reports[0].Status)
	}

	if len(runArgs) != 1 {
		t.Fatalf("expected 1 run command, got %d", len(runArgs))
	}
}

func TestEngineRunHandlesCancellation(t *testing.T) {
	deps := Dependencies{
		ExecutableDir: `C:\tools`,
		ResolveBinary: func(dir string) (string, error) {
			return `C:\ffmpeg\ffmpeg.exe`, nil
		},
		RunCommand: func(name string, args []string) error {
			time.Sleep(50 * time.Millisecond)
			return errors.New("signal: killed")
		},
		GetDuration: func(ffmpegPath string, inputPath string) (time.Duration, error) {
			return 10 * time.Second, nil
		},
		DetectGPUEncoder: func(ffmpegPath string, codec string, runCmd func(string, []string) error) string {
			return "libx264"
		},
		MkdirAll: func(path string, perm os.FileMode) error {
			return nil
		},
		PathExists: func(path string) bool {
			return false
		},
		InputAccessible: func(path string) bool {
			return true
		},
	}

	eng := NewCompressEngine(deps)

	ctx, cancel := context.WithCancel(context.Background())
	// Cancel immediately/soon
	go func() {
		time.Sleep(10 * time.Millisecond)
		cancel()
	}()

	reports, err := eng.Run(ctx, JobRequest{
		Files:       []string{`C:\videos\test.mp4`},
		Preset:      "standard",
		Concurrency: 1,
	}, nil)

	if err != nil {
		t.Fatalf("expected nil error (engine reports error via JobReport), got %v", err)
	}

	if len(reports) != 1 {
		t.Fatalf("expected 1 report, got %d", len(reports))
	}

	if reports[0].Status != "failed" {
		t.Fatalf("expected status failed, got %s", reports[0].Status)
	}

	if !strings.Contains(reports[0].ErrMessage, "任务已取消") {
		t.Fatalf("expected error message to contain '任务已取消', got %q", reports[0].ErrMessage)
	}
}

func TestEngineRunAppliesCrop(t *testing.T) {
	var runArgs [][]string
	deps := Dependencies{
		ExecutableDir: `C:\tools`,
		ResolveBinary: func(dir string) (string, error) {
			return `C:\ffmpeg\ffmpeg.exe`, nil
		},
		RunCommand: func(name string, args []string) error {
			runArgs = append(runArgs, args)
			return nil
		},
		GetDuration: func(ffmpegPath string, inputPath string) (time.Duration, error) {
			return 10 * time.Second, nil
		},
		GetDimensions: func(ffmpegPath string, inputPath string) (int, int, error) {
			return 1600, 900, nil
		},
		DetectGPUEncoder: func(ffmpegPath string, codec string, runCmd func(string, []string) error) string {
			return "libx264"
		},
		MkdirAll:        func(path string, perm os.FileMode) error { return nil },
		PathExists:      func(path string) bool { return false },
		InputAccessible: func(path string) bool { return true },
	}
	eng := NewCompressEngine(deps)

	tests := []struct {
		name     string
		crop     *CropSpec
		expected string
	}{
		{"exact", &CropSpec{Exact: true, W: 100, H: 100, X: 0, Y: 0}, "crop=100:100:0:0"},
		{"center 16:9 keep 90%", &CropSpec{RatioW: 16, RatioH: 9, Scale: 0.9}, "crop=1440:810:80:45"},
	}
	for _, tt := range tests {
		runArgs = nil
		reports, err := eng.Run(context.Background(), JobRequest{
			Files:       []string{`C:\videos\test.mp4`},
			Preset:      "standard",
			Concurrency: 1,
			Crop:        tt.crop,
		}, nil)
		if err != nil {
			t.Fatalf("%s: expected nil error, got %v", tt.name, err)
		}
		if reports[0].Status != "success" {
			t.Fatalf("%s: expected success, got %s (%s)", tt.name, reports[0].Status, reports[0].ErrMessage)
		}
		joined := strings.Join(runArgs[0], " ")
		if !strings.Contains(joined, "-vf "+tt.expected+",scale='") {
			t.Errorf("%s: expected vf to start with %q, got %s", tt.name, tt.expected, joined)
		}
	}
}

func TestEngineRunExactCropOutOfBounds(t *testing.T) {
	deps := Dependencies{
		ExecutableDir: `C:\tools`,
		ResolveBinary: func(dir string) (string, error) {
			return `C:\ffmpeg\ffmpeg.exe`, nil
		},
		RunCommand: func(name string, args []string) error {
			t.Errorf("should not invoke ffmpeg when crop is out of bounds")
			return nil
		},
		GetDuration: func(ffmpegPath string, inputPath string) (time.Duration, error) {
			return 10 * time.Second, nil
		},
		GetDimensions: func(ffmpegPath string, inputPath string) (int, int, error) {
			return 1600, 900, nil
		},
		DetectGPUEncoder: func(ffmpegPath string, codec string, runCmd func(string, []string) error) string {
			return "libx264"
		},
		MkdirAll:        func(path string, perm os.FileMode) error { return nil },
		PathExists:      func(path string) bool { return false },
		InputAccessible: func(path string) bool { return true },
	}
	eng := NewCompressEngine(deps)
	reports, err := eng.Run(context.Background(), JobRequest{
		Files:       []string{`C:\videos\test.mp4`},
		Preset:      "standard",
		Concurrency: 1,
		Crop:        &CropSpec{Exact: true, W: 2000, H: 900, X: 0, Y: 0},
	}, nil)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if reports[0].Status != "failed" {
		t.Fatalf("expected failed status, got %s", reports[0].Status)
	}
}
