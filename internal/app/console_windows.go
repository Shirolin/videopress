//go:build windows

package app

import (
	"os"
	"syscall"
	"unsafe"
)

var (
	procAllocConsole       = kernel32.NewProc("AllocConsole")
	procFreeConsole        = kernel32.NewProc("FreeConsole")
	procSetConsoleCP       = kernel32.NewProc("SetConsoleCP")
	procSetConsoleOutputCP = kernel32.NewProc("SetConsoleOutputCP")
	procSetConsoleTitleW   = kernel32.NewProc("SetConsoleTitleW")
)

// AttachSendToConsole 为 --sendto 模式分配可见控制台窗口，使无头启动的 GUI 子系统
// 进程也能实时显示压缩进度。已存在有效控制台（终端启动）时不做任何处理并返回 false。
func AttachSendToConsole() bool {
	// 标准输出已连接到有效控制台（cmd/PowerShell 直接运行），无需新开窗口
	if isTerminalFd(os.Stdout.Fd()) {
		return false
	}

	r, _, _ := procAllocConsole.Call()
	if r == 0 {
		return false
	}

	// 新控制台默认使用系统 OEM 代码页，强制 UTF-8 避免中文乱码
	procSetConsoleCP.Call(65001)
	procSetConsoleOutputCP.Call(65001)

	title := syscall.StringToUTF16Ptr("Videopress 快速压缩")
	procSetConsoleTitleW.Call(uintptr(unsafe.Pointer(title)))

	// 将标准流重定向到新控制台
	if conOut, err := os.OpenFile("CONOUT$", os.O_WRONLY, 0); err == nil {
		os.Stdout = conOut
		os.Stderr = conOut
	}
	if conIn, err := os.OpenFile("CONIN$", os.O_RDONLY, 0); err == nil {
		os.Stdin = conIn
	}

	// 启用 VT 处理以支持彩色输出与倒计时回车刷新
	_ = enableVirtualTerminal()
	return true
}

// DetachConsole 释放由 AttachSendToConsole 分配的控制台窗口。
func DetachConsole() {
	procFreeConsole.Call()
}
