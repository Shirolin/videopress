package ffmpeg

import (
	"errors"
	"sync"
	"testing"
)

func TestDetectGPUEncoder(t *testing.T) {
	tests := []struct {
		name          string
		runCmd        func(name string, args []string) error
		expectedCodec string
	}{
		{
			name: "NVIDIA available",
			runCmd: func(name string, args []string) error {
				for _, arg := range args {
					if arg == "h264_nvenc" {
						return nil
					}
				}
				return errors.New("failed")
			},
			expectedCodec: "h264_nvenc",
		},
		{
			name: "QSV available",
			runCmd: func(name string, args []string) error {
				for _, arg := range args {
					if arg == "h264_qsv" {
						return nil
					}
				}
				return errors.New("failed")
			},
			expectedCodec: "h264_qsv",
		},
		{
			name: "fallback to libx264",
			runCmd: func(name string, args []string) error {
				return errors.New("failed")
			},
			expectedCodec: "libx264",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			codec := DetectGPUEncoder("ffmpeg", tt.runCmd)
			if codec != tt.expectedCodec {
				t.Errorf("expected codec %s, got %s", tt.expectedCodec, codec)
			}
		})
	}
}

func TestGPUEncoderCache(t *testing.T) {
	ResetGPUEncoderCache()

	// 模拟第一次探测，使用自定义的 runCmd。因为传入了自定义的 runCmd，它应当被判断为测试而绕过缓存逻辑
	var callCount int
	runCmd := func(name string, args []string) error {
		callCount++
		for _, arg := range args {
			if arg == "h264_qsv" {
				return nil
			}
		}
		return errors.New("failed")
	}

	codec1 := DetectGPUEncoder("ffmpeg", runCmd)
	if codec1 != "h264_qsv" {
		t.Fatalf("expected h264_qsv, got %s", codec1)
	}

	// 因为探测顺序是 h264_nvenc -> h264_qsv，所以第一次成功时 callCount 应为 2
	if callCount != 2 {
		t.Errorf("expected 2 calls for first detection (nvenc fails, qsv succeeds), got %d", callCount)
	}

	// 再次探测，因为 runCmd 不为 nil，再次绕过缓存重新执行
	codec2 := DetectGPUEncoder("ffmpeg", runCmd)
	if codec2 != "h264_qsv" {
		t.Fatalf("expected h264_qsv, got %s", codec2)
	}
	if callCount != 4 {
		t.Errorf("expected 4 calls total after second detection, got %d", callCount)
	}

	// 在实际逻辑中，测试非 RunCmd（即生产环境中的默认行为）下的缓存
	ResetGPUEncoderCache()
	cacheMu.Lock()
	cachedGPUEncoder = "h264_nvenc" // 模拟内存中已有缓存结果
	cacheMu.Unlock()

	// 生产调用（不传 runCmd）
	codec3 := DetectGPUEncoder("ffmpeg", nil)
	if codec3 != "h264_nvenc" {
		t.Errorf("expected cached h264_nvenc, got %s", codec3)
	}

	ResetGPUEncoderCache()
}

func TestGPUEncoderCacheConcurrency(t *testing.T) {
	ResetGPUEncoderCache()
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = DetectGPUEncoder("ffmpeg", func(name string, args []string) error {
				return errors.New("failed")
			})
		}()
	}
	wg.Wait()
}
