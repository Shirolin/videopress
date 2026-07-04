package util

import (
	"os"
	"path/filepath"
	"strings"
)

// IsVideoFile 检查文件扩展名是否属于视频类型
func IsVideoFile(path string) bool {
	switch strings.ToLower(filepath.Ext(path)) {
	case ".mp4", ".mov", ".mkv", ".avi", ".m4v", ".wmv", ".webm",
		".ts", ".flv", ".mpg", ".mpeg", ".3gp":
		return true
	default:
		return false
	}
}

// GetFileSize 获取文件大小，如果出错则返回 0
func GetFileSize(path string) int64 {
	info, err := os.Stat(path)
	if err != nil {
		return 0
	}
	return info.Size()
}
