package env

import (
	"bytes"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
)

type GetPathFunc func() (string, error)
type SetPathFunc func(string) error

func AddToPath(dir string) (bool, error) {
	return AddToPathAt(dir, getPath, setPath)
}

func RemoveFromPath(dir string) (bool, error) {
	return RemoveFromPathAt(dir, getPath, setPath)
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
	cmd := exec.Command("powershell", "-NoProfile", "-Command", `[Environment]::GetEnvironmentVariable("Path", "User")`)
	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("读取环境变量失败: %w", err)
	}
	return strings.TrimSpace(out.String()), nil
}

func setPath(newPath string) error {
	escapedPath := strings.ReplaceAll(newPath, `'`, `''`)
	cmdStr := fmt.Sprintf(`[Environment]::SetEnvironmentVariable("Path", '%s', "User")`, escapedPath)
	cmd := exec.Command("powershell", "-NoProfile", "-Command", cmdStr)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("写入环境变量失败: %w", err)
	}
	return nil
}
