package ffmpeg

import (
	"fmt"
	"path/filepath"
)

type LookPathFunc func(file string) (string, error)

func ResolveBinary(executableDir string, lookPath LookPathFunc) (string, error) {
	localBinary := filepath.Clean(filepath.Join(executableDir, "ffmpeg.exe"))
	if path, err := lookPath(localBinary); err == nil {
		return filepath.Clean(path), nil
	}

	if path, err := lookPath("ffmpeg"); err == nil {
		return filepath.Clean(path), nil
	}

	return "", fmt.Errorf("ffmpeg not found; please install ffmpeg and ensure it is available in PATH or next to videopress.exe")
}
