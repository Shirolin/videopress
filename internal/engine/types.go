package engine

import "time"

// JobRequest defines the parameters for a video compression batch.
type JobRequest struct {
	Files        []string
	Preset       string // "small" | "standard" | "quality"
	HWAccel      bool
	CopyAudio    bool
	ForceMode    bool
	SkipExisting bool
	Concurrency  int
	OutputDir    string
	// Advanced configuration parameters
	VideoCodec string    // "" (auto) | "h264" | "h265" | "av1"
	MaxFPS     int       // 0 = unlimited, or positive integer like 30, 60
	AudioMode  string    // "" (fallback to CopyAudio) | "compress" | "copy" | "mute"
	CRF        int       // 0 = use preset default, or positive integer override
	Crop       *CropSpec // nil = 不裁剪，保持现有行为
}

// CropSpec 描述压缩前的可选几何裁剪配置。
type CropSpec struct {
	// Exact=true 时使用精确像素裁剪 W:H:X:Y（与 ffmpeg crop 一致）；
	// Exact=false 时为居中按比例裁剪（RatioW:RatioH 最大内接矩形，再乘 Scale 保留并居中）。
	Exact  bool
	W, H   int
	X, Y   int
	RatioW int
	RatioH int
	Scale  float64
}

// ProgressEvent represents progress updates from the compression engine.
type ProgressEvent struct {
	File    string  // 文件 base name（用于显示）
	Path    string  // 输入文件完整路径（用于精确匹配队列项）
	Percent float64 // 0 ~ 100
	Done    bool
	Error   string
}

// JobReport represents the summary of a completed video compression job.
type JobReport struct {
	InputName  string
	InputPath  string
	OutputDir  string
	Status     string // "success", "skipped", "failed"
	SourceSize int64
	TargetSize int64
	Duration   time.Duration
	ErrMessage string
}
