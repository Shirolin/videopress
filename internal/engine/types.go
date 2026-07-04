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
	VideoCodec   string // "" (auto) | "h264" | "h265" | "av1"
	MaxFPS       int    // 0 = unlimited, or positive integer like 30, 60
	AudioMode    string // "" (fallback to CopyAudio) | "compress" | "copy" | "mute"
	CRF          int    // 0 = use preset default, or positive integer override
}

// ProgressEvent represents progress updates from the compression engine.
type ProgressEvent struct {
	File    string
	Percent float64 // 0 ~ 100
	Done    bool
	Error   string
}

// JobReport represents the summary of a completed video compression job.
type JobReport struct {
	InputName  string
	OutputDir  string
	Status     string // "成功", "跳过", "失败"
	SourceSize int64
	TargetSize int64
	Duration   time.Duration
	ErrMessage string
}
