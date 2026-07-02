package app

import (
	"os"
	"os/exec"
	"syscall"
	"unsafe"
)

func init() {
	EnableConsoleColors = func() {
		k32 := syscall.NewLazyDLL("kernel32.dll")
		getStdHandle := k32.NewProc("GetStdHandle")
		setConsoleMode := k32.NewProc("SetConsoleMode")
		getConsoleMode := k32.NewProc("GetConsoleMode")

		ret, _, _ := getStdHandle.Call(uintptr(0xfffffff5)) // STD_OUTPUT_HANDLE
		handle := syscall.Handle(ret)
		if handle == syscall.InvalidHandle {
			return
		}

		var mode uint32
		ret, _, _ = getConsoleMode.Call(uintptr(handle), uintptr(unsafe.Pointer(&mode)))
		if ret != 0 {
			mode |= 0x0004 // ENABLE_VIRTUAL_TERMINAL_PROCESSING
			setConsoleMode.Call(uintptr(handle), uintptr(mode))
		}
	}
}

func execLookPath(file string) (string, error) {
	return exec.LookPath(file)
}

func runCommand(name string, args []string) error {
	cmd := exec.Command(name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func pathExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func checkInputAccessible(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}
