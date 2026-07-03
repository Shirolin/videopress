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
