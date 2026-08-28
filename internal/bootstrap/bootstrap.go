package bootstrap

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"videopress/internal/cli"
	"videopress/internal/desktop"
	"videopress/internal/env"
	"videopress/internal/gui"
	"videopress/internal/sendto"
)

var validCLIFlags = map[string]bool{
	"--preset":           true,
	"--concurrency":      true,
	"-c":                 true,
	"--hw":               true,
	"--force":            true,
	"-f":                 true,
	"--skip-existing":    true,
	"--copy-audio":       true,
	"-a":                 true,
	"--codec":            true,
	"--max-fps":          true,
	"--audio":            true,
	"--crf":              true,
	"--sendto":           true,
	"--install-sendto":   true,
	"--uninstall-sendto": true,
	"--install-path":     true,
	"--uninstall-path":   true,
	"--version":          true,
	"-h":                 true,
	"--help":             true,
}

// Run dispatches to CLI or GUI based on arguments. Returns process exit code.
func Run(version string, args []string) int {
	if version != "" {
		cli.Version = version
	}

	executablePath, err := os.Executable()
	if err != nil {
		fmt.Fprintln(os.Stderr, "无法获取可执行文件路径:", err)
		return 1
	}
	execDir := filepath.Dir(executablePath)

	if isCLIMode(args) {
		return cli.Execute(args[1:], cli.Dependencies{
			ExecutableDir:   execDir,
			ExecutablePath:  executablePath,
			Stdout:          os.Stdout,
			Stderr:          os.Stderr,
			InstallSendTo:   sendto.Install,
			UninstallSendTo: sendto.Uninstall,
			AddToPath:       env.AddToPath,
			RemoveFromPath:  env.RemoveFromPath,
		})
	}

	var initialFiles []string
	if len(args) > 1 {
		initialFiles = args[1:]
	}

	guiApp := gui.New(execDir, executablePath, initialFiles)
	if err := desktop.Run(guiApp); err != nil {
		fmt.Fprintln(os.Stderr, "GUI 启动错误:", err)
		return 1
	}

	return 0
}

func isCLIMode(args []string) bool {
	for _, arg := range args[1:] {
		if validCLIFlags[strings.ToLower(arg)] {
			return true
		}
	}
	return false
}
