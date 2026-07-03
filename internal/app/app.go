package app

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"videopress/internal/compress"
	"videopress/internal/engine"
	"videopress/internal/ffmpeg"
)

const Version = "0.1.0"

var EnableConsoleColors = func() {}

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
	ExecutableDir          string
	ExecutablePath         string
	ResolveBinary          func(dir string) (string, error)
	RunCommand             func(name string, args []string) error
	RunCommandWithProgress func(ffmpegPath string, args []string, duration time.Duration, prefix string, stdout io.Writer, simpleProgress bool) error
	GetDuration            func(ffmpegPath string, inputPath string) (time.Duration, error)
	DetectGPUEncoder       func(ffmpegPath string, runCmd func(string, []string) error) string
	MkdirAll               func(path string, perm os.FileMode) error
	PathExists             func(path string) bool
	InputAccessible        func(path string) bool
	Stdout                 io.Writer
	Stderr                 io.Writer
	Stdin                  io.Reader
	InstallSendTo          func(executablePath string) (string, error)
	UninstallSendTo        func() error
	AddToPath              func(dir string) (bool, error)
	RemoveFromPath         func(dir string) (bool, error)
}

type JobReport = engine.JobReport

func printUsage(w io.Writer) {
	fmt.Fprintln(w, magenta("========================================"))
	fmt.Fprintln(w, magenta("用法: videopress.exe [选项] <视频文件...>"))
	fmt.Fprintln(w, magenta("========================================"))
	fmt.Fprintln(w, "选项:")
	fmt.Fprintf(w, "  --preset %s  压缩规格（默认 standard，大小写不敏感）\n", cyan("small|standard|quality"))
	fmt.Fprintf(w, "  --concurrency, -c %s         最大并发压缩任务数（默认 1）\n", cyan("<数字>"))
	fmt.Fprintf(w, "  --hw                             %s\n", cyan("尝试使用 GPU 硬件加速编码"))
	fmt.Fprintln(w, "  --force, -f                      强制覆盖已存在的输出文件")
	fmt.Fprintln(w, "  --skip-existing                  如果输出文件已存在则跳过")
	fmt.Fprintln(w, "  --copy-audio, -a                 直接复制音频流，不重编码")
	fmt.Fprintln(w, "  --install-sendto                 安装 SendTo 右键快捷方式")
	fmt.Fprintln(w, "  --uninstall-sendto               移除 SendTo 快捷方式")
	fmt.Fprintln(w, "  --install-path                   添加程序目录到用户 Path 环境变量")
	fmt.Fprintln(w, "  --uninstall-path                 从用户 Path 环境变量移除程序目录")
	fmt.Fprintln(w, "  --version                        显示版本号")
	fmt.Fprintln(w, "  -h, --help                       显示此帮助")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "退出码:")
	fmt.Fprintf(w, "  0  %s\n", green("全部成功"))
	fmt.Fprintf(w, "  1  %s\n", red("存在失败、全部跳过或非视频文件"))
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
	if deps.RunCommand == nil {
		deps.RunCommand = runCommand
	}
	if deps.RunCommandWithProgress == nil {
		if hasCustomRunCommand {
			deps.RunCommandWithProgress = func(ffmpegPath string, args []string, duration time.Duration, prefix string, stdout io.Writer, simpleProgress bool) error {
				return deps.RunCommand(ffmpegPath, args)
			}
		} else {
			deps.RunCommandWithProgress = func(ffmpegPath string, args []string, duration time.Duration, prefix string, stdout io.Writer, simpleProgress bool) error {
				finalArgs := make([]string, 0, len(args)+2)
				finalArgs = append(finalArgs, args...)
				finalArgs = append(finalArgs, "-progress", "-")

				cmd := exec.Command(ffmpegPath, finalArgs...)
				stdoutPipe, err := cmd.StdoutPipe()
				if err != nil {
					return err
				}
				var stderrBuf bytes.Buffer
				cmd.Stderr = &stderrBuf

				if err := cmd.Start(); err != nil {
					return err
				}

				if simpleProgress {
					fmt.Fprintf(stdout, "[%s] 开始压缩...\n", prefix)
					io.Copy(io.Discard, stdoutPipe)
				} else {
					if duration > 0 {
						ffmpeg.TrackProgress(stdoutPipe, duration, func(percent float64) {
							ffmpeg.RenderProgressBar(stdout, percent, prefix)
						})
					} else {
						io.Copy(io.Discard, stdoutPipe)
					}
				}

				err = cmd.Wait()
				if err != nil {
					if !simpleProgress {
						fmt.Fprintln(stdout)
					}
					return fmt.Errorf("%w: %s", err, stderrBuf.String())
				}
				if !simpleProgress {
					fmt.Fprintln(stdout)
				} else {
					fmt.Fprintf(stdout, "[%s] 压缩完成\n", prefix)
				}
				return nil
			}
		}
	}
	if deps.GetDuration == nil {
		deps.GetDuration = ffmpeg.GetDuration
	}
	if deps.DetectGPUEncoder == nil {
		deps.DetectGPUEncoder = func(ffmpegPath string, runCmd func(string, []string) error) string {
			return ffmpeg.DetectGPUEncoder(ffmpegPath, runCmd)
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
			return false, fmt.Errorf("未配置 AddToPath")
		}
	}
	if deps.RemoveFromPath == nil {
		deps.RemoveFromPath = func(dir string) (bool, error) {
			return false, fmt.Errorf("未配置 RemoveFromPath")
		}
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
	concurrency := fs.Int("concurrency", 1, "number of concurrent compressions")
	fs.IntVar(concurrency, "c", 1, "number of concurrent compressions (shorthand)")
	hwAccel := fs.Bool("hw", false, "enable GPU hardware acceleration")
	forceMode := fs.Bool("force", false, "force overwrite output files")
	fs.BoolVar(forceMode, "f", false, "force overwrite output files (shorthand)")
	skipExisting := fs.Bool("skip-existing", false, "skip compression if output file already exists")
	copyAudio := fs.Bool("copy-audio", false, "copy audio stream instead of re-encoding")
	fs.BoolVar(copyAudio, "a", false, "copy audio stream instead of re-encoding (shorthand)")
	sendToMode := fs.Bool("sendto", false, "enable SendTo prompt on exit")
	installSendTo := fs.Bool("install-sendto", false, "install SendTo shortcut")
	uninstallSendTo := fs.Bool("uninstall-sendto", false, "remove SendTo shortcut")
	installPath := fs.Bool("install-path", false, "add executable directory to User Path")
	uninstallPath := fs.Bool("uninstall-path", false, "remove executable directory from User Path")
	showVersion := fs.Bool("version", false, "show version")

	if err := fs.Parse(args); err != nil {
		fmt.Fprintf(deps.Stderr, "%s %v\n\n", red("未知选项:"), err)
		printUsage(deps.Stderr)
		return 1
	}
	if *showVersion {
		fmt.Fprintln(deps.Stdout, Version)
		return 0
	}
	if *installSendTo {
		if deps.InstallSendTo == nil {
			fmt.Fprintln(deps.Stderr, red("当前构建未启用 SendTo 安装"))
			return 1
		}
		path, err := deps.InstallSendTo(deps.ExecutablePath)
		if err != nil {
			fmt.Fprintln(deps.Stderr, red(err.Error()))
			return 1
		}
		fmt.Fprintf(deps.Stdout, "%s 已安装 SendTo 快捷方式: %s\n", green("【成功】"), green(path))
		fmt.Fprintln(deps.Stdout, "\n提示：现在您可以在资源管理器中右键任意视频，选择「发送到 > 快速压缩视频」进行一键压缩！")
		fmt.Fprintln(deps.Stdout, "\n处理完成。按回车键退出...")
		var b [1]byte
		deps.Stdin.Read(b[:])
		return 0
	}
	if *uninstallSendTo {
		if deps.UninstallSendTo == nil {
			fmt.Fprintln(deps.Stderr, red("当前构建未启用 SendTo 卸载"))
			return 1
		}
		if err := deps.UninstallSendTo(); err != nil {
			fmt.Fprintln(deps.Stderr, red(err.Error()))
			return 1
		}
		fmt.Fprintln(deps.Stdout, green("【成功】已成功移除 SendTo 右键快捷方式。"))
		fmt.Fprintln(deps.Stdout, "\n处理完成。按回车键退出...")
		var b [1]byte
		deps.Stdin.Read(b[:])
		return 0
	}
	if *installPath {
		if deps.AddToPath == nil {
			fmt.Fprintln(deps.Stderr, red("当前构建未启用环境变量配置"))
			return 1
		}
		added, err := deps.AddToPath(deps.ExecutableDir)
		if err != nil {
			fmt.Fprintln(deps.Stderr, red(err.Error()))
			return 1
		}
		if added {
			fmt.Fprintf(deps.Stdout, "%s 已将目录添加至当前用户的 Path 环境变量: %s\n", green("【成功】"), green(deps.ExecutableDir))
			fmt.Fprintln(deps.Stdout, "\n提示：现在您可以在新打开的终端（CMD/PowerShell）中，直接通过 `videopress` 命令运行本软件！")
		} else {
			fmt.Fprintf(deps.Stdout, "%s 环境变量 Path 中已存在该目录: %s\n", yellow("【提示】"), deps.ExecutableDir)
		}
		fmt.Fprintln(deps.Stdout, "\n处理完成。按回车键退出...")
		var b [1]byte
		deps.Stdin.Read(b[:])
		return 0
	}
	if *uninstallPath {
		if deps.RemoveFromPath == nil {
			fmt.Fprintln(deps.Stderr, red("当前构建未启用环境变量配置"))
			return 1
		}
		removed, err := deps.RemoveFromPath(deps.ExecutableDir)
		if err != nil {
			fmt.Fprintln(deps.Stderr, red(err.Error()))
			return 1
		}
		if removed {
			fmt.Fprintf(deps.Stdout, "%s 已成功从 Path 环境变量中移除该目录: %s\n", green("【成功】"), green(deps.ExecutableDir))
		} else {
			fmt.Fprintf(deps.Stdout, "%s 环境变量 Path 中未找到该目录: %s\n", yellow("【提示】"), deps.ExecutableDir)
		}
		fmt.Fprintln(deps.Stdout, "\n处理完成。按回车键退出...")
		var b [1]byte
		deps.Stdin.Read(b[:])
		return 0
	}

	files := fs.Args()
	if len(files) == 0 {
		printUsage(deps.Stderr)
		if *sendToMode {
			fmt.Fprintln(deps.Stdout, "\n未指定视频文件。按回车键退出...")
			var b [1]byte
			deps.Stdin.Read(b[:])
		}
		return 1
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
			fmt.Fprintln(deps.Stdout, "\n预设无效。按回车键退出...")
			var b [1]byte
			deps.Stdin.Read(b[:])
		}
		return 1
	}

	ffmpegPath, err := deps.ResolveBinary(deps.ExecutableDir)
	if err != nil {
		fmt.Fprintln(deps.Stderr, red(err.Error()))
		if *sendToMode {
			fmt.Fprintln(deps.Stdout, "\n未找到 FFmpeg。按回车键退出...")
			var b [1]byte
			deps.Stdin.Read(b[:])
		}
		return 1
	}

	hwEncoder := "libx264"
	if *hwAccel {
		fmt.Fprintln(deps.Stdout, "正在检测可用 GPU 编码器...")
		hwEncoder = deps.DetectGPUEncoder(ffmpegPath, nil)
		if hwEncoder != "libx264" {
			fmt.Fprintf(deps.Stdout, "检测到 GPU 编码器: %s，将启用硬件加速\n", green(hwEncoder))
		} else {
			fmt.Fprintln(deps.Stdout, "未检测到可用 GPU 编码器，将使用 CPU 编码 (libx264)")
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
	fmt.Fprintln(deps.Stdout, magenta("========================================"))
	fmt.Fprintln(deps.Stdout, magenta("         Videopress 视频压缩工具        "))
	fmt.Fprintln(deps.Stdout, magenta("========================================"))
	var maxDimStr string
	if preset.MaxDimension > 0 {
		maxDimStr = fmt.Sprintf("%dpx", preset.MaxDimension)
	} else {
		maxDimStr = "无限制"
	}
	fmt.Fprintf(deps.Stdout, " 预设规格: %s\n", cyan(preset.Name))
	fmt.Fprintf(deps.Stdout, " 缩放比例: %s\n", cyan(fmt.Sprintf("%.0f%%", preset.ScaleFactor*100)))
	fmt.Fprintf(deps.Stdout, " 最大分辨率限制: %s\n", cyan(maxDimStr))
	fmt.Fprintf(deps.Stdout, " 并发限制: %s\n", cyan(fmt.Sprintf("%d", limit)))
	if *hwAccel {
		fmt.Fprintf(deps.Stdout, " 硬件编码: %s\n", green(hwEncoder))
	} else {
		fmt.Fprintf(deps.Stdout, " 硬件编码: %s\n", gray("已禁用"))
	}
	fmt.Fprintln(deps.Stdout, magenta("========================================"))
	fmt.Fprintln(deps.Stdout)

	// 建立输入到输出的映射，用于单任务模式下打印完成信息
	outputMap := make(map[string]string)
	var engineFiles []string

	var mu sync.Mutex
	failures := 0
	successes := 0
	var allReports []JobReport

	for _, input := range files {
		if !isVideoFile(input) {
			fmt.Fprintf(deps.Stdout, "跳过非视频文件: %s\n", gray(input))
			continue
		}

		if !deps.InputAccessible(input) {
			fmt.Fprintf(deps.Stderr, "%s 输入文件不存在或不可读: %s\n", red("警告:"), input)
			mu.Lock()
			failures++
			mu.Unlock()
			continue
		}

		defaultOutput, err := compress.BuildOutputPath(input, preset.Name, nil, true, "")
		if err == nil && *skipExisting && !*forceMode && deps.PathExists(defaultOutput) {
			fmt.Fprintf(deps.Stdout, "跳过已存在的文件: %s\n", yellow(defaultOutput))
			mu.Lock()
			successes++
			allReports = append(allReports, JobReport{
				InputName:  filepath.Base(input),
				OutputDir:  filepath.Dir(defaultOutput),
				Status:     "跳过",
				SourceSize: getFileSize(input),
			})
			mu.Unlock()
			continue
		}

		output, err := compress.BuildOutputPath(input, preset.Name, deps.PathExists, *forceMode, "")
		if err != nil {
			fmt.Fprintf(deps.Stderr, "%s 生成输出路径失败 %s: %v\n", red("错误:"), input, err)
			mu.Lock()
			failures++
			mu.Unlock()
			continue
		}

		outputMap[filepath.Base(input)] = output
		engineFiles = append(engineFiles, input)
	}

	if len(engineFiles) > 0 {
		engDeps := engine.Dependencies{
			ExecutableDir:    deps.ExecutableDir,
			ResolveBinary:    deps.ResolveBinary,
			RunCommand:       deps.RunCommand,
			GetDuration:      deps.GetDuration,
			DetectGPUEncoder: deps.DetectGPUEncoder,
			MkdirAll:         deps.MkdirAll,
			PathExists:       deps.PathExists,
			InputAccessible:  deps.InputAccessible,
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
				fmt.Fprintf(deps.Stderr, "\n%s %s: %s\n", red("压缩失败:"), ev.File, red(ev.Error))
				return
			}

			if ev.Done {
				if simpleProgress {
					fmt.Fprintf(deps.Stdout, "[%s] 压缩完成\n", ev.File)
				} else {
					fmt.Fprintln(deps.Stdout)
					outPath := outputMap[ev.File]
					// 还原原有的单文件压缩完成打印，ev.File 此时是 basename
					// 寻找完整的 input path
					var fullInput string
					for _, in := range engineFiles {
						if filepath.Base(in) == ev.File {
							fullInput = in
							break
						}
					}
					fmt.Fprintf(deps.Stdout, "压缩完成: %s -> %s\n", fullInput, outPath)
				}
				return
			}

			if simpleProgress {
				if ev.Percent == 0 {
					fmt.Fprintf(deps.Stdout, "[%s] 开始压缩...\n", ev.File)
				}
			} else {
				ffmpeg.RenderProgressBar(deps.Stdout, ev.Percent, ev.File)
			}
		}

		reports, err := eng.Run(engine.JobRequest{
			Files:        engineFiles,
			Preset:       *presetName,
			HWAccel:      *hwAccel,
			CopyAudio:    *copyAudio,
			ForceMode:    *forceMode,
			SkipExisting: *skipExisting,
			Concurrency:  limit,
		}, onProgress)

		if err != nil {
			fmt.Fprintf(deps.Stderr, "%s 引擎运行失败: %v\n", red("错误:"), err)
			return 1
		}

		for _, r := range reports {
			if r.Status == "成功" {
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
		if err == nil {
			fmt.Fprintf(f, "[%s] 开始压缩任务\n", time.Now().Format("2006-01-02 15:04:05"))
			for _, r := range reps {
				switch r.Status {
				case "成功":
					saved := 0.0
					if r.SourceSize > 0 {
						saved = float64(r.SourceSize-r.TargetSize) / float64(r.SourceSize) * 100.0
					}
					fmt.Fprintf(f, "成功: %s (%.1fMB -> %.1fMB, 节省 %.1f%%, 耗时 %s)\n",
						r.InputName,
						float64(r.SourceSize)/(1024*1024),
						float64(r.TargetSize)/(1024*1024),
						saved,
						r.Duration.Round(time.Millisecond).String(),
					)
				case "跳过":
					fmt.Fprintf(f, "跳过: %s (已存在)\n", r.InputName)
				case "失败":
					fmt.Fprintf(f, "失败: %s (错误信息: %s)\n", r.InputName, r.ErrMessage)
				}
			}
			fmt.Fprintln(f, "----------------------------------------")
			f.Close()
		}
	}

	// 打印最后的色彩汇总表格
	fmt.Fprintln(deps.Stdout, "\n"+magenta("================================ TASK SUMMARY ================================"))
	fmt.Fprintf(deps.Stdout, " %s | %s | %s | %s | %s | %s\n",
		padString("视频文件", 20, true),
		"状态  ",
		padString("原始大小", 12, true),
		padString("压缩大小", 12, true),
		padString("节省", 8, true),
		padString("耗时", 8, true),
	)
	fmt.Fprintln(deps.Stdout, magenta("------------------------------------------------------------------------------"))
	for _, r := range allReports {
		var statusStr string
		switch r.Status {
		case "成功":
			statusStr = green("成功") + "  "
		case "跳过":
			statusStr = yellow("跳过") + "  "
		case "失败":
			statusStr = red("失败") + "  "
		}

		displayInput := r.InputName
		runes := []rune(displayInput)
		if len(runes) > 18 {
			displayInput = string(runes[:15]) + "..."
		}

		savedStr := "-"
		if r.Status == "成功" && r.SourceSize > 0 {
			saved := float64(r.SourceSize-r.TargetSize) / float64(r.SourceSize) * 100.0
			savedStr = fmt.Sprintf("%.1f%%", saved)
		}

		targetSizeStr := "-"
		if r.Status == "成功" {
			targetSizeStr = fmt.Sprintf("%.1fMB", float64(r.TargetSize)/(1024*1024))
		}

		durationStr := "-"
		if r.Status != "跳过" {
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

	fmt.Fprintf(deps.Stdout, "处理完成: 成功 %s, 失败 %s\n", green(fmt.Sprintf("%d", successes)), red(fmt.Sprintf("%d", failures)))

	exitCode := 0
	if failures > 0 || successes == 0 {
		exitCode = 1
	}

	if *sendToMode {
		fmt.Fprintln(deps.Stdout)
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()
		countdown := 5

		doneChan := make(chan bool)
		inputChan := make(chan bool)

		go func() {
			for {
				select {
				case <-doneChan:
					return
				default:
					if hasKey() {
						_ = readKey() // 消费按键
						inputChan <- true
						return
					}
					time.Sleep(50 * time.Millisecond)
				}
			}
		}()

		fmt.Fprintf(deps.Stdout, "处理完成。%d 秒后自动关闭（或按任意键立即关闭）...   \r", countdown)

		for countdown > 0 {
			select {
			case <-ticker.C:
				countdown--
				if countdown <= 0 {
					close(doneChan)
					fmt.Fprint(deps.Stdout, "\n")
					return exitCode
				}
				fmt.Fprintf(deps.Stdout, "处理完成。%d 秒后自动关闭（或按任意键立即关闭）...   \r", countdown)
			case <-inputChan:
				close(doneChan)
				fmt.Fprint(deps.Stdout, "\n")
				return exitCode
			}
		}
	}

	return exitCode
}

func getFileSize(path string) int64 {
	info, err := os.Stat(path)
	if err != nil {
		return 0
	}
	return info.Size()
}

func isVideoFile(path string) bool {
	switch strings.ToLower(filepath.Ext(path)) {
	case ".mp4", ".mov", ".mkv", ".avi", ".m4v", ".wmv", ".webm",
		".ts", ".flv", ".mpg", ".mpeg", ".3gp":
		return true
	default:
		return false
	}
}

func getDisplayWidth(s string) int {
	w := 0
	for _, r := range s {
		if r > 127 {
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
