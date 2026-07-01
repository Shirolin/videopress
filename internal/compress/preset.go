package compress

import "fmt"

type Preset struct {
	Name         string
	MaxDimension int
	CRF          int
	Preset       string
	AudioBitrate string
}

var presets = map[string]Preset{
	"small": {
		Name:         "small",
		MaxDimension: 720,
		CRF:          30,
		Preset:       "veryfast",
		AudioBitrate: "64k",
	},
	"standard": {
		Name:         "standard",
		MaxDimension: 1080,
		CRF:          27,
		Preset:       "faster",
		AudioBitrate: "96k",
	},
	"quality": {
		Name:         "quality",
		MaxDimension: 1440,
		CRF:          24,
		Preset:       "medium",
		AudioBitrate: "128k",
	},
}

func PresetByName(name string) (Preset, error) {
	preset, ok := presets[name]
	if !ok {
		return Preset{}, fmt.Errorf("unknown preset: %s", name)
	}
	return preset, nil
}
