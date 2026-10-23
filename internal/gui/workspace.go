package gui

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/egoist/mygo/ui"

	"github.com/ELY-Open-Source/MujicaUI/core"
	"github.com/ELY-Open-Source/MujicaUI/data"
	"github.com/ELY-Open-Source/MujicaUI/display"
	"github.com/ELY-Open-Source/MujicaUI/feedback"
	"github.com/ELY-Open-Source/MujicaUI/input"
	"github.com/ELY-Open-Source/MujicaUI/layout"
	"github.com/ELY-Open-Source/MujicaUI/navigation"
	"github.com/ELY-Open-Source/MujicaUI/theme"

	"videopress/internal/gif"
)

// workspaceView is the main page: a scrolling body with the actions pinned below.
func (a *App) workspaceView(c *ui.Context) {
	u := core.Density(c).Unit()
	ui.Column(c).FillHeight().Gap(u * 3).Children(func() {
		sc := layout.ScrollArea(c, a.pageScroll, layout.ScrollAreaOptions{
			Direction: layout.ScrollVertical, Label: a.t("nav.workspace"),
		}, func() {
			ui.Column(c).Gap(u * 3).Children(func() {
				a.statsBar(c)
				a.quickControls(c)
				a.advancedPanel(c)
				a.dropZone(c)
				a.queueList(c)
				a.gifPanel(c)
			})
		})
		// Leave the overlay scrollbar its own gutter so it does not draw
		// across the cards' right edges.
		sc.Grow(1).Shrink(1).MinHeight(0).Padding(0, u*3, 0, 0)
		a.actionPanel(c)
	})
}

// statsBar summarises the batch.
func (a *App) statsBar(c *ui.Context) {
	total, done, saved, original := a.stats()
	if total == 0 {
		return
	}
	k := core.Tokens(c)
	u := core.Density(c).Unit()

	var compressLeft int64
	for _, it := range a.queue {
		switch it.Status {
		case "waiting", "compressing":
			compressLeft++
		}
	}
	status := a.t("stats.idle")
	if compressLeft > 0 {
		status = a.t("stats.processing")
	}

	data.Card(c, data.CardOptions{}, func() {
		ui.Row(c).Gap(u * 6).Children(func() {
			statCell(c, strconv.FormatInt(total, 10), a.t("stats.total"), k.Text)
			statCell(c, strconv.FormatInt(compressLeft, 10), status, k.AccentText)
			statCell(c, strconv.FormatInt(done, 10), a.t("stats.completed"), k.Success)
			statCell(c, sizeText(saved), a.t("stats.saved_space"), k.Success)
			statCell(c, ratioText(saved, original), a.t("stats.avg_ratio"), k.TextMuted)
		})
	})
}

// statCell is one number with its caption.
func statCell(c *ui.Context, value, label string, color ui.Color) {
	ui.Column(c).Gap(2).MinWidth(90).Children(func() {
		ui.Text(c, value).FontSize(core.FontSize(c, theme.H3Size)).FontWeight(600).TextColor(color)
		ui.Text(c, label).FontSize(core.FontSize(c, theme.CaptionSize)).TextColor(core.Tokens(c).TextMuted)
	})
}

// quickControls picks the preset and the output folder.
func (a *App) quickControls(c *ui.Context) {
	u := core.Density(c).Unit()
	k := core.Tokens(c)

	data.Card(c, data.CardOptions{}, func() {
		ui.Row(c).Gap(u * 5).AlignItems(ui.End).Children(func() {
			ui.Column(c).Gap(u).Children(func() {
				a.sectionLabel(c, "controls.preset")
				values := []string{"small", "standard", "quality"}
				selected := 1
				for i, v := range values {
					if v == a.prefs.Preset {
						selected = i
					}
				}
				before := selected
				navigation.SegmentedControl(c, &selected, []navigation.Segment{
					{Label: a.t("preset.small")},
					{Label: a.t("preset.standard")},
					{Label: a.t("preset.quality")},
				}, navigation.SegmentedControlOptions{Label: a.t("controls.preset"), Disabled: a.isCompressing})
				if selected != before && !a.isCompressing {
					a.prefs.Preset = values[selected]
					a.savePrefs()
				}
			})

			ui.Column(c).Grow(1).MinWidth(0).Gap(u).Children(func() {
				ui.Row(c).Gap(u).AlignItems(ui.Center).Children(func() {
					a.sectionLabel(c, "controls.output")
					if a.prefs.CRFEnabled {
						display.Tag(c, a.t("advanced.crf.preset_override"), display.TagOptions{})
					}
				})
				ui.Row(c).Gap(u * 2).AlignItems(ui.Center).Children(func() {
					path := a.prefs.OutputDir
					text := path
					color := k.Text
					if text == "" {
						text = a.t("controls.default_output_path")
						color = k.TextMuted
					}
					ui.Row(c).Grow(1).MinWidth(0).Gap(u).AlignItems(ui.Center).Padding(0, u*2).
						Height(core.Density(c).ControlHeight()).Radius(theme.ControlRadius).
						Border(theme.BorderWidth, k.ControlBorder).Background(k.Background).Children(func() {
						ui.Icon(c, ic("folder")).Size(16, 16)
						ui.Text(c, text).TextColor(color).SingleLine().Grow(1).Shrink(1).MinWidth(0)
					})
					if input.Button(c, a.t("btn.change"), input.ButtonOptions{Variant: input.Secondary, Disabled: a.isCompressing}).Clicked() {
						a.pickOutputDir()
					}
					if path != "" && input.Button(c, a.t("btn.reset"), input.ButtonOptions{
						Variant: input.Ghost, Disabled: a.isCompressing,
					}).Clicked() {
						a.prefs.OutputDir = ""
						a.savePrefs()
					}
				})
			})
		})
	})
}

// advancedPanel folds the codec, frame rate, audio, CRF and crop settings.
func (a *App) advancedPanel(c *ui.Context) {
	u := core.Density(c).Unit()
	open := a.showAdvanced
	opts := data.CollapsibleOptions{Title: a.t("advanced.title")}
	res := data.Collapsible(c, &open, opts, func() {
		ui.Column(c).Gap(u*3).Padding(u*3, 0).Children(func() {
			ui.Row(c).Gap(u * 4).Children(func() {
				a.codecField(c)
				a.fpsField(c)
				a.audioField(c)
			})
			a.crfField(c)
			a.cropField(c)
		})
	})
	if res.Changed() {
		a.showAdvanced = open
		a.prefs.ShowAdvanced = open
		a.savePrefs()
	}
}

// codecField chooses the video codec.
func (a *App) codecField(c *ui.Context) {
	u := core.Density(c).Unit()
	codecs := []string{"", "h264", "h265", "av1"}
	labels := map[string]string{
		"":     a.t("advanced.video_codec_auto"),
		"h264": "H.264 (AVC)",
		"h265": "H.265 (HEVC)",
		"av1":  "AV1 (Next-Gen)",
	}
	options := make([]input.SelectOption[string], 0, len(codecs))
	for _, code := range codecs {
		options = append(options, input.SelectOption[string]{Value: code, Label: labels[code]})
	}
	ui.Column(c).Grow(1).Gap(u).Children(func() {
		a.sectionLabel(c, "advanced.video_codec")
		value := a.prefs.VideoCodec
		res := input.Select(c, &value, options, input.SelectOptions{Label: a.t("advanced.video_codec"), Disabled: a.isCompressing})
		if res.Changed() {
			a.onCodecChange(value)
		}
		ui.Text(c, a.t("advanced.video_codec_desc")).FontSize(core.FontSize(c, theme.CaptionSize)).
			TextColor(core.Tokens(c).TextMuted)
	})
}

// onCodecChange keeps the CRF value inside the new codec's range.
func (a *App) onCodecChange(codec string) {
	prevDefault := defaultCRF(a.prefs.VideoCodec)
	a.prefs.VideoCodec = codec
	d, lo, hi := crfRange(codec)
	switch {
	case a.prefs.CRFValue == 0 || a.prefs.CRFValue == prevDefault:
		a.prefs.CRFValue = d
	default:
		a.prefs.CRFValue = min(hi, max(lo, a.prefs.CRFValue))
	}
	a.savePrefs()
	a.detectGPU()
}

// fpsField limits the frame rate.
func (a *App) fpsField(c *ui.Context) {
	u := core.Density(c).Unit()
	options := []input.SelectOption[int]{
		{Value: 0, Label: a.t("advanced.max_fps_auto")},
		{Value: 30, Label: "30 FPS"},
		{Value: 60, Label: "60 FPS"},
	}
	ui.Column(c).Grow(1).Gap(u).Children(func() {
		a.sectionLabel(c, "advanced.max_fps")
		value := a.prefs.MaxFPS
		if input.Select(c, &value, options, input.SelectOptions{Label: a.t("advanced.max_fps"), Disabled: a.isCompressing}).Changed() {
			a.prefs.MaxFPS = value
			a.savePrefs()
		}
		ui.Text(c, a.t("advanced.max_fps_desc")).FontSize(core.FontSize(c, theme.CaptionSize)).
			TextColor(core.Tokens(c).TextMuted)
	})
}

// audioField chooses the audio handling.
func (a *App) audioField(c *ui.Context) {
	u := core.Density(c).Unit()
	options := []input.SelectOption[string]{
		{Value: "", Label: a.t("advanced.audio_mode_auto")},
		{Value: "copy", Label: a.t("advanced.audio_mode_copy")},
		{Value: "mute", Label: a.t("advanced.audio_mode_mute")},
	}
	ui.Column(c).Grow(1).Gap(u).Children(func() {
		a.sectionLabel(c, "advanced.audio_mode")
		value := a.prefs.AudioMode
		if input.Select(c, &value, options, input.SelectOptions{Label: a.t("advanced.audio_mode"), Disabled: a.isCompressing}).Changed() {
			a.prefs.AudioMode = value
			a.savePrefs()
		}
		ui.Text(c, a.t("advanced.audio_mode_desc")).FontSize(core.FontSize(c, theme.CaptionSize)).
			TextColor(core.Tokens(c).TextMuted)
	})
}

// crfField is the quality override with its slider.
func (a *App) crfField(c *ui.Context) {
	u := core.Density(c).Unit()
	k := core.Tokens(c)
	d, lo, hi := crfRange(a.prefs.VideoCodec)

	ui.Column(c).Gap(u * 2).Children(func() {
		ui.Row(c).Gap(u * 3).AlignItems(ui.Center).Children(func() {
			enabled := a.prefs.CRFEnabled
			if input.Checkbox(c, &enabled, a.t("advanced.crf.title"), input.CheckboxOptions{Disabled: a.isCompressing}).Changed() {
				a.prefs.CRFEnabled = enabled
				if enabled && (a.prefs.CRFValue < lo || a.prefs.CRFValue > hi) {
					a.prefs.CRFValue = d
				}
				a.savePrefs()
			}
			if !a.prefs.CRFEnabled {
				return
			}
			raw := float64(a.prefs.crfValueOrDefault())
			if input.NumberInput(c, &raw, input.NumberInputOptions{
				Min: float64(lo), Max: float64(hi), Step: 1, Label: "CRF",
			}).Changed() {
				a.prefs.CRFValue = int(raw)
				a.savePrefs()
			}
			display.Tag(c, crfWord(a.prefs.crfValueOrDefault(), a.prefs.VideoCodec, a), display.TagOptions{})
			if input.Button(c, a.t("btn.reset"), input.ButtonOptions{Variant: input.Ghost}).Clicked() {
				a.prefs.CRFValue = d
				a.savePrefs()
			}
		})
		if !a.prefs.CRFEnabled {
			ui.Text(c, crfDesc(a.prefs.VideoCodec, a)).FontSize(core.FontSize(c, theme.CaptionSize)).TextColor(k.TextMuted)
			return
		}
		raw := float64(a.prefs.CRFValue)
		res := input.Slider(c, &raw, input.SliderOptions{
			Min: float64(lo), Max: float64(hi), Step: 1, ShowValue: true,
			Marks:    []input.SliderMark{{Value: float64(d), Label: a.t("btn.reset")}},
			Disabled: a.isCompressing, Label: a.t("advanced.crf.title"),
		})
		if res.Changed() {
			a.prefs.CRFValue = int(raw)
			a.savePrefs()
		}
		ui.Text(c, crfDesc(a.prefs.VideoCodec, a)).FontSize(core.FontSize(c, theme.CaptionSize)).TextColor(k.TextMuted)
	})
}

// cropField picks the crop mode and its parameters.
func (a *App) cropField(c *ui.Context) {
	u := core.Density(c).Unit()
	options := []input.SelectOption[string]{
		{Value: "off", Label: a.t("advanced.crop.mode_off")},
		{Value: "exact", Label: a.t("advanced.crop.mode_exact")},
		{Value: "center", Label: a.t("advanced.crop.mode_center")},
	}

	ui.Column(c).Gap(u * 2).Children(func() {
		ui.Row(c).Gap(u * 3).AlignItems(ui.Center).Children(func() {
			a.sectionLabel(c, "advanced.crop.title")
			value := a.prefs.CropMode
			if input.Select(c, &value, options, input.SelectOptions{Label: a.t("advanced.crop.mode"), Disabled: a.isCompressing}).Changed() {
				a.prefs.CropMode = value
				a.savePrefs()
			}
		})
		ui.Text(c, a.t("advanced.crop.desc")).FontSize(core.FontSize(c, theme.CaptionSize)).TextColor(core.Tokens(c).TextMuted)

		switch a.prefs.CropMode {
		case "exact":
			ui.Row(c).Gap(u * 3).AlignItems(ui.Center).Children(func() {
				a.intField(c, "advanced.crop.size_label", &a.prefs.CropW, 2, 8192, "1920")
				a.intField(c, "advanced.crop.offset_label", &a.prefs.CropX, 0, 8192, "0")
				a.intField(c, "", &a.prefs.CropY, 0, 8192, "0")
			})
		case "center":
			ui.Row(c).Gap(u * 3).AlignItems(ui.Center).Children(func() {
				a.intField(c, "advanced.crop.center", &a.prefs.CropCenterW, 1, 100, "16")
				a.intField(c, "", &a.prefs.CropCenterH, 1, 100, "9")
				raw := a.prefs.CropScale
				res := input.Slider(c, &raw, input.SliderOptions{
					Min: 0.5, Max: 1, Step: 0.05, ShowValue: true,
					Format:   func(v float64) string { return strconv.Itoa(int(v*100)) + "%" },
					Disabled: a.isCompressing, Label: a.t("advanced.crop.scale"),
				})
				if res.Changed() {
					a.prefs.CropScale = raw
					a.savePrefs()
				}
			})
		}
	})
}

// intField is a labelled integer input bound to a preference.
func (a *App) intField(c *ui.Context, labelKey string, target *int, lo, hi int, placeholder string) {
	u := core.Density(c).Unit()
	ui.Column(c).Gap(u).Children(func() {
		if labelKey != "" {
			a.sectionLabel(c, labelKey)
		}
		raw := float64(*target)
		if input.NumberInput(c, &raw, input.NumberInputOptions{
			Min: float64(lo), Max: float64(hi), Step: 1, Placeholder: placeholder,
			Label: placeholder, Disabled: a.isCompressing,
		}).Changed() {
			*target = int(raw)
			a.savePrefs()
		}
	})
}

// dropZone takes dropped files; compact when the queue has rows.
func (a *App) dropZone(c *ui.Context) {
	u := core.Density(c).Unit()
	k := core.Tokens(c)
	compact := len(a.queue) > 0

	e := ui.Box(c).FillWidth().Column().Gap(u*1.5).Center().Padding(u*3).
		Radius(theme.CardRadius).Border(theme.BorderWidth, k.ControlBorder).Background(k.Surface).
		Disabled(a.isCompressing)
	if compact {
		e.PaddingY(u * 2.5)
	}
	if e.FileDragOver() {
		e.Background(k.Selection).Border(theme.BorderWidth*2, k.AccentText)
	}
	var label, sub string
	if e.FileDragOver() {
		label, sub = a.t("dropzone.drag_hint"), a.t("dropzone.formats")
	} else if compact || len(a.queue) > 0 {
		label, sub = a.t("dropzone.add_more"), a.t("dropzone.formats")
	} else {
		label, sub = a.t("dropzone.title"), a.t("dropzone.desc")
	}
	if a.isCompressing {
		sub = a.t("dropzone.disabled")
	}

	e.Children(func() {
		if !compact {
			ui.Icon(c, ic("file")).Size(28, 28).TextColor(k.TextMuted)
		}
		ui.Text(c, label).FontSize(core.FontSize(c, theme.BodySize+1)).FontWeight(600)
		ui.Text(c, sub).FontSize(core.FontSize(c, theme.CaptionSize)).TextColor(k.TextMuted)
		if input.Button(c, a.t("dropzone.browse"), input.ButtonOptions{
			Variant: input.Outline, Icon: ic("folder"), Disabled: a.isCompressing,
		}).Clicked() {
			a.pickFiles()
		}
	})

	if files := e.DroppedFiles(); len(files) > 0 {
		a.addFiles(files)
	}
}

// queueList shows the queued files with progress and results.
func (a *App) queueList(c *ui.Context) {
	u := core.Density(c).Unit()
	k := core.Tokens(c)

	ui.Column(c).Gap(u * 2).Children(func() {
		// With no rows the drop zone above already says it all.
		if len(a.queue) == 0 {
			return
		}
		ui.Row(c).Gap(u * 2).AlignItems(ui.Center).Children(func() {
			h3(c, a.t("queue.title"))
			ui.Text(c, strconv.Itoa(len(a.queue))).FontSize(core.FontSize(c, theme.CaptionSize)).
				FontWeight(600).TextColor(k.TextMuted)
			ui.Spacer(c)
			if input.Button(c, a.t("queue.clear"), input.ButtonOptions{
				Variant: input.Ghost, Disabled: a.isCompressing || a.isGifExporting,
			}).Clicked() {
				a.clearQueue()
			}
		})

		ui.Column(c).Gap(u * 2).Children(func() {
			for i := range a.queue {
				a.queueRow(c, i)
			}
		})
	})
}

// queueRow draws one queued file.
func (a *App) queueRow(c *ui.Context, index int) {
	u := core.Density(c).Unit()
	k := core.Tokens(c)
	item := &a.queue[index]

	row := ui.Column(c).Gap(u).Padding(u*2, u*3).Radius(theme.CardRadius).
		Border(theme.BorderWidth, k.Border).Background(k.Surface).Key(strconv.Itoa(index))

	status, tone, glyph := queueStatus(c, item, a)
	row.Children(func() {
		ui.Row(c).Gap(u * 2).AlignItems(ui.Center).Children(func() {
			ui.Icon(c, glyph).Size(16, 16).TextColor(tone)
			ui.Column(c).Grow(1).MinWidth(0).Gap(1).Children(func() {
				ui.Text(c, item.Name).SingleLine().FontWeight(500)
				ui.Text(c, item.Path).FontSize(core.FontSize(c, theme.CaptionSize)).
					TextColor(k.TextMuted).SingleLine()
			})
			ui.Text(c, status).FontSize(core.FontSize(c, theme.CaptionSize)).FontWeight(600).TextColor(tone)
			if item.Status == "waiting" && !a.isCompressing {
				if input.IconButton(c, ic("x"), a.t("queue.remove"), input.IconButtonOptions{Variant: input.Ghost}).Clicked() {
					a.removeQueue(index)
				}
			}
			if item.Status == "success" || item.Status == "skipped" {
				if input.IconButton(c, ic("folder"), a.t("queue.open_folder"), input.IconButtonOptions{Variant: input.Ghost}).Clicked() {
					a.openFolder(outputDirFor(item.Path, a.prefs.OutputDir))
				}
			}
		})

		bar := item.Percent / 100
		switch item.Status {
		case "success", "skipped", "failed":
			bar = 1
		}
		if item.Status != "waiting" {
			feedback.Progress(c, clamp01(bar), feedback.ProgressOptions{Shape: feedback.ProgressLinear, Label: item.Name})
		}

		if item.Error != "" {
			tone := k.Danger
			if item.Warning && item.Status == "waiting" {
				tone = k.Warning
			}
			ui.Text(c, item.Error).FontSize(core.FontSize(c, theme.CaptionSize)).TextColor(tone)
		}

		if item.Status == "success" || item.Status == "skipped" || item.Status == "failed" {
			ui.Row(c).Gap(u * 4).Children(func() {
				resultStat(c, a.t("queue.origin_size"), sizeText(item.Size), k.Text)
				if item.TargetSize > 0 {
					resultStat(c, a.t("queue.target_size"), sizeText(item.TargetSize), k.Success)
				}
				if item.Size > 0 && item.TargetSize > 0 {
					resultStat(c, a.t("queue.save_ratio"), ratioText(item.Size-item.TargetSize, item.Size), k.TextMuted)
				}
				if item.Duration > 0 {
					resultStat(c, a.t("queue.duration"), durationText(item.Duration), k.TextMuted)
				}
			})
		}
	})
}

func resultStat(c *ui.Context, label, value string, color ui.Color) {
	ui.Row(c).Gap(4).AlignItems(ui.Baseline).Children(func() {
		ui.Text(c, label).FontSize(core.FontSize(c, theme.CaptionSize)).TextColor(core.Tokens(c).TextMuted)
		ui.Text(c, value).FontSize(core.FontSize(c, theme.CaptionSize)).FontWeight(600).TextColor(color)
	})
}

// actionPanel holds the start/cancel buttons.
func (a *App) actionPanel(c *ui.Context) {
	u := core.Density(c).Unit()
	ui.Row(c).Gap(u * 3).AlignItems(ui.Center).Children(func() {
		hasResult := false
		for _, it := range a.queue {
			switch it.Status {
			case "success", "skipped":
				hasResult = true
			}
		}
		if hasResult && input.Button(c, a.t("btn.open_output_folder"), input.ButtonOptions{
			Variant: input.Secondary, Icon: ic("folder"),
		}).Clicked() {
			a.openResultFolder()
		}
		ui.Spacer(c)
		if a.isCompressing {
			if input.Button(c, a.t("btn.cancel_compress"), input.ButtonOptions{Variant: input.Danger, Icon: ic("square")}).Clicked() {
				a.cancelCompress()
			}
			return
		}
		disabled := len(a.queue) == 0 || a.isGifExporting || a.ffmpegError != ""
		if input.Button(c, a.t("btn.start_compress"), input.ButtonOptions{
			Variant: input.Primary, Icon: ic("play"), Disabled: disabled,
		}).Clicked() && !disabled {
			a.startCompress()
		}
	})
}

// gifPanel exports the queue as animated clips.
func (a *App) gifPanel(c *ui.Context) {
	u := core.Density(c).Unit()
	k := core.Tokens(c)

	data.Card(c, data.CardOptions{Title: a.t("gif.title")}, func() {
		ui.Column(c).Gap(u * 2).Children(func() {
			ui.Text(c, a.t("gif.subtitle")).FontSize(core.FontSize(c, theme.CaptionSize)).TextColor(k.TextMuted)

			ui.Row(c).Gap(u * 4).AlignItems(ui.End).Children(func() {
				ui.Column(c).Gap(u).Children(func() {
					a.sectionLabel(c, "gif.tier")
					tiers := gifTierNames()
					selected := 1
					for i, t := range tiers {
						if t == a.gifTier {
							selected = i
						}
					}
					before := selected
					navigation.SegmentedControl(c, &selected, []navigation.Segment{
						{Label: a.t("gif.tier_smooth")},
						{Label: a.t("gif.tier_balanced")},
						{Label: a.t("gif.tier_hd")},
					}, navigation.SegmentedControlOptions{Label: a.t("gif.tier"), Disabled: a.gifBusy || a.isCompressing})
					if selected != before && !a.gifBusy {
						a.gifTier = tiers[selected]
					}
				})

				ui.Column(c).Gap(u).Children(func() {
					a.sectionLabel(c, "gif.format")
					ui.Row(c).Gap(u * 3).AlignItems(ui.Center).Children(func() {
						for _, name := range []string{"gif", "apng", "webp"} {
							on := a.gifFormats[name]
							label := strings.ToUpper(name)
							if input.Checkbox(c, &on, label, input.CheckboxOptions{Disabled: a.gifBusy || a.isCompressing}).Changed() {
								a.gifFormats[name] = on
							}
						}
					})
				})

				ui.Spacer(c)
				if a.gifBusy {
					if input.Button(c, a.t("gif.cancel"), input.ButtonOptions{Variant: input.Danger}).Clicked() {
						a.cancelGifExport()
					}
				} else {
					disabled := len(a.queue) == 0 || a.isCompressing || a.ffmpegError != ""
					if input.Button(c, a.t("gif.export_btn"), input.ButtonOptions{
						Variant: input.Secondary, Icon: ic("image"), Disabled: disabled,
					}).Clicked() && !disabled {
						a.startGifExport()
					}
				}
			})

			if spec, ok := a.gifTierSpec(a.gifTier); ok {
				ui.Text(c, spec).FontSize(core.FontSize(c, theme.CaptionSize)).TextColor(k.TextMuted)
			}
			if a.isCompressing && !a.gifBusy {
				ui.Text(c, a.t("gif.busy_note")).FontSize(core.FontSize(c, theme.CaptionSize)).TextColor(k.TextMuted)
			}
			if a.gifBusy {
				feedback.Progress(c, 0, feedback.ProgressOptions{Shape: feedback.ProgressLinear, Indeterminate: true, Label: a.t("gif.export_btn")})
			}
			if a.gifError != "" {
				ui.Text(c, a.gifError).FontSize(core.FontSize(c, theme.CaptionSize)).TextColor(k.Danger)
			}
			if len(a.gifResults) > 0 {
				a.gifSummary(c)
			}
		})
	})
}

// gifSummary lists the exported clips.
func (a *App) gifSummary(c *ui.Context) {
	u := core.Density(c).Unit()
	k := core.Tokens(c)
	var ok, skipped, over, failed int
	for _, r := range a.gifResults {
		switch {
		case r.OverBudget:
			over++
		case r.Status == "success":
			ok++
		case r.Status == "skipped":
			skipped++
		default:
			failed++
		}
	}
	ui.Column(c).Gap(u).Children(func() {
		ui.Text(c, fmt.Sprintf("%s %d · %s %d · %s %d · %s %d",
			a.t("gif.ok"), ok, a.t("gif.skipped"), skipped, a.t("gif.over_budget"), over, a.t("gif.failed"), failed)).
			FontSize(core.FontSize(c, theme.CaptionSize)).FontWeight(600)
		for _, r := range a.gifResults {
			tone := k.Success
			switch {
			case r.OverBudget, r.Status == "failed":
				tone = k.Warning
			case r.Status == "skipped":
				tone = k.TextMuted
			}
			ui.Row(c).Gap(u * 2).Children(func() {
				ui.Text(c, fmt.Sprintf("%s → %s.%s", r.InputName, r.Tier, r.Format)).
					FontSize(core.FontSize(c, theme.CaptionSize)).Grow(1).MinWidth(0).
					SingleLine()
				ui.Text(c, sizeText(r.Size)).FontSize(core.FontSize(c, theme.CaptionSize)).FontWeight(600).TextColor(tone)
			})
		}
	})
}

// setupView is the FFmpeg bootstrap screen.
func (a *App) setupView(c *ui.Context) {
	u := core.Density(c).Unit()
	k := core.Tokens(c)
	ui.Column(c).Fill().Center().Gap(u * 3).Children(func() {
		data.Card(c, data.CardOptions{Title: a.t("setup.title")}, func() {
			ui.Column(c).Gap(u * 3).Children(func() {
				ui.Row(c).Gap(u * 2).AlignItems(ui.Center).Children(func() {
					ui.Icon(c, ic("triangle-alert")).Size(20, 20).TextColor(k.Warning)
					ui.Text(c, a.t("setup.desc")).Grow(1).Shrink(1).MinWidth(0)
				})

				if a.downloading {
					ui.Column(c).Gap(u).Children(func() {
						ui.Text(c, a.tf("setup.downloading", map[string]string{"percent": fmt.Sprintf("%.1f", a.downloadPercent)}))
						feedback.Progress(c, clamp01(a.downloadPercent/100), feedback.ProgressOptions{
							Shape: feedback.ProgressLinear, ShowValue: true, Label: a.t("setup.title"),
						})
						ui.Text(c, a.t("setup.download_tip")).FontSize(core.FontSize(c, theme.CaptionSize)).TextColor(k.TextMuted)
					})
				} else {
					ui.Row(c).Gap(u * 2).Children(func() {
						if input.Button(c, a.t("setup.btn_auto"), input.ButtonOptions{Variant: input.Primary, Icon: ic("download")}).Clicked() {
							a.downloadFFmpeg()
						}
						if input.Button(c, a.t("setup.btn_manual"), input.ButtonOptions{Variant: input.Secondary}).Clicked() {
							a.page = 1
							a.settingsPage = "integration"
						}
					})
				}

				if a.downloadError != "" {
					ui.Text(c, a.t("setup.failed")).FontSize(core.FontSize(c, theme.CaptionSize)).TextColor(k.Danger)
				}
			})
		})
	})
}

// ---- helpers shared by the panels ----

func clamp01(v float64) float64 { return min(1, max(0, v)) }

// defaultCRF is the recommended quality for a codec.
func defaultCRF(codec string) int {
	switch codec {
	case "h265":
		return 29
	case "av1":
		return 32
	}
	return 27
}

// crfRange is the recommended default and bounds for a codec.
func crfRange(codec string) (def, lo, hi int) {
	switch codec {
	case "h265":
		return 29, 18, 38
	case "av1":
		return 32, 20, 45
	}
	return 27, 15, 35
}

// crfWord describes where a CRF sits in its range.
func crfWord(v int, codec string, a *App) string {
	switch codec {
	case "h265":
		switch {
		case v <= 23:
			return a.t("advanced.crf.status.high")
		case v <= 32:
			return a.t("advanced.crf.status.balanced")
		}
	case "av1":
		switch {
		case v <= 26:
			return a.t("advanced.crf.status.high")
		case v <= 38:
			return a.t("advanced.crf.status.balanced")
		}
	default:
		switch {
		case v <= 20:
			return a.t("advanced.crf.status.high")
		case v <= 29:
			return a.t("advanced.crf.status.balanced")
		}
	}
	return a.t("advanced.crf.status.low")
}

func crfDesc(codec string, a *App) string {
	switch codec {
	case "h265":
		return a.t("advanced.crf.desc.h265")
	case "av1":
		return a.t("advanced.crf.desc.av1")
	}
	return a.t("advanced.crf.desc.h264")
}

// queueStatus picks the word, the tone and the glyph of a queue row.
func queueStatus(c *ui.Context, item *QueueItem, a *App) (string, ui.Color, *ui.SVG) {
	k := core.Tokens(c)
	switch item.Status {
	case "compressing":
		return fmt.Sprintf("%s %.0f%%", a.t("queue.status_compressing"), item.Percent), k.AccentText, ic("loader-circle")
	case "success":
		return a.t("queue.status_succeeded"), k.Success, ic("circle-check")
	case "failed":
		return a.t("queue.status_failed"), k.Danger, ic("circle-x")
	case "skipped":
		return a.t("queue.status_skipped"), k.TextMuted, ic("minus")
	}
	if item.Warning {
		return a.t("queue.status_waiting"), k.Warning, ic("triangle-alert")
	}
	return a.t("queue.status_waiting"), k.TextMuted, ic("clock")
}

func durationText(ms int64) string {
	if ms < 1000 {
		return fmt.Sprintf("%dms", ms)
	}
	return strconv.FormatFloat(float64(ms)/1000, 'f', 1, 64) + "s"
}

// outputDirFor resolves where a finished file went.
func outputDirFor(input, override string) string {
	if override != "" {
		return override
	}
	return dirOf(input) + string(sepOf(input)) + "compressed"
}

func dirOf(path string) string {
	for i := len(path) - 1; i >= 0; i-- {
		if path[i] == '\\' || path[i] == '/' {
			return path[:i]
		}
	}
	return ""
}

func sepOf(path string) rune {
	for i := range path {
		if path[i] == '/' {
			return '/'
		}
		if path[i] == '\\' {
			return '\\'
		}
	}
	return '\\'
}

// gifTierNames lists the animated export tiers in display order.
func gifTierNames() []string { return []string{"smooth", "balanced", "hd"} }

// gifTierSpec describes a tier in the current language; its numbers come
// from the engine so the copy cannot drift from the export.
func (a *App) gifTierSpec(name string) (string, bool) {
	t, err := gif.TierByName(name)
	if err != nil {
		return "", false
	}
	return fmt.Sprintf(a.t("gif.spec"), t.MaxSizeMB, t.MaxWidth, t.FPS, a.t("gif.tier_desc_"+t.Name)), true
}
