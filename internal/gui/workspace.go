package gui

import (
	"fmt"
	"path/filepath"
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
				a.dropZone(c)
				a.queueList(c)
				a.advancedPanel(c)
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

	var running int64
	for _, it := range a.queue {
		if it.Status == "compressing" {
			running++
		}
	}
	// stats() counts skipped rows as done; the bar shows the actual successes.
	skipped := int64(a.skippedCount())
	succeeded := done - skipped

	data.Card(c, data.CardOptions{}, func() {
		ui.Column(c).Gap(u * 3).Children(func() {
			ui.Row(c).Gap(u * 6).Children(func() {
				statCell(c, strconv.FormatInt(total, 10), a.t("stats.total"), k.Text)
				statCell(c, strconv.FormatInt(running, 10), a.t("stats.processing"), k.AccentText)
				statCell(c, strconv.FormatInt(succeeded, 10), a.t("stats.completed"), k.Success)
				if skipped > 0 {
					statCell(c, strconv.FormatInt(skipped, 10), a.t("stats.skipped"), k.TextMuted)
				}
				statCell(c, sizeText(saved), a.t("stats.saved_space"), k.Success)
				statCell(c, ratioText(saved, original), a.t("stats.savings_rate"), k.TextMuted)
			})
			// One bar for the batch, since each row already shows its own.
			if a.isCompressing {
				feedback.Progress(c, a.batchProgress(), feedback.ProgressOptions{
					Shape: feedback.ProgressLinear, ShowValue: true, Label: a.t("stats.overall"),
				})
			}
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
				values, labels := a.presetOptions()
				selected := 1
				for i, v := range values {
					if v == a.prefs.Preset {
						selected = i
					}
				}
				segments := make([]navigation.Segment, len(labels))
				for i, l := range labels {
					segments[i] = navigation.Segment{Label: l}
				}
				before := selected
				navigation.SegmentedControl(c, &selected, segments, navigation.SegmentedControlOptions{Label: a.t("controls.preset"), Disabled: a.isCompressing})
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
	open := a.showAdvanced
	opts := data.CollapsibleOptions{Title: a.t("advanced.title")}
	res := data.Collapsible(c, &open, opts, func() {
		a.advancedFields(c)
	})
	if res.Changed() {
		a.showAdvanced = open
		a.prefs.ShowAdvanced = open
		a.savePrefs()
	}
}

// advancedFields is the shared body of the workspace panel and the settings
// page, so both entry points edit the same preferences.
func (a *App) advancedFields(c *ui.Context) {
	u := core.Density(c).Unit()
	ui.Column(c).Gap(u*3).Padding(u*3, 0).Children(func() {
		ui.Row(c).Gap(u * 4).Wrap().Children(func() {
			a.codecField(c)
			a.fpsField(c)
			a.audioField(c)
		})
		a.crfField(c)
		a.cropField(c)
	})
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
		ui.Column(c).Gap(u).Children(func() {
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
		Radius(theme.CardRadius).Border(theme.BorderWidth, k.ControlBorder).Background(k.Surface)
	if compact {
		e.PaddingY(u * 2.5)
	}
	// Drops anywhere else in the page land on the body, which reports its own
	// hover; show the same target hint here so the whole window reads as one.
	dragging := e.FileDragOver() || a.dropHover
	if dragging {
		e.Background(k.Selection).Border(theme.BorderWidth*2, k.AccentText)
	}
	var label, sub string
	if dragging {
		label, sub = a.t("dropzone.drag_hint"), a.t("dropzone.formats")
	} else if compact || len(a.queue) > 0 {
		label, sub = a.t("dropzone.add_more"), a.t("dropzone.formats")
	} else {
		label, sub = a.t("dropzone.title"), a.t("dropzone.desc")
	}
	if a.isCompressing {
		sub = a.t("dropzone.compressing")
	}

	e.Children(func() {
		if !compact {
			ui.Icon(c, ic("file")).Size(28, 28).TextColor(k.TextMuted)
		}
		ui.Text(c, label).FontSize(core.FontSize(c, theme.BodySize+1)).FontWeight(600)
		ui.Text(c, sub).FontSize(core.FontSize(c, theme.CaptionSize)).TextColor(k.TextMuted)
		if input.Button(c, a.t("dropzone.browse"), input.ButtonOptions{
			Variant: input.Outline, Icon: ic("folder"),
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
			// The filter only earns its space once the list is long enough to
			// hide a failure, but stays while a filter is active so it can
			// always be switched back.
			if len(a.queue) > 3 || (a.queueFilter != "" && a.queueFilter != "all") {
				filters := []string{"all", "failed", "done"}
				labels := []string{a.t("queue.filter_all"), a.t("queue.filter_failed"), a.t("queue.filter_done")}
				selected := 0
				for i, f := range filters {
					if f == a.queueFilter {
						selected = i
					}
				}
				before := selected
				navigationSegments(c, &selected, labels, a.t("queue.title"))
				if selected != before {
					a.queueFilter = filters[selected]
				}
			}
			if input.Button(c, a.t("queue.clear"), input.ButtonOptions{
				Variant: input.Ghost, Disabled: a.isCompressing || a.isGifExporting,
			}).Clicked() {
				a.clearQueue()
			}
		})

		ui.Column(c).Gap(u * 2).Children(func() {
			visible := 0
			for i := range a.queue {
				if !a.queueVisible(&a.queue[i]) {
					continue
				}
				visible++
				a.queueRow(c, i)
			}
			if visible == 0 {
				ui.Text(c, a.t("queue.filter_empty")).FontSize(core.FontSize(c, theme.CaptionSize)).TextColor(k.TextMuted)
			}
		})
	})
}

// queueRow draws one queued file.
func (a *App) queueRow(c *ui.Context, index int) {
	// 行内的移除按钮会在本次渲染中缩短队列。
	if index >= len(a.queue) {
		return
	}
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
			switch {
			case a.isCompressing && (item.Status == "compressing" || item.Status == "waiting"):
				// One row can be stopped without ending the whole batch.
				if input.IconButton(c, ic("x"), a.t("queue.cancel_one"), input.IconButtonOptions{Variant: input.Ghost}).Clicked() {
					a.cancelQueueItem(item.Path)
				}
			case item.Status != "compressing" && !a.isCompressing:
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

		// 只有进行中的行画进度条；终态已由状态文字和结果数据表达。
		if item.Status == "compressing" {
			feedback.Progress(c, clamp01(item.Percent/100), feedback.ProgressOptions{Shape: feedback.ProgressLinear})
			if item.Speed > 0 && item.ETA > 0 {
				ui.Text(c, a.tf("queue.progress_detail", map[string]string{
					"speed": strconv.FormatFloat(item.Speed, 'f', 1, 64) + "x",
					"eta":   clockFromMs(item.ETA),
				})).FontSize(core.FontSize(c, theme.CaptionSize)).TextColor(k.TextMuted)
			}
		}

		if msg, tone := queueNote(item, a, k); msg != "" {
			ui.Row(c).Gap(u).AlignItems(ui.Center).Children(func() {
				ui.Text(c, msg).FontSize(core.FontSize(c, theme.CaptionSize)).TextColor(tone)
				if item.Status == "failed" {
					if input.IconButton(c, ic("copy"), a.t("queue.copy_error"), input.IconButtonOptions{Variant: input.Ghost}).Clicked() {
						a.copyError(item.Error)
					}
				}
			})
		}

		hasResult := item.Status == "success" || item.Status == "skipped" || item.Status == "failed"
		if hasResult && (item.Size > 0 || item.TargetSize > 0) {
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

// actionPanel holds the batch summary and the start/cancel buttons.
func (a *App) actionPanel(c *ui.Context) {
	u := core.Density(c).Unit()
	k := core.Tokens(c)
	ui.Column(c).Gap(u * 2).Children(func() {
		layout.Divider(c, layout.DividerOptions{})
		ui.Row(c).Gap(u * 3).AlignItems(ui.Center).Children(func() {
			if len(a.queue) > 0 {
				ui.Text(c, a.tf("action.summary", map[string]string{
					"num":    strconv.Itoa(len(a.queue)),
					"preset": a.presetLabel(),
				})).FontSize(core.FontSize(c, theme.CaptionSize)).TextColor(k.TextMuted)
				if failed := a.failedCount(); failed > 0 {
					ui.Text(c, a.tf("action.failed", map[string]string{
						"num": strconv.Itoa(failed),
					})).FontSize(core.FontSize(c, theme.CaptionSize)).FontWeight(600).TextColor(k.Danger)
				}
			} else {
				ui.Text(c, a.t("action.empty")).FontSize(core.FontSize(c, theme.CaptionSize)).TextColor(k.TextMuted)
			}
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
			label := a.t("btn.start_compress")
			pending, waiting := 0, 0
			for _, it := range a.queue {
				if !isPendingStatus(it.Status) {
					continue
				}
				pending++
				if it.Status == "waiting" {
					waiting++
				}
			}
			switch {
			case pending == 0 && len(a.queue) > 0:
				label = a.t("btn.recompress_all")
			case pending > 0 && waiting == 0:
				label = a.t("btn.retry_failed")
			}
			if input.Button(c, label, input.ButtonOptions{
				Variant: input.Primary, Icon: ic("play"), Disabled: disabled,
			}).Clicked() && !disabled {
				a.startCompress()
			}
		})
	})
}

// presetLabel is the display name of the current preset.
func (a *App) presetLabel() string {
	ids, labels := a.presetOptions()
	for i, id := range ids {
		if id == a.prefs.Preset {
			return labels[i]
		}
	}
	return a.t("preset.standard")
}

// gifPanel exports the queue as animated clips.
func (a *App) gifPanel(c *ui.Context) {
	u := core.Density(c).Unit()
	k := core.Tokens(c)

	data.Card(c, data.CardOptions{Title: a.t("gif.title")}, func() {
		ui.Column(c).Gap(u * 2).Children(func() {
			// 有文件时把用途说明换成实际的导出范围。
			subtitle := a.t("gif.subtitle")
			if len(a.queue) > 0 {
				if formats := a.gifFormatCount(); formats > 0 {
					subtitle = a.tf("gif.pending", map[string]string{
						"n": strconv.Itoa(len(a.queue)),
						"m": strconv.Itoa(formats),
					})
				}
			}
			ui.Text(c, subtitle).FontSize(core.FontSize(c, theme.CaptionSize)).TextColor(k.TextMuted)

			ui.Row(c).Gap(u * 4).AlignItems(ui.End).Children(func() {
				ui.Column(c).Gap(u).Children(func() {
					a.sectionLabel(c, "gif.tier")
					tiers := gifTierNames()
					selected := 1
					for i, t := range tiers {
						if t == a.prefs.GifTier {
							selected = i
						}
					}
					before := selected
					navigation.SegmentedControl(c, &selected, []navigation.Segment{
						{Label: a.t("gif.tier_smooth")},
						{Label: a.t("gif.tier_balanced")},
						{Label: a.t("gif.tier_hd")},
						{Label: a.t("gif.tier_custom")},
					}, navigation.SegmentedControlOptions{Label: a.t("gif.tier"), Disabled: a.gifBusy || a.isCompressing})
					if selected != before && !a.gifBusy {
						a.prefs.GifTier = tiers[selected]
						a.savePrefs()
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

			if a.prefs.GifTier == "custom" {
				a.ensureGifDurations()
				a.gifCustomFields(c)
				ui.Text(c, a.gifCustomSummary()).FontSize(core.FontSize(c, theme.CaptionSize)).TextColor(k.TextMuted)
				if note, warn := a.gifDurationNote(); note != "" {
					color := k.TextMuted
					if warn {
						color = k.Warning
					}
					ui.Text(c, note).FontSize(core.FontSize(c, theme.CaptionSize)).TextColor(color)
				}
			} else if spec, ok := a.gifTierSpec(a.prefs.GifTier); ok {
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
	hasSuccess := ok > 0 || skipped > 0 || over > 0
	ui.Column(c).Gap(u).Children(func() {
		ui.Row(c).AlignItems(ui.Center).Children(func() {
			ui.Text(c, fmt.Sprintf("%s %d · %s %d · %s %d · %s %d",
				a.t("gif.ok"), ok, a.t("gif.skipped"), skipped, a.t("gif.over_budget"), over, a.t("gif.failed"), failed)).
				FontSize(core.FontSize(c, theme.CaptionSize)).FontWeight(600)
			ui.Spacer(c)
			if hasSuccess && input.Button(c, a.t("btn.open_output_folder"), input.ButtonOptions{
				Variant: input.Secondary, Icon: ic("folder"),
			}).Clicked() {
				a.openGifResultFolder()
			}
		})
		for _, r := range a.gifResults {
			tone := k.Success
			switch {
			case r.OverBudget, r.Status == "failed":
				tone = k.Warning
			case r.Status == "skipped":
				tone = k.TextMuted
			}
			ui.Column(c).Gap(u).Children(func() {
				ui.Row(c).Gap(u * 2).AlignItems(ui.Center).Children(func() {
					displayPath := fmt.Sprintf("%s → %s.%s", r.InputName, r.Tier, r.Format)
					if r.OutputPath != "" {
						dirName := filepath.Base(filepath.Dir(r.OutputPath))
						displayPath = fmt.Sprintf("%s → %s/%s", r.InputName, dirName, filepath.Base(r.OutputPath))
					}
					ui.Text(c, displayPath).
						FontSize(core.FontSize(c, theme.CaptionSize)).Grow(1).MinWidth(0).
						SingleLine()
					switch {
					case r.Status == "failed":
						ui.Text(c, a.t("gif.failed")).FontSize(core.FontSize(c, theme.CaptionSize)).FontWeight(600).TextColor(k.Danger)
					case r.OverBudget:
						ui.Text(c, a.t("gif.over_budget")).FontSize(core.FontSize(c, theme.CaptionSize)).FontWeight(600).TextColor(k.Warning)
					}
					if r.Size > 0 {
						ui.Text(c, sizeText(r.Size)).FontSize(core.FontSize(c, theme.CaptionSize)).FontWeight(600).TextColor(tone)
					}
				})
				if r.Error != "" {
					ui.Text(c, r.Error).FontSize(core.FontSize(c, theme.CaptionSize)).TextColor(k.Danger)
				}
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
					ui.Column(c).Gap(u * 2).Children(func() {
						ui.Row(c).Gap(u * 2).Children(func() {
							if input.Button(c, a.t("setup.btn_auto"), input.ButtonOptions{Variant: input.Primary, Icon: ic("download")}).Clicked() {
								a.downloadFFmpeg()
							}
							if input.Button(c, a.t("setup.btn_pick"), input.ButtonOptions{Variant: input.Secondary, Icon: ic("folder")}).Clicked() {
								a.pickFFmpeg()
							}
						})
						ui.Row(c).Gap(u * 2).Children(func() {
							if input.Button(c, a.t("setup.btn_recheck"), input.ButtonOptions{Variant: input.Ghost, Icon: ic("refresh-cw")}).Clicked() {
								a.detectFFmpeg()
								a.detectGPU()
							}
							if input.Button(c, a.t("setup.btn_manual"), input.ButtonOptions{Variant: input.Ghost}).Clicked() {
								a.page = 1
								a.settingsPage = "integration"
							}
						})
					})
					if a.prefs.FFmpegPath != "" {
						ui.Row(c).Gap(u).AlignItems(ui.Center).Children(func() {
							ui.Text(c, a.tf("setup.custom_path", map[string]string{"path": a.prefs.FFmpegPath})).
								FontSize(core.FontSize(c, theme.CaptionSize)).TextColor(k.TextMuted)
							if input.Button(c, a.t("setup.btn_clear_path"), input.ButtonOptions{Variant: input.Ghost, Icon: ic("refresh-cw")}).Clicked() {
								a.prefs.FFmpegPath = ""
								a.savePrefs()
								a.detectFFmpeg()
								a.detectGPU()
							}
						})
					}
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

// queueNote is the note under a queue row and its tone. A waiting row that
// already looks compressed carries the warning instead of an engine error;
// its text comes from the copy table so it follows the current language.
func queueNote(item *QueueItem, a *App, k theme.Tokens) (string, ui.Color) {
	if item.Status == "waiting" {
		if item.Warning {
			return a.t("queue.warn_recompressed"), k.Warning
		}
		return "", k.TextMuted
	}
	if item.Status == "skipped" {
		return a.t("queue.skip_tip"), k.TextMuted
	}
	return item.Error, k.Danger
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

// gifFormatCount is how many output formats are selected.
func (a *App) gifFormatCount() int {
	n := 0
	for _, f := range gif.AllFormats() {
		if a.gifFormats[string(f)] {
			n++
		}
	}
	return n
}

// gifTierNames lists the animated export tiers in display order.
func gifTierNames() []string { return []string{"smooth", "balanced", "hd", "custom"} }

// gifTierSpec describes a tier in the current language; its numbers come
// from the engine so the copy cannot drift from the export.
func (a *App) gifTierSpec(name string) (string, bool) {
	t, err := gif.TierByName(name)
	if err != nil {
		return "", false
	}
	return a.tf("gif.spec", map[string]string{
		"mb":      strconv.Itoa(t.MaxSizeMB),
		"width":   strconv.Itoa(t.MaxWidth),
		"fps":     strconv.Itoa(t.FPS),
		"seconds": strconv.Itoa(t.DurationSec()),
		"desc":    a.t("gif.tier_desc_" + t.Name),
	}), true
}

// gifCustomFields edits the custom tier parameters.
func (a *App) gifCustomFields(c *ui.Context) {
	u := core.Density(c).Unit()
	disabled := a.gifBusy || a.isCompressing

	field := func(key string, target *int, lo, hi int) {
		ui.Column(c).Grow(1).Gap(u).Children(func() {
			a.sectionLabel(c, key)
			raw := float64(*target)
			if input.NumberInput(c, &raw, input.NumberInputOptions{
				Min: float64(lo), Max: float64(hi), Step: 1,
				Label: a.t(key), Disabled: disabled,
			}).Changed() {
				*target = int(raw)
				a.savePrefs()
			}
		})
	}

	ui.Column(c).Gap(u*2).Padding(u*2, 0).Children(func() {
		ui.Row(c).Gap(u * 4).Children(func() {
			field("gif.start", &a.prefs.GifStart, 0, 600)
			field("gif.duration", &a.prefs.GifDuration, 1, 30)
			field("gif.width", &a.prefs.GifWidth, 120, 1080)
		})
		ui.Row(c).Gap(u * 4).AlignItems(ui.End).Children(func() {
			ui.Column(c).Grow(1).Gap(u).Children(func() {
				a.sectionLabel(c, "gif.fps")
				value := a.prefs.GifFPS
				options := []input.SelectOption[int]{
					{Value: 5, Label: "5 fps"}, {Value: 8, Label: "8 fps"},
					{Value: 10, Label: "10 fps"}, {Value: 12, Label: "12 fps"},
					{Value: 15, Label: "15 fps"}, {Value: 20, Label: "20 fps"},
					{Value: 24, Label: "24 fps"}, {Value: 30, Label: "30 fps"},
				}
				if input.Select(c, &value, options, input.SelectOptions{Label: a.t("gif.fps"), Disabled: disabled}).Changed() {
					a.prefs.GifFPS = value
					a.savePrefs()
				}
			})
			field("gif.budget", &a.prefs.GifBudgetMB, 1, 50)
			ui.Column(c).Grow(1).Gap(u).Children(func() {
				a.sectionLabel(c, "gif.quality")
				raw := float64(a.prefs.GifQuality)
				if input.NumberInput(c, &raw, input.NumberInputOptions{
					Min: 30, Max: 90, Step: 5,
					Label: a.t("gif.quality"), Disabled: disabled || !a.gifFormats["webp"],
				}).Changed() {
					a.prefs.GifQuality = int(raw)
					a.savePrefs()
				}
			})
		})
	})
}

// gifCustomSummary states what the custom parameters produce.
func (a *App) gifCustomSummary() string {
	return a.tf("gif.custom_summary", map[string]string{
		"start":  strconv.Itoa(a.prefs.GifStart),
		"end":    strconv.Itoa(a.prefs.GifStart + a.prefs.GifDuration),
		"width":  strconv.Itoa(a.prefs.GifWidth),
		"fps":    strconv.Itoa(a.prefs.GifFPS),
		"budget": strconv.Itoa(a.prefs.GifBudgetMB),
	})
}

// gifDurationNote states the source length and flags a selection that runs
// past its end. It returns warn=true when the note needs the warning color.
func (a *App) gifDurationNote() (string, bool) {
	if a.ffmpegPath == "" || len(a.queue) == 0 {
		return "", false
	}
	st := a.gifDurationStats()
	if st.known == 0 {
		if st.pending > 0 {
			return a.t("gif.duration_probing"), false
		}
		return a.t("gif.duration_unknown"), false
	}
	ranged := st.maxMs > st.minMs
	var source string
	if ranged {
		source = a.tf("gif.source_duration_range", map[string]string{
			"min": clockFromMs(st.minMs),
			"max": clockFromMs(st.maxMs),
		})
	} else {
		source = a.tf("gif.source_duration", map[string]string{"duration": clockFromMs(st.minMs)})
	}
	if st.unknown > 0 {
		source += " · " + a.t("gif.duration_partial")
	}
	if st.pending > 0 {
		source += " · " + a.t("gif.duration_pending_more")
	}
	startMs := int64(a.prefs.GifStart) * 1000
	if startMs >= st.minMs {
		key := "gif.duration_start_beyond"
		if ranged {
			key = "gif.duration_start_beyond_range"
		}
		note := a.tf(key, map[string]string{
			"start":    strconv.Itoa(a.prefs.GifStart),
			"duration": clockFromMs(st.minMs),
		})
		return source + " · " + note, true
	}
	endMs := startMs + int64(a.prefs.GifDuration)*1000
	if endMs > st.minMs {
		key := "gif.duration_overflow"
		if ranged {
			key = "gif.duration_overflow_range"
		}
		note := a.tf(key, map[string]string{
			"end":    strconv.Itoa(a.prefs.GifStart + a.prefs.GifDuration),
			"length": clockFromMs(st.minMs),
		})
		return source + " · " + note, true
	}
	return source, false
}

// clockFromMs renders a duration as m:ss (h:mm:ss past an hour).
func clockFromMs(ms int64) string {
	secs := ms / 1000
	if secs < 0 {
		secs = 0
	}
	if secs >= 3600 {
		return fmt.Sprintf("%d:%02d:%02d", secs/3600, secs%3600/60, secs%60)
	}
	return fmt.Sprintf("%d:%02d", secs/60, secs%60)
}

// gifCustomTier converts the custom preferences into an engine tier.
func (a *App) gifCustomTier() *gif.Tier {
	t := &gif.Tier{
		Name:        a.gifCustomName(),
		MaxWidth:    a.prefs.GifWidth,
		MaxSizeMB:   a.prefs.GifBudgetMB,
		FPS:         a.prefs.GifFPS,
		MaxDuration: clockText(a.prefs.GifDuration),
		WebPQuality: a.prefs.GifQuality,
	}
	if a.prefs.GifStart > 0 {
		t.Start = clockText(a.prefs.GifStart)
	}
	return t
}

// gifCustomName is the output name segment for a custom export: the window
// range and the encode parameters, so repeated exports stay distinguishable.
func (a *App) gifCustomName() string {
	return fmt.Sprintf("custom-%dpx-%dfps-%d-%ds",
		a.prefs.GifWidth, a.prefs.GifFPS, a.prefs.GifStart, a.prefs.GifStart+a.prefs.GifDuration)
}

// clockText renders whole seconds in ffmpeg's HH:MM:SS form.
func clockText(seconds int) string {
	if seconds < 0 {
		seconds = 0
	}
	return fmt.Sprintf("%02d:%02d:%02d", seconds/3600, seconds%3600/60, seconds%60)
}
