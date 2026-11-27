package main

import (
	"fmt"
	"math"

	"github.com/ELY-Open-Source/MujicaUI/theme"
	"github.com/egoist/mygo/plugins/render"
	"github.com/egoist/mygo/ui"
)

// BuildPart2 mounts Acts 4–6 of the Videopress product promo.
func BuildPart2(c *ui.Context, f render.Frame, assets *Assets) {
	f.Sequence(c, Scene4Start, Scene4Duration, func(sf render.Frame) {
		scene4GifAndCrop(c, sf, assets)
	})
	f.Sequence(c, Scene5Start, Scene5Duration, func(sf render.Frame) {
		scene5CLIAndIntegration(c, sf, assets)
	})
	f.Sequence(c, Scene6Start, Scene6Duration, func(sf render.Frame) {
		scene6ThemeAndFinale(c, sf, assets)
	})
}

// =============================================================================
// Act 4: Animated Export (GIF/APNG/WebP) & Geometric Crop (Light Ivory Theme)
// Global frames 378..525, local frames 0..147, CueCropLock at local 78 (global 456).
// =============================================================================

func scene4GifAndCrop(c *ui.Context, f render.Frame, assets *Assets) {
	p := LightPalette()
	stageAlpha := SceneFade(f, CrossfadeFrames, CrossfadeFrames)

	const (
		cropLockLocal = float64(CueCropLock - Scene4Start) // 78
		viewportW     = float32(840)
		viewportH     = float32(912)
		shotRenderH   = float32(1167) // 840 * 1250 / 900 preserves gif-custom.png exact aspect
		maxScrollY    = shotRenderH - viewportH
	)

	// Smooth scroll inside a static Clip() viewport so the GPU never recomputes clip coverage.
	scrollProgress := RangeProgress(f, 16, 96, EaseInOutQuart)
	scrollY := maxScrollY * float32(scrollProgress)

	cropT := RangeProgress(f, 26, cropLockLocal, EaseInOutQuart)
	cropLocked := f.Exact() >= cropLockLocal

	eyeOp, eyeY := EnterOffset(f, 4, 18, 14)
	headOp, headY := EnterOffset(f, 10, 20, 18)
	ruleP := RangeProgress(f, 16, 38, EaseOutQuart)
	bodyOp, bodyY := EnterOffset(f, 18, 18, 12)
	tagsOp, tagsY := EnterOffset(f, 24, 18, 12)
	cropCardOp, cropCardY := EnterOffset(f, 16, 22, 16)
	leftFade := float32(RangeProgress(f, 2, 18, EaseOutCubic))

	SceneStage(c, p.Background, stageAlpha, func() {
		SubtleBackdropFrame(c, p)

		ui.Row(c).
			Fill().
			Padding(76, 88).
			Gap(48).
			AlignItems(ui.Center).
			Children(func() {
				// Left column: static Clip() viewport with smooth vertical pan over gif-custom.png
				ui.Box(c).
					Size(viewportW, viewportH).
					Radius(theme.PopupRadius).
					Background(p.Surface).
					Border(theme.BorderWidth, p.ControlBorder.Alpha(0.82)).
					Shadow(0, 20, 44, 0, ui.RGBA(33, 28, 29, 0.14)).
					Opacity(leftFade).
					Clip().
					Shrink(0).
					Children(func() {
						if assets.GifCustom != nil {
							ui.Image(c, assets.GifCustom).
								Size(viewportW, shotRenderH).
								Fit(ui.Cover).
								Translate(0, -scrollY)
						}

						// Subtle floating viewport caption & scroll rail indicator centered in the blank bottom bar area
						statusText := "主工作区 · 向下定位动图面板"
						if scrollProgress >= 0.85 {
							statusText = "实机界面 · 动图导出自定义面板"
						}
						ui.Row(c).
							Absolute().
							Bottom(14).
							Left(210).
							Padding(7, 14).
							Gap(10).
							Radius(theme.ControlRadius + 1).
							Background(p.Surface.Alpha(0.95)).
							Border(theme.BorderWidth, p.Border).
							AlignItems(ui.Center).
							Children(func() {
								ui.Box(c).Size(7, 7).Radius(4).Background(p.Accent).Shrink(0)
								ui.Text(c, statusText).
									Font(assets.Fonts.Sans).
									FontSize(14).
									FontWeight(500).
									TextColor(p.TextMuted).
									SingleLine()
								// Mini scroll progress indicator
								ui.Box(c).Size(44, 4).Radius(2).Background(p.Border.Alpha(0.65)).Clip().Children(func() {
									thumbW := float32(16)
									thumbX := (44 - thumbW) * float32(scrollProgress)
									ui.Box(c).Size(thumbW, 4).Radius(2).Background(p.Accent).Translate(thumbX, 0)
								})
							})
					})

				// Right column: Editorial copy + format/tier specs + live geometric crop diagram
				ui.Column(c).
					Grow(1).
					Height(viewportH).
					Gap(22).
					Children(func() {
						// Top section: Copy & animated export formats
						ui.Column(c).FillWidth().Gap(14).Children(func() {
							Eyebrow(c, "03 · 动图与裁剪", p, assets.Fonts).
								Opacity(eyeOp).
								Translate(0, eyeY)

							Heading(c, "多格式动图导出，\n支持压缩前几何裁剪", 44, p, assets.Fonts).
								Opacity(headOp).
								Translate(0, headY)

							EditorialRule(c, p, float32(280*ruleP))

							BodyText(c, "将队列视频直接导出为动图，内置流畅、均衡、高清三档，并支持自定义起始时间、宽度、帧率与体积目标。", 19, p.TextMuted, assets.Fonts).
								Opacity(bodyOp).
								Translate(0, bodyY)

							// Output formats & tier selector rows
							ui.Column(c).
								FillWidth().
								Gap(12).
								Margin(4, 0, 0, 0).
								Opacity(tagsOp).
								Translate(0, tagsY).
								Children(func() {
									ui.Row(c).Gap(12).AlignItems(ui.Center).Children(func() {
										ui.Text(c, "输出格式").
											Font(assets.Fonts.Sans).
											FontSize(16).
											FontWeight(600).
											TextColor(p.TextMuted).
											Width(78).
											SingleLine()
										for idx, formatName := range []string{"GIF", "APNG", "WebP"} {
											tagOp, _ := EnterOffset(f, float64(26+idx*4), 14, 0)
											PillTag(c, formatName, p, assets.Fonts, true).Opacity(tagOp)
										}
									})

									ui.Row(c).Gap(12).AlignItems(ui.Center).Children(func() {
										ui.Text(c, "预设档位").
											Font(assets.Fonts.Sans).
											FontSize(16).
											FontWeight(600).
											TextColor(p.TextMuted).
											Width(78).
											SingleLine()
										tierSegmentedBar(c, p, assets.Fonts, []string{"流畅", "均衡", "高清", "自定义"}, 3)
									})
								})
						})

						// Bottom section: Geometric crop visualizer card
						SurfaceCard(c, p, 22, 26, func() {
							// Card header
							ui.Row(c).FillWidth().AlignItems(ui.Center).Margin(0, 0, 14, 0).Children(func() {
								ui.Row(c).Gap(8).AlignItems(ui.Center).Children(func() {
									ui.Icon(c, IconSVG("maximize")).Size(18, 18).TextColor(p.AccentText)
									ui.Text(c, "几何裁剪示意 · 去除安全黑边").
										Font(assets.Fonts.Sans).
										FontSize(19).
										FontWeight(600).
										TextColor(p.Text).
										SingleLine()
								})
								ui.Spacer(c)
								cropStatusBadge(c, p, assets.Fonts, cropLocked)
							})

							// Visualizer stage + live readout
							geometricCropDiagram(c, p, assets.Fonts, cropT, cropLocked)

							// Two real CLI crop parameter examples
							ui.Column(c).FillWidth().Gap(8).Margin(14, 0, 0, 0).Children(func() {
								cropParamLine(c, p, assets.Fonts, "居中比例", "--crop-center 16:9 --crop-scale 0.9", true)
								cropParamLine(c, p, assets.Fonts, "精确像素", "--crop 1600:900:160:90", false)
							})

							BodyText(c, "裁剪在缩放前执行，硬编与 CRF 参数不受影响，默认不裁剪。", 15, p.TextMuted, assets.Fonts).
								Margin(10, 0, 0, 0)
						}).
							FillWidth().
							Grow(1).
							Opacity(cropCardOp).
							Translate(0, cropCardY)
					})
			})
	})
}

func tierSegmentedBar(c *ui.Context, p theme.Tokens, fonts FontStacks, labels []string, activeIdx int) {
	ui.Row(c).
		Padding(3, 3).
		Gap(2).
		Radius(theme.ControlRadius + 1).
		Background(p.Background).
		Border(theme.BorderWidth, p.ControlBorder).
		AlignItems(ui.Center).
		Children(func() {
			for i, label := range labels {
				bg := ui.Transparent
				fg := p.TextMuted
				fw := 500
				if i == activeIdx {
					bg = p.Accent
					fg = p.OnAccent
					fw = 600
				}
				ui.Box(c).
					Padding(6, 16).
					Radius(theme.ControlRadius).
					Background(bg).
					Children(func() {
						ui.Text(c, label).
							Font(fonts.Sans).
							FontSize(16).
							FontWeight(fw).
							TextColor(fg).
							SingleLine()
					})
			}
		})
}

func cropStatusBadge(c *ui.Context, p theme.Tokens, fonts FontStacks, locked bool) {
	bg := p.Background
	fg := p.TextMuted
	bd := p.Border
	label := "计算居中比例框..."
	if locked {
		bg = p.Success.Alpha(0.14)
		fg = p.Success
		bd = p.Success.Alpha(0.55)
		label = "16:9 · 90% 已锁定"
	}
	ui.Row(c).
		Padding(5, 12).
		Gap(6).
		Radius(theme.ControlRadius + 1).
		Background(bg).
		Border(theme.BorderWidth, bd).
		AlignItems(ui.Center).
		Children(func() {
			if locked {
				ui.Icon(c, IconSVG("check")).Size(14, 14).TextColor(fg)
			} else {
				ui.Box(c).Size(6, 6).Radius(3).Background(p.Accent)
			}
			ui.Text(c, label).
				Font(fonts.Sans).
				FontSize(14).
				FontWeight(600).
				TextColor(fg).
				SingleLine()
		})
}

func geometricCropDiagram(c *ui.Context, p theme.Tokens, fonts FontStacks, cropT float64, locked bool) {
	const (
		canvasW = float32(476)
		canvasH = float32(236)
	)
	// Interpolate from outer raw frame (with safety margin) to centered 16:9 @ 90% (336x189).
	startW, startH := float32(436), float32(216)
	endW, endH := float32(336), float32(189) // exact 16:9 ratio with generous outer margin for labels
	curW := Lerp32(startW, endW, float32(cropT))
	curH := Lerp32(startH, endH, float32(cropT))
	left := (canvasW - curW) / 2
	top := (canvasH - curH) / 2
	rightW := canvasW - left - curW
	bottomH := canvasH - top - curH

	// Live source-pixel readout corresponding to 1920x1080 -> 1728x972 (90% centered).
	outPxW := int(math.Round(Lerp(1920, 1728, cropT)))
	outPxH := int(math.Round(Lerp(1080, 972, cropT)))
	offsetX := (1920 - outPxW) / 2
	offsetY := (1080 - outPxH) / 2
	scalePct := int(math.Round(Lerp(100, 90, cropT)))

	maskCol := p.Text.Alpha(float32(0.14 + 0.16*cropT))
	frameStroke := p.Accent
	borderW := float32(2)
	if locked {
		borderW = 2.5
	}

	ui.Row(c).FillWidth().Gap(20).AlignItems(ui.Center).Children(func() {
		// Left: Animated 16:9 crop frame diagram
		ui.Box(c).
			Size(canvasW, canvasH).
			Radius(theme.ControlRadius + 1).
			Background(p.Background).
			Border(theme.BorderWidth, p.Border).
			Clip().
			Shrink(0).
			Children(func() {
				// Subtle reference grid lines inside original video frame
				ui.Box(c).Absolute().Left(canvasW / 3).Top(0).Size(1, canvasH).Background(p.Border.Alpha(0.42))
				ui.Box(c).Absolute().Left(canvasW * 2 / 3).Top(0).Size(1, canvasH).Background(p.Border.Alpha(0.42))
				ui.Box(c).Absolute().Left(0).Top(canvasH / 3).Size(canvasW, 1).Background(p.Border.Alpha(0.42))
				ui.Box(c).Absolute().Left(0).Top(canvasH * 2 / 3).Size(canvasW, 1).Background(p.Border.Alpha(0.42))

				// Four exact dimmed letterbox/safety-margin strips outside the crop box
				if top > 0 {
					ui.Box(c).Absolute().Left(0).Top(0).Size(canvasW, top).Background(maskCol)
				}
				if bottomH > 0 {
					ui.Box(c).Absolute().Left(0).Top(top + curH).Size(canvasW, bottomH).Background(maskCol)
				}
				if left > 0 {
					ui.Box(c).Absolute().Left(0).Top(top).Size(left, curH).Background(maskCol)
				}
				if rightW > 0 {
					ui.Box(c).Absolute().Left(left + curW).Top(top).Size(rightW, curH).Background(maskCol)
				}

				// Retained active viewport frame (bright ivory interior + burgundy frame)
				ui.Box(c).
					Absolute().
					Left(left).
					Top(top).
					Size(curW, curH).
					Background(p.Surface.Alpha(0.88)).
					Border(borderW, frameStroke).
					Children(func() {
						// Center editorial crosshair
						ui.Box(c).Absolute().Left(curW/2 - 12).Top(curH/2 - 0.5).Size(24, 1).Background(p.Ornament.Alpha(0.75))
						ui.Box(c).Absolute().Left(curW/2 - 0.5).Top(curH/2 - 12).Size(1, 24).Background(p.Ornament.Alpha(0.75))

						// Crisp L-shaped corner brackets
						const bLen, bThick = float32(16), float32(3)
						ui.Box(c).Absolute().Left(6).Top(6).Size(bLen, bThick).Background(frameStroke)
						ui.Box(c).Absolute().Left(6).Top(6).Size(bThick, bLen).Background(frameStroke)

						ui.Box(c).Absolute().Right(6).Top(6).Size(bLen, bThick).Background(frameStroke)
						ui.Box(c).Absolute().Right(6).Top(6).Size(bThick, bLen).Background(frameStroke)

						ui.Box(c).Absolute().Left(6).Bottom(6).Size(bLen, bThick).Background(frameStroke)
						ui.Box(c).Absolute().Left(6).Bottom(6).Size(bThick, bLen).Background(frameStroke)

						ui.Box(c).Absolute().Right(6).Bottom(6).Size(bLen, bThick).Background(frameStroke)
						ui.Box(c).Absolute().Right(6).Bottom(6).Size(bThick, bLen).Background(frameStroke)

						// Center ratio & dimension badge
						ui.Column(c).Fill().AlignItems(ui.Center).Justify(ui.Center).Gap(4).Children(func() {
							ui.Text(c, fmt.Sprintf("16:9 · 保留 %d%%", scalePct)).
								Font(fonts.Serif).
								FontSize(22).
								FontWeight(600).
								TextColor(p.Text).
								SingleLine()
							MonoText(c, fmt.Sprintf("%d × %d", outPxW, outPxH), 15, p.AccentText, fonts).SingleLine()
						})
					})

				// Top-left raw dimension tag (cleanly tucked inside dimmed margin strip)
				ui.Box(c).
					Absolute().
					Left(8).
					Top(4).
					Padding(1, 6).
					Radius(2).
					Background(p.Surface.Alpha(0.92)).
					Children(func() {
						MonoText(c, "源 1920×1080", 11, p.TextMuted, fonts).SingleLine()
					})

				// Bottom-right margin label (cleanly inside dimmed margin strip)
				if cropT > 0.35 {
					ui.Box(c).
						Absolute().
						Right(8).
						Bottom(4).
						Padding(1, 6).
						Radius(2).
						Background(p.Surface.Alpha(0.92)).
						Children(func() {
							ui.Text(c, "裁除安全边").
								Font(fonts.Sans).
								FontSize(11).
								FontWeight(500).
								TextColor(p.TextMuted).
								SingleLine()
						})
				}
			})

		// Right: Live geometric telemetry readout
		ui.Column(c).
			Grow(1).
			Height(canvasH).
			Justify(ui.SpaceBetween).
			Children(func() {
				cropReadoutItem(c, p, fonts, "目标画幅", fmt.Sprintf("%d × %d", outPxW, outPxH), p.Text)
				Hairline(c, p.Border.Alpha(0.75), 270, 1)
				cropReadoutItem(c, p, fonts, "裁切偏移 (X, Y)", fmt.Sprintf("%d, %d px", offsetX, offsetY), p.Text)
				Hairline(c, p.Border.Alpha(0.75), 270, 1)
				accentVal := p.AccentText
				if locked {
					accentVal = p.Success
				}
				cropReadoutItem(c, p, fonts, "画面保留比例", fmt.Sprintf("16:9 · %d%%", scalePct), accentVal)
			})
	})
}

func cropReadoutItem(c *ui.Context, p theme.Tokens, fonts FontStacks, label, value string, valColor ui.Color) {
	ui.Column(c).Gap(3).Children(func() {
		ui.Text(c, label).
			Font(fonts.Sans).
			FontSize(14).
			TextColor(p.TextMuted).
			SingleLine()
		MonoText(c, value, 19, valColor, fonts).FontWeight(600).SingleLine()
	})
}

func cropParamLine(c *ui.Context, p theme.Tokens, fonts FontStacks, tag, cmd string, primary bool) {
	tagBg := p.Background
	tagFg := p.TextMuted
	if primary {
		tagBg = p.Accent.Alpha(0.12)
		tagFg = p.AccentText
	}
	ui.Row(c).
		FillWidth().
		Padding(8, 12).
		Gap(12).
		Radius(theme.ControlRadius).
		Background(p.Background).
		Border(theme.BorderWidth, p.Border.Alpha(0.85)).
		AlignItems(ui.Center).
		Children(func() {
			ui.Box(c).
				Padding(3, 8).
				Radius(theme.ControlRadius).
				Background(tagBg).
				Shrink(0).
				Children(func() {
					ui.Text(c, tag).
						Font(fonts.Sans).
						FontSize(13).
						FontWeight(600).
						TextColor(tagFg).
						SingleLine()
				})
			MonoText(c, cmd, 15, p.Text, fonts).SingleLine()
		})
}

// =============================================================================
// Act 5: CLI Dual Entry & Windows System Integration (Dark Charcoal Theme)
// Global frames 513..651, local frames 0..138, CueCLISummary at local 81 (global 594).
// =============================================================================

func scene5CLIAndIntegration(c *ui.Context, f render.Frame, assets *Assets) {
	p := DarkPalette()
	stageAlpha := SceneFade(f, CrossfadeFrames, CrossfadeFrames)

	const (
		summaryLocal = float64(CueCLISummary - Scene5Start) // 81
		termW        = float32(980)
		sideW        = float32(720)
		bodyH        = float32(782)
	)

	headOp, headY := EnterOffset(f, 4, 18, 14)
	termOp, termY := EnterOffset(f, 8, 20, 16)
	menuOp, menuY := EnterOffset(f, 16, 20, 16)
	intgOp, intgY := EnterOffset(f, 24, 20, 16)

	SceneStage(c, p.Background, stageAlpha, func() {
		SubtleBackdropFrame(c, p)

		ui.Column(c).
			Fill().
			Padding(70, 88).
			Gap(26).
			Children(func() {
				// Top header bar
				ui.Row(c).
					FillWidth().
					AlignItems(ui.End).
					Opacity(headOp).
					Translate(0, headY).
					Children(func() {
						ui.Column(c).Gap(10).Children(func() {
							Eyebrow(c, "04 · 双入口与系统集成", p, assets.Fonts)
							Heading(c, "桌面与终端同源，融入 Windows 日常操作", 42, p, assets.Fonts)
						})
						ui.Spacer(c)
						ui.Row(c).Gap(10).AlignItems(ui.Center).Margin(0, 0, 6, 0).Children(func() {
							PillTag(c, "GUI · 桌面窗口", p, assets.Fonts, false)
							PillTag(c, "CLI · 终端工作流", p, assets.Fonts, true)
						})
					})

				// Two-column main body: Left PowerShell Terminal (56%), Right Windows Integration (44%)
				ui.Row(c).
					FillWidth().
					Height(bodyH).
					Gap(44).
					Children(func() {
						// Left: PowerShell CLI demonstration inside WindowFrame
						WindowFrame(c, "Windows PowerShell — videopress CLI", termW, bodyH, p, assets.Fonts, func() {
							cliTerminalBody(c, f, p, assets.Fonts, summaryLocal)
						}).
							Opacity(termOp).
							Translate(0, termY).
							Shrink(0)

						// Right: Explorer context menu card + System integration toggles card
						ui.Column(c).
							Width(sideW).
							Height(bodyH).
							Gap(24).
							Shrink(0).
							Children(func() {
								windowsContextMenuCard(c, f, p, assets).
									Opacity(menuOp).
									Translate(0, menuY)
								windowsIntegrationCard(c, f, p, assets.Fonts).
									Opacity(intgOp).
									Translate(0, intgY)
							})
					})
			})
	})
}

func cliTerminalBody(c *ui.Context, f render.Frame, p theme.Tokens, fonts FontStacks, summaryLocal float64) {
	const fullCmd = "videopress --preset standard --hw -c 3 *.mp4 *.mkv"
	typed := sliceRunes(fullCmd, RangeProgress(f, 10, 40, nil))
	showCursor := f.Exact() >= 8 && f.Exact() < 44 && (f.Index/6)%2 == 0
	if showCursor {
		typed += "▌"
	}

	runStarted := f.Exact() >= 42
	overallProg := RangeProgress(f, 42, summaryLocal, EaseInOutCubic)
	summaryShown := f.Exact() >= summaryLocal

	type cliTask struct {
		name     string
		orig     string
		out      string
		saved    string
		dur      string
		doneAt   float64
		startAt  float64
		speedStr string
	}
	tasks := []cliTask{
		{name: "holiday.mp4", orig: "240.0 MB", out: "46.0 MB", saved: "-80.8%", dur: "18.4s", startAt: 45, doneAt: 76, speedStr: "3.1x"},
		{name: "long-webinar.mkv", orig: "128.4 MB", out: "38.2 MB", saved: "-70.2%", dur: "14.1s", startAt: 47, doneAt: 79, speedStr: "2.4x"},
		{name: "demo.mp4", orig: "88.0 MB", out: "41.0 MB", saved: "-53.4%", dur: "6.8s", startAt: 49, doneAt: 71, speedStr: "2.9x"},
	}

	ui.Column(c).
		Fill().
		Padding(24, 28).
		Background(p.Background).
		Gap(14).
		Children(func() {
			// Command prompt line
			ui.Row(c).FillWidth().Gap(10).AlignItems(ui.Center).Children(func() {
				MonoText(c, "PS D:\\clips>", 18, p.Ornament, fonts).FontWeight(600).SingleLine()
				MonoText(c, typed, 18, p.Text, fonts).SingleLine()
			})

			// Hardware probe & worker pool status log
			if runStarted {
				probeOp, _ := EnterOffset(f, 42, 8, 0)
				ui.Column(c).FillWidth().Gap(6).Opacity(probeOp).Children(func() {
					ui.Row(c).Gap(10).AlignItems(ui.Center).Children(func() {
						MonoText(c, "[HW]", 15, p.Success, fonts).FontWeight(600).SingleLine()
						MonoText(c, "已探测硬件编码器: h264_nvenc (NVIDIA) · 预设: standard", 15, p.TextMuted, fonts).SingleLine()
					})
					ui.Row(c).Gap(10).AlignItems(ui.Center).Children(func() {
						MonoText(c, "[POOL]", 15, p.AccentText, fonts).FontWeight(600).SingleLine()
						MonoText(c, "并发队列启动 (-c 3) · 3 个输入文件 · 输出目录 compressed/", 15, p.TextMuted, fonts).SingleLine()
					})
				})

				// Clean non-flickering concurrent worker status lines
				ui.Column(c).
					FillWidth().
					Padding(12, 16).
					Gap(8).
					Radius(theme.ControlRadius).
					Background(p.Surface.Alpha(0.75)).
					Border(theme.BorderWidth, p.Border.Alpha(0.7)).
					Children(func() {
						for _, t := range tasks {
							rowOp, _ := EnterOffset(f, t.startAt, 6, 0)
							taskProg := RangeProgress(f, t.startAt, t.doneAt, EaseInOutCubic)
							done := f.Exact() >= t.doneAt
							statusLabel := fmt.Sprintf("压缩中 %3.0f%%  %s", taskProg*100, t.speedStr)
							statusCol := p.AccentText
							if done {
								statusLabel = fmt.Sprintf("已完成  %s → %s (%s)", t.orig, t.out, t.saved)
								statusCol = p.Success
							}
							ui.Row(c).FillWidth().AlignItems(ui.Center).Opacity(rowOp).Children(func() {
								dotCol := p.Accent
								if done {
									dotCol = p.Success
								}
								ui.Box(c).Size(7, 7).Radius(4).Background(dotCol).Margin(0, 10, 0, 0).Shrink(0)
								MonoText(c, t.name, 15, p.Text, fonts).Width(210).SingleLine()
								MonoText(c, statusLabel, 15, statusCol, fonts).SingleLine()
							})
						}
					})

				// Overall progress bar
				progCol := p.Accent
				progStatus := fmt.Sprintf("总进度  %3.0f%%", overallProg*100)
				if summaryShown {
					progCol = p.Success
					progStatus = "总进度  100%  · 全部任务完成"
				}
				ui.Column(c).FillWidth().Gap(6).Children(func() {
					ui.Row(c).FillWidth().AlignItems(ui.Center).Children(func() {
						MonoText(c, progStatus, 15, progCol, fonts).FontWeight(600).SingleLine()
						ui.Spacer(c)
						MonoText(c, "已写入 compressed/compress_summary.log", 14, p.TextMuted, fonts).
							Opacity(float32(RangeProgress(f, summaryLocal, summaryLocal+10, EaseOutCubic))).
							SingleLine()
					})
					ProgressBar(c, overallProg, 924, 8, progCol, p)
				})
			}

			// Summary table aligned to CueCLISummary (local frame 81)
			if summaryShown {
				tableOp, tableY := EnterOffset(f, summaryLocal, 12, 8)
				ui.Column(c).
					FillWidth().
					Padding(14, 16).
					Gap(8).
					Radius(theme.ControlRadius).
					Background(p.Surface).
					Border(theme.BorderWidth, p.Ornament.Alpha(0.55)).
					Opacity(tableOp).
					Translate(0, tableY).
					Children(func() {
						ui.Row(c).FillWidth().AlignItems(ui.Center).Children(func() {
							MonoText(c, "TASK SUMMARY", 15, p.Ornament, fonts).FontWeight(600).SingleLine()
							ui.Spacer(c)
							MonoText(c, "3 成功 · 0 失败", 14, p.Success, fonts).FontWeight(600).SingleLine()
						})
						Hairline(c, p.Border, 890, 1)

						// Table header (fixed column widths keep CJK + Latin aligned cleanly)
						cliSummaryRow(c, fonts, p.TextMuted, p.TextMuted, p.TextMuted,
							"文件", "状态", "原始大小", "压缩后", "节省", "耗时", true)
						Hairline(c, p.Border.Alpha(0.55), 890, 1)

						for i, t := range tasks {
							rOp, _ := EnterOffset(f, summaryLocal+float64(3+i*4), 8, 0)
							ui.Box(c).FillWidth().Opacity(rOp).Children(func() {
								cliSummaryRow(c, fonts, p.Text, p.Success, p.TextMuted,
									t.name, "成功", t.orig, t.out, t.saved, t.dur, false)
							})
						}
					})
			}

			ui.Spacer(c)

			// Footer CLI flag strip
			Hairline(c, p.Border.Alpha(0.55), 924, 1)
			ui.Row(c).FillWidth().Gap(10).AlignItems(ui.Center).Children(func() {
				ui.Text(c, "常用选项").
					Font(fonts.Sans).
					FontSize(14).
					TextColor(p.TextMuted).
					SingleLine()
				for _, fl := range []string{"--skip-existing", "--force", "--copy-audio", "--install-path"} {
					ui.Box(c).
						Padding(4, 10).
						Radius(theme.ControlRadius).
						Background(p.Surface).
						Border(theme.BorderWidth, p.Border).
						Children(func() {
							MonoText(c, fl, 14, p.TextMuted, fonts).SingleLine()
						})
				}
			})
		})
}

func cliSummaryRow(
	c *ui.Context,
	fonts FontStacks,
	mainCol, statusCol, mutedCol ui.Color,
	file, status, orig, out, saved, dur string,
	header bool,
) {
	fw := 400
	if header {
		fw = 600
	}
	ui.Row(c).FillWidth().AlignItems(ui.Center).Children(func() {
		MonoText(c, file, 15, mainCol, fonts).FontWeight(fw).Width(250).SingleLine()
		MonoText(c, status, 15, statusCol, fonts).FontWeight(600).Width(100).SingleLine()
		MonoText(c, orig, 15, mutedCol, fonts).FontWeight(fw).Width(145).SingleLine()
		MonoText(c, out, 15, mainCol, fonts).FontWeight(fw).Width(145).SingleLine()
		MonoText(c, saved, 15, statusCol, fonts).FontWeight(600).Width(130).SingleLine()
		MonoText(c, dur, 15, mutedCol, fonts).FontWeight(fw).Grow(1).SingleLine()
	})
}

func windowsContextMenuCard(c *ui.Context, f render.Frame, p theme.Tokens, assets *Assets) ui.Element {
	highlightT := float32(RangeProgress(f, 34, 52, EaseOutQuart))
	rowBg := p.SurfaceHover.Mix(p.Accent, highlightT)
	rowFg := p.Text.Mix(p.OnAccent, highlightT)
	subFg := p.TextMuted.Mix(p.OnAccent.Alpha(0.85), highlightT)

	return SurfaceCard(c, p, 22, 26, func() {
		// Explorer multi-file selection strip
		ui.Row(c).FillWidth().AlignItems(ui.Center).Margin(0, 0, 14, 0).Children(func() {
			ui.Row(c).Gap(8).AlignItems(ui.Center).Children(func() {
				ui.Icon(c, IconSVG("folder")).Size(18, 18).TextColor(p.Ornament)
				ui.Text(c, "资源管理器右键菜单").
					Font(assets.Fonts.Sans).
					FontSize(19).
					FontWeight(600).
					TextColor(p.Text).
					SingleLine()
			})
			ui.Spacer(c)
			ui.Row(c).Gap(8).Children(func() {
				for _, fn := range []string{"holiday.mp4", "demo.mp4"} {
					ui.Row(c).
						Padding(4, 10).
						Gap(6).
						Radius(theme.ControlRadius).
						Background(p.Background).
						Border(theme.BorderWidth, p.ControlBorder).
						AlignItems(ui.Center).
						Children(func() {
							ui.Icon(c, IconSVG("file")).Size(13, 13).TextColor(p.AccentText)
							MonoText(c, fn, 13, p.TextMuted, assets.Fonts).SingleLine()
						})
				}
			})
		})

		// Mock Windows Explorer context menu panel
		ui.Column(c).
			FillWidth().
			Padding(8, 8).
			Gap(4).
			Radius(theme.CardRadius).
			Background(p.Background).
			Border(theme.BorderWidth, p.ControlBorder).
			Children(func() {
				contextMenuPlainItem(c, p, assets.Fonts, "打开", "Enter")

				// Highlighted primary action: Compress with Videopress
				ui.Row(c).
					FillWidth().
					Padding(10, 14).
					Gap(10).
					Radius(theme.ControlRadius + 1).
					Background(rowBg).
					AlignItems(ui.Center).
					Children(func() {
						if assets.AppIcon != nil {
							ui.Box(c).Size(20, 20).Radius(4).Clip().Shrink(0).Children(func() {
								ui.Image(c, assets.AppIcon).Size(20, 20).Fit(ui.Cover)
							})
						}
						ui.Text(c, "使用 Videopress 压缩").
							Font(assets.Fonts.Sans).
							FontSize(17).
							FontWeight(600).
							TextColor(rowFg).
							SingleLine()
						ui.Spacer(c)
						ui.Text(c, "单实例合并队列").
							Font(assets.Fonts.Sans).
							FontSize(13).
							TextColor(subFg).
							SingleLine()
					})

				// SendTo item
				ui.Row(c).
					FillWidth().
					Padding(9, 14).
					Gap(10).
					Radius(theme.ControlRadius).
					AlignItems(ui.Center).
					Children(func() {
						ui.Icon(c, IconSVG("share-2")).Size(16, 16).TextColor(p.Ornament)
						ui.Text(c, "发送到 (SendTo)").
							Font(assets.Fonts.Sans).
							FontSize(16).
							TextColor(p.Text).
							SingleLine()
						ui.Icon(c, IconSVG("chevron-right")).Size(14, 14).TextColor(p.TextMuted)
						ui.Text(c, "快速压缩视频").
							Font(assets.Fonts.Sans).
							FontSize(16).
							FontWeight(500).
							TextColor(p.AccentText).
							SingleLine()
					})

				Hairline(c, p.Border.Alpha(0.55), 650, 1)
				contextMenuPlainItem(c, p, assets.Fonts, "属性", "Alt+Enter")
			})

		BodyText(c, "多选文件右键打开仅出现一个窗口，其余文件自动并入已有队列。", 15, p.TextMuted, assets.Fonts).
			Margin(10, 0, 0, 0)
	}).FillWidth().Height(346)
}

func contextMenuPlainItem(c *ui.Context, p theme.Tokens, fonts FontStacks, label, shortcut string) {
	ui.Row(c).
		FillWidth().
		Padding(7, 14).
		AlignItems(ui.Center).
		Children(func() {
			ui.Text(c, label).
				Font(fonts.Sans).
				FontSize(15).
				TextColor(p.TextMuted.Alpha(0.75)).
				SingleLine()
			ui.Spacer(c)
			MonoText(c, shortcut, 13, p.TextMuted.Alpha(0.55), fonts).SingleLine()
		})
}

func windowsIntegrationCard(c *ui.Context, f render.Frame, p theme.Tokens, fonts FontStacks) ui.Element {
	type integrationRow struct {
		icon    string
		title   string
		desc    string
		startAt float64
	}
	rows := []integrationRow{
		{icon: "layout-grid", title: "资源管理器右键菜单", desc: "右键视频直接启动压缩", startAt: 42},
		{icon: "share-2", title: "发送到菜单 (SendTo)", desc: "静默压缩并在完成后推送通知", startAt: 50},
		{icon: "terminal", title: "用户 Path 环境变量", desc: "任意终端目录直接调用 videopress", startAt: 58},
		{icon: "globe", title: "中英双语菜单热切换", desc: "切换界面语言时自动同步右键名称", startAt: 66},
	}

	return SurfaceCard(c, p, 18, 24, func() {
		ui.Row(c).FillWidth().AlignItems(ui.Center).Margin(0, 0, 12, 0).Children(func() {
			ui.Row(c).Gap(8).AlignItems(ui.Center).Children(func() {
				ui.Icon(c, IconSVG("settings")).Size(18, 18).TextColor(p.AccentText)
				ui.Text(c, "Windows 系统集成").
					Font(fonts.Sans).
					FontSize(19).
					FontWeight(600).
					TextColor(p.Text).
					SingleLine()
			})
			ui.Spacer(c)
			ui.Row(c).
				Padding(4, 10).
				Gap(6).
				Radius(theme.ControlRadius).
				Background(p.Success.Alpha(0.14)).
				Border(theme.BorderWidth, p.Success.Alpha(0.45)).
				AlignItems(ui.Center).
				Children(func() {
					ui.Icon(c, IconSVG("shield-check")).Size(13, 13).TextColor(p.Success)
					ui.Text(c, "无需管理员权限").
						Font(fonts.Sans).
						FontSize(13).
						FontWeight(600).
						TextColor(p.Success).
						SingleLine()
				})
		})

		ui.Column(c).FillWidth().Gap(8).Children(func() {
			for i, item := range rows {
				toggleT := float32(RangeProgress(f, item.startAt, item.startAt+14, EaseOutQuart))
				ui.Row(c).
					FillWidth().
					Padding(8, 14).
					Gap(12).
					Radius(theme.ControlRadius).
					Background(p.Background.Alpha(0.75)).
					Border(theme.BorderWidth, p.Border.Alpha(0.75)).
					AlignItems(ui.Center).
					Children(func() {
						ui.Icon(c, IconSVG(item.icon)).Size(17, 17).TextColor(p.AccentText).Shrink(0)
						ui.Column(c).Grow(1).Gap(2).Children(func() {
							ui.Text(c, item.title).
								Font(fonts.Sans).
								FontSize(16).
								FontWeight(600).
								TextColor(p.Text).
								SingleLine()
							ui.Text(c, item.desc).
								Font(fonts.Sans).
								FontSize(13).
								TextColor(p.TextMuted).
								SingleLine()
						})
						promoToggleSwitch(c, p, toggleT)
					})
				_ = i
			}
		})

		BodyText(c, "全部配置仅写入当前用户，支持界面或 CLI 一键移除。", 14, p.TextMuted, fonts).
			Margin(12, 0, 0, 0)
	}).FillWidth().Grow(1)
}

func promoToggleSwitch(c *ui.Context, p theme.Tokens, t float32) {
	const (
		trackW = float32(42)
		trackH = float32(24)
		knobD  = float32(18)
	)
	trackBg := p.Border.Mix(p.Accent, t)
	knobCol := p.TextMuted.Mix(p.OnAccent, t)
	knobX := Lerp32(3, trackW-knobD-3, t)

	ui.Box(c).
		Size(trackW, trackH).
		Radius(trackH / 2).
		Background(trackBg).
		Shrink(0).
		Children(func() {
			ui.Box(c).
				Absolute().
				Left(knobX).
				Top((trackH - knobD) / 2).
				Size(knobD, knobD).
				Radius(knobD / 2).
				Background(knobCol)
		})
}

// =============================================================================
// Act 6: Light/Dark Theme Showcase & Brand Finale End-Card
// Global frames 639..780, local frames 0..141, CueFinaleLock at local 69 (global 708).
// No trailing fade-out: final brand card locks crisp through frame 780.
// =============================================================================

func scene6ThemeAndFinale(c *ui.Context, f render.Frame, assets *Assets) {
	dark := DarkPalette()
	light := LightPalette()

	// Fade in smoothly from Act 5; hold crisp (fadeOut = 0) through the very last frame.
	stageAlpha := SceneFade(f, CrossfadeFrames, 0)

	// Two-phase envelope inside Act 6:
	// Phase A (local 0..72): Dark Charcoal vs Light Ivory side-by-side showcase.
	// Phase B (local 66..141): Brand finale lock-in aligned to CueFinaleLock (local 69).
	phaseAAlpha := float32(1.0 - RangeProgress(f, 58, 72, EaseInOutCubic))
	phaseBAlpha := float32(RangeProgress(f, 66, 80, EaseOutCubic))

	SceneStage(c, dark.Background, stageAlpha, func() {
		SubtleBackdropFrame(c, dark)

		// Phase A: Theme comparison showcase
		if f.Exact() <= 73 && phaseAAlpha > 0.001 {
			render.Layer(c).Opacity(phaseAAlpha).Children(func() {
				scene6ThemeSplit(c, f, dark, light, assets)
			})
		}

		// Phase B: Finale brand end-card
		if f.Exact() >= 65 && phaseBAlpha > 0.001 {
			render.Layer(c).Opacity(phaseBAlpha).Children(func() {
				scene6FinaleCard(c, f, dark, assets)
			})
		}
	})
}

func scene6ThemeSplit(c *ui.Context, f render.Frame, dark, light theme.Tokens, assets *Assets) {
	headOp, headY := EnterOffset(f, 2, 16, 14)
	leftSlide := float32(28 * (1 - RangeProgress(f, 4, 24, EaseOutQuart)))
	leftOp := float32(RangeProgress(f, 4, 18, EaseOutCubic))
	rightSlide := float32(28 * (1 - RangeProgress(f, 10, 30, EaseOutQuart)))
	rightOp := float32(RangeProgress(f, 10, 24, EaseOutCubic))
	dividerH := float32(580 * RangeProgress(f, 8, 28, EaseOutQuart))
	capsOp, capsY := EnterOffset(f, 18, 16, 10)

	const (
		stageW = float32(822)
		stageH = float32(676)
		shotW  = float32(660)
		shotH  = float32(602) // 900x820 aspect ratio
	)

	ui.Column(c).
		Fill().
		Padding(68, 88).
		Gap(24).
		AlignItems(ui.Center).
		Children(func() {
			// Centered heading
			ui.Column(c).
				AlignItems(ui.Center).
				Gap(10).
				Opacity(headOp).
				Translate(0, headY).
				Children(func() {
					Eyebrow(c, "05 · 外观与主题", dark, assets.Fonts)
					Heading(c, "深炭与象牙，默认跟随系统明暗", 44, dark, assets.Fonts)
				})

			// Dual-theme stage with center warm ornament divider
			ui.Row(c).
				FillWidth().
				Height(stageH).
				Gap(32).
				AlignItems(ui.Center).
				Justify(ui.Center).
				Children(func() {
					// Left dark charcoal panel
					ui.Column(c).
						Size(stageW, stageH).
						Padding(20, 24).
						Gap(14).
						Radius(theme.CardRadius).
						Background(dark.Surface).
						Border(theme.BorderWidth, dark.Border).
						AlignItems(ui.Center).
						Opacity(leftOp).
						Translate(-leftSlide, 0).
						Children(func() {
							ui.Row(c).FillWidth().AlignItems(ui.Center).Children(func() {
								ui.Row(c).Gap(8).AlignItems(ui.Center).Children(func() {
									ui.Icon(c, IconSVG("moon")).Size(16, 16).TextColor(dark.Ornament)
									ui.Text(c, "深炭主题 · Dark").
										Font(assets.Fonts.Sans).
										FontSize(17).
										FontWeight(600).
										TextColor(dark.Text).
										SingleLine()
								})
								ui.Spacer(c)
								ui.Text(c, "主工作区 · 夜间低眩光").
									Font(assets.Fonts.Sans).
									FontSize(14).
									TextColor(dark.TextMuted).
									SingleLine()
							})
							ScreenshotCard(c, assets.WorkspaceDark, shotW, shotH, dark)
						})

					// Center warm-gold vertical editorial divider
					ui.Box(c).
						Size(2, dividerH).
						Background(dark.Ornament.Alpha(0.75)).
						Shrink(0)

					// Right light ivory panel (true light background mat for crisp contrast)
					ui.Column(c).
						Size(stageW, stageH).
						Padding(20, 24).
						Gap(14).
						Radius(theme.CardRadius).
						Background(light.Background).
						Border(theme.BorderWidth, light.Border).
						AlignItems(ui.Center).
						Opacity(rightOp).
						Translate(rightSlide, 0).
						Children(func() {
							ui.Row(c).FillWidth().AlignItems(ui.Center).Children(func() {
								ui.Row(c).Gap(8).AlignItems(ui.Center).Children(func() {
									ui.Icon(c, IconSVG("sun")).Size(16, 16).TextColor(light.Accent)
									ui.Text(c, "象牙浅色 · Light").
										Font(assets.Fonts.Sans).
										FontSize(17).
										FontWeight(600).
										TextColor(light.Text).
										SingleLine()
								})
								ui.Spacer(c)
								ui.Text(c, "配置面板 · 外观设置").
									Font(assets.Fonts.Sans).
									FontSize(14).
									TextColor(light.TextMuted).
									SingleLine()
							})
							ScreenshotCard(c, assets.AppearanceLight, shotW, shotH, light)
						})
				})

			// Bottom three feature capsules
			ui.Row(c).
				Gap(18).
				AlignItems(ui.Center).
				Opacity(capsOp).
				Translate(0, capsY).
				Children(func() {
					finaleFeaturePill(c, dark, assets.Fonts, "palette", "跟随系统 / 浅色 / 深色")
					finaleFeaturePill(c, dark, assets.Fonts, "globe", "中英双语切换")
					finaleFeaturePill(c, dark, assets.Fonts, "archive", "单文件便携与安装包")
				})
		})
}

func finaleFeaturePill(c *ui.Context, p theme.Tokens, fonts FontStacks, iconName, text string) {
	ui.Row(c).
		Padding(9, 18).
		Gap(8).
		Radius(theme.ControlRadius + 1).
		Background(p.Surface).
		Border(theme.BorderWidth, p.Border).
		AlignItems(ui.Center).
		Children(func() {
			ui.Icon(c, IconSVG(iconName)).Size(16, 16).TextColor(p.AccentText)
			ui.Text(c, text).
				Font(fonts.Sans).
				FontSize(17).
				FontWeight(500).
				TextColor(p.Text).
				SingleLine()
		})
}

func scene6FinaleCard(c *ui.Context, f render.Frame, p theme.Tokens, assets *Assets) {
	// CueFinaleLock is global 708 -> Scene 6 local 69.
	const finaleLocal = float64(CueFinaleLock - Scene6Start) // 69

	logoP := RangeProgress(f, finaleLocal-3, finaleLocal+17, EaseOutQuart)
	logoScale := Lerp32(0.92, 1.0, float32(logoP))
	titleOp, titleY := EnterOffset(f, finaleLocal, 18, 16)
	ruleW := float32(380 * RangeProgress(f, finaleLocal+4, finaleLocal+24, EaseOutQuart))
	subOp, subY := EnterOffset(f, finaleLocal+8, 18, 12)
	metaOp, metaY := EnterOffset(f, finaleLocal+15, 18, 10)
	repoOp, repoY := EnterOffset(f, finaleLocal+20, 18, 10)

	ui.Column(c).
		Fill().
		AlignItems(ui.Center).
		Justify(ui.Center).
		Padding(80, 120).
		Children(func() {
			// App icon + Serif Wordmark
			BrandMark(c, assets.AppIcon, 112, p).
				Scale(logoScale, logoScale).
				Opacity(float32(logoP)).
				Margin(0, 0, 26, 0)

			Heading(c, "Videopress", 82, p, assets.Fonts).
				SingleLine().
				Opacity(titleOp).
				Translate(0, titleY).
				Margin(0, 0, 20, 0)

			// Two-tone burgundy + warm ornament divider
			ui.Box(c).Height(6).Margin(0, 0, 24, 0).Children(func() {
				EditorialRule(c, p, ruleW)
			})

			// Core positioning sentence
			ui.Text(c, "Windows 原生视频压缩工具 · GUI 与 CLI 双入口").
				Font(assets.Fonts.Sans).
				FontSize(28).
				FontWeight(400).
				TextColor(p.Text.Alpha(0.92)).
				SingleLine().
				Opacity(subOp).
				Translate(0, subY).
				Margin(0, 0, 38, 0)

			// Architecture & open-source stack strip
			ui.Row(c).
				Gap(14).
				AlignItems(ui.Center).
				Opacity(metaOp).
				Translate(0, metaY).
				Margin(0, 0, 26, 0).
				Children(func() {
					PillTag(c, "Go + MyGO + MujicaUI", p, assets.Fonts, false)
					PillTag(c, "FFmpeg 编码引擎", p, assets.Fonts, false)
					PillTag(c, "MIT License", p, assets.Fonts, false)
				})

			// Repository address bar
			ui.Row(c).
				Padding(12, 24).
				Gap(12).
				Radius(theme.ControlRadius + 1).
				Background(p.Surface).
				Border(theme.BorderWidth, p.Ornament.Alpha(0.65)).
				AlignItems(ui.Center).
				Opacity(repoOp).
				Translate(0, repoY).
				Children(func() {
					ui.Icon(c, IconSVG("git-branch")).Size(18, 18).TextColor(p.Ornament)
					MonoText(c, "github.com/Shirolin/videopress", 22, p.Text, assets.Fonts).
						FontWeight(500).
						SingleLine()
				})
		})
}

// sliceRunes returns the first floor(progress * len(runes)) characters of s safely.
func sliceRunes(s string, progress float64) string {
	r := []rune(s)
	n := int(math.Floor(Clamp01(progress) * float64(len(r))))
	if n <= 0 {
		return ""
	}
	if n >= len(r) {
		return s
	}
	return string(r[:n])
}
