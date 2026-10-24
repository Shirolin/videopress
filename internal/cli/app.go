package cli

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"videopress/internal/compress"
	"videopress/internal/engine"
	"videopress/internal/ffmpeg"
	"videopress/internal/gif"
	"videopress/internal/locale"
	"videopress/internal/notify"
	"videopress/internal/util"
)

// Version 应用版本号，可在构建时通过 -ldflags "-X videopress/internal/cli.Version=..." 注入。
var Version = "0.1.0"

var EnableConsoleColors = func() {}

var (
	systemLangOnce sync.Once
	cachedLang     string
	testLang       string // 仅供测试注入，非空时优先返回，避免 CI 多语言环境断言漂移
)

// systemLanguage 返回系统 UI 语言（zh/en），仅首次调用触发 syscall，之后走缓存。
func systemLanguage() string {
	systemLangOnce.Do(func() {
		if testLang != "" {
			cachedLang = testLang
			return
		}
		cachedLang = locale.System()
	})
	return cachedLang
}

// setSystemLanguageForTest 仅供测试覆盖系统语言检测结果，保证多语言环境下的断言稳定。
func setSystemLanguageForTest(lang string) {
	testLang = lang
	systemLangOnce = sync.Once{}
	cachedLang = ""
}

func colorize(text string, colorCode string) string {
	return fmt.Sprintf("\033[%sm%s\033[0m", colorCode, text)
}

func green(text string) string   { return colorize(text, "32") }
func red(text string) string     { return colorize(text, "31") }
func yellow(text string) string  { return colorize(text, "33") }
func magenta(text string) string { return colorize(text, "35") }
func cyan(text string) string    { return colorize(text, "36") }
func gray(text string) string    { return colorize(text, "90") }

type Dependencies struct {
	ExecutableDir    string
	ExecutablePath   string
	ResolveBinary    func(dir string) (string, error)
	RunCommand       func(name string, args []string) error
	DetectGPUEncoder func(ffmpegPath string, codec string, runCmd func(string, []string) error) string
	MkdirAll         func(path string, perm os.FileMode) error
	PathExists       func(path string) bool
	InputAccessible  func(path string) bool
	Stdout           io.Writer
	Stderr           io.Writer
	Stdin            io.Reader
	InstallSendTo    func(executablePath string) (string, error)
	UninstallSendTo  func() error
	AddToPath        func(dir string) (bool, error)
	RemoveFromPath   func(dir string) (bool, error)
	// SendToNotify 在 SendTo 静默压缩结束时发送系统通知（title/body 已按语言生成）。
	SendToNotify func(executablePath string, title, body string) error
}

type JobReport = engine.JobReport

func getMsg(zh, en string) string {
	if systemLanguage() == "en" {
		return en
	}
	return zh
}

// sendToNotify 发送 SendTo 静默压缩的系统通知，并确保 Toast 所需的 AUMID 快捷方式存在。
func sendToNotify(executablePath string, title, body string) error {
	if executablePath != "" {
		_ = notify.EnsureShortcut(executablePath)
	}
	return notify.Notify(title, body)
}

func sendToResultTitle(successes, failures int) string {
	if failures > 0 {
		return msg("notify.done_failures")
	}
	return msg("notify.done")
}

func sendToResultBody(successes, failures int) string {
	return msg("notify.result_body", successes, failures)
}

func printUsage(w io.Writer) {
	rule := magenta("========================================")
	fmt.Fprintln(w, rule)
	fmt.Fprintln(w, magenta(msg("usage.title")))
	fmt.Fprintln(w, rule)
	fmt.Fprintln(w, msg("usage.options"))
	fmt.Fprintf(w, "  --preset %s  %s\n", cyan("small|standard|quality"), msg("usage.preset"))
	fmt.Fprintf(w, "  --concurrency, -c %s         %s\n", cyan(msg("usage.number")), msg("usage.concurrency"))
	fmt.Fprintf(w, "  --hw                             %s\n", cyan(msg("usage.hw")))
	fmt.Fprintln(w, "  --force, -f                      "+msg("usage.force"))
	fmt.Fprintln(w, "  --skip-existing                  "+msg("usage.skip_existing"))
	fmt.Fprintf(w, "  --codec %s         %s\n", cyan("h264|h265|av1"), msg("usage.codec"))
	fmt.Fprintf(w, "  --max-fps %s           %s\n", cyan(msg("usage.number")), msg("usage.max_fps"))
	fmt.Fprintf(w, "  --audio %s   %s\n", cyan("copy|compress|mute"), msg("usage.audio"))
	fmt.Fprintln(w, "  --copy-audio, -a                 "+msg("usage.copy_audio"))
	fmt.Fprintf(w, "  --crf %s               %s\n", cyan(msg("usage.number")), msg("usage.crf"))
	fmt.Fprintf(w, "  --crop %s           %s\n", cyan("W:H:X:Y"), msg("usage.crop"))
	fmt.Fprintf(w, "  --crop-center %s       %s\n", cyan("W:H"), msg("usage.crop_center"))
	fmt.Fprintf(w, "  --crop-scale %s      %s\n", cyan("<0.5-1.0>"), msg("usage.crop_scale"))
	fmt.Fprintf(w, "  --gif                             %s\n", cyan(msg("usage.gif")))
	fmt.Fprintf(w, "  --tier %s          %s\n", cyan("smooth|balanced|hd"), msg("usage.tier"))
	fmt.Fprintf(w, "  --format %s    %s\n", cyan("gif|apng|webp"), msg("usage.format"))
	fmt.Fprintln(w, "  --install-sendto                 "+msg("usage.install_sendto"))
	fmt.Fprintln(w, "  --uninstall-sendto               "+msg("usage.uninstall_sendto"))
	fmt.Fprintln(w, "  --install-path                   "+msg("usage.install_path"))
	fmt.Fprintln(w, "  --uninstall-path                 "+msg("usage.uninstall_path"))
	fmt.Fprintln(w, "  --version                        "+msg("usage.version"))
	fmt.Fprintln(w, "  -h, --help                       "+msg("usage.help"))
	fmt.Fprintln(w)
	fmt.Fprintln(w, msg("usage.exit_codes"))
	fmt.Fprintf(w, "  0  %s\n", green(msg("usage.exit_0")))
	fmt.Fprintf(w, "  1  %s\n", red(msg("usage.exit_1")))
}

// gifFormatList 收集 --format 多次指定的动图格式。
type gifFormatList []gif.Format

func (l *gifFormatList) String() string {
	return fmt.Sprint([]gif.Format(*l))
}

func (l *gifFormatList) Set(v string) error {
	f := gif.Format(strings.ToLower(v))
	switch f {
	case gif.FormatGIF, gif.FormatAPNG, gif.FormatWebP:
		*l = append(*l, f)
		return nil
	default:
		return fmt.Errorf(msg("err.gif_format"), v)
	}
}

// runGifExport 执行 GIF/APNG/WebP 动图导出。
func runGifExport(deps Dependencies, files []string, tierName string, formats gifFormatList, force bool) int {
	tier, err := gif.TierByName(tierName)
	if err != nil {
		fmt.Fprintln(deps.Stderr, red(err.Error()))
		return 1
	}

	// ffmpeg 二进制解析交给引擎（Run 内统一 resolve），此处不预检。

	rule := magenta("========================================")
	fmt.Fprintln(deps.Stdout, rule)
	fmt.Fprintln(deps.Stdout, magenta(msg("gif.title")))
	fmt.Fprintln(deps.Stdout, rule)
	fmt.Fprintf(deps.Stdout, msg("gif.tier"), cyan(tier.Name))
	fmt.Fprintf(deps.Stdout, msg("gif.max_width"), cyan(fmt.Sprintf("%dpx", tier.MaxWidth)))
	fmt.Fprintf(deps.Stdout, msg("gif.max_size"), cyan(fmt.Sprintf("%dMB", tier.MaxSizeMB)))
	fmt.Fprintf(deps.Stdout, msg("gif.fps"), cyan(fmt.Sprintf("%dfps", tier.FPS)))
	fmt.Fprintf(deps.Stdout, msg("gif.max_duration"), cyan(tier.MaxDuration))
	fmt.Fprintln(deps.Stdout, rule)
	fmt.Fprintln(deps.Stdout)

	engDeps := gif.Dependencies{
		ExecutableDir: deps.ExecutableDir,
		ResolveBinary: deps.ResolveBinary,
		MkdirAll:      deps.MkdirAll,
		PathExists:    deps.PathExists,
	}
	if deps.RunCommand != nil {
		// 仅测试注入 mock 时走此分支（生产恒为 nil，用引擎默认 ctx 实现）；
		// mock 同步执行，ctx 仅用于探测/编码流程控制。
		cliRun := deps.RunCommand
		engDeps.RunCommand = func(ctx context.Context, name string, args []string) error {
			return cliRun(name, args)
		}
	}
	eng := gif.New(engDeps)

	req := gif.ExportRequest{
		Files:   files,
		Tier:    tier.Name,
		Formats: formats,
		Force:   force,
	}
	if len(req.Formats) == 0 {
		req.Formats = gif.DefaultFormats()
	}

	// Ctrl+C 可中断探测与编码（生产路径引擎默认实现尊重 ctx）。
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	results, err := eng.Run(ctx, req)

	successes := 0
	skipped := 0
	failures := 0
	for _, r := range results {
		disp := fmt.Sprintf("%s [%s/%s]", filepath.Base(r.OutputPath), r.Tier, r.Format)
		switch r.Status {
		case "success":
			successes++
			over := ""
			if r.OverBudget {
				over = yellow(msg("gif.over_budget"))
			}
			fmt.Fprintf(deps.Stdout, "%s %s -> %s%s\n", green(msg("gif.success")), disp, fmt.Sprintf("%.2fMB", float64(r.Size)/(1024*1024)), over)
		case "skipped":
			skipped++
			overSkipped := ""
			if r.OverBudget {
				overSkipped = yellow(msg("gif.over_budget"))
			}
			fmt.Fprintf(deps.Stdout, "%s %s %s%s\n", yellow(msg("gif.skipped")), disp, msg("gif.exists"), overSkipped)
		default:
			failures++
			fmt.Fprintf(deps.Stderr, "%s %s: %s\n", red(msg("gif.failed")), disp, red(r.Error))
		}
	}

	fmt.Fprintln(deps.Stdout, magenta("=========================================="))
	fmt.Fprintf(deps.Stdout, msg("gif.summary"), green(fmt.Sprintf("%d", successes)), yellow(fmt.Sprintf("%d", skipped)), red(fmt.Sprintf("%d", failures)))

	exitCode := 0
	if failures > 0 || successes == 0 {
		exitCode = 1
	}
	return exitCode
}

func Execute(args []string, deps Dependencies) int {
	EnableConsoleColors()

	if deps.Stdout == nil {
		deps.Stdout = io.Discard
	}
	if deps.Stderr == nil {
		deps.Stderr = io.Discard
	}
	if deps.ResolveBinary == nil {
		deps.ResolveBinary = func(dir string) (string, error) {
			return ffmpeg.ResolveBinary(dir, func(name string) (string, error) {
				return execLookPath(name)
			})
		}
	}
	hasCustomRunCommand := deps.RunCommand != nil
	if deps.DetectGPUEncoder == nil {
		deps.DetectGPUEncoder = func(ffmpegPath string, codec string, runCmd func(string, []string) error) string {
			return ffmpeg.DetectGPUEncoder(ffmpegPath, codec, runCmd)
		}
	}
	if deps.MkdirAll == nil {
		deps.MkdirAll = os.MkdirAll
	}
	if deps.PathExists == nil {
		deps.PathExists = pathExists
	}
	if deps.InputAccessible == nil {
		deps.InputAccessible = checkInputAccessible
	}
	if deps.Stdin == nil {
		deps.Stdin = os.Stdin
	}
	if deps.AddToPath == nil {
		deps.AddToPath = func(dir string) (bool, error) {
			return false, errors.New(msg("err.add_to_path_missing"))
		}
	}
	if deps.RemoveFromPath == nil {
		deps.RemoveFromPath = func(dir string) (bool, error) {
			return false, errors.New(msg("err.remove_path_missing"))
		}
	}
	if deps.SendToNotify == nil {
		deps.SendToNotify = sendToNotify
	}

	for _, arg := range args {
		if arg == "-h" || arg == "--help" {
			printUsage(deps.Stdout)
			return 0
		}
	}

	fs := flag.NewFlagSet("videopress", flag.ContinueOnError)
	fs.SetOutput(deps.Stderr)

	presetName := fs.String("preset", "standard", "compression preset")
	cropExact := fs.String("crop", "", "exact crop W:H:X:Y")
	cropCenter := fs.String("crop-center", "", "center crop to ratio W:H")
	cropScale := fs.Float64("crop-scale", 1.0, "keep scale inside center crop (0.5..1.0)")
	concurrency := fs.Int("concurrency", 1, "number of concurrent compressions")
	fs.IntVar(concurrency, "c", 1, "number of concurrent compressions (shorthand)")
	hwAccel := fs.Bool("hw", false, "enable GPU hardware acceleration")
	forceMode := fs.Bool("force", false, "force overwrite output files")
	fs.BoolVar(forceMode, "f", false, "force overwrite output files (shorthand)")
	skipExisting := fs.Bool("skip-existing", false, "skip compression if output file already exists")
	copyAudio := fs.Bool("copy-audio", false, "copy audio stream instead of re-encoding")
	fs.BoolVar(copyAudio, "a", false, "copy audio stream instead of re-encoding (shorthand)")
	codecName := fs.String("codec", "", "video encoder format (h264|h265|av1)")
	maxFPS := fs.Int("max-fps", 0, "limit max frames per second")
	audioMode := fs.String("audio", "", "audio mode (copy|compress|mute)")
	crf := fs.Int("crf", 0, "override preset CRF quality value")
	gifMode := fs.Bool("gif", false, "enable animated export mode")
	gifTier := fs.String("tier", gif.DefaultTier, "animated tier")
	var gifFormats gifFormatList
	fs.Var(&gifFormats, "format", "animated output format (gif|apng|webp)")
	sendToMode := fs.Bool("sendto", false, "enable SendTo prompt on exit")
	installSendTo := fs.Bool("install-sendto", false, "install SendTo shortcut")
	uninstallSendTo := fs.Bool("uninstall-sendto", false, "remove SendTo shortcut")
	installPath := fs.Bool("install-path", false, "add executable directory to User Path")
	uninstallPath := fs.Bool("uninstall-path", false, "remove executable directory from User Path")
	showVersion := fs.Bool("version", false, "show version")

	if err := fs.Parse(args); err != nil {
		fmt.Fprintf(deps.Stderr, "%s %v\n\n", red(msg("err.unknown_option")), err)
		printUsage(deps.Stderr)
		return 1
	}

	// 几何裁剪参数解析与互斥校验（校验失败直接退出，不调用压片）
	var cropSpec *engine.CropSpec
	if *cropExact != "" && *cropCenter != "" {
		fmt.Fprintln(deps.Stderr, red(msg("err.crop_conflict")))
		return 1
	}
	if cropScaleFlagGiven(fs) && *cropCenter == "" {
		fmt.Fprintln(deps.Stderr, red(msg("err.crop_scale_alone")))
		return 1
	}
	if *cropExact != "" {
		spec, err := parseExactCrop(*cropExact)
		if err != nil {
			fmt.Fprintln(deps.Stderr, red(err.Error()))
			return 1
		}
		cropSpec = spec
	}
	if *cropCenter != "" {
		if *cropScale < 0.5 || *cropScale > 1.0 {
			fmt.Fprintln(deps.Stderr, red(msg("err.crop_scale_range")))
			return 1
		}
		spec, err := parseCenterCrop(*cropCenter)
		if err != nil {
			fmt.Fprintln(deps.Stderr, red(err.Error()))
			return 1
		}
		spec.Scale = *cropScale
		cropSpec = spec
	}
	if *showVersion {
		fmt.Fprintln(deps.Stdout, Version)
		return 0
	}
	if *installSendTo {
		if deps.InstallSendTo == nil {
			fmt.Fprintln(deps.Stderr, red(msg("err.sendto_install_off")))
			return 1
		}
		path, err := deps.InstallSendTo(deps.ExecutablePath)
		if err != nil {
			fmt.Fprintln(deps.Stderr, red(err.Error()))
			return 1
		}
		fmt.Fprintf(deps.Stdout, "%s %s\n", green(msg("install.sendto_ok")), green(path))
		fmt.Fprintln(deps.Stdout, msg("install.sendto_tip"))
		fmt.Fprintln(deps.Stdout, msg("install.done"))
		var b [1]byte
		deps.Stdin.Read(b[:])
		return 0
	}
	if *uninstallSendTo {
		if deps.UninstallSendTo == nil {
			fmt.Fprintln(deps.Stderr, red(msg("err.sendto_uninstall_off")))
			return 1
		}
		if err := deps.UninstallSendTo(); err != nil {
			fmt.Fprintln(deps.Stderr, red(err.Error()))
			return 1
		}
		fmt.Fprintln(deps.Stdout, green(msg("install.sendto_removed")))
		fmt.Fprintln(deps.Stdout, msg("install.done"))
		var b [1]byte
		deps.Stdin.Read(b[:])
		return 0
	}
	if *installPath {
		if deps.AddToPath == nil {
			fmt.Fprintln(deps.Stderr, red(msg("err.path_config_off")))
			return 1
		}
		added, err := deps.AddToPath(deps.ExecutableDir)
		if err != nil {
			fmt.Fprintln(deps.Stderr, red(err.Error()))
			return 1
		}
		if added {
			fmt.Fprintf(deps.Stdout, "%s %s\n", green(msg("install.path_ok")), green(deps.ExecutableDir))
			fmt.Fprintln(deps.Stdout, msg("install.path_tip"))
		} else {
			fmt.Fprintf(deps.Stdout, "%s %s\n", yellow(msg("install.path_exists")), deps.ExecutableDir)
		}
		fmt.Fprintln(deps.Stdout, msg("install.done"))
		var b [1]byte
		deps.Stdin.Read(b[:])
		return 0
	}
	if *uninstallPath {
		if deps.RemoveFromPath == nil {
			fmt.Fprintln(deps.Stderr, red(msg("err.path_config_off")))
			return 1
		}
		removed, err := deps.RemoveFromPath(deps.ExecutableDir)
		if err != nil {
			fmt.Fprintln(deps.Stderr, red(err.Error()))
			return 1
		}
		if removed {
			fmt.Fprintf(deps.Stdout, "%s %s\n", green(msg("install.path_removed")), green(deps.ExecutableDir))
		} else {
			fmt.Fprintf(deps.Stdout, "%s %s\n", yellow(msg("install.path_missing")), deps.ExecutableDir)
		}
		fmt.Fprintln(deps.Stdout, msg("install.done"))
		var b [1]byte
		deps.Stdin.Read(b[:])
		return 0
	}

	files := fs.Args()
	if len(files) == 0 {
		printUsage(deps.Stderr)
		if *sendToMode {
			_ = deps.SendToNotify(deps.ExecutablePath, msg("notify.failed"), msg("notify.no_files"))
		}
		return 1
	}

	// 动图导出模式：GIF/APNG/WebP 三档
	if *gifMode {
		return runGifExport(deps, files, *gifTier, gifFormats, *forceMode)
	}
	if len(gifFormats) > 0 || *gifTier != gif.DefaultTier {
		fmt.Fprintln(deps.Stderr, yellow(msg("err.tier_format_ignored")))
	}
	presetExplicit := false
	fs.Visit(func(f *flag.Flag) {
		if f.Name == "preset" {
			presetExplicit = true
		}
	})

	if *sendToMode && !presetExplicit && isTerminal(deps.Stdout) {
		*presetName = showInteractiveMenu(deps.Stdout)
	}

	preset, err := compress.PresetByName(*presetName)
	if err != nil {
		fmt.Fprintln(deps.Stderr, red(err.Error()))
		if *sendToMode {
			_ = deps.SendToNotify(deps.ExecutablePath, msg("notify.failed"), msg("notify.invalid_preset"))
		}
		return 1
	}

	ffmpegPath, err := deps.ResolveBinary(deps.ExecutableDir)
	if err != nil {
		fmt.Fprintln(deps.Stderr, red(err.Error()))
		if *sendToMode {
			_ = deps.SendToNotify(deps.ExecutablePath, msg("notify.failed"), msg("notify.no_ffmpeg"))
		}
		return 1
	}

	hwEncoder := "libx264"
	if *hwAccel {
		fmt.Fprintln(deps.Stdout, msg("gpu.detecting"))
		hwEncoder = deps.DetectGPUEncoder(ffmpegPath, *codecName, nil)
		if hwEncoder != "libx264" && hwEncoder != "libx265" && hwEncoder != "libsvtav1" {
			fmt.Fprintln(deps.Stdout, msg("gpu.detected", green(hwEncoder)))
		} else {
			fmt.Fprintln(deps.Stdout, msg("gpu.none"))
		}
	}

	limit := *concurrency
	if limit < 1 {
		limit = 1
	}
	if limit > len(files) {
		limit = len(files)
	}

	// 打印任务配置摘要卡片
	rule := magenta("========================================")
	fmt.Fprintln(deps.Stdout, rule)
	fmt.Fprintln(deps.Stdout, magenta(msg("run.title")))
	fmt.Fprintln(deps.Stdout, rule)
	var maxDimStr string
	if preset.MaxDimension > 0 {
		maxDimStr = fmt.Sprintf("%dpx", preset.MaxDimension)
	} else {
		maxDimStr = msg("run.unlimited")
	}
	fmt.Fprintf(deps.Stdout, msg("run.preset"), cyan(preset.Name))
	fmt.Fprintf(deps.Stdout, msg("run.scale"), cyan(fmt.Sprintf("%.0f%%", preset.ScaleFactor*100)))
	fmt.Fprintf(deps.Stdout, msg("run.max_dim"), cyan(maxDimStr))
	fmt.Fprintf(deps.Stdout, msg("run.concurrency"), cyan(fmt.Sprintf("%d", limit)))

	codecPrint := *codecName
	if codecPrint == "" {
		codecPrint = msg("run.codec_auto")
	}
	fmt.Fprintf(deps.Stdout, msg("run.codec"), cyan(codecPrint))

	if *crf > 0 {
		fmt.Fprintf(deps.Stdout, msg("run.crf"), cyan(msg("run.crf_overridden", *crf)))
	} else {
		fmt.Fprintf(deps.Stdout, msg("run.crf"), gray(msg("run.crf_off")))
	}

	if *maxFPS > 0 {
		fmt.Fprintf(deps.Stdout, msg("run.fps"), cyan(fmt.Sprintf("%dfps", *maxFPS)))
	} else {
		fmt.Fprintf(deps.Stdout, msg("run.fps"), cyan(msg("run.unlimited")))
	}

	var audioModePrint string
	if *audioMode != "" {
		audioModePrint = *audioMode
	} else if *copyAudio {
		audioModePrint = msg("run.audio_copy")
	} else {
		audioModePrint = msg("run.audio_compress")
	}
	fmt.Fprintf(deps.Stdout, msg("run.audio"), cyan(audioModePrint))

	if *hwAccel {
		fmt.Fprintf(deps.Stdout, msg("run.hw"), green(hwEncoder))
	} else {
		fmt.Fprintf(deps.Stdout, msg("run.hw"), gray(msg("run.hw_disabled")))
	}

	if cropSpec != nil {
		var cropDesc string
		if cropSpec.Exact {
			cropDesc = msg("run.crop_exact", cropSpec.W, cropSpec.H, cropSpec.X, cropSpec.Y)
		} else {
			cropDesc = msg("run.crop_center", cropSpec.RatioW, cropSpec.RatioH, cropSpec.Scale*100)
		}
		fmt.Fprintf(deps.Stdout, msg("run.crop"), cyan(cropDesc))
	} else {
		fmt.Fprintf(deps.Stdout, msg("run.crop"), gray(msg("run.crop_off")))
	}
	fmt.Fprintln(deps.Stdout, rule)
	fmt.Fprintln(deps.Stdout)

	// 建立输入到输出的映射，用于单任务模式下打印完成信息
	outputMap := make(map[string]string)
	var engineFiles []string

	var mu sync.Mutex
	failures := 0
	successes := 0
	var allReports []JobReport

	// 同批次内已分配的输出路径：不同目录下的同名输入会落到同一个输出文件，
	// 只查磁盘会漏掉它们之间的冲突。
	reserved := make(map[string]bool, len(files))
	pathTaken := func(path string) bool {
		return reserved[filepath.Clean(path)] || deps.PathExists(path)
	}

	for _, input := range files {
		if !util.IsVideoFile(input) {
			fmt.Fprintf(deps.Stdout, msg("file.skip_non_video"), gray(input))
			continue
		}

		if !deps.InputAccessible(input) {
			fmt.Fprintf(deps.Stderr, "%s %s", red(msg("file.unreadable")), msg("file.unreadable_detail", input))
			mu.Lock()
			failures++
			mu.Unlock()
			continue
		}

		defaultOutput, err := compress.BuildOutputPath(input, preset.Name, nil, true, "")
		if err == nil && *skipExisting && !*forceMode && deps.PathExists(defaultOutput) {
			fmt.Fprintf(deps.Stdout, msg("file.skip_existing"), yellow(defaultOutput))
			mu.Lock()
			successes++
			allReports = append(allReports, JobReport{
				InputName:  filepath.Base(input),
				OutputDir:  filepath.Dir(defaultOutput),
				Status:     "skipped",
				SourceSize: util.GetFileSize(input),
			})
			mu.Unlock()
			continue
		}

		output, err := compress.BuildOutputPath(input, preset.Name, pathTaken, *forceMode, "")
		if err != nil {
			fmt.Fprintf(deps.Stderr, "%s %s", red(msg("file.output_path_failed")), msg("file.output_path_detail", input, err))
			mu.Lock()
			failures++
			mu.Unlock()
			continue
		}

		reserved[filepath.Clean(output)] = true
		outputMap[input] = output
		engineFiles = append(engineFiles, input)
	}

	if len(engineFiles) > 0 {
		engDeps := engine.Dependencies{
			ExecutableDir:    deps.ExecutableDir,
			ResolveBinary:    deps.ResolveBinary,
			RunCommand:       nil, // 生产路径传 nil 解锁引擎进度分析流水线；测试通过注入 mock 保留
			DetectGPUEncoder: deps.DetectGPUEncoder,
			MkdirAll:         deps.MkdirAll,
			PathExists:       deps.PathExists,
			InputAccessible:  deps.InputAccessible,
		}
		if hasCustomRunCommand {
			engDeps.RunCommand = deps.RunCommand
		}
		eng := engine.NewCompressEngine(engDeps)

		simpleProgress := limit > 1
		var muProgress sync.Mutex

		onProgress := func(ev engine.ProgressEvent) {
			muProgress.Lock()
			defer muProgress.Unlock()

			if ev.Error != "" {
				if !simpleProgress {
					fmt.Fprintln(deps.Stdout)
				}
				fmt.Fprintf(deps.Stderr, "\n%s %s: %s\n", red(msg("progress.failed")), ev.File, red(ev.Error))
				return
			}

			if ev.Done {
				if simpleProgress {
					fmt.Fprintf(deps.Stdout, msg("progress.done"), ev.File)
				} else {
					fmt.Fprintln(deps.Stdout)
					outPath := outputMap[ev.Path]
					fmt.Fprintf(deps.Stdout, msg("progress.done_path"), ev.Path, outPath)
				}
				return
			}

			if simpleProgress {
				if ev.Percent == 0 {
					fmt.Fprintf(deps.Stdout, msg("progress.start"), ev.File)
				}
			} else {
				ffmpeg.RenderProgressBar(deps.Stdout, ev.Percent, ev.File)
			}
		}

		reports, err := eng.Run(context.Background(), engine.JobRequest{
			Files:        engineFiles,
			Preset:       *presetName,
			HWAccel:      *hwAccel,
			CopyAudio:    *copyAudio,
			ForceMode:    *forceMode,
			SkipExisting: *skipExisting,
			Concurrency:  limit,
			VideoCodec:   *codecName,
			MaxFPS:       *maxFPS,
			AudioMode:    *audioMode,
			CRF:          *crf,
			Crop:         cropSpec,
		}, onProgress)

		if err != nil {
			fmt.Fprintf(deps.Stderr, "%s %s\n", red(msg("file.output_path_failed")), msg("err.engine_failed", err))
			return 1
		}

		for _, r := range reports {
			if r.Status == "success" {
				successes++
			} else {
				failures++
			}
			allReports = append(allReports, r)
		}
	}

	// 归档日志
	reportsByDir := make(map[string][]JobReport)
	for _, r := range allReports {
		reportsByDir[r.OutputDir] = append(reportsByDir[r.OutputDir], r)
	}
	for outDir, reps := range reportsByDir {
		logPath := filepath.Join(outDir, "compress_summary.log")
		f, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
		if err != nil {
			fmt.Fprintf(deps.Stderr, "%s %s\n", red(msg("file.unreadable")), msg("err.summary_log", logPath, err))
		} else {
			fmt.Fprintf(f, msg("log.start"), time.Now().Format("2006-01-02 15:04:05"))
			for _, r := range reps {
				switch r.Status {
				case "success":
					saved := 0.0
					if r.SourceSize > 0 {
						saved = float64(r.SourceSize-r.TargetSize) / float64(r.SourceSize) * 100.0
					}
					fmt.Fprintf(f, msg("log.success"),
						r.InputName,
						float64(r.SourceSize)/(1024*1024),
						float64(r.TargetSize)/(1024*1024),
						saved,
						r.Duration.Round(time.Millisecond).String(),
					)
				case "skipped":
					fmt.Fprintf(f, msg("log.skipped"), r.InputName)
				case "failed":
					fmt.Fprintf(f, msg("log.failed"), r.InputName, r.ErrMessage)
				}
			}
			fmt.Fprint(f, msg("log.separator"))
			f.Close()
		}
	}

	// 打印最后的色彩汇总表格
	fmt.Fprintln(deps.Stdout, "\n"+magenta("================================ TASK SUMMARY ================================"))
	fmt.Fprintf(deps.Stdout, " %s | %s | %s | %s | %s | %s\n",
		padString(msg("table.file"), 20, true),
		padString(msg("table.status"), 8, true),
		padString(msg("table.original"), 12, true),
		padString(msg("table.compressed"), 12, true),
		padString(msg("table.saved"), 8, true),
		padString(msg("table.time"), 8, true),
	)
	fmt.Fprintln(deps.Stdout, magenta("------------------------------------------------------------------------------"))
	for _, r := range allReports {
		var statusStr string
		switch r.Status {
		case "success":
			statusStr = green(msg("table.success")) + "  "
		case "skipped":
			statusStr = yellow(msg("table.skipped")) + "  "
		case "failed":
			statusStr = red(msg("table.failed")) + "  "
		}

		displayInput := r.InputName
		runes := []rune(displayInput)
		if len(runes) > 18 {
			displayInput = string(runes[:15]) + "..."
		}

		savedStr := "-"
		if r.Status == "success" && r.SourceSize > 0 {
			saved := float64(r.SourceSize-r.TargetSize) / float64(r.SourceSize) * 100.0
			savedStr = fmt.Sprintf("%.1f%%", saved)
		}

		targetSizeStr := "-"
		if r.Status == "success" {
			targetSizeStr = fmt.Sprintf("%.1fMB", float64(r.TargetSize)/(1024*1024))
		}

		durationStr := "-"
		if r.Status != "skipped" {
			durationStr = r.Duration.Round(time.Millisecond).String()
		}

		fmt.Fprintf(deps.Stdout, " %s | %s | %s | %s | %s | %s\n",
			padString(displayInput, 20, true),
			statusStr,
			padString(fmt.Sprintf("%.1fMB", float64(r.SourceSize)/(1024*1024)), 12, true),
			padString(targetSizeStr, 12, true),
			padString(savedStr, 8, true),
			padString(durationStr, 8, true),
		)
	}
	fmt.Fprintln(deps.Stdout, magenta("=============================================================================="))

	fmt.Fprintf(deps.Stdout, msg("table.total"), green(fmt.Sprintf("%d", successes)), red(fmt.Sprintf("%d", failures)))

	exitCode := 0
	if failures > 0 || successes == 0 {
		exitCode = 1
	}

	if *sendToMode {
		_ = deps.SendToNotify(deps.ExecutablePath, sendToResultTitle(successes, failures), sendToResultBody(successes, failures))
	}
	return exitCode
}

// cropScaleFlagGiven 判断 --crop-scale 是否被显式指定。
func cropScaleFlagGiven(fs *flag.FlagSet) bool {
	given := false
	fs.Visit(func(f *flag.Flag) {
		if f.Name == "crop-scale" {
			given = true
		}
	})
	return given
}

// parseExactCrop 解析 --crop W:H:X:Y 精确像素裁剪（与 ffmpeg crop 一致）。
func parseExactCrop(s string) (*engine.CropSpec, error) {
	parts := strings.Split(s, ":")
	if len(parts) != 4 {
		return nil, fmt.Errorf(msg("err.crop_format"), s)
	}
	nums := make([]int, 4)
	for i, p := range parts {
		n, err := strconv.Atoi(strings.TrimSpace(p))
		if err != nil || n < 0 {
			return nil, fmt.Errorf(msg("err.crop_numbers"), s)
		}
		nums[i] = n
	}
	return &engine.CropSpec{Exact: true, W: nums[0], H: nums[1], X: nums[2], Y: nums[3]}, nil
}

// parseCenterCrop 解析 --crop-center W:H 居中比例裁剪。
func parseCenterCrop(s string) (*engine.CropSpec, error) {
	parts := strings.Split(s, ":")
	if len(parts) != 2 {
		return nil, fmt.Errorf(msg("err.crop_ratio_format"), s)
	}
	rw, err1 := strconv.Atoi(strings.TrimSpace(parts[0]))
	rh, err2 := strconv.Atoi(strings.TrimSpace(parts[1]))
	if err1 != nil || err2 != nil || rw <= 0 || rh <= 0 {
		return nil, fmt.Errorf(msg("err.crop_ratio_numbers"), s)
	}
	return &engine.CropSpec{RatioW: rw, RatioH: rh, Scale: 1.0}, nil
}

func getDisplayWidth(s string) int {
	w := 0
	for _, r := range s {
		// 精确匹配 CJK 统一汉字、标点符号、兼容汉字及全角字符区间
		if (r >= 0x4e00 && r <= 0x9fff) ||
			(r >= 0x3000 && r <= 0x303f) ||
			(r >= 0xf900 && r <= 0xfaff) ||
			(r >= 0xff00 && r <= 0xffef) {
			w += 2
		} else {
			w += 1
		}
	}
	return w
}

func padString(s string, width int, leftAlign bool) string {
	curW := getDisplayWidth(s)
	if curW >= width {
		return s
	}
	padding := strings.Repeat(" ", width-curW)
	if leftAlign {
		return s + padding
	}
	return padding + s
}
