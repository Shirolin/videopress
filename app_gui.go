package main

import (
	"context"
	"os/exec"
	"path/filepath"
	"sync"

	"videopress/internal/app"
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
	initialFiles   []string
	mu             sync.Mutex
	cancelFunc     context.CancelFunc
}

// NewApp creates a new App struct instance
func NewApp(execDir, execPath string, initialFiles []string) *App {
	return &App{
		executableDir:  execDir,
		executablePath: execPath,
		initialFiles:   initialFiles,
	}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

// GetInitialFiles returns the initial file paths passed during application launch
func (a *App) GetInitialFiles() []string {
	a.mu.Lock()
	defer a.mu.Unlock()
	files := a.initialFiles
	a.initialFiles = nil // clear to prevent duplicate loads
	return files
}

// GetVersion returns the application version
func (a *App) GetVersion() string {
	return app.Version
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
		{Name: "small", ScaleFactor: 0.33, MaxDimension: 480, Description: "小文件规格，适合社交媒体快速分享"},
		{Name: "standard", ScaleFactor: 0.50, MaxDimension: 720, Description: "标准规格，画质与体积的完美平衡"},
		{Name: "quality", ScaleFactor: 1.00, MaxDimension: 0, Description: "高画质规格，保留极致视频细节"},
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
	deps.RunCommand = nil // 解锁引擎进度分析流水线
	eng := engine.NewCompressEngine(deps)

	onProgress := func(ev engine.ProgressEvent) {
		// Emit progress events to Svelte frontend
		runtime.EventsEmit(a.ctx, "progress", ev)
	}

	a.mu.Lock()
	ctx, cancel := context.WithCancel(context.Background())
	a.cancelFunc = cancel
	a.mu.Unlock()

	reports, err := eng.Run(ctx, req, onProgress)

	a.mu.Lock()
	if a.cancelFunc != nil {
		a.cancelFunc = nil
	}
	a.mu.Unlock()

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

// CancelCompress cancels the ongoing compression task
func (a *App) CancelCompress() {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.cancelFunc != nil {
		a.cancelFunc()
		a.cancelFunc = nil
	}
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

// SelectFolder opens a directory dialog and returns the selected folder path
func (a *App) SelectFolder() (string, error) {
	options := runtime.OpenDialogOptions{
		Title: "选择压缩后视频的保存目录",
	}
	return runtime.OpenDirectoryDialog(a.ctx, options)
}

// OpenFolder opens the target directory in explorer
func (a *App) OpenFolder(path string) error {
	cmd := exec.Command("explorer.exe", filepath.Clean(path))
	return cmd.Run()
}

// DownloadFFmpeg triggers the download and extraction of the ffmpeg binary
func (a *App) DownloadFFmpeg() error {
	onProgress := func(percent float64) {
		runtime.EventsEmit(a.ctx, "download-progress", percent)
	}

	err := ffmpeg.DownloadFFmpeg(a.executableDir, onProgress)
	if err != nil {
		runtime.EventsEmit(a.ctx, "download-progress", -1.0)
		return err
	}

	runtime.EventsEmit(a.ctx, "download-progress", 100.0)
	return nil
}

// InstallDesktopShortcut creates a desktop shortcut pointing to the application executable
func (a *App) InstallDesktopShortcut() error {
	return sendto.InstallDesktop(a.executablePath)
}

// UninstallDesktopShortcut removes the application shortcut from the user's desktop
func (a *App) UninstallDesktopShortcut() error {
	return sendto.UninstallDesktop()
}

// InstallStartMenuShortcut creates a shortcut directory in the Start Menu for the application
func (a *App) InstallStartMenuShortcut() error {
	return sendto.InstallStartMenu(a.executablePath)
}

// UninstallStartMenuShortcut removes the application shortcut directory from the Start Menu
func (a *App) UninstallStartMenuShortcut() error {
	return sendto.UninstallStartMenu()
}

// InstallContextMenu registers the "Compress with Videopress" context menu entry for all files
func (a *App) InstallContextMenu() error {
	return sendto.RegisterContextMenu(a.executablePath)
}

// UninstallContextMenu removes the "Compress with Videopress" context menu entry from the system registry
func (a *App) UninstallContextMenu() error {
	return sendto.UnregisterContextMenu()
}

// GetIntegrationStatus queries the current installation status of various desktop integrations
func (a *App) GetIntegrationStatus() (map[string]bool, error) {
	status := make(map[string]bool)
	status["sendto"] = sendto.IsSendToInstalled()
	status["desktop"] = sendto.IsDesktopInstalled()
	status["startmenu"] = sendto.IsStartMenuInstalled()
	status["contextmenu"] = sendto.IsContextMenuInstalled()

	isPath, _ := env.IsPathConfigured(a.executableDir)
	status["path"] = isPath

	return status, nil
}
