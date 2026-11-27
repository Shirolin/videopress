package main

import (
	"math"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestPromoSFXAssetsLoadable(t *testing.T) {
	assetsDir, err := resolveAudioAssetsDir()
	if err != nil {
		t.Fatalf("resolveAudioAssetsDir: %v", err)
	}

	seen := make(map[string]bool)
	for _, cue := range promoSFXCues() {
		if seen[cue.file] {
			continue
		}
		seen[cue.file] = true
		p := filepath.Join(assetsDir, "sfx", cue.file)
		trk, err := loadWAVStereo48k(p)
		if err != nil {
			t.Fatalf("load SFX %s: %v", p, err)
		}
		if trk.rate != AudioSampleRate {
			t.Errorf("%s rate = %d, want %d", cue.file, trk.rate, AudioSampleRate)
		}
		if len(trk.l) == 0 || len(trk.l) != len(trk.r) {
			t.Errorf("%s invalid lengths: L=%d R=%d", cue.file, len(trk.l), len(trk.r))
		}
		var peak float64
		for i := range trk.l {
			if math.IsNaN(trk.l[i]) || math.IsInf(trk.l[i], 0) || math.IsNaN(trk.r[i]) || math.IsInf(trk.r[i], 0) {
				t.Fatalf("%s has NaN/Inf at frame %d", cue.file, i)
			}
			peak = max(peak, math.Abs(trk.l[i]), math.Abs(trk.r[i]))
		}
		if peak < 0.05 || peak > 0.99 {
			t.Errorf("%s unexpected peak %.4f", cue.file, peak)
		}
	}
}

func TestBuildSoundtrackTechnicalSpecs(t *testing.T) {
	assetsDir, err := resolveAudioAssetsDir()
	if err != nil {
		t.Fatalf("resolveAudioAssetsDir: %v", err)
	}
	if _, err := os.Stat(filepath.Join(assetsDir, "bgm.wav")); err != nil {
		t.Fatalf("missing required BGM asset cmd/promo/assets/bgm.wav: %v", err)
	}

	sound := BuildSoundtrack()
	if sound == nil {
		t.Fatal("BuildSoundtrack returned nil")
	}

	wantDuration := time.Duration(TotalFrames) * time.Second / time.Duration(FPS)
	gotDuration := sound.Duration()
	if diff := gotDuration - wantDuration; diff < -time.Millisecond || diff > time.Millisecond {
		t.Fatalf("sound duration = %v, want %v (diff %v)", gotDuration, wantDuration, diff)
	}

	samples, err := buildSoundtrackSamples()
	if err != nil {
		t.Fatalf("buildSoundtrackSamples: %v", err)
	}
	wantSamples := audioTotalStereoFrames() * AudioChannels
	if len(samples) != wantSamples {
		t.Fatalf("sample count = %d, want %d", len(samples), wantSamples)
	}

	var peak float64
	var sumL, sumR float64
	var sumSq float64
	nFrames := len(samples) / AudioChannels
	for i := range nFrames {
		l := float64(samples[2*i])
		r := float64(samples[2*i+1])
		if math.IsNaN(l) || math.IsInf(l, 0) || math.IsNaN(r) || math.IsInf(r, 0) {
			t.Fatalf("invalid sample at stereo frame %d: L=%v R=%v", i, l, r)
		}
		if a := math.Abs(l); a > peak {
			peak = a
		}
		if a := math.Abs(r); a > peak {
			peak = a
		}
		sumL += l
		sumR += r
		sumSq += l*l + r*r
	}

	peakDB := 20 * math.Log10(max(peak, 1e-12))
	rms := math.Sqrt(sumSq / float64(len(samples)))
	rmsDB := 20 * math.Log10(max(rms, 1e-12))
	dcL := math.Abs(sumL / float64(nFrames))
	dcR := math.Abs(sumR / float64(nFrames))

	t.Logf("soundtrack metrics: frames=%d samples=%d duration=%v peak=%.4f (%.2f dBFS) rms=%.4f (%.2f dBFS) dc=(%.6f, %.6f)",
		nFrames, len(samples), gotDuration, peak, peakDB, rms, rmsDB, dcL, dcR)

	if peak >= 0.95 {
		t.Errorf("peak %.4f (%.2f dBFS) too close to 0 dBFS; risk of AAC clipping", peak, peakDB)
	}
	if peak < 0.35 {
		t.Errorf("peak %.4f (%.2f dBFS) too quiet", peak, peakDB)
	}
	if rmsDB < -22.0 || rmsDB > -11.0 {
		t.Errorf("RMS %.2f dBFS outside balanced range [-22, -11] dBFS", rmsDB)
	}
	if dcL > 1e-3 || dcR > 1e-3 {
		t.Errorf("excessive DC offset: L=%.6f R=%.6f", dcL, dcR)
	}

	if math.Abs(float64(samples[0])) > 1e-4 || math.Abs(float64(samples[1])) > 1e-4 {
		t.Errorf("non-silent start sample: (%v, %v)", samples[0], samples[1])
	}
	lastL := math.Abs(float64(samples[len(samples)-2]))
	lastR := math.Abs(float64(samples[len(samples)-1]))
	if lastL > 1e-5 || lastR > 1e-5 {
		t.Errorf("non-silent final sample: (%v, %v)", lastL, lastR)
	}

	fadeStartIdx := int(int64(CueAudioFadeOutStart) * int64(AudioSampleRate) / int64(FPS))
	earlyTailRMS := windowRMS(samples, fadeStartIdx, fadeStartIdx+AudioSampleRate/4)
	lateTailRMS := windowRMS(samples, nFrames-AudioSampleRate/10, nFrames)
	if earlyTailRMS <= 1e-3 {
		t.Errorf("fade-out start region unexpectedly silent: rms=%.6f", earlyTailRMS)
	}
	if lateTailRMS >= earlyTailRMS*0.08 {
		t.Errorf("fade-out tail did not decay enough: early=%.5f late=%.5f", earlyTailRMS, lateTailRMS)
	}

	scenes := []struct {
		name       string
		startFrame int
		durFrames  int
	}{
		{"scene1", Scene1Start, Scene1Duration},
		{"scene2", Scene2Start, Scene2Duration},
		{"scene3", Scene3Start, Scene3Duration},
		{"scene4", Scene4Start, Scene4Duration},
		{"scene5", Scene5Start, Scene5Duration},
		{"scene6", Scene6Start, Scene6Duration},
	}
	for _, sc := range scenes {
		s0 := int(int64(sc.startFrame) * int64(AudioSampleRate) / int64(FPS))
		s1 := int(int64(sc.startFrame+sc.durFrames) * int64(AudioSampleRate) / int64(FPS))
		sRMS := windowRMS(samples, s0, min(s1, nFrames))
		t.Logf("%s: rms=%.4f (%.2f dBFS)", sc.name, sRMS, 20*math.Log10(max(sRMS, 1e-12)))
		if sRMS < 0.03 {
			t.Errorf("%s audio too quiet: rms=%.5f", sc.name, sRMS)
		}
	}
}

func windowRMS(samples []float32, startFrame, endFrame int) float64 {
	if endFrame <= startFrame {
		return 0
	}
	var sumSq float64
	for i := startFrame; i < endFrame; i++ {
		l := float64(samples[2*i])
		r := float64(samples[2*i+1])
		sumSq += l*l + r*r
	}
	return math.Sqrt(sumSq / float64(2*(endFrame-startFrame)))
}
