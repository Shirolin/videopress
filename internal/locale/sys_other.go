//go:build !windows

package locale

// System 返回系统 UI 语言（zh 或 en）。
func System() string {
	return "zh"
}
