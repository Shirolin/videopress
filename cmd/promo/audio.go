package main

import (
	"fmt"
	"io"
	"math"
	"os"
	"path/filepath"
	"runtime"

	"github.com/egoist/mygo/audio"
	"github.com/egoist/mygo/plugins/render"
)

// Mix tuning constants calibrated for real BGM + studio UI SFX layering.
const (
	// bgmStartOffsetSec selects 12.0s inside assets/bgm.wav (44s, 120 BPM Ab Major):
	// 0.0s-4.0s plays the warm 2-bar Rhodes intro, 4.0s lands cleanly on the 16.0s groove entry,
	// and 20.0s-26.0s rides the 32.0s+ melodic lift into the finale fade-out.
	bgmStartOffsetSec = 12.0
	// bgmTargetRMS sets the nominal BGM groove loudness (~ -16.0 dBFS RMS) leaving clean headroom for SFX.
	bgmTargetRMS = 0.158
	// masterLimiterThreshold and masterLimiterCeiling keep peaks cleanly around -1.05 dBFS for AAC safety.
	masterLimiterThreshold = 0.74
	masterLimiterCeiling   = 0.886
)

type stereoTrack struct {
	l, r []float64
	rate int
}

type sfxCue struct {
	file       string
	timeSec    float64
	gain       float64
	pan        float64 // -1 (left) .. 0 (center) .. +1 (right)
	duckDB     float64 // gentle negative dB dip applied to BGM bus during this cue (e.g. -2.2)
	reverbSend float64 // subtle room glue send [0, 1]
}

// promoSFXCues defines the 12-cue studio sound-effect map aligned to spec.go.
func promoSFXCues() []sfxCue {
	return []sfxCue{
		// Scene 1: Brand reveal (frame 12 = 0.40s)
		{file: "brand_reveal.wav", timeSec: FrameSeconds(CueBrandReveal) - 0.04, gain: 0.40, pan: 0.00, duckDB: -1.8, reverbSend: 0.30},

		// Scene 2: Enter workspace (frame 108 = 3.60s) & screenshot swap (frame 174 = 5.80s)
		{file: "whoosh_soft.wav", timeSec: FrameSeconds(CueScene2In) - 0.12, gain: 0.36, pan: -0.12, duckDB: -1.5, reverbSend: 0.22},
		{file: "ui_thock.wav", timeSec: FrameSeconds(CueWorkspaceSwap), gain: 0.48, pan: 0.10, duckDB: -2.0, reverbSend: 0.14},
		{file: "ui_soft_click.wav", timeSec: FrameSeconds(CueWorkspaceSwap) + 0.015, gain: 0.26, pan: 0.14, duckDB: 0.0, reverbSend: 0.14},

		// Scene 3: Enter GPU & concurrency (frame 243 = 8.10s) & worker completion (frame 336 = 11.20s)
		{file: "whoosh_soft.wav", timeSec: FrameSeconds(CueScene3In) - 0.12, gain: 0.38, pan: 0.12, duckDB: -1.5, reverbSend: 0.22},
		{file: "worker_chime.wav", timeSec: FrameSeconds(CueWorkerComplete), gain: 0.50, pan: 0.00, duckDB: -2.6, reverbSend: 0.34},

		// Scene 4: Enter Light Ivory GIF & crop (frame 378 = 12.60s) & crop lock (frame 456 = 15.20s)
		{file: "whoosh_stereo_light.wav", timeSec: FrameSeconds(CueScene4In) - 0.18, gain: 0.42, pan: 0.00, duckDB: -1.8, reverbSend: 0.24},
		{file: "crop_shutter.wav", timeSec: FrameSeconds(CueCropLock), gain: 0.50, pan: 0.16, duckDB: -2.2, reverbSend: 0.18},

		// Scene 5: Enter dark CLI (frame 513 = 17.10s), subtle typing bed, and summary table (frame 594 = 19.80s)
		{file: "whoosh_tight.wav", timeSec: FrameSeconds(CueScene5In) - 0.14, gain: 0.38, pan: -0.10, duckDB: -1.5, reverbSend: 0.20},
		{file: "cli_typing.wav", timeSec: FrameSeconds(Scene5Start + 10), gain: 0.28, pan: -0.14, duckDB: -1.2, reverbSend: 0.08},
		{file: "cli_summary_tick.wav", timeSec: FrameSeconds(CueCLISummary), gain: 0.44, pan: -0.12, duckDB: -2.0, reverbSend: 0.22},
		{file: "ui_soft_click.wav", timeSec: FrameSeconds(CueCLISummary), gain: 0.20, pan: -0.08, duckDB: 0.0, reverbSend: 0.12},

		// Scene 6: Theme split (frame 639 = 21.30s) & Finale lockup (frame 708 = 23.60s)
		{file: "panel_slide.wav", timeSec: FrameSeconds(CueScene6In) - 0.12, gain: 0.34, pan: 0.00, duckDB: -1.6, reverbSend: 0.22},
		{file: "whoosh_stereo_light.wav", timeSec: FrameSeconds(CueScene6In) - 0.16, gain: 0.22, pan: 0.00, duckDB: 0.0, reverbSend: 0.24},
		{file: "finale_settle.wav", timeSec: FrameSeconds(CueFinaleLock), gain: 0.42, pan: 0.00, duckDB: -2.0, reverbSend: 0.24},
		{file: "brand_reveal.wav", timeSec: FrameSeconds(CueFinaleLock) - 0.04, gain: 0.30, pan: 0.00, duckDB: 0.0, reverbSend: 0.32},
	}
}

// BuildSoundtrack loads the mastered BGM track and curated UI/motion SFX WAVs
// from cmd/promo/assets/, mixes them at 48 kHz stereo aligned to spec.go Cues,
// applies gentle cue ducking, DC blocking, soft limiting, and smooth fades.
func BuildSoundtrack() *render.Sound {
	samples, err := buildSoundtrackSamples()
	if err != nil {
		panic(fmt.Sprintf("promo audio: %v", err))
	}
	return render.SoundOf(AudioSampleRate, AudioChannels, samples)
}

// audioTotalStereoFrames returns the exact number of stereo sample frames
// covering TotalFrames at FPS and AudioSampleRate.
func audioTotalStereoFrames() int {
	return int(int64(TotalFrames) * int64(AudioSampleRate) / int64(FPS))
}

func resolveAudioAssetsDir() (string, error) {
	if root, err := findRepoRoot(""); err == nil {
		dir := filepath.Join(root, "cmd", "promo", "assets")
		if st, err := os.Stat(dir); err == nil && st.IsDir() {
			return dir, nil
		}
	}
	if _, file, _, ok := runtime.Caller(0); ok {
		dir := filepath.Join(filepath.Dir(file), "assets")
		if st, err := os.Stat(dir); err == nil && st.IsDir() {
			return dir, nil
		}
	}
	return "", fmt.Errorf("could not locate cmd/promo/assets directory")
}

func buildSoundtrackSamples() ([]float32, error) {
	assetsDir, err := resolveAudioAssetsDir()
	if err != nil {
		return nil, err
	}

	const sr = float64(AudioSampleRate)
	n := audioTotalStereoFrames()

	bgmL := make([]float64, n)
	bgmR := make([]float64, n)
	sfxDryL := make([]float64, n)
	sfxDryR := make([]float64, n)
	sfxWetL := make([]float64, n)
	sfxWetR := make([]float64, n)
	duckEnv := make([]float64, n)
	for i := range duckEnv {
		duckEnv[i] = 1.0
	}

	// 1. Load & level-calibrate BGM bed (cmd/promo/assets/bgm.wav)
	bgmPath := filepath.Join(assetsDir, "bgm.wav")
	bgmTrack, err := loadWAVStereo48k(bgmPath)
	if err != nil {
		return nil, fmt.Errorf("load BGM %s: %w", bgmPath, err)
	}
	copyBGMWithCalibration(bgmL, bgmR, bgmTrack, bgmStartOffsetSec, bgmTargetRMS)

	// 2. Load & mix curated studio SFX clips (cached per filename)
	sfxCache := make(map[string]*stereoTrack)
	for _, cue := range promoSFXCues() {
		trk, ok := sfxCache[cue.file]
		if !ok {
			p := filepath.Join(assetsDir, "sfx", cue.file)
			trk, err = loadWAVStereo48k(p)
			if err != nil {
				return nil, fmt.Errorf("load SFX %s: %w", p, err)
			}
			sfxCache[cue.file] = trk
		}
		mixCue(sfxDryL, sfxDryR, sfxWetL, sfxWetR, trk, cue, sr)
		if cue.duckDB < 0 {
			applyDuckDip(duckEnv, cue.timeSec, float64(len(trk.l))/sr, cue.duckDB, sr)
		}
	}

	// 3. Subtle studio early-reflection glue on SFX wet bus only (keeps BGM pristine)
	revL, revR := applyShortRoomReverb(sfxWetL, sfxWetR)

	// 4. Combine ducked BGM + SFX + room glue
	mixL := make([]float64, n)
	mixR := make([]float64, n)
	const sfxReverbReturn = 0.22
	for i := range n {
		mixL[i] = bgmL[i]*duckEnv[i] + sfxDryL[i] + revL[i]*sfxReverbReturn
		mixR[i] = bgmR[i]*duckEnv[i] + sfxDryR[i] + revR[i]*sfxReverbReturn
	}

	// 5. DC blocker (20 Hz HPF)
	applyDCBlocker(mixL)
	applyDCBlocker(mixR)

	// 6. Smooth start fade-in, CueAudioFadeOutStart..TotalFrames fade-out, and C1 soft-knee limiter
	fadeInSamples := int(0.18 * sr) // 180ms smooth opening
	fadeOutStart := int(int64(CueAudioFadeOutStart) * int64(AudioSampleRate) / int64(FPS))
	fadeOutLen := n - fadeOutStart

	out := make([]float32, n*AudioChannels)
	for i := range n {
		env := 1.0
		if i < fadeInSamples {
			u := float64(i) / float64(fadeInSamples)
			env *= 0.5 * (1.0 - math.Cos(math.Pi*u))
		}
		if i >= fadeOutStart && fadeOutLen > 1 {
			u := float64(i-fadeOutStart) / float64(fadeOutLen-1)
			if u >= 1.0 {
				env = 0.0
			} else {
				c := 0.5 * (1.0 + math.Cos(math.Pi*u))
				env *= c * c
			}
		}

		l := softLimit(mixL[i], masterLimiterThreshold, masterLimiterCeiling) * env
		r := softLimit(mixR[i], masterLimiterThreshold, masterLimiterCeiling) * env

		if math.IsNaN(l) || math.IsInf(l, 0) {
			l = 0
		}
		if math.IsNaN(r) || math.IsInf(r, 0) {
			r = 0
		}
		out[2*i] = float32(l)
		out[2*i+1] = float32(r)
	}

	return out, nil
}

// loadWAVStereo48k decodes a WAV file via mygo's built-in audio.NewDecoder,
// converting mono to stereo and resampling linearly if needed to AudioSampleRate.
func loadWAVStereo48k(path string) (*stereoTrack, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	dec, _, err := audio.NewDecoder(f)
	if err != nil {
		return nil, err
	}
	ch := dec.Channels()
	rate := dec.SampleRate()
	if ch < 1 || rate <= 0 {
		return nil, fmt.Errorf("invalid audio format: channels=%d rate=%d", ch, rate)
	}

	var srcL, srcR []float64
	if n := dec.Len(); n > 0 && n < 1<<26 {
		srcL = make([]float64, 0, n)
		srcR = make([]float64, 0, n)
	}
	buf := make([]float32, 4096/ch*ch)
	for {
		n, rerr := dec.Read(buf)
		n -= n % ch
		for i := 0; i < n; i += ch {
			l := float64(buf[i])
			r := l
			if ch >= 2 {
				r = float64(buf[i+1])
			}
			srcL = append(srcL, l)
			srcR = append(srcR, r)
		}
		if rerr == io.EOF {
			break
		}
		if rerr != nil {
			return nil, rerr
		}
		if n == 0 {
			return nil, io.ErrNoProgress
		}
	}

	if rate == AudioSampleRate {
		return &stereoTrack{l: srcL, r: srcR, rate: AudioSampleRate}, nil
	}
	return resampleLinear(srcL, srcR, rate, AudioSampleRate), nil
}

func resampleLinear(srcL, srcR []float64, srcRate, dstRate int) *stereoTrack {
	if len(srcL) == 0 {
		return &stereoTrack{rate: dstRate}
	}
	outLen := int(math.Round(float64(len(srcL)) * float64(dstRate) / float64(srcRate)))
	if outLen < 1 {
		outLen = 1
	}
	dstL := make([]float64, outLen)
	dstR := make([]float64, outLen)
	step := float64(srcRate) / float64(dstRate)
	last := len(srcL) - 1
	for i := range outLen {
		pos := float64(i) * step
		idx := int(pos)
		if idx >= last {
			dstL[i] = srcL[last]
			dstR[i] = srcR[last]
			continue
		}
		frac := pos - float64(idx)
		dstL[i] = srcL[idx]*(1-frac) + srcL[idx+1]*frac
		dstR[i] = srcR[idx]*(1-frac) + srcR[idx+1]*frac
	}
	return &stereoTrack{l: dstL, r: dstR, rate: dstRate}
}

// copyBGMWithCalibration copies the BGM window starting at offsetSec into dstL/dstR
// (with gentle crossfaded loop only if BGM is shorter than the composition) and
// calibrates body RMS to targetRMS without crushing dynamics.
func copyBGMWithCalibration(dstL, dstR []float64, bgm *stereoTrack, offsetSec, targetRMS float64) {
	n := len(dstL)
	srcN := len(bgm.l)
	if srcN == 0 || n == 0 {
		return
	}
	startIdx := int(math.Round(offsetSec * float64(AudioSampleRate)))
	if startIdx < 0 {
		startIdx = 0
	}
	if startIdx >= srcN {
		startIdx = 0
	}

	for i := range n {
		si := startIdx + i
		if si < srcN {
			dstL[i] = bgm.l[si]
			dstR[i] = bgm.r[si]
		}
	}

	// Calibrate active groove RMS (4.0s .. CueAudioFadeOutStart) so the soft intro
	// does not skew gain and the main body hits targetRMS cleanly.
	grooveStart := min(n, int(4.0*float64(AudioSampleRate)))
	bodyEnd := min(n, int(FrameSeconds(CueAudioFadeOutStart)*float64(AudioSampleRate)))
	var sumSq float64
	for i := grooveStart; i < bodyEnd; i++ {
		l, r := dstL[i], dstR[i]
		sumSq += l*l + r*r
	}
	rms := math.Sqrt(sumSq / float64(max(1, 2*(bodyEnd-grooveStart))))
	gain := 1.0
	if rms > 1e-6 {
		gain = targetRMS / rms
		if gain > 3.0 {
			gain = 3.0
		} else if gain < 0.25 {
			gain = 0.25
		}
	}

	// Gently lift the quieter 0.0s..4.0s Rhodes intro (+5.2 dB tapering smoothly to 0 dB across 3.1s..4.0s)
	// so Scene 1 is warm and clearly audible while preserving a natural swell at the 4.0s beat drop.
	// Soft-knee tame isolated BGM transients above 0.66 so SFX transients always have clean headroom.
	introLift := math.Pow(10.0, 5.2/20.0)
	taperStart := int(3.1 * float64(AudioSampleRate))
	taperEnd := int(4.0 * float64(AudioSampleRate))
	for i := range n {
		g := gain
		if i < taperStart {
			g *= introLift
		} else if i < taperEnd {
			u := float64(i-taperStart) / float64(taperEnd-taperStart)
			w := 0.5 * (1.0 + math.Cos(math.Pi*u))
			g *= 1.0 + (introLift-1.0)*w
		}
		dstL[i] = softLimit(dstL[i]*g, 0.66, 0.78)
		dstR[i] = softLimit(dstR[i]*g, 0.66, 0.78)
	}
}

// mixCue adds a stereo SFX clip into dry/wet buses with gentle balance panning.
func mixCue(dryL, dryR, wetL, wetR []float64, trk *stereoTrack, cue sfxCue, sr float64) {
	startIdx := int(math.Round(cue.timeSec * sr))
	panL, panR := stereoBalanceGains(cue.pan)
	send := cue.reverbSend
	if send < 0 {
		send = 0
	} else if send > 1 {
		send = 1
	}

	for i := range trk.l {
		dst := startIdx + i
		if dst < 0 {
			continue
		}
		if dst >= len(dryL) {
			break
		}
		l := trk.l[i] * cue.gain * panL
		r := trk.r[i] * cue.gain * panR
		dryL[dst] += l
		dryR[dst] += r
		if send > 0 {
			wetL[dst] += l * send
			wetR[dst] += r * send
		}
	}
}

// stereoBalanceGains preserves stereo width while gently steering L/R balance.
func stereoBalanceGains(pan float64) (l, r float64) {
	if pan < -1 {
		pan = -1
	} else if pan > 1 {
		pan = 1
	}
	if pan >= 0 {
		return 1.0 - 0.35*pan, 1.0 + 0.15*pan
	}
	return 1.0 - 0.15*pan, 1.0 + 0.35*pan
}

// applyDuckDip smoothly lowers BGM gain envelope around a key SFX event
// (fast 25ms cosine attack, hold across active cue core, 220ms cosine release).
func applyDuckDip(env []float64, startSec, durSec, duckDB, sr float64) {
	if duckDB >= 0 {
		return
	}
	dipGain := math.Pow(10.0, duckDB/20.0)
	attack := 0.025
	hold := min(max(durSec*0.55, 0.06), 0.35)
	release := 0.22

	t0 := startSec - attack
	t1 := startSec
	t2 := startSec + hold
	t3 := t2 + release

	i0 := max(0, int(math.Floor(t0*sr)))
	i3 := min(len(env), int(math.Ceil(t3*sr)))
	for i := i0; i < i3; i++ {
		t := float64(i) / sr
		var g float64
		switch {
		case t < t1:
			u := (t - t0) / attack
			w := 0.5 * (1.0 - math.Cos(math.Pi*u))
			g = 1.0 - (1.0-dipGain)*w
		case t < t2:
			g = dipGain
		default:
			u := (t - t2) / release
			if u >= 1.0 {
				g = 1.0
			} else {
				w := 0.5 * (1.0 + math.Cos(math.Pi*u))
				g = 1.0 - (1.0-dipGain)*w
			}
		}
		if g < env[i] {
			env[i] = g
		}
	}
}

// softLimit applies a C1-continuous soft-knee tanh ceiling strictly bounded by [-ceil, +ceil].
func softLimit(x, threshold, ceiling float64) float64 {
	ax := math.Abs(x)
	if ax <= threshold {
		return x
	}
	headroom := ceiling - threshold
	if headroom <= 0 {
		if x > 0 {
			return ceiling
		}
		return -ceiling
	}
	excess := (ax - threshold) / headroom
	y := threshold + headroom*math.Tanh(excess)
	if x < 0 {
		return -y
	}
	return y
}

// applyDCBlocker removes DC offset using a first-order 20 Hz high-pass filter.
func applyDCBlocker(buf []float64) {
	const r = 0.9974
	var prevX, prevY float64
	for i, x := range buf {
		y := x - prevX + r*prevY
		prevX = x
		prevY = y
		buf[i] = y
	}
}

// -----------------------------------------------------------------------------
// Short Studio Room Reverb (gentle glue for SFX bus)
// -----------------------------------------------------------------------------

type combFilter struct {
	buf      []float64
	idx      int
	feedback float64
	damp     float64
	lpState  float64
}

func newCombFilter(delay int, feedback, damp float64) combFilter {
	return combFilter{buf: make([]float64, delay), feedback: feedback, damp: damp}
}

func (c *combFilter) process(in float64) float64 {
	out := c.buf[c.idx]
	c.lpState = out*(1.0-c.damp) + c.lpState*c.damp
	c.buf[c.idx] = in + c.lpState*c.feedback
	c.idx++
	if c.idx >= len(c.buf) {
		c.idx = 0
	}
	return out
}

type allpassFilter struct {
	buf      []float64
	idx      int
	feedback float64
}

func newAllpassFilter(delay int, feedback float64) allpassFilter {
	return allpassFilter{buf: make([]float64, delay), feedback: feedback}
}

func (a *allpassFilter) process(in float64) float64 {
	delayed := a.buf[a.idx]
	out := -in + delayed
	a.buf[a.idx] = in + delayed*a.feedback
	a.idx++
	if a.idx >= len(a.buf) {
		a.idx = 0
	}
	return out
}

func applyShortRoomReverb(inL, inR []float64) (outL, outR []float64) {
	n := len(inL)
	outL = make([]float64, n)
	outR = make([]float64, n)

	// Tight studio room (~0.45s warm decay) so UI clicks and chimes feel cohesive, not cavernous
	combsL := []combFilter{
		newCombFilter(1117, 0.62, 0.48),
		newCombFilter(1277, 0.61, 0.48),
		newCombFilter(1429, 0.60, 0.50),
		newCombFilter(1559, 0.59, 0.52),
	}
	combsR := []combFilter{
		newCombFilter(1151, 0.62, 0.48),
		newCombFilter(1307, 0.61, 0.48),
		newCombFilter(1459, 0.60, 0.50),
		newCombFilter(1597, 0.59, 0.52),
	}
	apL := []allpassFilter{
		newAllpassFilter(211, 0.48),
		newAllpassFilter(433, 0.48),
	}
	apR := []allpassFilter{
		newAllpassFilter(229, 0.48),
		newAllpassFilter(461, 0.48),
	}

	for i := range n {
		xL := inL[i]*0.80 + inR[i]*0.20
		xR := inR[i]*0.80 + inL[i]*0.20
		var sumL, sumR float64
		for k := range combsL {
			sumL += combsL[k].process(xL)
			sumR += combsR[k].process(xR)
		}
		sumL *= 0.25
		sumR *= 0.25
		for k := range apL {
			sumL = apL[k].process(sumL)
			sumR = apR[k].process(sumR)
		}
		outL[i] = sumL
		outR[i] = sumR
	}
	return outL, outR
}
