package ffmpeg

import (
	"errors"
	"sync"
	"testing"
)

func TestDetectGPUEncoder(t *testing.T) {
	tests := []struct {
		name          string
		codec         string
		runCmd        func(name string, args []string) error
		expectedCodec string
	}{
		{
			name:  "NVIDIA H.264 available",
			codec: "h264",
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
			name:  "Intel H.265 QSV available",
			codec: "h265",
			runCmd: func(name string, args []string) error {
				for _, arg := range args {
					if arg == "hevc_qsv" {
						return nil
					}
				}
				return errors.New("failed")
			},
			expectedCodec: "hevc_qsv",
		},
		{
			name:  "fallback to libsvtav1 for AV1",
			codec: "av1",
			runCmd: func(name string, args []string) error {
				return errors.New("failed")
			},
			expectedCodec: "libsvtav1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			codec := DetectGPUEncoder("ffmpeg", tt.codec, tt.runCmd)
			if codec != tt.expectedCodec {
				t.Errorf("expected codec %s, got %s", tt.expectedCodec, codec)
			}
		})
	}
}

func TestGPUEncoderCache(t *testing.T) {
	ResetGPUEncoderCache()

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

	codec1 := DetectGPUEncoder("ffmpeg", "h264", runCmd)
	if codec1 != "h264_qsv" {
		t.Fatalf("expected h264_qsv, got %s", codec1)
	}

	if callCount != 2 {
		t.Errorf("expected 2 calls for first detection, got %d", callCount)
	}

	// 再次探测，因为 runCmd 不为 nil，再次绕过缓存重新执行
	codec2 := DetectGPUEncoder("ffmpeg", "h264", runCmd)
	if codec2 != "h264_qsv" {
		t.Fatalf("expected h264_qsv, got %s", codec2)
	}
	if callCount != 4 {
		t.Errorf("expected 4 calls total after second detection, got %d", callCount)
	}

	// 内存缓存测试
	ResetGPUEncoderCache()
	cacheMu.Lock()
	cachedGPUEncoders["h264"] = "h264_nvenc"
	cacheMu.Unlock()

	codec3 := DetectGPUEncoder("ffmpeg", "h264", nil)
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
			_ = DetectGPUEncoder("ffmpeg", "h264", func(name string, args []string) error {
				return errors.New("failed")
			})
		}()
	}
	wg.Wait()
}
