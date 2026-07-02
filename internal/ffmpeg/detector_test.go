package ffmpeg

import (
	"errors"
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
