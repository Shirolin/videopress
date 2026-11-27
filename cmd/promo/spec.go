package main

import (
	"math"

	"github.com/egoist/mygo/plugins/render"
	"github.com/egoist/mygo/ui"
)

// Video composition geometry and timing constants (1920x1080 @ 30fps, 26.0s).
const (
	Width       = 1920
	Height      = 1080
	FPS         = 30
	Scale       = 1
	TotalFrames = 780 // 26.0 seconds at 30fps

	BPM           = 120
	FramesPerBeat = 15 // 0.5s
	FramesPerBar  = 60 // 2.0s

	CrossfadeFrames = 12 // 0.4s smooth dissolve between adjacent scenes

	AudioSampleRate = 48000
	AudioChannels   = 2
)

// Scene start frames and durations. Adjacent scenes overlap by CrossfadeFrames.
const (
	Scene1Start    = 0
	Scene1Duration = 120 // 0..120 (4.0s)

	Scene2Start    = 108
	Scene2Duration = 147 // 108..255 (4.9s)

	Scene3Start    = 243
	Scene3Duration = 147 // 243..390 (4.9s)

	Scene4Start    = 378
	Scene4Duration = 147 // 378..525 (4.9s)

	Scene5Start    = 513
	Scene5Duration = 138 // 513..651 (4.6s)

	Scene6Start    = 639
	Scene6Duration = 141 // 639..780 (4.7s)
)

// Audio and visual cue points in global composition frames (aligned to 120 BPM).
const (
	CueBrandReveal       = 12  // 0.40s: Logo & wordmark reveal
	CueScene2In          = 108 // 3.60s: Enter Workspace scene
	CueWorkspaceSwap     = 174 // 5.80s: Cross-dissolve to batch-progress screenshot (Scene2 local 66)
	CueScene3In          = 243 // 8.10s: Enter GPU & Concurrency scene
	CueWorkerComplete    = 336 // 11.20s: Concurrent worker pool completion (Scene3 local 93)
	CueScene4In          = 378 // 12.60s: Enter Light Ivory GIF & Crop scene
	CueCropLock          = 456 // 15.20s: Geometric crop frame locks in (Scene4 local 78)
	CueScene5In          = 513 // 17.10s: Enter CLI & Windows Integration scene
	CueCLISummary        = 594 // 19.80s: CLI summary table printed (Scene5 local 81)
	CueScene6In          = 639 // 21.30s: Enter Theme Split & Finale scene
	CueFinaleLock        = 708 // 23.60s: Finale end-card lock-in (Scene6 local 69)
	CueAudioFadeOutStart = 735 // 24.50s: 1.5s fade-out to silence at frame 780
)

// FrameSeconds converts a frame index to composition seconds.
func FrameSeconds(frame int) float64 {
	return float64(frame) / float64(FPS)
}

// TotalDurationSeconds returns the exact video duration in seconds.
func TotalDurationSeconds() float64 {
	return float64(TotalFrames) / float64(FPS)
}

// Clamp01 restricts v to [0, 1].
func Clamp01(v float64) float64 {
	if v < 0 {
		return 0
	}
	if v > 1 {
		return 1
	}
	return v
}

// Lerp linearly interpolates between a and b by t in [0, 1].
func Lerp(a, b, t float64) float64 {
	return a + (b-a)*t
}

// Lerp32 linearly interpolates float32 values between a and b by t.
func Lerp32(a, b, t float32) float32 {
	return a + (b-a)*t
}

// EaseOutCubic provides a crisp deceleration curve.
func EaseOutCubic(t float32) float32 {
	u := 1 - t
	return 1 - u*u*u
}

// EaseInOutCubic provides smooth acceleration and deceleration.
func EaseInOutCubic(t float32) float32 {
	if t < 0.5 {
		return 4 * t * t * t
	}
	u := 2 - 2*t
	return 1 - u*u*u/2
}

// EaseOutQuart provides a slightly tighter editorial settle curve.
func EaseOutQuart(t float32) float32 {
	u := 1 - t
	u2 := u * u
	return 1 - u2*u2
}

// EaseOutQuint settles cleanly with minimal tail drift.
func EaseOutQuint(t float32) float32 {
	u := 1 - t
	u2 := u * u
	return 1 - u2*u2*u
}

// EaseInOutQuart provides smooth symmetric acceleration and deceleration.
func EaseInOutQuart(t float32) float32 {
	if t < 0.5 {
		return 8 * t * t * t * t
	}
	u := -2*t + 2
	u2 := u * u
	return 1 - (u2*u2)/2
}

// RangeProgress maps f.Exact() over [startFrame, endFrame] into [0, 1] using ease.
// If ease is nil, EaseOutCubic is used.
func RangeProgress(f render.Frame, startFrame, endFrame float64, ease ui.Easing) float64 {
	if endFrame <= startFrame {
		if f.Exact() >= startFrame {
			return 1
		}
		return 0
	}
	t := Clamp01((f.Exact() - startFrame) / (endFrame - startFrame))
	if ease == nil {
		ease = EaseOutCubic
	}
	return float64(ease(float32(t)))
}

// EnterOffset computes a restrained editorial entrance (opacity 0->1 and subtle upward glide).
func EnterOffset(f render.Frame, startFrame, duration float64, distance float32) (opacity float32, offsetY float32) {
	p := RangeProgress(f, startFrame, startFrame+duration, EaseOutQuart)
	op := Clamp01(RangeProgress(f, startFrame, startFrame+math.Max(duration*0.75, 1), EaseOutCubic))
	return float32(op), distance * float32(1-p)
}

// SceneFade returns the scene-level opacity envelope [0, 1] for fadeIn frames at
// the start of f and fadeOut frames at the end of f.
func SceneFade(f render.Frame, fadeIn, fadeOut int) float32 {
	alpha := 1.0
	if fadeIn > 0 && f.Exact() < float64(fadeIn) {
		alpha = min(alpha, RangeProgress(f, 0, float64(fadeIn), EaseInOutCubic))
	}
	if fadeOut > 0 && f.Count > fadeOut {
		outStart := float64(f.Count - fadeOut)
		outEnd := float64(f.Count - 1)
		if f.Exact() > outStart {
			alpha = min(alpha, 1.0-RangeProgress(f, outStart, outEnd, EaseInOutCubic))
		}
	}
	return float32(Clamp01(alpha))
}
