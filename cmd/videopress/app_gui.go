package main

import (
	"context"
	"fmt"
	"path/filepath"

	"videopress/internal/compress"
	"videopress/internal/engine"
	"videopress/internal/env"
	"videopress/internal/ffmpeg"
	"videopress/internal/sendto"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// App struct handles GUI bindings
type App struct {
	ctx            context.Context
	executableDir  string
	executablePath string
}

// NewApp creates a new App struct instance
func NewApp(execDir, execPath string) *App {
	return &App{
		executableDir:  execDir,
		executablePath: execPath,
	}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

// PresetInfo represents preset metadata returned to frontend
type PresetInfo struct {
	Name         string  `json:"name"`
	ScaleFactor  float64 `json:"scaleFactor"`
	MaxDimension int     `json:"maxDimension"`
	Description  string  `json:"description"`
}

// GetPresets returns the list of compression presets
func (a *App) GetPresets() []PresetInfo {
	return []PresetInfo{
		{Name: "small", ScaleFactor: 0.5, MaxDimension: 720, Description: "小文件规格，适合社交媒体快速分享"},
		{Name: "standard", ScaleFactor: 1.0, MaxDimension: 1080, Description: "标准规格，画质与体积的完美平衡"},
		{Name: "quality", ScaleFactor: 1.0, MaxDimension: 0, Description: "高画质规格，保留极致视频细节"},
	}
}

// DetectFFmpeg checks if FFmpeg is available and returns its path
func (a *App) DetectFFmpeg() (string, error) {
	deps := engine.DefaultDependencies(a.executableDir)
	return deps.ResolveBinary(a.executableDir)
}

// DetectGPUEncoder auto-detects GPU hardware acceleration support
func (a *App) DetectGPUEncoder() (string, error) {
	ffmpegPath, err := a.DetectFFmpeg()
	if err != nil {
		return "libx264", err
	}
	deps := engine.DefaultDependencies(a.executableDir)
	encoder := deps.DetectGPUEncoder(ffmpegPath, deps.RunCommand)
	return encoder, nil
}

// StartCompress starts the compression process for the given files
func (a *App) StartCompress(req engine.JobRequest) ([]engine.JobReport, error) {
	deps := engine.DefaultDependencies(a.executableDir)
	eng := engine.NewCompressEngine(deps)

	onProgress := func(ev engine.ProgressEvent) {
		// Emit progress events to Svelte frontend
		runtime.EventsEmit(a.ctx, "progress", ev)
	}

	reports, err := eng.Run(req, onProgress)
	if err != nil {
		return nil, err
	}

	// Summarize results log
	reportsByDir := make(map[string][]engine.JobReport)
	for _, r := range reports {
		if r.OutputDir != "" {
			reportsByDir[r.OutputDir] = append(reportsByDir[r.OutputDir], r)
		}
	}
	// We can write to summary log as well (similar to CLI version)
	// (Implementation omitted or kept simple)

	return reports, nil
}

// InstallSendTo installs Windows SendTo right click menu binding
func (a *App) InstallSendTo() (string, error) {
	return sendto.Install(a.executablePath)
}

// UninstallSendTo removes Windows SendTo right click menu binding
func (a *App) UninstallSendTo() error {
	return sendto.Uninstall()
}

// AddToPath adds executable directory to user Path env
func (a *App) AddToPath() (bool, error) {
	return env.AddToPath(a.executableDir)
}

// RemoveFromPath removes executable directory from user Path env
func (a *App) RemoveFromPath() (bool, error) {
	return env.RemoveFromPath(a.executableDir)
}

// SelectFiles opens a file dialog and returns selected video paths
func (a *App) SelectFiles() ([]string, error) {
	options := runtime.OpenDialogOptions{
		Title: "选择视频文件",
		Filters: []runtime.FileFilter{
			{
				DisplayName: "视频文件 (*.mp4; *.mov; *.mkv; *.avi; *.webm)",
				Pattern:     "*.mp4;*.mov;*.mkv;*.avi;*.webm;*.m4v;*.wmv;*.ts;*.flv;*.mpg;*.mpeg;*.3gp",
			},
		},
	}
	return runtime.OpenMultipleFilesDialog(a.ctx, options)
}

// OpenFolder opens the target directory in explorer
func (a *App) OpenFolder(path string) error {
	deps := engine.DefaultDependencies(a.executableDir)
	// In Windows, explorer.exe can be launched to show folder
	return deps.RunCommand("explorer.exe", []string{filepath.Clean(path)})
}
