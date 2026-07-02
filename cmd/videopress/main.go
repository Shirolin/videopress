package main

import (
	"fmt"
	"os"
	"path/filepath"

	"videopress/internal/app"
	"videopress/internal/env"
	"videopress/internal/sendto"
)

func main() {
	executablePath, err := os.Executable()
	if err != nil {
		fmt.Fprintln(os.Stderr, "无法获取可执行文件路径:", err)
		os.Exit(1)
	}
	exitCode := app.Execute(os.Args[1:], app.Dependencies{
		ExecutableDir:   filepath.Dir(executablePath),
		ExecutablePath:  executablePath,
		Stdout:          os.Stdout,
		Stderr:          os.Stderr,
		InstallSendTo:   sendto.Install,
		UninstallSendTo: sendto.Uninstall,
		AddToPath:       env.AddToPath,
		RemoveFromPath:  env.RemoveFromPath,
	})
	os.Exit(exitCode)
}
