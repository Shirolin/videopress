package compress

import (
	"fmt"
	"strings"
)

type Preset struct {
	Name         string
	ScaleFactor  float64
	MaxDimension int
	CRF          int
	Preset       string
	AudioBitrate string
}

var presets = map[string]Preset{
	"small": {
		Name:         "small",
		ScaleFactor:  0.33,
		MaxDimension: 480,
		CRF:          30,
		Preset:       "veryfast",
		AudioBitrate: "64k",
	},
	"standard": {
		Name:         "standard",
		ScaleFactor:  0.50,
		MaxDimension: 720,
		CRF:          27,
		Preset:       "faster",
		AudioBitrate: "96k",
	},
	"quality": {
		Name:         "quality",
		ScaleFactor:  1.00,
		MaxDimension: 0,
		CRF:          24,
		Preset:       "medium",
		AudioBitrate: "128k",
	},
}

func PresetByName(name string) (Preset, error) {
	preset, ok := presets[strings.ToLower(name)]
	if !ok {
		return Preset{}, fmt.Errorf("未知的预设: %s", name)
	}
	return preset, nil
}
