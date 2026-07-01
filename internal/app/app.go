package app

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"videopress/internal/compress"
	"videopress/internal/ffmpeg"
)

const Version = "0.1.0"

type Dependencies struct {
	ExecutableDir   string
	ExecutablePath  string
	ResolveBinary   func(dir string) (string, error)
	RunCommand      func(name string, args []string) error
	MkdirAll        func(path string, perm os.FileMode) error
	PathExists       func(path string) bool
	InputAccessible  func(path string) bool
	Stdout           io.Writer
	Stderr          io.Writer
	InstallSendTo   func(executablePath string) (string, error)
	UninstallSendTo func() error
}

func Execute(args []string, deps Dependencies) int {
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
	if deps.RunCommand == nil {
		deps.RunCommand = runCommand
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

	fs := flag.NewFlagSet("videopress", flag.ContinueOnError)
	fs.SetOutput(deps.Stderr)

	presetName := fs.String("preset", "standard", "compression preset")
	installSendTo := fs.Bool("install-sendto", false, "install SendTo shortcut")
	uninstallSendTo := fs.Bool("uninstall-sendto", false, "remove SendTo shortcut")
	showVersion := fs.Bool("version", false, "show version")

	if err := fs.Parse(args); err != nil {
		return 1
	}
	if *showVersion {
		fmt.Fprintln(deps.Stdout, Version)
		return 0
	}
	if *installSendTo {
		if deps.InstallSendTo == nil {
			fmt.Fprintln(deps.Stderr, "当前构建未启用 SendTo 安装")
			return 1
		}
		path, err := deps.InstallSendTo(deps.ExecutablePath)
		if err != nil {
			fmt.Fprintln(deps.Stderr, err.Error())
			return 1
		}
		fmt.Fprintf(deps.Stdout, "已安装 SendTo 快捷方式: %s\n", path)
		return 0
	}
	if *uninstallSendTo {
		if deps.UninstallSendTo == nil {
			fmt.Fprintln(deps.Stderr, "当前构建未启用 SendTo 卸载")
			return 1
		}
		if err := deps.UninstallSendTo(); err != nil {
			fmt.Fprintln(deps.Stderr, err.Error())
			return 1
		}
		fmt.Fprintln(deps.Stdout, "已移除 SendTo 快捷方式")
		return 0
	}

	files := fs.Args()
	if len(files) == 0 {
		fmt.Fprintln(deps.Stderr, "用法: videopress.exe [--preset small|standard|quality] <files...>")
		return 1
	}

	preset, err := compress.PresetByName(*presetName)
	if err != nil {
		fmt.Fprintln(deps.Stderr, err.Error())
		return 1
	}

	ffmpegPath, err := deps.ResolveBinary(deps.ExecutableDir)
	if err != nil {
		fmt.Fprintln(deps.Stderr, err.Error())
		return 1
	}

	failures := 0
	successes := 0
	for _, input := range files {
		if !isVideoFile(input) {
			fmt.Fprintf(deps.Stdout, "跳过非视频文件: %s\n", input)
			continue
		}

		if !deps.InputAccessible(input) {
			fmt.Fprintf(deps.Stderr, "输入文件不存在或不可读: %s\n", input)
			failures++
			continue
		}

		output, err := compress.BuildOutputPath(input, preset.Name, deps.PathExists)
		if err != nil {
			fmt.Fprintf(deps.Stderr, "生成输出路径失败 %s: %v\n", input, err)
			failures++
			continue
		}

		if err := deps.MkdirAll(filepath.Dir(output), 0o755); err != nil {
			fmt.Fprintf(deps.Stderr, "创建输出目录失败 %s: %v\n", output, err)
			failures++
			continue
		}

		args := ffmpeg.BuildArgs(input, output, preset)
		if err := deps.RunCommand(ffmpegPath, args); err != nil {
			fmt.Fprintf(deps.Stderr, "压缩失败 %s: %v\n", input, err)
			failures++
			continue
		}

		fmt.Fprintf(deps.Stdout, "压缩完成: %s -> %s\n", input, output)
		successes++
	}

	fmt.Fprintf(deps.Stdout, "处理完成: 成功 %d, 失败 %d\n", successes, failures)
	if failures > 0 {
		return 1
	}
	if successes == 0 {
		return 1
	}
	return 0
}

func isVideoFile(path string) bool {
	switch strings.ToLower(filepath.Ext(path)) {
	case ".mp4", ".mov", ".mkv", ".avi", ".m4v", ".wmv", ".webm":
		return true
	default:
		return false
	}
}
