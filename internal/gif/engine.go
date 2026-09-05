package gif

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"videopress/internal/ffmpeg"
	"videopress/internal/util"
)

// Dependencies holds external command dependencies for the animated export engine.
type Dependencies struct {
	ExecutableDir string
	ResolveBinary func(dir string) (string, error)
	RunCommand    func(ctx context.Context, name string, args []string) error
	MkdirAll      func(path string, perm os.FileMode) error
	PathExists    func(path string) bool
}

// DefaultDependencies provides default implementation of animated export dependencies.
func DefaultDependencies(execDir string) Dependencies {
	return Dependencies{
		ExecutableDir: execDir,
		ResolveBinary: func(dir string) (string, error) {
			return ffmpeg.ResolveBinary(dir, func(name string) (string, error) {
				return exec.LookPath(name)
			})
		},
		RunCommand: func(ctx context.Context, name string, args []string) error {
			cmd := exec.CommandContext(ctx, name, args...)
			util.HideConsoleWindow(cmd)
			return cmd.Run()
		},
		MkdirAll: os.MkdirAll,
		PathExists: func(path string) bool {
			_, err := os.Stat(path)
			return err == nil
		},
	}
}

// ExportResult 描述一次导出产物。
type ExportResult struct {
	InputName  string `json:"inputName"`
	InputPath  string `json:"inputPath"`
	OutputPath string `json:"outputPath"`
	Format     string `json:"format"`
	Tier       string `json:"tier"`
	Size       int64  `json:"size"`
	Status     string `json:"status"` // "success" | "failed" | "skipped"
	Error      string `json:"error,omitempty"`
	OverBudget bool   `json:"overBudget"` // 产物（含跳过）体积超过档位上限
}

// Engine 运行动图导出任务。
type Engine struct {
	deps Dependencies
}

// New 创建动图导出引擎。
func New(deps Dependencies) *Engine {
	if deps.RunCommand == nil {
		deps.RunCommand = func(ctx context.Context, name string, args []string) error {
			cmd := exec.CommandContext(ctx, name, args...)
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
	if deps.MkdirAll == nil {
		deps.MkdirAll = os.MkdirAll
	}
	if deps.PathExists == nil {
		deps.PathExists = func(path string) bool {
			_, err := os.Stat(path)
			return err == nil
		}
	}
	return &Engine{deps: deps}
}

// ExportRequest 描述一次动图导出请求。
type ExportRequest struct {
	Files     []string // 输入视频
	Tier      string   // "smooth" | "balanced" | "hd"
	Formats   []Format // 要导出的格式，空则用全部
	OutputDir string   // 自定义输出目录，空则输入目录下 gif_export/
	Force     bool     // 覆盖已存在输出
}

// outputPathFor 生成某格式的输出路径。
func outputPathFor(inputPath string, tier string, format Format, outputDir string) string {
	ext := filepath.Ext(inputPath)
	base := strings.TrimSuffix(filepath.Base(inputPath), ext)
	dir := filepath.Dir(inputPath)
	if outputDir != "" {
		dir = outputDir
	} else {
		dir = filepath.Join(dir, "gif_export")
	}
	return filepath.Join(dir, fmt.Sprintf("%s.%s%s", base, tier, format.Ext()))
}

// Run 执行动图导出，为每个输入文件、每个格式生成一个产物。
func (e *Engine) Run(ctx context.Context, req ExportRequest) ([]ExportResult, error) {
	tier, err := TierByName(req.Tier)
	if err != nil {
		return nil, err
	}

	formats := req.Formats
	if len(formats) == 0 {
		formats = DefaultFormats()
	}
	for _, f := range formats {
		switch f {
		case FormatGIF, FormatAPNG, FormatWebP:
		default:
			return nil, fmt.Errorf("未知的动图格式 unknown format: %s (gif|apng|webp)", string(f))
		}
	}
	formats = dedupeFormats(formats)
	budget := int64(tier.MaxSizeMB) * 1024 * 1024

	ffmpegPath, err := e.deps.ResolveBinary(e.deps.ExecutableDir)

	var results []ExportResult
	for _, input := range req.Files {
		if !util.IsVideoFile(input) {
			continue
		}
		// APNG 无损：每个输入探测一次，算出压进体积预算的目标帧率。
		apngFPS := tier.FPS
		if containsFormat(formats, FormatAPNG) {
			apngFPS = e.probeAPNGFPS(ctx, ffmpegPath, input, tier)
		}
		for _, format := range formats {
			out := outputPathFor(input, tier.Name, format, req.OutputDir)
			if err := e.deps.MkdirAll(filepath.Dir(out), 0o755); err != nil {
				results = append(results, ExportResult{
					InputName:  filepath.Base(input),
					InputPath:  input,
					OutputPath: out,
					Format:     string(format),
					Tier:       tier.Name,
					Status:     "failed",
					Error:      fmt.Sprintf("创建输出目录失败: %v", err),
				})
				continue
			}
			if !req.Force && e.deps.PathExists(out) {
				skippedSize := fileSize(out)
				results = append(results, ExportResult{
					InputName:  filepath.Base(input),
					InputPath:  input,
					OutputPath: out,
					Format:     string(format),
					Tier:       tier.Name,
					Status:     "skipped",
					Size:       skippedSize,
					OverBudget: skippedSize > budget,
				})
				continue
			}

			args := BuildArgs(input, out, tier, format, apngFPS)
			runErr := e.deps.RunCommand(ctx, ffmpegPath, args)

			if runErr != nil {
				results = append(results, ExportResult{
					InputName:  filepath.Base(input),
					InputPath:  input,
					OutputPath: out,
					Format:     string(format),
					Tier:       tier.Name,
					Status:     "failed",
					Error:      runErr.Error(),
				})
				continue
			}
		results = append(results, ExportResult{
			InputName:  filepath.Base(input),
			InputPath:  input,
			OutputPath: out,
			Format:     string(format),
			Tier:       tier.Name,
			Status:     "success",
			Size:       fileSize(out),
			OverBudget: fileSize(out) > budget,
		})
	}
	}
	return results, nil
}

// minAPNGFPS 是 APNG 自适应帧率下限，再低动画会明显卡顿。
const minAPNGFPS = 2

// apngBudgetHeadroom 是预算余量：真实编码含文件头等固定开销且字节与帧率
// 不严格线性，只用 95% 预算反推帧率，避免压线超限。
const apngBudgetHeadroom = 0.95

// containsFormat 检查格式列表是否包含目标格式。
func containsFormat(formats []Format, want Format) bool {
	for _, f := range formats {
		if f == want {
			return true
		}
	}
	return false
}

// dedupeFormats 按首次出现顺序去重，避免重复格式导致重复编码。
func dedupeFormats(formats []Format) []Format {
	seen := make(map[Format]struct{}, len(formats))
	out := make([]Format, 0, len(formats))
	for _, f := range formats {
		if _, ok := seen[f]; ok {
			continue
		}
		seen[f] = struct{}{}
		out = append(out, f)
	}
	return out
}

// tierDurationSec 解析档位时长上限（"HH:MM:SS"），解析失败回退 5 秒。
func tierDurationSec(t Tier) float64 {
	parts := strings.Split(t.MaxDuration, ":")
	if len(parts) != 3 {
		return 5
	}
	h, err1 := strconv.Atoi(parts[0])
	m, err2 := strconv.Atoi(parts[1])
	s, err3 := strconv.Atoi(parts[2])
	if err1 != nil || err2 != nil || err3 != nil || h < 0 || m < 0 || s <= 0 {
		return 5
	}
	return float64(h*3600 + m*60 + s)
}

// adaptiveAPNGFPS 是纯函数：按 1 秒探测的每秒字节数，反推完整时长能压进
// 预算的目标帧率，结果钳制到 [minAPNGFPS, maxFPS]。
func adaptiveAPNGFPS(budgetBytes, probeBytesPerSec int64, durationSec float64, maxFPS int) int {
	if maxFPS < 1 {
		return 1
	}
	if probeBytesPerSec <= 0 || durationSec <= 0 {
		return maxFPS
	}
	fps := int(float64(budgetBytes) * float64(maxFPS) / (float64(probeBytesPerSec) * durationSec))
	if fps < minAPNGFPS {
		fps = minAPNGFPS
	}
	if fps > maxFPS {
		fps = maxFPS
	}
	return fps
}

// probeAPNGFPS 用 1 秒全帧率探测估算无损 APNG 每秒字节数，再按档位预算
// 与完整时长反推目标帧率。探测经 deps.RunCommand 执行（mock 下无产物则
// 回退全帧率）；ctx 已取消时直接回下限，让后续编码立刻因取消而失败。
func (e *Engine) probeAPNGFPS(ctx context.Context, ffmpegPath string, input string, tier Tier) int {
	if ctx.Err() != nil {
		return minAPNGFPS
	}
	budget := int64(float64(tier.MaxSizeMB) * 1024 * 1024 * apngBudgetHeadroom)

	probeDir, err := os.MkdirTemp("", "apngprobe-")
	if err != nil {
		return tier.FPS
	}
	defer os.RemoveAll(probeDir)
	probePath := filepath.Join(probeDir, "probe.apng")

	probeArgs := []string{
		"-y", "-i", input,
		"-t", "1",
		"-vf", fmt.Sprintf("fps=%d,%s", tier.FPS, scaleFilter(tier.MaxWidth)),
		"-plays", "0",
		"-pix_fmt", "rgba",
		"-compression_level", "6",
		probePath,
	}
	if err := e.deps.RunCommand(ctx, ffmpegPath, probeArgs); err != nil {
		return tier.FPS
	}
	return adaptiveAPNGFPS(budget, fileSize(probePath), tierDurationSec(tier), tier.FPS)
}

func fileSize(path string) int64 {
	info, err := os.Stat(path)
	if err != nil {
		return 0
	}
	return info.Size()
}