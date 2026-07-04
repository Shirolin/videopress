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

	args := BuildArgs(`C:\videos\input.mov`, `C:\videos\compressed\output.mp4`, preset, "", false, 0, "")
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

	if !strings.Contains(joined, "scale='if(gt(iw,ih),trunc(min(iw*0.50,720)/2)*2,-2)':'if(gt(iw,ih),-2,trunc(min(ih*0.50,720)/2)*2)'") {
		t.Fatalf("expected scale filter for 0.50 scale factor and 720 max dimension, got %s", joined)
	}
}

func TestBuildArgsForGPUEncoders(t *testing.T) {
	preset, _ := compress.PresetByName("standard")

	tests := []struct {
		encoder  string
		expected []string
	}{
		{"h264_nvenc", []string{"-c:v h264_nvenc", "-rc constqp", "-qp 27"}},
		{"h264_qsv", []string{"-c:v h264_qsv", "-global_quality 27"}},
		{"h264_amf", []string{"-c:v h264_amf", "-rc cqp", "-qp_i 27", "-qp_p 27"}},
		{"hevc_nvenc", []string{"-c:v hevc_nvenc", "-rc constqp", "-qp 27"}},
		{"av1_qsv", []string{"-c:v av1_qsv", "-global_quality 27"}},
	}

	for _, tt := range tests {
		args := BuildArgs(`C:\videos\input.mov`, `C:\videos\compressed\output.mp4`, preset, tt.encoder, false, 0, "")
		joined := strings.Join(args, " ")
		for _, part := range tt.expected {
			if !strings.Contains(joined, part) {
				t.Errorf("expected encoder %s args to contain %q, got %s", tt.encoder, part, joined)
			}
		}
	}
}

func TestBuildArgsWithCopyAudio(t *testing.T) {
	preset, _ := compress.PresetByName("standard")
	args := BuildArgs(`C:\videos\input.mov`, `C:\videos\compressed\output.mp4`, preset, "", true, 0, "")
	joined := strings.Join(args, " ")

	if !strings.Contains(joined, "-c:a copy") {
		t.Fatalf("expected copy audio parameter, got %s", joined)
	}
	if strings.Contains(joined, "-c:a aac") {
		t.Fatalf("should not contain aac codec, got %s", joined)
	}
}

func TestBuildArgsAdvancedOptions(t *testing.T) {
	preset, _ := compress.PresetByName("standard")

	// 1. MaxFPS
	args := BuildArgs(`C:\videos\input.mov`, `C:\videos\compressed\output.mp4`, preset, "", false, 30, "")
	joined := strings.Join(args, " ")
	if !strings.Contains(joined, "-r 30") {
		t.Fatalf("expected -r 30 parameter, got %s", joined)
	}

	// 2. AudioMode mute
	argsMute := BuildArgs(`C:\videos\input.mov`, `C:\videos\compressed\output.mp4`, preset, "", false, 0, "mute")
	joinedMute := strings.Join(argsMute, " ")
	if !strings.Contains(joinedMute, "-an") {
		t.Fatalf("expected -an parameter for mute mode, got %s", joinedMute)
	}
	if strings.Contains(joinedMute, "-c:a") {
		t.Fatalf("should not set audio codec when muted, got %s", joinedMute)
	}

	// 3. AudioMode copy
	argsCopy := BuildArgs(`C:\videos\input.mov`, `C:\videos\compressed\output.mp4`, preset, "", false, 0, "copy")
	joinedCopy := strings.Join(argsCopy, " ")
	if !strings.Contains(joinedCopy, "-c:a copy") {
		t.Fatalf("expected -c:a copy parameter for copy mode, got %s", joinedCopy)
	}
}
