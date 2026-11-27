package main

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"

	"github.com/ELY-Open-Source/MujicaUI/icons"
	"github.com/ELY-Open-Source/MujicaUI/theme"
	"github.com/egoist/mygo/ui"

	"videopress/resources"
)

// FontStacks holds Windows CJK-safe font fallback lists resolved against installed system fonts.
type FontStacks struct {
	Serif string
	Sans  string
	Mono  string
}

// Assets holds all pre-decoded bitmaps and verified font stacks loaded once prior to rendering.
type Assets struct {
	AppIcon         *ui.Bitmap
	WorkspaceDark   *ui.Bitmap
	BatchProgress   *ui.Bitmap
	GifCustom       *ui.Bitmap
	AppearanceLight *ui.Bitmap
	Fonts           FontStacks
}

var (
	fontsOnce   sync.Once
	cachedFonts FontStacks
)

// DarkPalette returns MujicaUI's official dark charcoal token set.
func DarkPalette() theme.Tokens {
	return theme.Dark()
}

// LightPalette returns MujicaUI's official light ivory token set.
func LightPalette() theme.Tokens {
	return theme.Light()
}

// MixTokens linearly blends two MujicaUI token sets for smooth light/dark transitions.
func MixTokens(a, b theme.Tokens, t float32) theme.Tokens {
	if t <= 0 {
		return a
	}
	if t >= 1 {
		return b
	}
	return theme.Tokens{
		Dark:          t >= 0.5 && b.Dark || t < 0.5 && a.Dark,
		Background:    a.Background.Mix(b.Background, t),
		Surface:       a.Surface.Mix(b.Surface, t),
		SurfaceHover:  a.SurfaceHover.Mix(b.SurfaceHover, t),
		Text:          a.Text.Mix(b.Text, t),
		TextMuted:     a.TextMuted.Mix(b.TextMuted, t),
		Border:        a.Border.Mix(b.Border, t),
		ControlBorder: a.ControlBorder.Mix(b.ControlBorder, t),
		Accent:        a.Accent.Mix(b.Accent, t),
		AccentHover:   a.AccentHover.Mix(b.AccentHover, t),
		AccentPressed: a.AccentPressed.Mix(b.AccentPressed, t),
		OnAccent:      a.OnAccent.Mix(b.OnAccent, t),
		AccentText:    a.AccentText.Mix(b.AccentText, t),
		Selection:     a.Selection.Mix(b.Selection, t),
		Focus:         a.Focus.Mix(b.Focus, t),
		Ornament:      a.Ornament.Mix(b.Ornament, t),
		Success:       a.Success.Mix(b.Success, t),
		Warning:       a.Warning.Mix(b.Warning, t),
		Danger:        a.Danger.Mix(b.Danger, t),
		Info:          a.Info.Mix(b.Info, t),
		DangerFill:    a.DangerFill.Mix(b.DangerFill, t),
		OnDanger:      a.OnDanger.Mix(b.OnDanger, t),
	}
}

// ResolveFonts inspects installed fonts via ui.FontFamilies() and builds verified
// CJK-safe fallback stacks for Windows DirectWrite so Chinese glyphs never turn into tofu.
func ResolveFonts() FontStacks {
	fontsOnce.Do(func() {
		installed := make(map[string]string)
		for _, fam := range ui.FontFamilies() {
			installed[strings.ToLower(fam.Name)] = fam.Name
		}
		pick := func(candidates []string, generic string) string {
			var list []string
			seen := make(map[string]bool)
			for _, name := range candidates {
				if actual, ok := installed[strings.ToLower(name)]; ok && !seen[strings.ToLower(actual)] {
					list = append(list, actual)
					seen[strings.ToLower(actual)] = true
				}
			}
			if len(list) == 0 {
				return strings.Join(append(candidates, generic), ", ")
			}
			return strings.Join(append(list, generic), ", ")
		}

		cachedFonts = FontStacks{
			Serif: pick([]string{
				"Georgia",
				"Noto Serif SC",
				"Source Han Serif SC",
				"STSong",
				"Songti SC",
				"SimSun",
				"Microsoft YaHei UI",
			}, "serif"),
			Sans: pick([]string{
				"Segoe UI",
				"Microsoft YaHei UI",
				"Microsoft YaHei",
				"Noto Sans SC",
				"Source Han Sans SC",
				"DengXian",
			}, "sans-serif"),
			Mono: pick([]string{
				"Cascadia Mono",
				"Cascadia Code",
				"Consolas",
				"Sarasa Term SC Nerd",
				"Microsoft YaHei UI",
			}, "monospace"),
		}
	})
	return cachedFonts
}

// LoadAssets locates the repository root and decodes the app icon and real product screenshots once.
func LoadAssets(repoRoot string) (*Assets, error) {
	root, err := findRepoRoot(repoRoot)
	if err != nil {
		return nil, err
	}

	appIcon, err := ui.DecodeBitmap(resources.IconPNG)
	if err != nil {
		return nil, fmt.Errorf("promo: decode embedded icon: %w", err)
	}

	loadShot := func(name string) (*ui.Bitmap, error) {
		p := filepath.Join(root, "docs", "screenshots", name)
		data, err := os.ReadFile(p)
		if err != nil {
			return nil, fmt.Errorf("promo: read screenshot %s: %w", p, err)
		}
		bmp, err := ui.DecodeBitmap(data)
		if err != nil {
			return nil, fmt.Errorf("promo: decode screenshot %s: %w", p, err)
		}
		return bmp, nil
	}

	wsDark, err := loadShot("workspace-dark.png")
	if err != nil {
		return nil, err
	}
	batchProg, err := loadShot("batch-progress.png")
	if err != nil {
		return nil, err
	}
	gifCustom, err := loadShot("gif-custom.png")
	if err != nil {
		return nil, err
	}
	appLight, err := loadShot("appearance-light.png")
	if err != nil {
		return nil, err
	}

	return &Assets{
		AppIcon:         appIcon,
		WorkspaceDark:   wsDark,
		BatchProgress:   batchProg,
		GifCustom:       gifCustom,
		AppearanceLight: appLight,
		Fonts:           ResolveFonts(),
	}, nil
}

func findRepoRoot(explicit string) (string, error) {
	candidates := make([]string, 0, 6)
	if explicit != "" {
		candidates = append(candidates, explicit)
	}
	if wd, err := os.Getwd(); err == nil {
		candidates = append(candidates, wd, filepath.Dir(wd), filepath.Dir(filepath.Dir(wd)))
	}
	if _, file, _, ok := runtime.Caller(0); ok {
		// cmd/promo/kit.go -> repo root is two levels up
		candidates = append(candidates, filepath.Dir(filepath.Dir(filepath.Dir(file))))
	}
	for _, dir := range candidates {
		probe := filepath.Join(dir, "docs", "screenshots", "workspace-dark.png")
		if _, err := os.Stat(probe); err == nil {
			return dir, nil
		}
	}
	return "", fmt.Errorf("promo: could not locate repo root containing docs/screenshots/workspace-dark.png")
}

// IconSVG fetches a bundled MujicaUI Lucide icon, falling back safely to "circle".
func IconSVG(name string) *ui.SVG {
	if svg, ok := icons.Get(name); ok {
		return svg
	}
	svg, _ := icons.Get("circle")
	return svg
}

// SceneStage wraps a scene in a full-canvas 1920x1080 container with background fill and opacity.
func SceneStage(c *ui.Context, bg ui.Color, opacity float32, build func()) {
	ui.Box(c).Fill().Background(bg).Opacity(opacity).Children(build)
}

// SubtleBackdropFrame draws restrained editorial margin hairlines matching MujicaUI's print aesthetic.
func SubtleBackdropFrame(c *ui.Context, p theme.Tokens) {
	line := p.Border.Alpha(0.32)
	accentTick := p.Ornament.Alpha(0.45)

	// Top and bottom inset guide hairlines
	ui.Box(c).Absolute().Top(44).Left(64).Right(64).Height(1).Background(line)
	ui.Box(c).Absolute().Bottom(44).Left(64).Right(64).Height(1).Background(line)

	// Subtle corner ticks
	ui.Box(c).Absolute().Top(44).Left(64).Size(24, 2).Background(accentTick)
	ui.Box(c).Absolute().Bottom(44).Right(64).Size(24, 2).Background(accentTick)
}

// Hairline draws a crisp 1px-style architectural divider.
func Hairline(c *ui.Context, color ui.Color, w, h float32) ui.Element {
	return ui.Box(c).Size(w, h).Background(color).Shrink(0)
}

// EditorialRule draws a two-tone burgundy + warm ornament divider rule.
func EditorialRule(c *ui.Context, p theme.Tokens, width float32) {
	accentW := width * 0.36
	restW := width - accentW
	ui.Row(c).Gap(0).AlignItems(ui.Center).Children(func() {
		ui.Box(c).Size(accentW, 2).Background(p.Accent)
		ui.Box(c).Size(restW, 1).Background(p.Ornament.Alpha(0.65))
	})
}

// Eyebrow renders a restrained section label with a short leading accent rule.
func Eyebrow(c *ui.Context, text string, p theme.Tokens, fonts FontStacks) ui.Element {
	return ui.Row(c).Gap(12).AlignItems(ui.Center).Children(func() {
		ui.Box(c).Size(22, 2).Background(p.Accent).Shrink(0)
		ui.Text(c, text).
			Font(fonts.Sans).
			FontSize(20).
			FontWeight(600).
			TextColor(p.AccentText).
			SingleLine()
	})
}

// Heading renders a Serif display heading aligned with MujicaUI's editorial typography.
func Heading(c *ui.Context, text string, size float32, p theme.Tokens, fonts FontStacks) ui.Element {
	return ui.Text(c, text).
		Font(fonts.Serif).
		FontSize(size).
		FontWeight(600).
		LineHeight(1.24).
		TextColor(p.Text)
}

// BodyText renders clean sans-serif body copy with safe Windows Chinese fallback.
func BodyText(c *ui.Context, text string, size float32, color ui.Color, fonts FontStacks) ui.Element {
	return ui.Text(c, text).
		Font(fonts.Sans).
		FontSize(size).
		LineHeight(1.5).
		TextColor(color)
}

// MonoText renders clean monospaced text for CLI commands, presets, and parameters.
func MonoText(c *ui.Context, text string, size float32, color ui.Color, fonts FontStacks) ui.Element {
	return ui.Text(c, text).
		Font(fonts.Mono).
		FontSize(size).
		LineHeight(1.45).
		TextColor(color)
}

// PillTag renders a restrained MujicaUI-styled status or capability tag (sharp 2-4px radius, no neon).
func PillTag(c *ui.Context, text string, p theme.Tokens, fonts FontStacks, active bool) ui.Element {
	bg := p.Surface
	fg := p.TextMuted
	bd := p.Border
	if active {
		bg = p.Accent
		fg = p.OnAccent
		bd = p.Accent
	}
	return ui.Row(c).
		Padding(8, 16).
		Radius(theme.ControlRadius + 1).
		Background(bg).
		Border(theme.BorderWidth, bd).
		AlignItems(ui.Center).
		Children(func() {
			ui.Text(c, text).
				Font(fonts.Sans).
				FontSize(18).
				FontWeight(500).
				TextColor(fg).
				SingleLine()
		})
}

// SurfaceCard creates a flat, crisp MujicaUI card container with 1px border and standard CardRadius.
func SurfaceCard(c *ui.Context, p theme.Tokens, padY, padX float32, build func()) ui.Element {
	return ui.Column(c).
		Padding(padY, padX).
		Radius(theme.CardRadius).
		Background(p.Surface).
		Border(theme.BorderWidth, p.Border).
		Children(build)
}

// MetricTile renders a serif stat value above a muted caption, matching Videopress's workspace stats bar.
func MetricTile(c *ui.Context, value, label string, valueColor ui.Color, valueSize float32, p theme.Tokens, fonts FontStacks) {
	ui.Column(c).Gap(4).Children(func() {
		ui.Text(c, value).
			Font(fonts.Serif).
			FontSize(valueSize).
			FontWeight(600).
			TextColor(valueColor).
			SingleLine()
		ui.Text(c, label).
			Font(fonts.Sans).
			FontSize(17).
			TextColor(p.TextMuted).
			SingleLine()
	})
}

// ProgressBar renders a clean linear progress track and flat fill bar.
func ProgressBar(c *ui.Context, progress float64, width, height float32, fillColor ui.Color, p theme.Tokens) {
	prog := float32(Clamp01(progress))
	trackBg := p.SurfaceHover
	if p.Dark {
		trackBg = p.Border.Alpha(0.45)
	}
	ui.Box(c).Size(width, height).Radius(height / 2).Background(trackBg).Clip().Children(func() {
		fillW := width * prog
		if fillW > 0 {
			ui.Box(c).Size(fillW, height).Radius(height / 2).Background(fillColor)
		}
	})
}

// ScreenshotCard frames a real Videopress screenshot cleanly without duplicating its built-in title bar.
func ScreenshotCard(c *ui.Context, bmp *ui.Bitmap, width, height float32, p theme.Tokens) ui.Element {
	shadowCol := ui.RGBA(0, 0, 0, 0.36)
	if !p.Dark {
		shadowCol = ui.RGBA(33, 28, 29, 0.14)
	}
	return ui.Box(c).
		Size(width, height).
		Radius(theme.PopupRadius).
		Background(p.Surface).
		Border(theme.BorderWidth, p.ControlBorder.Alpha(0.75)).
		Shadow(0, 22, 48, 0, shadowCol).
		Clip().
		Children(func() {
			if bmp != nil {
				ui.Image(c, bmp).Size(width, height).Fit(ui.Contain)
			}
		})
}

// WindowFrame renders a clean desktop window chrome container (for terminal or custom showcase views).
func WindowFrame(c *ui.Context, title string, width, height float32, p theme.Tokens, fonts FontStacks, build func()) ui.Element {
	shadowCol := ui.RGBA(0, 0, 0, 0.38)
	if !p.Dark {
		shadowCol = ui.RGBA(33, 28, 29, 0.14)
	}
	barBg := p.SurfaceHover.Mix(p.Background, 0.45)
	return ui.Column(c).
		Size(width, height).
		Radius(theme.PopupRadius).
		Background(p.Surface).
		Border(theme.BorderWidth, p.ControlBorder.Alpha(0.75)).
		Shadow(0, 20, 44, 0, shadowCol).
		Clip().
		Children(func() {
			// Minimal Windows-style title strip
			ui.Row(c).
				FillWidth().
				Height(44).
				Padding(0, 18).
				Background(barBg).
				AlignItems(ui.Center).
				Children(func() {
					ui.Box(c).Size(8, 8).Radius(4).Background(p.Accent).Margin(0, 10, 0, 0)
					ui.Text(c, title).
						Font(fonts.Sans).
						FontSize(16).
						FontWeight(500).
						TextColor(p.TextMuted).
						SingleLine()
					ui.Spacer(c)
					ui.Row(c).Gap(10).AlignItems(ui.Center).Children(func() {
						ui.Box(c).Size(12, 1.5).Background(p.TextMuted.Alpha(0.65))
						ui.Box(c).Size(10, 10).Border(1.5, p.TextMuted.Alpha(0.65))
						ui.Text(c, "×").Font(fonts.Sans).FontSize(18).TextColor(p.TextMuted.Alpha(0.7))
					})
				})
			ui.Box(c).FillWidth().Height(1).Background(p.Border)
			ui.Box(c).FillWidth().Grow(1).MinHeight(0).Clip().Children(build)
		})
}

// BrandMark renders the Videopress application icon with subtle border and framing.
func BrandMark(c *ui.Context, bmp *ui.Bitmap, size float32, p theme.Tokens) ui.Element {
	shadowCol := ui.RGBA(0, 0, 0, 0.32)
	if !p.Dark {
		shadowCol = ui.RGBA(33, 28, 29, 0.14)
	}
	radius := size * 0.18
	return ui.Box(c).
		Size(size, size).
		Radius(radius).
		Background(p.Surface).
		Border(1, p.Ornament.Alpha(0.45)).
		Shadow(0, 12, 30, 0, shadowCol).
		Clip().
		Children(func() {
			if bmp != nil {
				ui.Image(c, bmp).Size(size, size).Fit(ui.Cover)
			}
		})
}
