package gui

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"time"

	"videopress/internal/cli"
	"videopress/internal/engine"
	"videopress/internal/env"
	"videopress/internal/ffmpeg"
	"videopress/internal/gif"
	"videopress/internal/locale"
	"videopress/internal/sendto"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// App handles Wails GUI bindings.
type App struct {
	ctx            context.Context
	executableDir  string
	executablePath string
	initialFiles   []string
	mu             sync.Mutex
	compressCancel context.CancelFunc
	compressGen    uint64
	gifCancel      context.CancelFunc
	gifGen         uint64
	enableDebug    bool
	language       string
}

// New creates a GUI application instance.
func New(execDir, execPath string, initialFiles []string) *App {
	return &App{
		executableDir:  execDir,
		executablePath: execPath,
		initialFiles:   initialFiles,
		language:       locale.System(),
	}
}

// Startup is called when the Wails app starts.
func (a *App) Startup(ctx context.Context) {
	a.ctx = ctx
}

// GetInitialFiles returns the initial file paths passed during application launch.
func (a *App) GetInitialFiles() []string {
	a.mu.Lock()
	defer a.mu.Unlock()
	files := a.initialFiles
	a.initialFiles = nil
	return files
}

// GetVersion returns the application version.
func (a *App) GetVersion() string {
	return cli.Version
}

// PresetInfo represents preset metadata returned to frontend.
type PresetInfo struct {
	Name         string  `json:"name"`
	ScaleFactor  float64 `json:"scaleFactor"`
	MaxDimension int     `json:"maxDimension"`
	Description  string  `json:"description"`
}

// GetPresets returns the list of compression presets.
func (a *App) GetPresets() []PresetInfo {
	a.mu.Lock()
	lang := a.language
	a.mu.Unlock()

	descSmall := "小文件规格，适合社交媒体快速分享"
	descStandard := "标准规格，画质与体积的完美平衡"
	descQuality := "高画质规格，保留极致视频细节"

	if lang == "en" {
		descSmall = "Small file spec, perfect for social media sharing"
		descStandard = "Standard spec, ideal balance of quality and size"
		descQuality = "High quality spec, preserves maximum video details"
	}

	return []PresetInfo{
		{Name: "small", ScaleFactor: 0.33, MaxDimension: 480, Description: descSmall},
		{Name: "standard", ScaleFactor: 0.50, MaxDimension: 720, Description: descStandard},
		{Name: "quality", ScaleFactor: 1.00, MaxDimension: 0, Description: descQuality},
	}
}

// DetectFFmpeg checks if FFmpeg is available and returns its path.
func (a *App) DetectFFmpeg() (string, error) {
	deps := engine.DefaultDependencies(a.executableDir)
	return deps.ResolveBinary(a.executableDir)
}

// DetectGPUEncoder auto-detects GPU hardware acceleration support.
func (a *App) DetectGPUEncoder() (string, error) {
	ffmpegPath, err := a.DetectFFmpeg()
	if err != nil {
		return "libx264", err
	}
	deps := engine.DefaultDependencies(a.executableDir)
	encoder := deps.DetectGPUEncoder(ffmpegPath, "h264", deps.RunCommand)
	return encoder, nil
}

// StartCompress starts the compression process for the given files.
func (a *App) StartCompress(req engine.JobRequest) ([]engine.JobReport, error) {
	deps := engine.DefaultDependencies(a.executableDir)
	deps.RunCommand = nil
	eng := engine.NewCompressEngine(deps)

	onProgress := func(ev engine.ProgressEvent) {
		runtime.EventsEmit(a.ctx, "progress", ev)
	}

	a.mu.Lock()
	ctx, cancel := context.WithCancel(context.Background())
	a.compressGen++
	compressGen := a.compressGen
	a.compressCancel = cancel
	a.mu.Unlock()

	reports, err := eng.Run(ctx, req, onProgress)

	a.mu.Lock()
	if a.compressGen == compressGen {
		a.compressCancel = nil
	}
	a.mu.Unlock()

	if err != nil {
		return nil, err
	}

	return reports, nil
}

// CancelCompress cancels the ongoing compression task.
func (a *App) CancelCompress() {
	a.mu.Lock()
	cancel := a.compressCancel
	a.compressCancel = nil
	a.mu.Unlock()
	if cancel != nil {
		cancel()
	}
}

// CancelGifExport cancels the ongoing animated export task.
func (a *App) CancelGifExport() {
	a.mu.Lock()
	cancel := a.gifCancel
	a.gifCancel = nil
	a.mu.Unlock()
	if cancel != nil {
		cancel()
	}
}

type GifTierInfo struct {
	Name        string `json:"name"`
	MaxWidth    int    `json:"maxWidth"`
	MaxSizeMB   int    `json:"maxSizeMB"`
	FPS         int    `json:"fps"`
	MaxDuration string `json:"maxDuration"`
	Description string `json:"description"`
	IsDefault   bool   `json:"isDefault"`
}

// GetGifTiers returns the list of animated export tiers.
func (a *App) GetGifTiers() []GifTierInfo {
	a.mu.Lock()
	lang := a.language
	a.mu.Unlock()

	infos := make([]GifTierInfo, 0, 3)
	for _, t := range gif.AllTiers() {
		desc := t.Description
		if lang == "en" {
			switch t.Name {
			case "smooth":
				desc = "Small smooth tier, instant sharing"
			case "balanced":
				desc = "Balanced tier (default), size and quality balance"
			case "hd":
				desc = "HD tier, keeps more detail"
			}
		}
		infos = append(infos, GifTierInfo{
			Name:        t.Name,
			MaxWidth:    t.MaxWidth,
			MaxSizeMB:   t.MaxSizeMB,
			FPS:         t.FPS,
			MaxDuration: t.MaxDuration,
			Description: desc,
			IsDefault:   t.Name == gif.DefaultTier,
		})
	}
	return infos
}

// GetGifFormats returns the supported animated output formats.
func (a *App) GetGifFormats() []string {
	out := make([]string, 0, 3)
	for _, f := range gif.AllFormats() {
		out = append(out, string(f))
	}
	return out
}

// StartGifExport starts animated (GIF/APNG/WebP) export for the given files.
func (a *App) StartGifExport(req gif.ExportRequest) ([]gif.ExportResult, error) {
	deps := gif.DefaultDependencies(a.executableDir)
	deps.RunCommand = nil
	eng := gif.New(deps)

	a.mu.Lock()
	ctx, cancel := context.WithCancel(context.Background())
	a.gifGen++
	gifGen := a.gifGen
	a.gifCancel = cancel
	a.mu.Unlock()

	results, err := eng.Run(ctx, req)

	a.mu.Lock()
	if a.gifGen == gifGen {
		a.gifCancel = nil
	}
	a.mu.Unlock()

	if err != nil {
		return nil, err
	}
	return results, nil
}

// InstallSendTo installs Windows SendTo right click menu binding.
func (a *App) InstallSendTo() (string, error) {
	return sendto.Install(a.executablePath)
}

// UninstallSendTo removes Windows SendTo right click menu binding.
func (a *App) UninstallSendTo() error {
	return sendto.Uninstall()
}

// AddToPath adds executable directory to user Path env.
func (a *App) AddToPath() (bool, error) {
	return env.AddToPath(a.executableDir)
}

// RemoveFromPath removes executable directory from user Path env.
func (a *App) RemoveFromPath() (bool, error) {
	return env.RemoveFromPath(a.executableDir)
}

// SelectFiles opens a file dialog and returns selected video paths.
func (a *App) SelectFiles() ([]string, error) {
	a.mu.Lock()
	lang := a.language
	a.mu.Unlock()

	title := "选择视频文件"
	filterName := "视频文件 (*.mp4; *.mov; *.mkv; *.avi; *.webm)"
	if lang == "en" {
		title = "Select Video Files"
		filterName = "Video Files (*.mp4; *.mov; *.mkv; *.avi; *.webm)"
	}

	options := runtime.OpenDialogOptions{
		Title: title,
		Filters: []runtime.FileFilter{
			{
				DisplayName: filterName,
				Pattern:     "*.mp4;*.mov;*.mkv;*.avi;*.webm;*.m4v;*.wmv;*.ts;*.flv;*.mpg;*.mpeg;*.3gp",
			},
		},
	}
	return runtime.OpenMultipleFilesDialog(a.ctx, options)
}

// SelectFolder opens a directory dialog and returns the selected folder path.
func (a *App) SelectFolder() (string, error) {
	a.mu.Lock()
	lang := a.language
	a.mu.Unlock()

	title := "选择压缩后视频的保存目录"
	if lang == "en" {
		title = "Select Save Directory for Compressed Videos"
	}

	options := runtime.OpenDialogOptions{
		Title: title,
	}
	return runtime.OpenDirectoryDialog(a.ctx, options)
}

// OpenFolder opens the target directory in explorer.
func (a *App) OpenFolder(path string) error {
	cmd := exec.Command("explorer.exe", filepath.Clean(path))
	return cmd.Run()
}

// DownloadFFmpeg triggers the download and extraction of the ffmpeg binary.
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

// InstallDesktopShortcut creates a desktop shortcut pointing to the application executable.
func (a *App) InstallDesktopShortcut() error {
	return sendto.InstallDesktop(a.executablePath)
}

// UninstallDesktopShortcut removes the application shortcut from the user's desktop.
func (a *App) UninstallDesktopShortcut() error {
	return sendto.UninstallDesktop()
}

// InstallStartMenuShortcut creates a shortcut directory in the Start Menu for the application.
func (a *App) InstallStartMenuShortcut() error {
	return sendto.InstallStartMenu(a.executablePath)
}

// UninstallStartMenuShortcut removes the application shortcut directory from the Start Menu.
func (a *App) UninstallStartMenuShortcut() error {
	return sendto.UninstallStartMenu()
}

// InstallContextMenu registers the context menu entry for all files.
func (a *App) InstallContextMenu() error {
	a.mu.Lock()
	lang := a.language
	a.mu.Unlock()
	return sendto.RegisterContextMenu(a.executablePath, lang)
}

// UninstallContextMenu removes the context menu entry from the system registry.
func (a *App) UninstallContextMenu() error {
	return sendto.UnregisterContextMenu()
}

// GetIntegrationStatus queries the current installation status of desktop integrations.
func (a *App) GetIntegrationStatus() (map[string]bool, error) {
	start := time.Now()
	status := make(map[string]bool)

	t := time.Now()
	status["sendto"] = sendto.IsSendToInstalled()
	sendToTime := time.Since(t)

	t = time.Now()
	status["desktop"] = sendto.IsDesktopInstalled()
	desktopTime := time.Since(t)

	t = time.Now()
	status["startmenu"] = sendto.IsStartMenuInstalled()
	startMenuTime := time.Since(t)

	t = time.Now()
	status["contextmenu"] = sendto.IsContextMenuInstalled()
	contextMenuTime := time.Since(t)

	t = time.Now()
	isPath, _ := env.IsPathConfigured(a.executableDir)
	status["path"] = isPath
	pathTime := time.Since(t)

	totalTime := time.Since(start)

	a.mu.Lock()
	debugEnabled := a.enableDebug
	a.mu.Unlock()

	if debugEnabled {
		logMsg := fmt.Sprintf("[%s] 配置项载入耗时统计 (总耗时: %v):\n"+
			"- SendTo 右键发送菜单检测: %v\n"+
			"- 桌面快捷方式检测: %v\n"+
			"- 开始菜单快捷方式检测: %v\n"+
			"- 右键注册表项菜单检测: %v\n"+
			"- 环境变量 Path 检测: %v\n\n",
			time.Now().Format("2006-01-02 15:04:05"),
			totalTime,
			sendToTime,
			desktopTime,
			startMenuTime,
			contextMenuTime,
			pathTime,
		)

		cacheDir, err := os.UserCacheDir()
		if err == nil {
			logFile := filepath.Join(cacheDir, "videopress_debug.log")
			f, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
			if err == nil {
				_, _ = f.WriteString(logMsg)
				_ = f.Close()
			}
		}
	}

	return status, nil
}

// SetDebugMode sets whether debug logging is enabled.
func (a *App) SetDebugMode(enable bool) {
	a.mu.Lock()
	a.enableDebug = enable
	a.mu.Unlock()
	ffmpeg.EnableDebugLog = enable
}

// GetDebugLogs returns the contents of the debug log.
func (a *App) GetDebugLogs() (string, error) {
	cacheDir, err := os.UserCacheDir()
	if err != nil {
		return "", err
	}
	logFile := filepath.Join(cacheDir, "videopress_debug.log")
	data, err := os.ReadFile(logFile)
	if err != nil {
		if os.IsNotExist(err) {
			return "当前暂无调试日志记录", nil
		}
		return "", err
	}
	return string(data), nil
}

// ClearDebugLogs clears all debug logs.
func (a *App) ClearDebugLogs() error {
	cacheDir, err := os.UserCacheDir()
	if err != nil {
		return err
	}
	logFile := filepath.Join(cacheDir, "videopress_debug.log")
	_ = os.Remove(logFile)
	gpuCache := filepath.Join(cacheDir, "videopress_gpu.cache")
	_ = os.Remove(gpuCache)
	ffmpeg.ResetGPUEncoderCache()
	return nil
}

// OpenDebugLogFile opens the debug log file in default text editor.
func (a *App) OpenDebugLogFile() error {
	cacheDir, err := os.UserCacheDir()
	if err != nil {
		return err
	}
	logFile := filepath.Join(cacheDir, "videopress_debug.log")
	if _, err := os.Stat(logFile); os.IsNotExist(err) {
		_ = os.WriteFile(logFile, []byte(""), 0o644)
	}
	cmd := exec.Command("cmd", "/c", "start", "", logFile)
	return cmd.Run()
}

// SetLanguage sets the UI language and updates hot-reloadable integrations.
func (a *App) SetLanguage(lang string) {
	a.mu.Lock()
	a.language = lang
	a.mu.Unlock()

	if sendto.IsContextMenuInstalled() {
		_ = sendto.RegisterContextMenu(a.executablePath, lang)
	}
}

// GetLanguage returns the active UI language.
func (a *App) GetLanguage() string {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.language
}
