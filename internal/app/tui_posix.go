//go:build !windows

package app

func enableVirtualTerminal() error {
	return nil
}

func hasKey() bool {
	return false
}

func readKey() int {
	return 0
}

func isTerminalFd(fd uintptr) bool {
	// POSIX 平台非 TTY 检测这里仅作编译占位，可后续补充
	return true
}
