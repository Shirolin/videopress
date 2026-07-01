package main

import (
	"os"
	"path/filepath"

	"videopress/internal/app"
	"videopress/internal/sendto"
)

func main() {
	executablePath, _ := os.Executable()
	exitCode := app.Execute(os.Args[1:], app.Dependencies{
		ExecutableDir:   filepath.Dir(executablePath),
		ExecutablePath:  executablePath,
		InstallSendTo:   sendto.Install,
		UninstallSendTo: sendto.Uninstall,
	})
	os.Exit(exitCode)
}
