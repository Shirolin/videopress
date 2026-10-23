package gui

import (
	"encoding/json"
	"os"
	"path/filepath"

	"videopress/internal/engine"
)

// prefs is the persisted user configuration, written to settings.json.
type prefs struct {
	Preset       string `json:"preset"`
	Concurrency  int    `json:"concurrency"`
	HWAccel      bool   `json:"hwAccel"`
	HWAccelSet   bool   `json:"hwAccelSet"`
	CopyAudio    bool   `json:"copyAudio"`
	ForceMode    bool   `json:"forceMode"`
	SkipExisting bool   `json:"skipExisting"`
	VideoCodec   string `json:"videoCodec"`
	MaxFPS       int    `json:"maxFps"`
	AudioMode    string `json:"audioMode"`
	CRFEnabled   bool   `json:"crfEnabled"`
	CRFValue     int    `json:"crfValue"`
	OutputDir    string `json:"outputDir"`
	Language     string `json:"language"`

	CropMode    string  `json:"cropMode"`
	CropW       int     `json:"cropW"`
	CropH       int     `json:"cropH"`
	CropX       int     `json:"cropX"`
	CropY       int     `json:"cropY"`
	CropCenterW int     `json:"cropCenterW"`
	CropCenterH int     `json:"cropCenterH"`
	CropScale   float64 `json:"cropScale"`

	ShowAdvanced   bool `json:"showAdvanced"`
	EnableDebugLog bool `json:"enableDebugLog"`
}

// defaultPrefs returns the recommended starting configuration.
func defaultPrefs() prefs {
	return prefs{
		Preset:      "standard",
		Concurrency: 1,
		HWAccel:     true,
		CopyAudio:   true,
		CropMode:    "off",
		CropCenterW: 16,
		CropCenterH: 9,
		CropScale:   1,
	}
}

// settingsPath is the preferences file next to the window state.
func settingsPath() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "Videopress", "settings.json"), nil
}

// loadPrefs reads settings.json, falling back to the defaults.
func loadPrefs() prefs {
	p := defaultPrefs()
	path, err := settingsPath()
	if err != nil {
		return p
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return p
	}
	if err := json.Unmarshal(data, &p); err != nil {
		return defaultPrefs()
	}
	if p.Concurrency < 1 {
		p.Concurrency = 1
	}
	if p.Preset == "" {
		p.Preset = "standard"
	}
	if p.CropMode == "" {
		p.CropMode = "off"
	}
	if p.CropScale <= 0 || p.CropScale > 1 {
		p.CropScale = 1
	}
	return p
}

// savePrefs writes settings.json atomically.
func (a *App) savePrefs() {
	path, err := settingsPath()
	if err != nil {
		return
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return
	}
	data, err := json.MarshalIndent(a.prefs, "", "  ")
	if err != nil {
		return
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return
	}
	_ = os.Rename(tmp, path)
}

// crfValueOrDefault is the effective CRF, the codec default until the user edits it.
func (p prefs) crfValueOrDefault() int {
	if p.CRFValue > 0 {
		return p.CRFValue
	}
	return defaultCRF(p.VideoCodec)
}

// crf is the CRF handed to the engine, zero when the user kept the preset.
func (p prefs) crf() int {
	if !p.CRFEnabled {
		return 0
	}
	return p.CRFValue
}

// cropSpec builds the engine's crop request, nil when cropping is off.
func (p prefs) cropSpec() *engine.CropSpec {
	switch p.CropMode {
	case "exact":
		if p.CropW > 0 && p.CropH > 0 {
			return &engine.CropSpec{Exact: true, W: p.CropW, H: p.CropH, X: positive(p.CropX), Y: positive(p.CropY), Scale: 1}
		}
	case "center":
		if p.CropCenterW > 0 && p.CropCenterH > 0 {
			scale := min(1, max(0.5, p.CropScale))
			return &engine.CropSpec{RatioW: p.CropCenterW, RatioH: p.CropCenterH, Scale: scale}
		}
	}
	return nil
}

func positive(v int) int {
	if v < 0 {
		return 0
	}
	return v
}
