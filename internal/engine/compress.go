package engine

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"time"

	"videopress/internal/compress"
	"videopress/internal/ffmpeg"
	"videopress/internal/util"
)

// Dependencies holds the file system and external command dependencies for the engine.
type Dependencies struct {
	ExecutableDir    string
	ResolveBinary    func(dir string) (string, error)
	RunCommand       func(name string, args []string) error
	GetDuration      func(ffmpegPath string, inputPath string) (time.Duration, error)
	GetDimensions    func(ffmpegPath string, inputPath string) (int, int, error)
	Probe            func(ffmpegPath string, inputPath string) (ffmpeg.VideoInfo, error)
	DetectGPUEncoder func(ffmpegPath string, codec string, runCmd func(string, []string) error) string
	MkdirAll         func(path string, perm os.FileMode) error
	PathExists       func(path string) bool
	InputAccessible  func(path string) bool
}

// DefaultDependencies provides default implementation of engine dependencies.
func DefaultDependencies(execDir string) Dependencies {
	return Dependencies{
		ExecutableDir: execDir,
		ResolveBinary: func(dir string) (string, error) {
			return ffmpeg.ResolveBinary(dir, func(name string) (string, error) {
				return exec.LookPath(name)
			})
		},
		RunCommand: func(name string, args []string) error {
			cmd := exec.Command(name, args...)
			util.HideConsoleWindow(cmd)
			return cmd.Run()
		},
		GetDuration:   ffmpeg.GetDuration,
		GetDimensions: ffmpeg.GetDimensions,
		Probe:         ffmpeg.Probe,
		DetectGPUEncoder: func(ffmpegPath string, codec string, runCmd func(string, []string) error) string {
			return ffmpeg.DetectGPUEncoder(ffmpegPath, codec, runCmd)
		},
		MkdirAll: os.MkdirAll,
		PathExists: func(path string) bool {
			_, err := os.Stat(path)
			return err == nil
		},
		InputAccessible: func(path string) bool {
			info, err := os.Stat(path)
			if err != nil {
				return false
			}
			return !info.IsDir()
		},
	}
}

// CompressEngine runs the video compression process.
type CompressEngine struct {
	deps         Dependencies
	hasCustomRun bool
}

// NewCompressEngine creates a new instance of CompressEngine.
func NewCompressEngine(deps Dependencies) *CompressEngine {
	hasCustomRun := deps.RunCommand != nil
	// Fallbacks
	if deps.RunCommand == nil {
		deps.RunCommand = func(name string, args []string) error {
			cmd := exec.Command(name, args...)
			util.HideConsoleWindow(cmd)
			return cmd.Run()
		}
	}
	if deps.ResolveBinary == nil {
		deps.ResolveBinary = func(dir string) (string, error) {
			return ffmpeg.ResolveBinary(dir, func(name string) (string, error) {
				return exec.LookPath(name)
			})
		}
	}
	if deps.GetDuration == nil {
		deps.GetDuration = ffmpeg.GetDuration
	}
	if deps.GetDimensions == nil {
		deps.GetDimensions = ffmpeg.GetDimensions
	}
	if deps.Probe == nil {
		deps.Probe = ffmpeg.Probe
	}
	if deps.DetectGPUEncoder == nil {
		deps.DetectGPUEncoder = func(ffmpegPath string, codec string, runCmd func(string, []string) error) string {
			return ffmpeg.DetectGPUEncoder(ffmpegPath, codec, runCmd)
		}
	}
	if deps.MkdirAll == nil {
		deps.MkdirAll = os.MkdirAll
	}
	if deps.PathExists == nil {
		deps.PathExists = func(path string) bool {
			_, err := os.Stat(path)
			return err == nil
		}
	}
	if deps.InputAccessible == nil {
		deps.InputAccessible = func(path string) bool {
			info, err := os.Stat(path)
			if err != nil {
				return false
			}
			return !info.IsDir()
		}
	}
	return &CompressEngine{
		deps:         deps,
		hasCustomRun: hasCustomRun,
	}
}

// Run executes the compression job and reports progress via onProgress callback.
func (e *CompressEngine) Run(ctx context.Context, req JobRequest, onProgress func(ProgressEvent)) ([]JobReport, error) {
	preset, err := compress.PresetByName(req.Preset)
	if err != nil {
		return nil, err
	}

	ffmpegPath, err := e.deps.ResolveBinary(e.deps.ExecutableDir)
	if err != nil {
		return nil, err
	}

	hwEncoder := "libx264"
	if req.VideoCodec == "h265" || req.VideoCodec == "hevc" {
		hwEncoder = "libx265"
	} else if req.VideoCodec == "av1" {
		hwEncoder = "libsvtav1"
	}

	if req.HWAccel {
		hwEncoder = e.deps.DetectGPUEncoder(ffmpegPath, req.VideoCodec, e.deps.RunCommand)
	}

	limit := req.Concurrency
	if limit < 1 {
		limit = 1
	}
	if limit > len(req.Files) {
		limit = len(req.Files)
	}

	type Task struct {
		input  string
		output string
	}

	var mu sync.Mutex
	var allReports []JobReport

	var validTasks []Task
	for _, input := range req.Files {
		if !util.IsVideoFile(input) {
			// Skip non-video files
			continue
		}

		if !e.deps.InputAccessible(input) {
			mu.Lock()
			allReports = append(allReports, JobReport{
				InputName:  filepath.Base(input),
				InputPath:  input,
				Status:     "failed",
				SourceSize: util.GetFileSize(input),
				ErrMessage: "输入文件不存在或不可读",
			})
			mu.Unlock()
			if onProgress != nil {
				onProgress(ProgressEvent{
					File:  filepath.Base(input),
					Path:  input,
					Done:  true,
					Error: "输入文件不存在或不可读",
				})
			}
			continue
		}

		defaultOutput, err := compress.BuildOutputPath(input, preset.Name, nil, true, req.OutputDir)
		if err == nil && req.SkipExisting && !req.ForceMode && e.deps.PathExists(defaultOutput) {
			mu.Lock()
			allReports = append(allReports, JobReport{
				InputName:  filepath.Base(input),
				InputPath:  input,
				OutputDir:  filepath.Dir(defaultOutput),
				Status:     "skipped",
				SourceSize: util.GetFileSize(input),
			})
			mu.Unlock()
			if onProgress != nil {
				onProgress(ProgressEvent{
					File:    filepath.Base(input),
					Path:    input,
					Percent: 100,
					Done:    true,
				})
			}
			continue
		}

		output, err := compress.BuildOutputPath(input, preset.Name, e.deps.PathExists, req.ForceMode, req.OutputDir)
		if err != nil {
			mu.Lock()
			allReports = append(allReports, JobReport{
				InputName:  filepath.Base(input),
				InputPath:  input,
				Status:     "failed",
				SourceSize: util.GetFileSize(input),
				ErrMessage: fmt.Sprintf("生成输出路径失败: %v", err),
			})
			mu.Unlock()
			if onProgress != nil {
				onProgress(ProgressEvent{
					File:  filepath.Base(input),
					Path:  input,
					Done:  true,
					Error: fmt.Sprintf("生成输出路径失败: %v", err),
				})
			}
			continue
		}

		if err := e.deps.MkdirAll(filepath.Dir(output), 0o755); err != nil {
			mu.Lock()
			allReports = append(allReports, JobReport{
				InputName:  filepath.Base(input),
				InputPath:  input,
				Status:     "failed",
				SourceSize: util.GetFileSize(input),
				ErrMessage: fmt.Sprintf("创建输出目录失败: %v", err),
			})
			mu.Unlock()
			if onProgress != nil {
				onProgress(ProgressEvent{
					File:  filepath.Base(input),
					Path:  input,
					Done:  true,
					Error: fmt.Sprintf("创建输出目录失败: %v", err),
				})
			}
			continue
		}

		validTasks = append(validTasks, Task{input: input, output: output})
	}

	tasksChan := make(chan Task, limit*2)
	go func() {
		for _, t := range validTasks {
			tasksChan <- t
		}
		close(tasksChan)
	}()

	var wg sync.WaitGroup
	for i := 0; i < limit; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for task := range tasksChan {
				if ctx.Err() != nil {
					mu.Lock()
					allReports = append(allReports, JobReport{
						InputName:  filepath.Base(task.input),
						InputPath:  task.input,
						OutputDir:  filepath.Dir(task.output),
						Status:     "failed",
						SourceSize: util.GetFileSize(task.input),
						ErrMessage: "任务已取消",
					})
					if onProgress != nil {
						onProgress(ProgressEvent{
							File:  filepath.Base(task.input),
							Path:  task.input,
							Done:  true,
							Error: "任务已取消",
						})
					}
					mu.Unlock()
					continue
				}

				startTime := time.Now()
				cropExpr := ""
				var duration time.Duration
				if req.Crop != nil {
					// 裁剪开启时单次探测同时取时长与分辨率，避免逐项各起一个 ffmpeg 进程
					var info ffmpeg.VideoInfo
					var perr error
					info, perr = e.deps.Probe(ffmpegPath, task.input)
					duration = info.Duration
					var cerr error
					cropExpr, cerr = e.resolveCropExpr(req.Crop, info.Width, info.Height, perr)
					if cerr != nil {
						mu.Lock()
						allReports = append(allReports, JobReport{
							InputName:  filepath.Base(task.input),
							InputPath:  task.input,
							Status:     "failed",
							SourceSize: util.GetFileSize(task.input),
							ErrMessage: cerr.Error(),
						})
						mu.Unlock()
						if onProgress != nil {
							onProgress(ProgressEvent{
								File:  filepath.Base(task.input),
								Path:  task.input,
								Done:  true,
								Error: cerr.Error(),
							})
						}
						continue
					}
				}
				args := ffmpeg.BuildArgs(task.input, task.output, preset, hwEncoder, req.CopyAudio, req.MaxFPS, req.AudioMode, req.CRF, cropExpr)

				err := e.runCommandWithProgress(ctx, ffmpegPath, args, duration, task.input, filepath.Base(task.input), onProgress)
				elapsed := time.Since(startTime)

				mu.Lock()
				if err != nil {
					friendlyErr := "任务已取消"
					if ctx.Err() == nil {
						friendlyErr = ffmpeg.ParseFFmpegError(err.Error())
					}
					allReports = append(allReports, JobReport{
						InputName:  filepath.Base(task.input),
						InputPath:  task.input,
						OutputDir:  filepath.Dir(task.output),
						Status:     "failed",
						SourceSize: util.GetFileSize(task.input),
						Duration:   elapsed,
						ErrMessage: friendlyErr,
					})
					if onProgress != nil {
						onProgress(ProgressEvent{
							File:  filepath.Base(task.input),
							Path:  task.input,
							Done:  true,
							Error: friendlyErr,
						})
					}
				} else {
					allReports = append(allReports, JobReport{
						InputName:  filepath.Base(task.input),
						InputPath:  task.input,
						OutputDir:  filepath.Dir(task.output),
						Status:     "success",
						SourceSize: util.GetFileSize(task.input),
						TargetSize: util.GetFileSize(task.output),
						Duration:   elapsed,
					})
					if onProgress != nil {
						onProgress(ProgressEvent{
							File:    filepath.Base(task.input),
							Path:    task.input,
							Percent: 100,
							Done:    true,
						})
					}
				}
				mu.Unlock()
			}
		}()
	}
	wg.Wait()

	return allReports, nil
}

// resolveCropExpr 根据裁剪配置生成 crop 滤镜表达式（空串表示不裁剪）。
// 精确裁剪时对奇数宽高/偏移自动落偶（与居中裁策略一致，避免 yuv420p 运行时报错），
// 并校验归一化后的区域是否越界；居中比例裁剪由 ffmpeg.ComputeCenterCrop 计算内接矩形。
func (e *CompressEngine) resolveCropExpr(crop *CropSpec, iw int, ih int, probeErr error) (string, error) {
	if probeErr != nil {
		return "", fmt.Errorf("探测视频分辨率失败: %v", probeErr)
	}
	if crop.Exact {
		if crop.W <= 0 || crop.H <= 0 {
			return "", fmt.Errorf("裁剪宽高必须为正数: %d:%d", crop.W, crop.H)
		}
		w := ffmpeg.EvenFloor(crop.W)
		h := ffmpeg.EvenFloor(crop.H)
		x := crop.X / 2 * 2
		y := crop.Y / 2 * 2
		if w > iw || h > ih || x < 0 || y < 0 || x+w > iw || y+h > ih {
			return "", fmt.Errorf("裁剪区域越界: crop=%d:%d:%d:%d 超出画面 %dx%d", w, h, x, y, iw, ih)
		}
		return fmt.Sprintf("crop=%d:%d:%d:%d", w, h, x, y), nil
	}

	w, h, x, y := ffmpeg.ComputeCenterCrop(iw, ih, crop.RatioW, crop.RatioH, crop.Scale)
	return fmt.Sprintf("crop=%d:%d:%d:%d", w, h, x, y), nil
}

func (e *CompressEngine) runCommandWithProgress(ctx context.Context, ffmpegPath string, args []string, duration time.Duration, inputPath string, prefix string, onProgress func(ProgressEvent)) error {
	if e.hasCustomRun {
		return e.deps.RunCommand(ffmpegPath, args)
	}

	finalArgs := make([]string, 0, len(args)+2)
	finalArgs = append(finalArgs, args...)
	finalArgs = append(finalArgs, "-progress", "-")

	cmd := exec.CommandContext(ctx, ffmpegPath, finalArgs...)
	util.HideConsoleWindow(cmd)
	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	var stderrBuf bytes.Buffer
	cmd.Stderr = &stderrBuf

	if err := cmd.Start(); err != nil {
		return err
	}

	if onProgress != nil {
		onProgress(ProgressEvent{
			File:    prefix,
			Path:    inputPath,
			Percent: 0,
		})
	}

	if duration > 0 {
		lastPercent := -1.0
		ffmpeg.TrackProgress(stdoutPipe, duration, func(percent float64) {
			if onProgress != nil {
				// 进度变化 >= 0.5% 或者达到 100% 才发送，避免频繁 IPC
				if percent-lastPercent >= 0.5 || percent >= 100.0 {
					lastPercent = percent
					onProgress(ProgressEvent{
						File:    prefix,
						Path:    inputPath,
						Percent: percent,
					})
				}
			}
		})
	} else {
		_, _ = io.Copy(io.Discard, stdoutPipe)
	}

	err = cmd.Wait()
	if err != nil {
		return fmt.Errorf("%w: %s", err, stderrBuf.String())
	}
	return nil
}
