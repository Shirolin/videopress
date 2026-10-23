package gui

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"

	"videopress/internal/cli"
	"videopress/internal/engine"
	"videopress/internal/env"
	"videopress/internal/ffmpeg"
	"videopress/internal/gif"
	"videopress/internal/sendto"
	"videopress/internal/util"
)

// QueueItem is one row of the file queue.
type QueueItem struct {
	Path       string
	Name       string
	Size       int64
	TargetSize int64
	Percent    float64
	Status     string // waiting | compressing | success | failed | skipped
	Duration   int64  // milliseconds
	Error      string
	Warning    bool
}

// App holds the GUI state and the services the view calls.
type App struct {
	executableDir  string
	executablePath string
	win            *mygo.Window

	mu             sync.Mutex
	compressCancel context.CancelFunc
	compressGen    uint64
	gifCancel      context.CancelFunc
	gifGen         uint64

	// View state, touched on the main thread only.
	page            int
	queue           []QueueItem
	pageScroll      *ui.ScrollState
	settingsScroll  *ui.ScrollState
	settingsPage    string
	showAdvanced    bool
	isCompressing   bool
	isGifExporting  bool
	ffmpegError     string
	ffmpegPath      string
	gpuEncoder      string
	downloading     bool
	downloadPercent float64
	downloadError   string
	gifTier         string
	gifFormats      map[string]bool
	gifResults      []gif.ExportResult
	gifError        string
	integration     map[string]bool
	integrationBusy bool
	gifBusy         bool
	lastOutputDir   string
	toast           string
	toastKind       toastKind

	// locale is the language MyGO resolved at startup, kept so the view
	// need not ask the app (that call waits for the main thread).
	locale string

	// Preferences, persisted to settings.json.
	prefs prefs
}

// New creates a GUI application instance.
func New(execDir, execPath string, initialFiles []string) *App {
	p := loadPrefs()
	a := &App{
		executableDir:  execDir,
		executablePath: execPath,
		pageScroll:     &ui.ScrollState{},
		settingsScroll: &ui.ScrollState{},
		settingsPage:   "general",
		gifTier:        gif.DefaultTier,
		locale:         p.Language,
		gifFormats:     map[string]bool{"gif": true, "apng": true, "webp": true},
		integration:    map[string]bool{},
		prefs:          p,
	}
	a.addFiles(initialFiles)
	return a
}

// Attach stores the window the view belongs to.
func (a *App) Attach(win *mygo.Window) { a.win = win }

// update runs fn on the main thread and repaints.
func (a *App) update(fn func()) {
	if a.win == nil {
		fn()
		return
	}
	a.win.Update(fn)
}

// Version returns the application version.
func (a *App) Version() string { return cli.Version }

// lang reports the active UI language, "zh" or "en".
func (a *App) lang() string {
	if a.locale == "zh-Hans" {
		return "zh"
	}
	return "en"
}

// ---- queue ----

// addFiles appends the video files that are not queued yet.
func (a *App) addFiles(paths []string) {
	seen := make(map[string]bool, len(a.queue))
	for _, it := range a.queue {
		seen[filepath.Clean(it.Path)] = true
	}
	for _, p := range paths {
		if !util.IsVideoFile(p) {
			continue
		}
		key := filepath.Clean(p)
		if seen[key] {
			continue
		}
		seen[key] = true
		name := filepath.Base(p)
		warn := containsCompressed(name)
		item := QueueItem{Path: p, Name: name, Status: "waiting", Warning: warn}
		if warn {
			item.Error = "该视频可能已经被压缩过，二次压缩会损害画质。"
		}
		a.queue = append(a.queue, item)
	}
}

func containsCompressed(name string) bool { return strings.Contains(name, ".compressed") }

// removeQueue removes the item at index.
func (a *App) removeQueue(index int) {
	if index < 0 || index >= len(a.queue) {
		return
	}
	a.queue = append(a.queue[:index], a.queue[index+1:]...)
}

// clearQueue empties the queue.
func (a *App) clearQueue() {
	if a.isCompressing {
		return
	}
	a.queue = nil
}

// applyProgress updates the queue row that matches an engine event.
func (a *App) applyProgress(ev engine.ProgressEvent) {
	for i := range a.queue {
		item := &a.queue[i]
		match := false
		if ev.Path != "" {
			match = filepath.Clean(item.Path) == filepath.Clean(ev.Path)
		}
		if !match {
			match = item.Name == ev.File
		}
		if !match {
			continue
		}
		item.Percent = ev.Percent
		switch {
		case ev.Error != "":
			item.Status = "failed"
			item.Error = ev.Error
		case ev.Done:
			item.Status = "success"
		default:
			item.Status = "compressing"
		}
		return
	}
}

// ---- compression ----

// startCompress runs the queue through the engine.
func (a *App) startCompress() {
	if a.isCompressing || a.isGifExporting || len(a.queue) == 0 {
		return
	}
	files := make([]string, len(a.queue))
	for i, it := range a.queue {
		files[i] = it.Path
		a.queue[i].Percent = 0
		a.queue[i].Status = "waiting"
		a.queue[i].TargetSize = 0
		a.queue[i].Duration = 0
		if a.queue[i].Warning {
			a.queue[i].Error = "该视频可能已经被压缩过，二次压缩会损害画质。"
		} else {
			a.queue[i].Error = ""
		}
	}

	req := engine.JobRequest{
		Files:        files,
		Preset:       a.prefs.Preset,
		HWAccel:      a.prefs.HWAccel,
		CopyAudio:    a.prefs.CopyAudio,
		ForceMode:    a.prefs.ForceMode,
		SkipExisting: a.prefs.SkipExisting,
		Concurrency:  a.prefs.Concurrency,
		OutputDir:    a.prefs.OutputDir,
		VideoCodec:   a.prefs.VideoCodec,
		MaxFPS:       a.prefs.MaxFPS,
		AudioMode:    a.prefs.AudioMode,
		CRF:          a.prefs.crf(),
		Crop:         a.prefs.cropSpec(),
	}

	a.isCompressing = true
	a.toast = ""
	go a.runCompress(req)
}

func (a *App) runCompress(req engine.JobRequest) {
	deps := engine.DefaultDependencies(a.executableDir)
	deps.RunCommand = nil
	eng := engine.NewCompressEngine(deps)

	ctx, cancel := context.WithCancel(context.Background())

	a.mu.Lock()
	a.compressGen++
	gen := a.compressGen
	a.compressCancel = cancel
	a.mu.Unlock()

	reports, err := eng.Run(ctx, req, func(ev engine.ProgressEvent) {
		a.update(func() { a.applyProgress(ev) })
	})

	a.mu.Lock()
	if a.compressGen == gen {
		a.compressCancel = nil
	}
	a.mu.Unlock()

	a.update(func() {
		a.isCompressing = false
		if err != nil {
			a.setToast(err.Error(), toastError)
			return
		}
		a.applyReports(reports)
	})
}

// applyReports copies the engine results onto the queue rows.
func (a *App) applyReports(reports []engine.JobReport) {
	for _, r := range reports {
		for i := range a.queue {
			item := &a.queue[i]
			match := item.Name == r.InputName
			if r.InputPath != "" {
				match = filepath.Clean(item.Path) == filepath.Clean(r.InputPath)
			}
			if !match {
				continue
			}
			item.Size = r.SourceSize
			item.TargetSize = r.TargetSize
			item.Duration = r.Duration.Milliseconds()
			item.Error = r.ErrMessage
			switch r.Status {
			case "skipped":
				item.Status = "skipped"
				item.Percent = 100
			case "failed":
				item.Status = "failed"
			default:
				item.Status = "success"
				item.Percent = 100
			}
			if r.Status == "success" && r.OutputDir != "" {
				a.lastOutputDir = r.OutputDir
			}
			break
		}
	}
}

// cancelCompress stops the running batch.
func (a *App) cancelCompress() {
	a.mu.Lock()
	cancel := a.compressCancel
	a.compressCancel = nil
	a.mu.Unlock()
	if cancel != nil {
		cancel()
	}
}

// ---- animated export ----

// startGifExport exports the queue as animated clips.
func (a *App) startGifExport() {
	if a.isGifExporting || a.isCompressing || len(a.queue) == 0 {
		return
	}
	var formats []gif.Format
	for _, f := range gif.AllFormats() {
		if a.gifFormats[string(f)] {
			formats = append(formats, f)
		}
	}
	if len(formats) == 0 {
		a.gifError = a.t("gif.need_format")
		return
	}
	files := make([]string, len(a.queue))
	for i, it := range a.queue {
		files[i] = it.Path
	}
	req := gif.ExportRequest{
		Files:     files,
		Tier:      a.gifTier,
		Formats:   formats,
		OutputDir: a.prefs.OutputDir,
		Force:     a.prefs.ForceMode,
	}
	a.isGifExporting = true
	a.gifBusy = true
	a.gifError = ""
	a.gifResults = nil
	go a.runGifExport(req)
}

func (a *App) runGifExport(req gif.ExportRequest) {
	deps := gif.DefaultDependencies(a.executableDir)
	deps.RunCommand = nil
	eng := gif.New(deps)

	ctx, cancel := context.WithCancel(context.Background())

	a.mu.Lock()
	a.gifGen++
	gen := a.gifGen
	a.gifCancel = cancel
	a.mu.Unlock()

	results, err := eng.Run(ctx, req)

	a.mu.Lock()
	if a.gifGen == gen {
		a.gifCancel = nil
	}
	a.mu.Unlock()

	a.update(func() {
		a.isGifExporting = false
		a.gifBusy = false
		if err != nil {
			a.gifError = err.Error()
			return
		}
		a.gifResults = results
	})
}

// cancelGifExport stops the running export.
func (a *App) cancelGifExport() {
	a.mu.Lock()
	cancel := a.gifCancel
	a.gifCancel = nil
	a.mu.Unlock()
	if cancel != nil {
		cancel()
	}
}

// ---- FFmpeg ----

// detectFFmpeg refreshes the FFmpeg path, the setup screen follows it.
func (a *App) detectFFmpeg() {
	path, err := engine.DefaultDependencies(a.executableDir).ResolveBinary(a.executableDir)
	if err != nil {
		a.ffmpegPath = ""
		a.ffmpegError = a.t("app.ffmpeg_not_found")
		return
	}
	a.ffmpegPath = path
	a.ffmpegError = ""
}

// detectGPU probes the graphics encoder in the background.
func (a *App) detectGPU() {
	go func() {
		deps := engine.DefaultDependencies(a.executableDir)
		path, err := deps.ResolveBinary(a.executableDir)
		if err != nil {
			return
		}
		enc := deps.DetectGPUEncoder(path, a.prefs.VideoCodec, deps.RunCommand)
		a.update(func() {
			a.gpuEncoder = enc
			if enc != "libx264" && !a.prefs.HWAccelSet {
				a.prefs.HWAccel = true
				a.prefs.HWAccelSet = true
				a.savePrefs()
			}
		})
	}()
}

// downloadFFmpeg fetches and installs FFmpeg next to the executable.
func (a *App) downloadFFmpeg() {
	if a.downloading {
		return
	}
	a.downloading = true
	a.downloadPercent = 0
	a.downloadError = ""
	go func() {
		err := ffmpeg.DownloadFFmpeg(a.executableDir, func(percent float64) {
			a.update(func() { a.downloadPercent = percent })
		})
		a.update(func() {
			a.downloading = false
			if err != nil {
				a.downloadError = err.Error()
				return
			}
			a.detectFFmpeg()
		})
	}()
}

// setDebug toggles the engine's debug log.
func (a *App) setDebug(on bool) {
	a.prefs.EnableDebugLog = on
	ffmpeg.EnableDebugLog = on
	a.savePrefs()
}

// debugLogPath is the diagnostics log file.
func debugLogPath() (string, error) {
	dir, err := os.UserCacheDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "videopress_debug.log"), nil
}

// openDebugLog opens the debug log in the default editor.
func (a *App) openDebugLog() {
	logFile, err := debugLogPath()
	if err != nil {
		a.setToast(a.tf("toast.log_open_failed", map[string]string{"err": err.Error()}), toastError)
		return
	}
	_ = os.WriteFile(logFile, nil, 0o644)
	if err := exec.Command("cmd", "/c", "start", "", logFile).Start(); err != nil {
		a.setToast(a.tf("toast.log_open_failed", map[string]string{"err": err.Error()}), toastError)
		return
	}
	a.setToast(a.t("toast.log_opened"), toastSuccess)
}

// clearDebugLog wipes the log and the GPU cache.
func (a *App) clearDebugLog() {
	dir, err := os.UserCacheDir()
	if err != nil {
		a.setToast(a.tf("toast.log_clear_failed", map[string]string{"err": err.Error()}), toastError)
		return
	}
	_ = os.Remove(filepath.Join(dir, "videopress_debug.log"))
	_ = os.Remove(filepath.Join(dir, "videopress_gpu.cache"))
	ffmpeg.ResetGPUEncoderCache()
	a.detectGPU()
	a.setToast(a.t("toast.log_cleared"), toastSuccess)
}

// ---- system integration ----

// refreshIntegration rereads the Windows integration state.
func (a *App) refreshIntegration() {
	go func() {
		status := map[string]bool{
			"sendto":      sendto.IsSendToInstalled(),
			"desktop":     sendto.IsDesktopInstalled(),
			"startmenu":   sendto.IsStartMenuInstalled(),
			"contextmenu": sendto.IsContextMenuInstalled(),
		}
		onPath, _ := env.IsPathConfigured(a.executableDir)
		status["path"] = onPath
		a.update(func() {
			a.integration = status
			a.integrationBusy = false
		})
	}()
}

// toggleIntegration installs or uninstalls one integration.
func (a *App) toggleIntegration(kind string) {
	if a.integrationBusy {
		return
	}
	a.integrationBusy = true
	install := !a.integration[kind]
	go func() {
		err := a.applyIntegration(kind, install)
		a.update(func() {
			if err != nil {
				a.integrationBusy = false
				a.setToast(a.tf("toast.action_failed", map[string]string{"err": err.Error()}), toastError)
				return
			}
			a.integrationBusy = false
			a.setToast(a.t("toast."+kind+"_"+verb(install)), toastSuccess)
			a.refreshIntegration()
		})
	}()
}

func verb(install bool) string {
	if install {
		return "added"
	}
	return "removed"
}

func (a *App) applyIntegration(kind string, install bool) error {
	switch kind {
	case "sendto":
		if install {
			_, err := sendto.Install(a.executablePath)
			return err
		}
		return sendto.Uninstall()
	case "desktop":
		if install {
			return sendto.InstallDesktop(a.executablePath)
		}
		return sendto.UninstallDesktop()
	case "startmenu":
		if install {
			return sendto.InstallStartMenu(a.executablePath)
		}
		return sendto.UninstallStartMenu()
	case "contextmenu":
		if install {
			return sendto.RegisterContextMenu(a.executablePath, a.lang())
		}
		return sendto.UnregisterContextMenu()
	case "path":
		if install {
			_, err := env.AddToPath(a.executableDir)
			return err
		}
		_, err := env.RemoveFromPath(a.executableDir)
		return err
	}
	return fmt.Errorf("未知的系统集成 unknown integration: %s", kind)
}

// ---- dialogs ----

// pickFiles asks for video files and queues them.
func (a *App) pickFiles() {
	go func() {
		paths, err := mygo.Dialog.Open(mygo.OpenDialogOptions{
			Parent:   a.win,
			Title:    a.t("dropzone.title"),
			Multiple: true,
			Filters:  []mygo.FileFilter{{Name: a.t("dropzone.formats"), Extensions: videoExtensions}},
		})
		if err != nil || len(paths) == 0 {
			return
		}
		a.update(func() { a.addFiles(paths) })
	}()
}

// pickOutputDir asks for the output folder.
func (a *App) pickOutputDir() {
	go func() {
		paths, err := mygo.Dialog.Open(mygo.OpenDialogOptions{Parent: a.win, Directory: true, Title: a.t("controls.output")})
		if err != nil || len(paths) == 0 {
			return
		}
		a.update(func() {
			a.prefs.OutputDir = paths[0]
			a.savePrefs()
		})
	}()
}

// openFolder reveals a directory in Explorer.
func (a *App) openFolder(path string) {
	if path == "" {
		return
	}
	_ = exec.Command("explorer.exe", filepath.Clean(path)).Start()
}

// openResultFolder opens the folder of the finished batch.
func (a *App) openResultFolder() {
	if a.prefs.OutputDir != "" {
		a.openFolder(a.prefs.OutputDir)
		return
	}
	if a.lastOutputDir != "" {
		a.openFolder(a.lastOutputDir)
		return
	}
	for _, it := range a.queue {
		if it.Status == "success" {
			a.openFolder(filepath.Join(filepath.Dir(it.Path), "compressed"))
			return
		}
	}
}

var videoExtensions = []string{"mp4", "mov", "mkv", "avi", "m4v", "wmv", "webm", "ts", "flv", "mpg", "mpeg", "3gp"}

// ---- stats ----

// stats sums the finished rows for the summary bar.
func (a *App) stats() (total, done, saved, original int64) {
	total = int64(len(a.queue))
	for _, it := range a.queue {
		if it.Status != "success" {
			continue
		}
		done++
		original += it.Size
		if it.Size > it.TargetSize {
			saved += it.Size - it.TargetSize
		}
	}
	return total, done, saved, original
}

// concurrencyOptions lists plausible worker counts for this machine.
func concurrencyOptions(cores int) []int {
	out := []int{1}
	for _, n := range []int{2, 4, 8} {
		if cores >= n {
			out = append(out, n)
		}
	}
	if cores > 8 {
		out = append(out, cores)
	}
	return out
}

// sizeText formats a byte count for the queue rows.
func sizeText(bytes int64) string {
	if bytes <= 0 {
		return "-"
	}
	mb := float64(bytes) / (1024 * 1024)
	if mb < 1024 {
		return strconv.FormatFloat(mb, 'f', 1, 64) + " MB"
	}
	return strconv.FormatFloat(mb/1024, 'f', 2, 64) + " GB"
}

// ratioText is the share of the source the batch saved.
func ratioText(saved, original int64) string {
	if original <= 0 {
		return "0.0%"
	}
	return strconv.FormatFloat(float64(saved)/float64(original)*100, 'f', 1, 64) + "%"
}
