package main

import (
	"embed"
	"fmt"
	"os"
	"path/filepath"

	"videopress/internal/app"
	"videopress/internal/env"
	"videopress/internal/sendto"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	executablePath, err := os.Executable()
	if err != nil {
		fmt.Fprintln(os.Stderr, "无法获取可执行文件路径:", err)
		os.Exit(1)
	}
	execDir := filepath.Dir(executablePath)

	// 如果有命令行参数，走 CLI 模式
	if len(os.Args) > 1 {
		exitCode := app.Execute(os.Args[1:], app.Dependencies{
			ExecutableDir:   execDir,
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

	// 否则启动 GUI 模式 (Wails)
	guiApp := NewApp(execDir, executablePath)

	err = wails.Run(&options.App{
		Title:  "Videopress",
		Width:  850,
		Height: 620,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 10, G: 10, B: 12, A: 1},
		OnStartup:        guiApp.startup,
		Bind: []interface{}{
			guiApp,
		},
	})

	if err != nil {
		fmt.Fprintln(os.Stderr, "GUI 启动错误:", err)
		os.Exit(1)
	}
}
