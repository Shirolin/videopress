package ffmpeg

import (
	"bytes"
	"strings"
	"testing"
	"time"
)

func TestParseDuration(t *testing.T) {
	tests := []struct {
		name     string
		metadata string
		expected time.Duration
	}{
		{
			name:     "2 decimal milliseconds",
			metadata: "  Duration: 00:02:15.50, start: 0.0",
			expected: 2*time.Minute + 15*time.Second + 500*time.Millisecond,
		},
		{
			name:     "3 decimal milliseconds",
			metadata: "  Duration: 00:02:15.123, start: 0.0",
			expected: 2*time.Minute + 15*time.Second + 123*time.Millisecond,
		},
		{
			name:     "1 decimal millisecond",
			metadata: "  Duration: 00:02:15.5, start: 0.0",
			expected: 2*time.Minute + 15*time.Second + 500*time.Millisecond,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dur, err := ParseDuration(tt.metadata)
			if err != nil {
				t.Fatalf("ParseDuration failed: %v", err)
			}
			if dur != tt.expected {
				t.Errorf("expected duration %v, got %v", tt.expected, dur)
			}
		})
	}
}

func TestTrackProgress(t *testing.T) {
	progressOutput := `frame=120
fps=30.00
stream_0_0_q=28.0
bitrate= 1200.5kbits/s
total_size=2048000
out_time_us=30000000
out_time_ms=30000000
out_time=00:00:30.000000
dup_frames=0
drop_frames=0
speed=1.5x
progress=continue`

	r := strings.NewReader(progressOutput)
	duration := 60 * time.Second // 30s / 60s = 50%

	var calledPercent float64
	TrackProgress(r, duration, func(percent float64) {
		calledPercent = percent
	})

	if calledPercent != 50.0 {
		t.Errorf("expected progress percentage to be 50.0, got %.1f", calledPercent)
	}
}

func TestRenderProgressBar(t *testing.T) {
	var buf bytes.Buffer
	RenderProgressBar(&buf, 50.0, "test.mp4")

	out := buf.String()
	if !strings.HasPrefix(out, "\rtest.mp4            ") { // 20个字符对齐
		t.Errorf("unexpected prefix in progress bar: %q", out)
	}
	if !strings.Contains(out, "50.0%") {
		t.Errorf("expected 50.0%% in progress bar: %q", out)
	}
}

func TestParseDimensions(t *testing.T) {
	videoLine := "Stream #0:0[0x1](und): Video: h264 (High) (avc1 / 0x31637661), yuv420p, 1600x900, 30 fps"
	coverLine := "Stream #0:1: Video: mjpeg (Progressive), yuvj444p, 640x360, attached_pic"

	tests := []struct {
		name         string
		metadata     string
		wantW, wantH int
		wantErr      bool
	}{
		{"ordinary single stream", videoLine, 1600, 900, false},
		{"attached pic first must be skipped", coverLine + "\n" + videoLine, 1600, 900, false},
		{"cover only", coverLine, 640, 360, false},
		{"no video stream", "Stream #0:0: Audio: aac, 44100 Hz", 0, 0, true},
	}
	for _, tt := range tests {
		w, h, err := ParseDimensions(tt.metadata)
		if tt.wantErr {
			if err == nil {
				t.Errorf("%s: expected error, got %dx%d", tt.name, w, h)
			}
			continue
		}
		if err != nil {
			t.Errorf("%s: unexpected error: %v", tt.name, err)
			continue
		}
		if w != tt.wantW || h != tt.wantH {
			t.Errorf("%s: expected %dx%d, got %dx%d", tt.name, tt.wantW, tt.wantH, w, h)
		}
	}
}
