//go:build windows

package app

import (
	"syscall"
	"unsafe"
)

var (
	kernel32        = syscall.NewLazyDLL("kernel32.dll")
	getConsoleMode  = kernel32.NewProc("GetConsoleMode")
	setConsoleMode  = kernel32.NewProc("SetConsoleMode")
	getStdHandle    = kernel32.NewProc("GetStdHandle")

	msvcrt = syscall.NewLazyDLL("msvcrt.dll")
	kbhit  = msvcrt.NewProc("_kbhit")
	getch  = msvcrt.NewProc("_getch")
)

const (
	stdOutputHandle                 = uint32(0xfffffff5) // -11
	enableVirtualTerminalProcessing = uint32(0x0004)
)

func enableVirtualTerminal() error {
	handle, _, _ := getStdHandle.Call(uintptr(stdOutputHandle))
	if syscall.Handle(handle) == syscall.InvalidHandle {
		return syscall.ENOTTY
	}
	var mode uint32
	r, _, _ := getConsoleMode.Call(handle, uintptr(unsafe.Pointer(&mode)))
	if r == 0 {
		return syscall.ENOTTY
	}
	mode |= enableVirtualTerminalProcessing
	r, _, _ = setConsoleMode.Call(handle, uintptr(mode))
	if r == 0 {
		return syscall.ENOTTY
	}
	return nil
}

func hasKey() bool {
	r, _, _ := kbhit.Call()
	return r != 0
}

func readKey() int {
	r, _, _ := getch.Call()
	return int(r)
}

func isTerminalFd(fd uintptr) bool {
	var mode uint32
	r, _, _ := getConsoleMode.Call(fd, uintptr(unsafe.Pointer(&mode)))
	return r != 0
}
