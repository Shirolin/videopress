package compress

import (
	"fmt"
	"path/filepath"
	"strings"
)

type PathExistsFunc func(path string) bool

func BuildOutputPath(inputPath string, preset string, exists PathExistsFunc) (string, error) {
	ext := filepath.Ext(inputPath)
	if ext == "" {
		return "", fmt.Errorf("input file has no extension: %s", inputPath)
	}

	baseDir := filepath.Dir(inputPath)
	baseName := strings.TrimSuffix(filepath.Base(inputPath), ext)
	outputDir := filepath.Join(baseDir, "compressed")
	targetBase := filepath.Join(outputDir, fmt.Sprintf("%s.%s.compressed", baseName, preset))

	candidate := targetBase + ".mp4"
	if exists == nil || !exists(candidate) {
		return filepath.Clean(candidate), nil
	}

	index := 1
	for {
		candidate = filepath.Join(outputDir, fmt.Sprintf("%s.%s.compressed-%d.mp4", baseName, preset, index))
		if !exists(candidate) {
			return filepath.Clean(candidate), nil
		}
		index++
	}
}
