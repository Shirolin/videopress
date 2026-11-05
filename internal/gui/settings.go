package gui

import (
	"runtime"
	"strconv"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"

	"github.com/ELY-Open-Source/MujicaUI/account"
	"github.com/ELY-Open-Source/MujicaUI/core"
	"github.com/ELY-Open-Source/MujicaUI/input"
	"github.com/ELY-Open-Source/MujicaUI/navigation"
	"github.com/ELY-Open-Source/MujicaUI/theme"

	"videopress/internal/sendto"
)

// settingsView is the settings page.
func (a *App) settingsView(c *ui.Context) {
	pages := []account.SettingsPage{
		{ID: "general", Label: a.t("settings.title.general"), Icon: ic("sliders-horizontal")},
		{ID: "integration", Label: a.t("settings.title.sys"), Icon: ic("layout-grid")},
		{ID: "language", Label: a.t("settings.title.lang"), Icon: ic("globe")},
		{ID: "debug", Label: a.t("settings.title.debug"), Icon: ic("file-text")},
	}
	// The FFmpeg screen leaves its setup button pointing here; show the page it means.
	if a.settingsPage == "setup" {
		a.settingsPage = "integration"
	}
	page := a.settingsPage
	res := account.SettingsLayout(c, &page, pages, account.SettingsLayoutOptions{Title: a.t("nav.settings")}, func(id string) {
		switch id {
		case "general":
			a.generalSettings(c)
		case "integration":
			a.integrationSettings(c)
		case "language":
			a.languageSettings(c)
		case "debug":
			a.debugSettings(c)
		}
	})
	if res.Changed() {
		a.settingsPage = page
		switch page {
		case "integration":
			a.integrationBusy = true
			a.refreshIntegration()
		}
	}
	if len(a.integration) == 0 && page == "integration" && !a.integrationBusy {
		a.integrationBusy = true
		a.refreshIntegration()
	}
}

// generalSettings holds the preset, performance and behaviour rows.
func (a *App) generalSettings(c *ui.Context) {
	presets, presetLabels := a.presetOptions()
	current := len(presets) - 1
	for i, p := range presets {
		if p == a.prefs.Preset {
			current = i
		}
	}
	account.SettingsSection(c, a.t("settings.section.preset"), func() {
		account.SettingsRow(c, a.t("settings.section.preset"),
			account.SettingsRowOptions{Description: a.t("settings.preset_desc")},
			func() ui.Element {
				sel := current
				navigationSegments(c, &sel, presetLabels, a.t("controls.preset"))
				if sel != current && !a.isCompressing {
					a.prefs.Preset = presets[sel]
					a.savePrefs()
				}
				return ui.Row(c)
			})
	})

	cores := runtime.NumCPU()
	account.SettingsSection(c, a.t("settings.section.perf"), func() {
		account.SettingsRow(c, a.t("settings.concurrency"),
			account.SettingsRowOptions{Description: a.t("settings.concurrency_desc")},
			func() ui.Element {
				values := concurrencyOptions(cores)
				labels := make([]string, len(values))
				for i, n := range values {
					labels[i] = strconv.Itoa(n)
					if n == cores {
						labels[i] = a.tf("settings.concurrency_all", map[string]string{"num": labels[i]})
					}
				}
				current := 0
				for i, n := range values {
					if n == a.prefs.Concurrency {
						current = i
					}
				}
				sel := current
				navigationSegments(c, &sel, labels, a.t("settings.concurrency"))
				if sel != current && !a.isCompressing {
					a.prefs.Concurrency = values[sel]
					a.savePrefs()
				}
				return ui.Row(c)
			})

		account.SettingsRow(c, a.t("settings.hw_accel"), account.SettingsRowOptions{Description: a.hwAccelNote()},
			func() ui.Element {
				on := a.prefs.HWAccel
				sw := settingsSwitch(c, &on, a.isCompressing)
				if sw.Changed() {
					a.prefs.HWAccel = on
					a.prefs.HWAccelSet = true
					a.savePrefs()
				}
				return sw
			})

		account.SettingsRow(c, a.t("settings.copy_audio"), account.SettingsRowOptions{Description: a.t("settings.copy_audio_desc")},
			func() ui.Element {
				on := a.prefs.CopyAudio
				sw := settingsSwitch(c, &on, a.isCompressing)
				if sw.Changed() {
					a.prefs.CopyAudio = on
					a.savePrefs()
				}
				return sw
			})

		account.SettingsRow(c, a.t("settings.force_mode"), account.SettingsRowOptions{Description: a.t("settings.force_mode_desc")},
			func() ui.Element {
				on := a.prefs.ForceMode
				sw := settingsSwitch(c, &on, a.isCompressing)
				if sw.Changed() {
					a.prefs.ForceMode = on
					a.savePrefs()
				}
				return sw
			})

		account.SettingsRow(c, a.t("settings.skip_existing"), account.SettingsRowOptions{Description: a.t("settings.skip_existing_desc")},
			func() ui.Element {
				on := a.prefs.SkipExisting
				sw := settingsSwitch(c, &on, a.isCompressing)
				if sw.Changed() {
					a.prefs.SkipExisting = on
					a.savePrefs()
				}
				return sw
			})
	})
}

// hwAccelNote explains the encoder state.
func (a *App) hwAccelNote() string {
	if a.gpuEncoder == "" {
		return a.t("settings.hw_accel_probing")
	}
	if a.gpuEncoder == "libx264" {
		return a.t("settings.hw_accel_off")
	}
	return a.tf("settings.hw_accel_active", map[string]string{"gpu": a.gpuEncoder})
}

// integrationSettings toggles the five Windows integrations.
func (a *App) integrationSettings(c *ui.Context) {
	kinds := []string{"contextmenu", "sendto", "desktop", "startmenu", "path"}
	account.SettingsSection(c, a.t("settings.title.sys"), func() {
		for _, kind := range kinds {
			kind := kind
			account.SettingsRow(c, a.t("integration."+kind),
				account.SettingsRowOptions{Description: a.t("integration." + kind + "_desc")},
				func() ui.Element {
					installed, known := a.integration[kind]
					state := a.t("status.checking")
					if known {
						if installed {
							state = a.t("status.enabled")
						} else {
							state = a.t("status.disabled")
						}
					}
					label := a.t("btn.enable")
					variant := input.Primary
					if installed {
						label = a.t("btn.disable")
						variant = input.Secondary
					}
					if a.integrationBusy {
						state = a.t("status.checking")
					}
					return ui.Row(c).Gap(core.Density(c).Unit() * 2).AlignItems(ui.Center).Children(func() {
						ui.Text(c, state).FontSize(core.FontSize(c, theme.CaptionSize)).TextColor(core.Tokens(c).TextMuted)
						if input.Button(c, label, input.ButtonOptions{
							Variant: variant, Disabled: a.integrationBusy, Loading: a.integrationBusy,
						}).Clicked() && !a.integrationBusy {
							a.toggleIntegration(kind)
						}
					})
				})
		}
	})
}

// languageSettings switches the interface language.
func (a *App) languageSettings(c *ui.Context) {
	account.SettingsSection(c, a.t("settings.title.lang"), func() {
		account.SettingsRow(c, a.t("settings.lang.display"), account.SettingsRowOptions{Description: a.t("settings.lang.desc")},
			func() ui.Element {
				zh := a.locale == "zh-Hans"
				sel := 0
				if !zh {
					sel = 1
				}
				before := sel
				navigationSegments(c, &sel, []string{"简体中文", "English"}, a.t("settings.title.lang"))
				if sel != before {
					tag := "zh-Hans"
					if sel == 1 {
						tag = "en"
					}
					mygo.App.SetLocale(tag)
					a.locale = tag
					a.prefs.Language = tag
					a.savePrefs()
					if a.integration["contextmenu"] {
						go func() { _ = sendto.RegisterContextMenu(a.executablePath, sendtoLang(tag)) }()
					}
				}
				return ui.Row(c)
			})
	})
}

// debugSettings toggles the diagnostics log.
func (a *App) debugSettings(c *ui.Context) {
	account.SettingsSection(c, a.t("settings.title.debug"), func() {
		account.SettingsRow(c, a.t("settings.debug.toggle"), account.SettingsRowOptions{Description: a.t("settings.debug.desc")},
			func() ui.Element {
				on := a.prefs.EnableDebugLog
				sw := settingsSwitch(c, &on, false)
				if sw.Changed() {
					a.setDebug(on)
				}
				return sw
			})

		account.SettingsRow(c, a.t("settings.debug.manage"), account.SettingsRowOptions{Description: a.t("settings.debug.file_desc")},
			func() ui.Element {
				// Stacked: side by side the pair leaves the description too
				// little width when the window is narrow.
				return ui.Column(c).Gap(core.Density(c).Unit()).Children(func() {
					if input.Button(c, a.t("btn.open_log"), input.ButtonOptions{Variant: input.Secondary, Icon: ic("file-text")}).Clicked() {
						a.openDebugLog()
					}
					if input.Button(c, a.t("btn.clear_log"), input.ButtonOptions{Variant: input.Ghost, Icon: ic("trash-2")}).Clicked() {
						a.clearDebugLog()
					}
				})
			})
	})

	account.SettingsSection(c, "Videopress", func() {
		account.SettingsRow(c, "FFmpeg", account.SettingsRowOptions{Description: a.ffmpegPath}, func() ui.Element {
			state := a.t("status.not_installed")
			if a.ffmpegPath != "" {
				state = a.t("status.up_to_date")
			}
			return ui.Text(c, state).TextColor(core.Tokens(c).TextMuted).FontSize(core.FontSize(c, theme.CaptionSize))
		})
		account.SettingsRow(c, a.t("settings.version"), account.SettingsRowOptions{Description: "MyGO + MujicaUI"}, func() ui.Element {
			v := a.Version()
			if v == "" {
				v = "dev"
			}
			return ui.Text(c, v).FontSize(core.FontSize(c, theme.CaptionSize))
		})
	})
}

// navigationSegments is a SegmentedControl over labels; a plain row is returned
// so it can sit inside a settings row's control slot.
func navigationSegments(c *ui.Context, selected *int, labels []string, label string) {
	segs := make([]navigation.Segment, len(labels))
	for i, l := range labels {
		segs[i] = navigation.Segment{Label: l}
	}
	navigation.SegmentedControl(c, selected, segs, navigation.SegmentedControlOptions{Label: label})
}

const (
	switchW float32 = 36
	switchH float32 = 20
	switchK float32 = 14
)

// settingsSwitch renders MujicaUI's switch with its thumb centered inside the
// track's padding box (accounting for theme.BorderWidth on each edge).
func settingsSwitch(c *ui.Context, on *bool, disabled bool) ui.Element {
	k := core.Tokens(c)
	d := core.Density(c)
	var sw ui.Element
	ui.Row(c).Shrink(0).Disabled(disabled).Children(func() {
		sw = ui.SwitchBase(c, on).Gap(d.Unit() * 2).AlignItems(ui.Center).
			MinHeight(d.ControlHeight()).FocusRing(false)
	})
	hover := sw.Hovered()
	target := float32(0)
	if *on {
		target = 1
	}
	pos := sw.Animate("on", target, core.Motion(c, theme.StateDuration))
	sw.Children(func() {
		track := ui.Box(c).Size(switchW, switchH).Shrink(0).Radius(theme.CardRadius)
		bg, border := k.Surface, k.ControlBorder
		thumb := k.ControlBorder
		if *on {
			bg, border, thumb = k.Accent, k.Accent, k.OnAccent
			if hover {
				bg, border = k.AccentHover, k.AccentHover
			}
		} else if hover {
			bg, border, thumb = k.SurfaceHover, k.Text, k.Text
		}
		track.Background(bg).Border(theme.BorderWidth, border)
		track.Draw(func(p *ui.Painter, r ui.Rect) {
			if *on {
				p.Fill(ui.Rect{X: r.X + 9, Y: r.Y + 6, W: 1.5, H: r.H - 12}, k.OnAccent, 0)
			}
		})
		track.DrawOver(func(p *ui.Painter, r ui.Rect) {
			if sw.FocusVisible() {
				o := theme.FocusGap + theme.FocusWidth
				p.Stroke(ui.Rect{X: r.X - o, Y: r.Y - o, W: r.W + 2*o, H: r.H + 2*o}, k.Focus, theme.CardRadius+o, theme.FocusWidth)
			}
		})
		innerW := switchW - 2*theme.BorderWidth
		innerH := switchH - 2*theme.BorderWidth
		inset := (innerH - switchK) / 2
		x := inset + pos*(innerW-switchK-2*inset)
		track.Children(func() {
			ui.Box(c).Absolute().Left(x).Top(inset).Size(switchK, switchK).Radius(theme.ControlRadius).Background(thumb)
		})
	})
	return sw
}

// sendtoLang maps a locale tag to the language the context menu is written in.
func sendtoLang(tag string) string {
	if tag == "zh-Hans" {
		return "zh"
	}
	return "en"
}
