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

	if reports[0].Status != "成功" {
		t.Fatalf("expected status 成功, got %s", reports[0].Status)
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

	if reports[0].Status != "失败" {
		t.Fatalf("expected status 失败, got %s", reports[0].Status)
	}

	if !strings.Contains(reports[0].ErrMessage, "任务已取消") {
		t.Fatalf("expected error message to contain '任务已取消', got %q", reports[0].ErrMessage)
	}
}
