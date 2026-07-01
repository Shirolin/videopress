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

	return "", fmt.Errorf("未找到 ffmpeg，请安装 ffmpeg 并确保其在 PATH 中或与 videopress.exe 同目录")
}
