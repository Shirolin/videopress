package ffmpeg

import (
	"errors"
	"path/filepath"
	"testing"
)

func TestResolveBinaryPrefersExecutableDirectory(t *testing.T) {
	lookups := 0
	path, err := ResolveBinary(`C:\tools\videopress`, func(name string) (string, error) {
		lookups++
		if name == filepath.Clean(`C:\tools\videopress\ffmpeg.exe`) {
			return name, nil
		}
		return "", errors.New("not found")
	})
	if err != nil {
		t.Fatalf("ResolveBinary returned error: %v", err)
	}
	if path != filepath.Clean(`C:\tools\videopress\ffmpeg.exe`) {
		t.Fatalf("expected local ffmpeg path, got %s", path)
	}
	if lookups != 1 {
		t.Fatalf("expected 1 lookup, got %d", lookups)
	}
}

func TestResolveBinaryFallsBackToPathLookup(t *testing.T) {
	path, err := ResolveBinary(`C:\tools\videopress`, func(name string) (string, error) {
		if name == filepath.Clean(`C:\tools\videopress\ffmpeg.exe`) {
			return "", errors.New("not found")
		}
		if name == "ffmpeg" {
			return `C:\ffmpeg\bin\ffmpeg.exe`, nil
		}
		return "", errors.New("unexpected lookup")
	})
	if err != nil {
		t.Fatalf("ResolveBinary returned error: %v", err)
	}
	if path != filepath.Clean(`C:\ffmpeg\bin\ffmpeg.exe`) {
		t.Fatalf("expected PATH ffmpeg, got %s", path)
	}
}

func TestResolveBinaryReturnsHelpfulError(t *testing.T) {
	_, err := ResolveBinary(`C:\tools\videopress`, func(name string) (string, error) {
		return "", errors.New("not found")
	})
	if err == nil {
		t.Fatal("expected error when ffmpeg cannot be found")
	}
}
