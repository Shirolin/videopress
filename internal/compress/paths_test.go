package compress

import (
	"path/filepath"
	"testing"
)

func TestBuildOutputPathUsesCompressedSubdirectory(t *testing.T) {
	output, err := BuildOutputPath(`C:\videos\demo.mp4`, "standard", nil, false, "")
	if err != nil {
		t.Fatalf("BuildOutputPath returned error: %v", err)
	}

	expected := filepath.Clean(`C:\videos\compressed\demo.standard.compressed.mp4`)
	if output != expected {
		t.Fatalf("expected %s, got %s", expected, output)
	}
}

func TestBuildOutputPathAppendsSequenceWhenTargetExists(t *testing.T) {
	exists := func(path string) bool {
		return filepath.Clean(path) == filepath.Clean(`C:\videos\compressed\demo.standard.compressed.mp4`)
	}

	output, err := BuildOutputPath(`C:\videos\demo.mp4`, "standard", exists, false, "")
	if err != nil {
		t.Fatalf("BuildOutputPath returned error: %v", err)
	}

	expected := filepath.Clean(`C:\videos\compressed\demo.standard.compressed-1.mp4`)
	if output != expected {
		t.Fatalf("expected %s, got %s", expected, output)
	}
}

func TestBuildOutputPathRejectsMissingExtension(t *testing.T) {
	_, err := BuildOutputPath(`C:\videos\demo`, "standard", nil, false, "")
	if err == nil {
		t.Fatal("expected error for missing extension")
	}
}

func TestBuildOutputPathWithForce(t *testing.T) {
	exists := func(path string) bool { return true }
	output, err := BuildOutputPath(`C:\videos\demo.mp4`, "standard", exists, true, "")
	if err != nil {
		t.Fatalf("BuildOutputPath returned error: %v", err)
	}
	expected := filepath.Clean(`C:\videos\compressed\demo.standard.compressed.mp4`)
	if output != expected {
		t.Fatalf("expected %s even if file exists because force=true, got %s", expected, output)
	}
}
