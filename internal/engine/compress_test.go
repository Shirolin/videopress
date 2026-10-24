package engine

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"videopress/internal/ffmpeg"
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
		Probe: func(ffmpegPath string, inputPath string) (ffmpeg.VideoInfo, error) {
			return ffmpeg.VideoInfo{Duration: 10 * time.Second}, nil
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
	output := filepath.Join(t.TempDir(), "test.compressed.mp4")
	deps := Dependencies{
		ExecutableDir: `C:\tools`,
		ResolveBinary: func(dir string) (string, error) {
			return `C:\ffmpeg\ffmpeg.exe`, nil
		},
		RunCommand: func(name string, args []string) error {
			time.Sleep(50 * time.Millisecond)
			// 模拟 ffmpeg 被中断时已写出的半截产物
			_ = os.WriteFile(output, []byte("partial"), 0o644)
			return errors.New("signal: killed")
		},
		Probe: func(ffmpegPath string, inputPath string) (ffmpeg.VideoInfo, error) {
			return ffmpeg.VideoInfo{Duration: 10 * time.Second}, nil
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
		OutputDir:   filepath.Dir(output),
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

	if _, err := os.Stat(output); !os.IsNotExist(err) {
		t.Fatalf("expected partial output to be removed, stat err = %v", err)
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
		Probe: func(ffmpegPath string, inputPath string) (ffmpeg.VideoInfo, error) {
			return ffmpeg.VideoInfo{Duration: 10 * time.Second, Width: 1600, Height: 900}, nil
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

	// 精确裁奇数自动落偶（与居中裁策略一致）
	tests = append(tests, struct {
		name     string
		crop     *CropSpec
		expected string
	}{"exact odd values fall to even", &CropSpec{Exact: true, W: 801, H: 701, X: 3, Y: 5}, "crop=800:700:2:4"})
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

// 不裁剪时也必须探测时长，否则进度回调拿不到百分比（回归：进度条不动）。
func TestEngineRunProbesDurationWithoutCrop(t *testing.T) {
	probes := 0
	deps := Dependencies{
		ExecutableDir: `C:\tools`,
		ResolveBinary: func(dir string) (string, error) {
			return `C:\ffmpeg\ffmpeg.exe`, nil
		},
		RunCommand: func(name string, args []string) error { return nil },
		Probe: func(ffmpegPath string, inputPath string) (ffmpeg.VideoInfo, error) {
			probes++
			return ffmpeg.VideoInfo{Duration: 10 * time.Second}, nil
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
	}, nil)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if reports[0].Status != "success" {
		t.Fatalf("expected success, got %s (%s)", reports[0].Status, reports[0].ErrMessage)
	}
	if probes != 1 {
		t.Fatalf("expected 1 probe without crop, got %d", probes)
	}
}

// 不同目录下的同名输入会映射到同一个输出文件，同批次必须错开，否则并发写同一文件。
func TestEngineRunAvoidsOutputCollisionInBatch(t *testing.T) {
	var mu sync.Mutex
	var runArgs [][]string
	deps := Dependencies{
		ExecutableDir: `C:\tools`,
		ResolveBinary: func(dir string) (string, error) {
			return `C:\ffmpeg\ffmpeg.exe`, nil
		},
		RunCommand: func(name string, args []string) error {
			mu.Lock()
			runArgs = append(runArgs, args)
			mu.Unlock()
			return nil
		},
		Probe: func(ffmpegPath string, inputPath string) (ffmpeg.VideoInfo, error) {
			return ffmpeg.VideoInfo{Duration: 10 * time.Second}, nil
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
		Files:       []string{`C:\a\demo.mp4`, `C:\b\demo.mp4`},
		Preset:      "standard",
		Concurrency: 2,
		OutputDir:   `C:\out`,
	}, nil)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if len(reports) != 2 || len(runArgs) != 2 {
		t.Fatalf("expected 2 reports and 2 runs, got %d and %d", len(reports), len(runArgs))
	}

	outputs := map[string]bool{}
	for _, args := range runArgs {
		outputs[args[len(args)-1]] = true
	}
	if len(outputs) != 2 {
		t.Fatalf("expected two distinct output paths, got %v", outputs)
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
		Probe: func(ffmpegPath string, inputPath string) (ffmpeg.VideoInfo, error) {
			return ffmpeg.VideoInfo{Duration: 10 * time.Second, Width: 1600, Height: 900}, nil
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
