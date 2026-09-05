package gif

import (
	"fmt"
	"strings"
)

// Tier 定义一档动图导出的体积与分辨率规格。
type Tier struct {
	Name        string // "smooth" | "balanced" | "hd"
	MaxWidth    int    // 宽度上限（px）
	MaxSizeMB   int    // 文件大小上限（MB）
	FPS         int    // 目标帧率
	MaxDuration string // 截取片段时长上限（ffmpeg 时间格式，如 "00:00:06"）
	Description string
}

// DefaultTier 是默认模板（均衡档）。
const DefaultTier = "balanced"

var tiers = map[string]Tier{
	"smooth": {
		Name:        "smooth",
		MaxWidth:    240,
		MaxSizeMB:   1,
		FPS:         15,
		MaxDuration: "00:00:05",
		Description: "小体积流畅档，适合极速分享",
	},
	"balanced": {
		Name:        "balanced",
		MaxWidth:    360,
		MaxSizeMB:   2,
		FPS:         15,
		MaxDuration: "00:00:05",
		Description: "均衡档（默认模板），画质与体积平衡",
	},
	"hd": {
		Name:        "hd",
		MaxWidth:    480,
		MaxSizeMB:   5,
		FPS:         15,
		MaxDuration: "00:00:05",
		Description: "高清档，保留更多细节",
	},
}

// Format 支持的动图输出格式。
type Format string

const (
	FormatGIF  Format = "gif"
	FormatAPNG Format = "apng"
	FormatWebP Format = "webp"
)

// AllFormats 返回三种受支持格式。
func AllFormats() []Format {
	return []Format{FormatGIF, FormatAPNG, FormatWebP}
}

// DefaultFormats 是默认模板一键导出覆盖的三种格式。
func DefaultFormats() []Format {
	return AllFormats()
}

// TierByName 按名称返回档位（大小写不敏感）。
func TierByName(name string) (Tier, error) {
	t, ok := tiers[strings.ToLower(name)]
	if !ok {
		return Tier{}, fmt.Errorf("未知的动图档位 unknown tier: %s (smooth|balanced|hd)", name)
	}
	return t, nil
}

// AllTiers 返回所有档位。
func AllTiers() []Tier {
	list := []Tier{tiers["smooth"], tiers["balanced"], tiers["hd"]}
	return list
}

// Ext 返回格式对应的文件扩展名（含点）。
func (f Format) Ext() string {
	switch f {
	case FormatGIF:
		return ".gif"
	case FormatAPNG:
		return ".apng"
	case FormatWebP:
		return ".webp"
	default:
		return ".gif"
	}
}