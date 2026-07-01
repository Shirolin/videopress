package compress

import "testing"

func TestPresetByNameReturnsExpectedConfig(t *testing.T) {
	small, err := PresetByName("small")
	if err != nil {
		t.Fatalf("PresetByName(small) returned error: %v", err)
	}
	if small.Name != "small" {
		t.Fatalf("expected preset name small, got %s", small.Name)
	}
	if small.MaxDimension != 720 {
		t.Fatalf("expected max dimension 720, got %d", small.MaxDimension)
	}
	if small.CRF != 30 {
		t.Fatalf("expected crf 30, got %d", small.CRF)
	}
	if small.Preset != "veryfast" {
		t.Fatalf("expected ffmpeg preset veryfast, got %s", small.Preset)
	}
	if small.AudioBitrate != "64k" {
		t.Fatalf("expected audio bitrate 64k, got %s", small.AudioBitrate)
	}

	standard, err := PresetByName("standard")
	if err != nil {
		t.Fatalf("PresetByName(standard) returned error: %v", err)
	}
	if standard.MaxDimension != 1080 {
		t.Fatalf("expected max dimension 1080, got %d", standard.MaxDimension)
	}
	if standard.CRF != 27 {
		t.Fatalf("expected crf 27, got %d", standard.CRF)
	}
	if standard.AudioBitrate != "96k" {
		t.Fatalf("expected audio bitrate 96k, got %s", standard.AudioBitrate)
	}

	quality, err := PresetByName("quality")
	if err != nil {
		t.Fatalf("PresetByName(quality) returned error: %v", err)
	}
	if quality.MaxDimension != 1440 {
		t.Fatalf("expected max dimension 1440, got %d", quality.MaxDimension)
	}
	if quality.CRF != 24 {
		t.Fatalf("expected crf 24, got %d", quality.CRF)
	}
	if quality.AudioBitrate != "128k" {
		t.Fatalf("expected audio bitrate 128k, got %s", quality.AudioBitrate)
	}

	upper, err := PresetByName("STANDARD")
	if err != nil {
		t.Fatalf("PresetByName(STANDARD) returned error: %v", err)
	}
	if upper.MaxDimension != 1080 {
		t.Fatalf("expected max dimension 1080 for STANDARD, got %d", upper.MaxDimension)
	}
}

func TestPresetByNameRejectsUnknownPreset(t *testing.T) {
	_, err := PresetByName("unknown")
	if err == nil {
		t.Fatal("expected error for unknown preset")
	}
}
