package env

import (
	"fmt"
	"path/filepath"
	"strings"

	"golang.org/x/sys/windows/registry"
)

type GetPathFunc func() (string, error)
type SetPathFunc func(string) error

func AddToPath(dir string) (bool, error) {
	return AddToPathAt(dir, getPath, setPath)
}

func RemoveFromPath(dir string) (bool, error) {
	return RemoveFromPathAt(dir, getPath, setPath)
}

func IsPathConfigured(dir string) (bool, error) {
	return IsPathConfiguredAt(dir, getPath)
}

func IsPathConfiguredAt(dir string, getPath GetPathFunc) (bool, error) {
	dirClean := filepath.Clean(dir)
	current, err := getPath()
	if err != nil {
		return false, err
	}
	parts := splitPath(current)
	for _, p := range parts {
		if filepath.Clean(p) == dirClean {
			return true, nil
		}
	}
	return false, nil
}

func AddToPathAt(dir string, getPath GetPathFunc, setPath SetPathFunc) (bool, error) {
	dirClean := filepath.Clean(dir)
	current, err := getPath()
	if err != nil {
		return false, err
	}

	parts := splitPath(current)
	for _, part := range parts {
		if filepath.Clean(part) == dirClean {
			return false, nil // 已经存在，无需添加
		}
	}

	parts = append(parts, dir)
	newPath := joinPath(parts)
	if err := setPath(newPath); err != nil {
		return false, err
	}
	return true, nil
}

func RemoveFromPathAt(dir string, getPath GetPathFunc, setPath SetPathFunc) (bool, error) {
	dirClean := filepath.Clean(dir)
	current, err := getPath()
	if err != nil {
		return false, err
	}

	parts := splitPath(current)
	var updated []string
	removed := false
	for _, part := range parts {
		if filepath.Clean(part) == dirClean {
			removed = true
			continue
		}
		updated = append(updated, part)
	}

	if !removed {
		return false, nil // 不存在，无需移除
	}

	newPath := joinPath(updated)
	if err := setPath(newPath); err != nil {
		return false, err
	}
	return true, nil
}

func splitPath(pathStr string) []string {
	if pathStr == "" {
		return nil
	}
	var res []string
	for _, p := range strings.Split(pathStr, ";") {
		p = strings.TrimSpace(p)
		if p != "" {
			res = append(res, p)
		}
	}
	return res
}

func joinPath(parts []string) string {
	return strings.Join(parts, ";")
}

func getPath() (string, error) {
	k, err := registry.OpenKey(registry.CURRENT_USER, `Environment`, registry.QUERY_VALUE)
	if err != nil {
		return "", fmt.Errorf("读取注册表失败: %w", err)
	}
	defer k.Close()
	val, _, err := k.GetStringValue("Path")
	if err != nil {
		// 如果注册表中不存在 Path (虽然罕见)，返回空字符串而不是报错
		if err == registry.ErrNotExist {
			return "", nil
		}
		return "", fmt.Errorf("读取 Path 变量失败: %w", err)
	}
	return val, nil
}

func setPath(newPath string) error {
	k, err := registry.OpenKey(registry.CURRENT_USER, `Environment`, registry.SET_VALUE)
	if err != nil {
		return fmt.Errorf("打开注册表失败: %w", err)
	}
	defer k.Close()
	err = k.SetStringValue("Path", newPath)
	if err != nil {
		return fmt.Errorf("写入 Path 变量失败: %w", err)
	}
	return nil
}
