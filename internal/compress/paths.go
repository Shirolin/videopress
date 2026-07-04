package compress

import (
	"fmt"
	"path/filepath"
	"strings"
)

type PathExistsFunc func(path string) bool

func BuildOutputPath(inputPath string, preset string, exists PathExistsFunc, force bool, customOutputDir string) (string, error) {
	ext := filepath.Ext(inputPath)
	if ext == "" {
		return "", fmt.Errorf("输入文件没有扩展名: %s", inputPath)
	}

	baseDir := filepath.Dir(inputPath)
	baseName := strings.TrimSuffix(filepath.Base(inputPath), ext)
	outputDir := filepath.Join(baseDir, "compressed")
	if customOutputDir != "" {
		outputDir = customOutputDir
	}
	targetBase := filepath.Join(outputDir, fmt.Sprintf("%s.compressed", baseName))

	candidate := targetBase + ".mp4"
	if force || exists == nil || !exists(candidate) {
		return filepath.Clean(candidate), nil
	}

	index := 1
	for {
		if index > 10000 {
			return "", fmt.Errorf("无法生成唯一的输出文件名: 已尝试 %d 次", index)
		}
		candidate = filepath.Join(outputDir, fmt.Sprintf("%s.compressed-%d.mp4", baseName, index))
		if !exists(candidate) {
			return filepath.Clean(candidate), nil
		}
		index++
	}
}
