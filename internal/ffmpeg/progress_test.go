package ffmpeg

import (
	"bytes"
	"strings"
	"testing"
	"time"
)

func TestParseDuration(t *testing.T) {
	metadata := `ffmpeg version 6.0 Copyright (c) 2000-2023 the FFmpeg developers
  built with gcc 12.2.0 (Rev10, Built by MSYS2 project)
  duration: N/A, bitrate: N/A
  Duration: 00:02:15.50, start: 0.000000, bitrate: 1205 kb/s
  Stream #0:0[0x1]: Video: h264`

	dur, err := ParseDuration(metadata)
	if err != nil {
		t.Fatalf("ParseDuration failed: %v", err)
	}

	expected := 2*time.Minute + 15*time.Second + 500*time.Millisecond
	if dur != expected {
		t.Errorf("expected duration %v, got %v", expected, dur)
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
