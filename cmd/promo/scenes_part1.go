package main

import (
	"fmt"

	"github.com/ELY-Open-Source/MujicaUI/theme"
	"github.com/egoist/mygo/plugins/render"
	"github.com/egoist/mygo/ui"
)

// BuildPart1 mounts Acts 1-3 of the Videopress product video:
//   - Act 1 (0..120): Brand opening and core positioning
//   - Act 2 (108..255): Drag-and-drop batch workspace with real app screenshots
//   - Act 3 (243..390): GPU hardware acceleration probe and Worker Pool pipeline
func BuildPart1(c *ui.Context, f render.Frame, assets *Assets) {
	f.Sequence(c, Scene1Start, Scene1Duration, func(sf render.Frame) {
		renderScene1(c, sf, assets)
	})
	f.Sequence(c, Scene2Start, Scene2Duration, func(sf render.Frame) {
		renderScene2(c, sf, assets)
	})
	f.Sequence(c, Scene3Start, Scene3Duration, func(sf render.Frame) {
		renderScene3(c, sf, assets)
	})
}

// renderScene1 builds Act 1: Brand Opening & Core Positioning (local frames 0..120).
func renderScene1(c *ui.Context, f render.Frame, assets *Assets) {
	p := DarkPalette()
	fonts := assets.Fonts

	// Frame 0 serves as a crisp, fully resolved poster/cover frame for players,
	// file explorers, and social previews, while frames 1..119 play the normal
	// editorial entrance without shifting total duration or audio cues.
	isCoverFrame := f.Exact() < 0.5

	// Keep the canvas grounded in charcoal from frame 0, then dissolve out cleanly
	// over the final CrossfadeFrames into Act 2.
	stageAlpha := SceneFade(f, 0, CrossfadeFrames)

	frameAlpha := float32(RangeProgress(f, 0, 14, EaseOutCubic))

	// Brand mark and wordmark reveal aligned to CueBrandReveal = 12.
	markOp, markY := EnterOffset(f, 12, 22, 18)
	markScale := float32(Lerp(0.93, 1.0, RangeProgress(f, 12, 34, EaseOutQuart)))

	titleOp, titleY := EnterOffset(f, 16, 24, 22)
	ruleW := float32(Lerp(0, 440, RangeProgress(f, 24, 48, EaseOutQuart)))

	subOp, subY := EnterOffset(f, 32, 24, 18)
	descOp, descY := EnterOffset(f, 44, 22, 14)

	if isCoverFrame {
		frameAlpha = 1
		markOp, markY, markScale = 1, 0, 1
		titleOp, titleY = 1, 0
		ruleW = 440
		subOp, subY = 1, 0
		descOp, descY = 1, 0
	}

	SceneStage(c, p.Background, stageAlpha, func() {
		// Editorial margin frame and subtle architectural metadata on its own full-bleed layer
		// so the centered hero column fills the entire 1920x1080 canvas cleanly.
		render.Layer(c).Opacity(frameAlpha).Children(func() {
			SubtleBackdropFrame(c, p)

			// Quiet top corner markers inside the editorial margins.
			ui.Row(c).
				Absolute().Top(58).Left(64).Right(64).
				AlignItems(ui.Center).
				Children(func() {
					ui.Text(c, "Videopress · Windows 桌面端").
						Font(fonts.Sans).
						FontSize(16).
						TextColor(p.TextMuted.Alpha(0.8)).
						SingleLine()
					ui.Spacer(c)
					MonoText(c, "Go + MyGO + MujicaUI", 15, p.TextMuted.Alpha(0.75), fonts).SingleLine()
				})

			// Quiet bottom edge indicator.
			ui.Row(c).
				Absolute().Bottom(58).Left(64).Right(64).
				AlignItems(ui.Center).
				Children(func() {
					MonoText(c, "Windows 10 / 11 (x64)", 15, p.TextMuted.Alpha(0.72), fonts).SingleLine()
					ui.Spacer(c)
					ui.Text(c, "基于 FFmpeg 视频编码引擎").
						Font(fonts.Sans).
						FontSize(15).
						TextColor(p.TextMuted.Alpha(0.72)).
						SingleLine()
				})
		})

		// Centered hero stack.
		render.Layer(c).Children(func() {
			ui.Column(c).
				Fill().
				Padding(96, 120).
				AlignItems(ui.Center).
				Justify(ui.Center).
				Children(func() {
				// Application icon.
				ui.Box(c).
					Opacity(markOp).
					Translate(0, markY).
					Scale(markScale, markScale).
					Children(func() {
						BrandMark(c, assets.AppIcon, 120, p)
					})

				// Serif brand wordmark.
				ui.Box(c).
					Margin(28, 0, 0, 0).
					Opacity(titleOp).
					Translate(0, titleY).
					Children(func() {
						ui.Text(c, "Videopress").
							Font(fonts.Serif).
							FontSize(92).
							FontWeight(600).
							LineHeight(1.12).
							TextColor(p.Text).
							SingleLine()
					})

				// Burgundy + warm ornament expanding editorial rule.
				ui.Box(c).
					Margin(22, 0, 26, 0).
					Height(4).
					AlignItems(ui.Center).
					Justify(ui.Center).
					Children(func() {
						if ruleW > 1 {
							EditorialRule(c, p, ruleW)
						}
					})

				// Primary positioning statement.
				ui.Box(c).
					Opacity(subOp).
					Translate(0, subY).
					Children(func() {
						ui.Text(c, "面向 Windows 的原生视频压缩工具").
							Font(fonts.Sans).
							FontSize(34).
							FontWeight(500).
							LineHeight(1.35).
							TextColor(p.Text.Mix(p.TextMuted, 0.16)).
							SingleLine()
					})

				// Calm factual sub-descriptor.
				ui.Box(c).
					Margin(12, 0, 0, 0).
					Opacity(descOp).
					Translate(0, descY).
					Children(func() {
						BodyText(c, "免记复杂参数 · 预设即用 · 硬件加速与并发队列", 21, p.TextMuted, fonts).SingleLine()
					})

				// Three staggered capability badges.
				type capBadge struct {
					icon  string
					label string
					start float64
				}
				badges := []capBadge{
					{icon: "layout-grid", label: "纯 Go 原生界面", start: 54},
					{icon: "sliders-horizontal", label: "FFmpeg 编码引擎", start: 64},
					{icon: "terminal", label: "GUI 与 CLI 双入口", start: 74},
				}

				ui.Row(c).
					Margin(48, 0, 0, 0).
					Gap(24).
					AlignItems(ui.Center).
					Children(func() {
						for _, b := range badges {
							op, dy := EnterOffset(f, b.start, 20, 16)
							if isCoverFrame {
								op, dy = 1, 0
							}
							ui.Row(c).
								Padding(12, 24).
								Gap(12).
								Radius(theme.ControlRadius + 2).
								Background(p.Surface).
								Border(theme.BorderWidth, p.Border).
								AlignItems(ui.Center).
								Opacity(op).
								Translate(0, dy).
								Children(func() {
									ui.Icon(c, IconSVG(b.icon)).Size(20, 20).TextColor(p.AccentText).Shrink(0)
									ui.Text(c, b.label).
										Font(fonts.Sans).
										FontSize(20).
										FontWeight(500).
										TextColor(p.Text).
										SingleLine()
								})
						}
					})
			})
		})
	})
}

// renderScene2 builds Act 2: Drag-and-Drop & Batch Compression Workspace (local frames 0..147).
func renderScene2(c *ui.Context, f render.Frame, assets *Assets) {
	p := DarkPalette()
	fonts := assets.Fonts

	stageAlpha := SceneFade(f, CrossfadeFrames, CrossfadeFrames)

	// Left editorial column choreography.
	eyebrowOp, eyebrowY := EnterOffset(f, 4, 18, 12)
	headOp, headY := EnterOffset(f, 10, 22, 18)
	ruleW := float32(Lerp(0, 280, RangeProgress(f, 18, 40, EaseOutQuart)))
	presetOp, presetY := EnterOffset(f, 28, 18, 12)
	metricOp, metricY := EnterOffset(f, 34, 20, 14)

	// Animated real metric count-up matching workspace-dark.png (241.0 MB, 73.5%)
	// so numbers settle cleanly while workspace-dark.png is still on screen.
	metricProg := RangeProgress(f, 34, 62, EaseOutQuart)
	savedMB := 241.0 * metricProg
	savedPct := 73.5 * metricProg

	// Right showcase frame choreography:
	// Local frame 66 aligns to global CueWorkspaceSwap = 174 (108 + 66).
	shotOp, shotY := EnterOffset(f, 8, 24, 18)
	swapT := float32(RangeProgress(f, 66, 84, EaseInOutCubic))
	zoom := float32(Lerp(1.00, 1.016, RangeProgress(f, 16, 136, EaseInOutCubic)))
	captionOp, captionY := EnterOffset(f, 24, 20, 10)

	const (
		shotW = float32(892)
		shotH = float32(812)
	)

	SceneStage(c, p.Background, stageAlpha, func() {
		SubtleBackdropFrame(c, p)

		ui.Row(c).
			Fill().
			Padding(78, 92).
			Gap(56).
			AlignItems(ui.Center).
			Children(func() {
				// Left editorial narrative column (660px).
				ui.Column(c).
					Width(660).
					Shrink(0).
					Children(func() {
						// Section eyebrow.
						ui.Box(c).Opacity(eyebrowOp).Translate(0, eyebrowY).Children(func() {
							Eyebrow(c, "01 · 批量工作区", p, fonts)
						})

						// Serif two-line display heading.
						ui.Box(c).
							Margin(18, 0, 0, 0).
							Opacity(headOp).
							Translate(0, headY).
							Children(func() {
								Heading(c, "拖入即压，\n队列进度清晰可控", 52, p, fonts)
							})

						// Editorial rule.
						ui.Box(c).Margin(22, 0, 26, 0).Height(3).Children(func() {
							if ruleW > 1 {
								EditorialRule(c, p, ruleW)
							}
						})

						// Concise functional bullets.
						type bullet struct {
							icon  string
							title string
							desc  string
							start float64
						}
						bullets := []bullet{
							{
								icon:  "sliders-horizontal",
								title: "三档内置预设",
								desc:  "小文件 / 标准 / 高画质，覆盖日常分享与存档",
								start: 22,
							},
							{
								icon:  "folder",
								title: "原目录安全输出",
								desc:  "默认写入 compressed/ 文件夹，不覆盖原片",
								start: 30,
							},
							{
								icon:  "list",
								title: "可控队列与状态筛选",
								desc:  "实时显示编码速率与预计剩余时间，支持逐文件取消",
								start: 38,
							},
						}

						ui.Column(c).Gap(16).Children(func() {
							for _, b := range bullets {
								op, dy := EnterOffset(f, b.start, 18, 12)
								ui.Row(c).
									Gap(14).
									AlignItems(ui.Start).
									Opacity(op).
									Translate(0, dy).
									Children(func() {
										ui.Box(c).
											Size(34, 34).
											Radius(theme.ControlRadius + 1).
											Background(p.Surface).
											Border(theme.BorderWidth, p.Border).
											AlignItems(ui.Center).
											Justify(ui.Center).
											Shrink(0).
											Margin(2, 0, 0, 0).
											Children(func() {
												ui.Icon(c, IconSVG(b.icon)).Size(17, 17).TextColor(p.AccentText)
											})
										ui.Column(c).Gap(2).Grow(1).MinWidth(0).Children(func() {
											ui.Text(c, b.title).
												Font(fonts.Sans).
												FontSize(20).
												FontWeight(600).
												TextColor(p.Text).
												SingleLine()
											BodyText(c, b.desc, 17, p.TextMuted, fonts)
										})
									})
							}
						})

						// Compact preset bar preview matching the real GUI segmented control.
						ui.Row(c).
							Margin(26, 0, 0, 0).
							Gap(10).
							AlignItems(ui.Center).
							Opacity(presetOp).
							Translate(0, presetY).
							Children(func() {
								ui.Text(c, "压缩预设").
									Font(fonts.Sans).
									FontSize(16).
									FontWeight(600).
									TextColor(p.TextMuted).
									Margin(0, 8, 0, 0).
									SingleLine()
								PillTag(c, "小文件", p, fonts, false)
								PillTag(c, "标准", p, fonts, true)
								PillTag(c, "高画质", p, fonts, false)
							})

						// Real workspace compression metrics card.
						ui.Box(c).
							Margin(28, 0, 0, 0).
							FillWidth().
							Opacity(metricOp).
							Translate(0, metricY).
							Children(func() {
								SurfaceCard(c, p, 22, 28, func() {
									ui.Row(c).FillWidth().AlignItems(ui.Center).Children(func() {
										ui.Box(c).Grow(1).Children(func() {
											MetricTile(c, fmt.Sprintf("%.1f MB", savedMB), "共节省空间", p.Success, 42, p, fonts)
										})
										Hairline(c, p.Border, 1, 64).Margin(0, 28, 0, 28)
										ui.Box(c).Grow(1).Children(func() {
											MetricTile(c, fmt.Sprintf("%.1f%%", savedPct), "整体节省率", p.Success, 42, p, fonts)
										})
									})
								}).FillWidth()
							})
					})

				// Right real-app showcase column (1020px).
				ui.Column(c).
					Grow(1).
					MinWidth(0).
					AlignItems(ui.Center).
					Justify(ui.Center).
					Children(func() {
						// Static clipped frame wrapping both screenshots; only inner images scale/fade
						// so the clip mask stays static across frames.
						shadowCol := ui.RGBA(0, 0, 0, 0.38)
						ui.Box(c).
							Size(shotW, shotH).
							Radius(theme.PopupRadius).
							Background(p.Surface).
							Border(theme.BorderWidth, p.ControlBorder.Alpha(0.8)).
							Shadow(0, 24, 52, 0, shadowCol).
							Opacity(shotOp).
							Translate(0, shotY).
							Clip().
							Children(func() {
								// Layer 1: workspace-dark.png
								if assets.WorkspaceDark != nil {
									ui.Image(c, assets.WorkspaceDark).
										Absolute().Top(0).Left(0).Size(shotW, shotH).
										Fit(ui.Cover).
										Scale(zoom, zoom)
								}
								// Layer 2: batch-progress.png cross-dissolves in at CueWorkspaceSwap (local 66)
								if assets.BatchProgress != nil && swapT > 0.001 {
									ui.Image(c, assets.BatchProgress).
										Absolute().Top(0).Left(0).Size(shotW, shotH).
										Fit(ui.Cover).
										Scale(zoom, zoom).
										Opacity(swapT)
								}
							})

						// Restrained status & output annotation strip directly beneath the screenshot.
						ui.Row(c).
							Margin(16, 0, 0, 0).
							Width(shotW).
							Height(46).
							Padding(0, 18).
							Radius(theme.CardRadius).
							Background(p.Surface).
							Border(theme.BorderWidth, p.Border).
							AlignItems(ui.Center).
							Opacity(captionOp).
							Translate(0, captionY).
							Children(func() {
								dotColor := p.AccentText.Mix(p.Success, swapT*0.35)
								ui.Box(c).Size(8, 8).Radius(4).Background(dotColor).Shrink(0).Margin(0, 10, 0, 0)

								// Smooth caption swap synced with CueWorkspaceSwap.
								ui.Box(c).Width(250).Height(24).Children(func() {
									ui.Text(c, "主工作区 · 队列与压缩统计").
										Absolute().Top(0).Left(0).
										Font(fonts.Sans).
										FontSize(16).
										FontWeight(500).
										TextColor(p.Text).
										Opacity(1 - swapT).
										SingleLine()
									ui.Text(c, "批量进度 · 整体完成 73%").
										Absolute().Top(0).Left(0).
										Font(fonts.Sans).
										FontSize(16).
										FontWeight(500).
										TextColor(p.Text).
										Opacity(swapT).
										SingleLine()
								})

								ui.Spacer(c)

								MonoText(c, "默认写入 compressed/ · 自动记录 compress_summary.log", 14, p.TextMuted, fonts).SingleLine()
							})
					})
			})
	})
}

// renderScene3 builds Act 3: GPU Hardware Acceleration & Concurrent Worker Pool (local frames 0..147).
func renderScene3(c *ui.Context, f render.Frame, assets *Assets) {
	p := DarkPalette()
	fonts := assets.Fonts

	stageAlpha := SceneFade(f, CrossfadeFrames, CrossfadeFrames)

	// Header animations.
	eyebrowOp, eyebrowY := EnterOffset(f, 4, 18, 12)
	headOp, headY := EnterOffset(f, 10, 22, 16)
	badgeOp, badgeY := EnterOffset(f, 16, 20, 12)
	ruleW := float32(Lerp(0, 320, RangeProgress(f, 16, 40, EaseOutQuart)))

	// Stage card entrances.
	col1Op, col1Y := EnterOffset(f, 12, 22, 16)
	col2Op, col2Y := EnterOffset(f, 18, 22, 16)
	col3Op, col3Y := EnterOffset(f, 24, 22, 16)

	// Worker progress curves:
	// Worker 1 finishes at local 84, Worker 3 at 89, Worker 2 locks right at local 93 (CueWorkerComplete = 336).
	w1Prog := RangeProgress(f, 24, 84, EaseInOutCubic)
	w2Prog := RangeProgress(f, 28, 93, EaseInOutCubic)
	w3Prog := RangeProgress(f, 32, 89, EaseInOutCubic)
	allDoneT := float32(RangeProgress(f, 92, 100, EaseOutCubic))

	// GPU probe locks on NVIDIA h264_nvenc at local 38..48.
	gpuLockT := float32(RangeProgress(f, 36, 48, EaseOutQuart))

	// Animated pulse along pipeline connectors while workers are active.
	pulseCycle := float32(RangeProgress(f, 20, 93, ui.Linear))
	pulsePos := (pulseCycle * 3.0) - float32(int(pulseCycle*3.0))

	SceneStage(c, p.Background, stageAlpha, func() {
		SubtleBackdropFrame(c, p)

		ui.Column(c).
			Fill().
			Padding(72, 120).
			Justify(ui.Center).
			Children(func() {
				// Top header bar (1680px wide).
				ui.Row(c).
					FillWidth().
					AlignItems(ui.End).
					Children(func() {
						ui.Column(c).Gap(10).Children(func() {
							ui.Box(c).Opacity(eyebrowOp).Translate(0, eyebrowY).Children(func() {
								Eyebrow(c, "02 · 硬件加速与并发", p, fonts)
							})
							ui.Box(c).Opacity(headOp).Translate(0, headY).Children(func() {
								Heading(c, "自动探测显卡硬编，多文件并发压缩", 46, p, fonts).SingleLine()
							})
						})

						ui.Spacer(c)

						// Right technical tag pill.
						ui.Row(c).
							Padding(10, 18).
							Gap(10).
							Radius(theme.ControlRadius + 2).
							Background(p.Surface).
							Border(theme.BorderWidth, p.Border).
							AlignItems(ui.Center).
							Opacity(badgeOp).
							Translate(0, badgeY).
							Children(func() {
								ui.Box(c).Size(8, 8).Radius(4).Background(p.Accent.Mix(p.Success, allDoneT)).Shrink(0)
								MonoText(c, "Worker Pool · -c 3 · --hw", 16, p.AccentText, fonts).SingleLine()
							})
					})

				// Full-width subtle rule with burgundy/ornament accent on the left.
				ui.Box(c).Margin(18, 0, 26, 0).FillWidth().Height(2).Children(func() {
					ui.Box(c).Absolute().Top(1).Left(0).Right(0).Height(1).Background(p.Border.Alpha(0.55))
					if ruleW > 1 {
						ui.Box(c).Absolute().Top(0).Left(0).Children(func() {
							EditorialRule(c, p, ruleW)
						})
					}
				})

				// Middle 3-stage architecture workflow (440 + 60 + 660 + 60 + 460 = 1680px).
				const stageH = float32(512)
				ui.Row(c).
					FillWidth().
					Height(stageH).
					AlignItems(ui.Center).
					Children(func() {
						// Stage 1: Input Queue Card (440px).
						ui.Box(c).Width(440).Height(stageH).Shrink(0).Opacity(col1Op).Translate(0, col1Y).Children(func() {
							renderScene3QueueCard(c, f, p, fonts, w1Prog, w2Prog, w3Prog, allDoneT)
						})

						// Connector 1 (60px).
						renderPipelineConnector(c, p, pulsePos, allDoneT, col2Op)

						// Stage 2: Worker Pool Card (660px).
						ui.Box(c).Width(660).Height(stageH).Shrink(0).Opacity(col2Op).Translate(0, col2Y).Children(func() {
							renderScene3WorkerPool(c, f, p, fonts, w1Prog, w2Prog, w3Prog, allDoneT)
						})

						// Connector 2 (60px).
						renderPipelineConnector(c, p, pulsePos, allDoneT, col3Op)

						// Stage 3: GPU Encoder Probe Card (460px).
						ui.Box(c).Width(460).Height(stageH).Shrink(0).Opacity(col3Op).Translate(0, col3Y).Children(func() {
							renderScene3GPUCard(c, f, p, fonts, gpuLockT)
						})
					})

				// Bottom 3-column feature strip (3 × 544px + 2 × 24px gap = 1680px).
				type featItem struct {
					icon  string
					title string
					desc  string
					start float64
				}
				feats := []featItem{
					{
						icon:  "sliders-horizontal",
						title: "可调并发数 (Worker Pool)",
						desc:  "自由设定并发上限，多文件稳定并行处理",
						start: 30,
					},
					{
						icon:  "volume-2",
						title: "音频流直通 (--copy-audio)",
						desc:  "直接复用原片音轨，省去音频重编码开销",
						start: 36,
					},
					{
						icon:  "file-text",
						title: "压缩摘要自动归档",
						desc:  "完成后自动写入 compress_summary.log 便于复盘",
						start: 42,
					},
				}

				ui.Row(c).
					Margin(24, 0, 0, 0).
					FillWidth().
					Gap(24).
					Children(func() {
						for _, item := range feats {
							op, dy := EnterOffset(f, item.start, 18, 12)
							ui.Box(c).Grow(1).MinWidth(0).Opacity(op).Translate(0, dy).Children(func() {
								SurfaceCard(c, p, 18, 22, func() {
									ui.Row(c).Gap(14).AlignItems(ui.Start).Children(func() {
										ui.Box(c).
											Size(36, 36).
											Radius(theme.ControlRadius + 1).
											Background(p.Background).
											Border(theme.BorderWidth, p.Border).
											AlignItems(ui.Center).
											Justify(ui.Center).
											Shrink(0).
											Children(func() {
												ui.Icon(c, IconSVG(item.icon)).Size(18, 18).TextColor(p.AccentText)
											})
										ui.Column(c).Gap(4).Grow(1).MinWidth(0).Children(func() {
											ui.Text(c, item.title).
												Font(fonts.Sans).
												FontSize(18).
												FontWeight(600).
												TextColor(p.Text).
												SingleLine()
											BodyText(c, item.desc, 15, p.TextMuted, fonts).SingleLine()
										})
									})
								}).FillWidth()
							})
						}
					})
			})
	})
}

func renderScene3QueueCard(
	c *ui.Context,
	f render.Frame,
	p theme.Tokens,
	fonts FontStacks,
	w1Prog, w2Prog, w3Prog float64,
	allDoneT float32,
) {
	type queueItem struct {
		name  string
		size  string
		state string
		done  bool
		run   bool
		start float64
	}

	q4State := "排队中"
	if allDoneT > 0.5 {
		q4State = "等待下轮"
	}

	items := []queueItem{
		{
			name:  "holiday.mp4",
			size:  "240.0 MB",
			state: workerQueueLabel(w1Prog, "Worker #1"),
			done:  w1Prog >= 0.999,
			run:   w1Prog > 0.01 && w1Prog < 0.999,
			start: 18,
		},
		{
			name:  "long-webinar.mkv",
			size:  "128.4 MB",
			state: workerQueueLabel(w2Prog, "Worker #2"),
			done:  w2Prog >= 0.999,
			run:   w2Prog > 0.01 && w2Prog < 0.999,
			start: 24,
		},
		{
			name:  "demo.mp4",
			size:  "88.0 MB",
			state: workerQueueLabel(w3Prog, "Worker #3"),
			done:  w3Prog >= 0.999,
			run:   w3Prog > 0.01 && w3Prog < 0.999,
			start: 30,
		},
		{
			name:  "capture_4k.mov",
			size:  "312.0 MB",
			state: q4State,
			done:  false,
			run:   false,
			start: 36,
		},
	}

	SurfaceCard(c, p, 22, 24, func() {
		// Card header.
		ui.Row(c).FillWidth().AlignItems(ui.Center).Children(func() {
			ui.Icon(c, IconSVG("list")).Size(20, 20).TextColor(p.AccentText).Margin(0, 10, 0, 0).Shrink(0)
			ui.Text(c, "输入任务队列").
				Font(fonts.Sans).
				FontSize(21).
				FontWeight(600).
				TextColor(p.Text).
				SingleLine()
			ui.Spacer(c)
			MonoText(c, "4 个文件", 15, p.TextMuted, fonts).SingleLine()
		})

		Hairline(c, p.Border, 392, 1).Margin(16, 0, 18, 0)

		// 4 task rows.
		ui.Column(c).FillWidth().Gap(14).Grow(1).Children(func() {
			for _, it := range items {
				op, dy := EnterOffset(f, it.start, 16, 10)
				badgeBg := p.SurfaceHover
				badgeFg := p.TextMuted
				rowBorder := p.Border
				iconName := "file"
				iconColor := p.TextMuted

				switch {
				case it.done:
					badgeBg = p.Success.Mix(p.Surface, 0.82)
					badgeFg = p.Success
					rowBorder = p.Success.Alpha(0.45)
					iconName = "circle-check"
					iconColor = p.Success
				case it.run:
					badgeBg = p.Accent.Mix(p.Surface, 0.72)
					badgeFg = p.AccentText
					rowBorder = p.Accent.Alpha(0.65)
					iconName = "play"
					iconColor = p.AccentText
				}

				ui.Row(c).
					FillWidth().
					Padding(14, 16).
					Radius(theme.CardRadius).
					Background(p.Background).
					Border(theme.BorderWidth, rowBorder).
					AlignItems(ui.Center).
					Opacity(op).
					Translate(0, dy).
					Children(func() {
						ui.Icon(c, IconSVG(iconName)).Size(18, 18).TextColor(iconColor).Margin(0, 12, 0, 0).Shrink(0)
						ui.Column(c).Gap(3).Grow(1).MinWidth(0).Children(func() {
							MonoText(c, it.name, 16, p.Text, fonts).SingleLine()
							MonoText(c, it.size, 14, p.TextMuted, fonts).SingleLine()
						})
						ui.Box(c).
							Padding(4, 10).
							Radius(theme.ControlRadius).
							Background(badgeBg).
							Shrink(0).
							Children(func() {
								ui.Text(c, it.state).
									Font(fonts.Sans).
									FontSize(13).
									FontWeight(600).
									TextColor(badgeFg).
									SingleLine()
							})
					})
			}
		})

		// Footer hint.
		Hairline(c, p.Border, 392, 1).Margin(14, 0, 12, 0)
		ui.Text(c, "支持同批路径冲突检测与增量跳过同名文件").
			Font(fonts.Sans).
			FontSize(14).
			TextColor(p.TextMuted).
			SingleLine()
	}).Fill()
}

func workerQueueLabel(prog float64, workerName string) string {
	switch {
	case prog >= 0.999:
		return "已完成"
	case prog > 0.01:
		return workerName
	default:
		return "待调度"
	}
}

func renderScene3WorkerPool(
	c *ui.Context,
	f render.Frame,
	p theme.Tokens,
	fonts FontStacks,
	w1Prog, w2Prog, w3Prog float64,
	allDoneT float32,
) {
	type workerLane struct {
		id      string
		file    string
		speed   string
		summary string
		prog    float64
	}

	lanes := []workerLane{
		{
			id:      "Worker #1",
			file:    "holiday.mp4",
			speed:   "2.4x",
			summary: "240.0 -> 46.0 MB (-80.8%)",
			prog:    w1Prog,
		},
		{
			id:      "Worker #2",
			file:    "long-webinar.mkv",
			speed:   "3.1x",
			summary: "128.4 -> 38.2 MB (-70.2%)",
			prog:    w2Prog,
		},
		{
			id:      "Worker #3",
			file:    "demo.mp4",
			speed:   "2.8x",
			summary: "88.0 -> 41.0 MB (-53.4%)",
			prog:    w3Prog,
		},
	}

	avgPct := (w1Prog + w2Prog + w3Prog) / 3.0 * 100.0

	SurfaceCard(c, p, 22, 26, func() {
		// Header.
		ui.Row(c).FillWidth().AlignItems(ui.Center).Children(func() {
			ui.Icon(c, IconSVG("sliders-horizontal")).Size(20, 20).TextColor(p.AccentText).Margin(0, 10, 0, 0).Shrink(0)
			ui.Text(c, "Worker Pool 并发池").
				Font(fonts.Sans).
				FontSize(21).
				FontWeight(600).
				TextColor(p.Text).
				SingleLine()
			ui.Spacer(c)
			ui.Box(c).
				Padding(4, 12).
				Radius(theme.ControlRadius).
				Background(p.Accent).
				Children(func() {
					MonoText(c, "-c 3", 14, p.OnAccent, fonts).SingleLine()
				})
		})

		Hairline(c, p.Border, 608, 1).Margin(16, 0, 18, 0)

		// 3 parallel worker lanes.
		ui.Column(c).FillWidth().Gap(22).Grow(1).Justify(ui.Center).Children(func() {
			for i, lane := range lanes {
				done := lane.prog >= 0.999
				borderCol := p.Border
				barColor := p.AccentText
				if done {
					borderCol = p.Success.Alpha(0.55)
					barColor = p.Success
				} else if lane.prog > 0.01 {
					borderCol = p.ControlBorder
				}

				op, dy := EnterOffset(f, float64(20+i*6), 16, 10)
				ui.Column(c).
					FillWidth().
					Padding(19, 20).
					Gap(13).
					Radius(theme.CardRadius).
					Background(p.Background).
					Border(theme.BorderWidth, borderCol).
					Opacity(op).
					Translate(0, dy).
					Children(func() {
						// Lane top row: Worker label + filename on left, speed or completion on right.
						ui.Row(c).FillWidth().AlignItems(ui.Center).Children(func() {
							ui.Box(c).
								Padding(3, 9).
								Radius(theme.ControlRadius).
								Background(p.Surface).
								Border(theme.BorderWidth, p.Border).
								Margin(0, 12, 0, 0).
								Shrink(0).
								Children(func() {
									MonoText(c, lane.id, 13, p.TextMuted, fonts).SingleLine()
								})

							MonoText(c, lane.file, 16, p.Text, fonts).SingleLine()

							ui.Spacer(c)

							if done {
								ui.Row(c).Gap(6).AlignItems(ui.Center).Children(func() {
									ui.Icon(c, IconSVG("circle-check")).Size(16, 16).TextColor(p.Success).Shrink(0)
									MonoText(c, lane.summary, 14, p.Success, fonts).SingleLine()
								})
							} else {
								ui.Row(c).Gap(10).AlignItems(ui.Center).Children(func() {
									MonoText(c, lane.speed, 14, p.AccentText, fonts).SingleLine()
									MonoText(c, fmt.Sprintf("%3.0f%%", lane.prog*100), 14, p.Text, fonts).SingleLine()
								})
							}
						})

						// Lane progress bar (608 - 40 inner padding = 568px).
						ProgressBar(c, lane.prog, 566, 9, barColor, p)
					})
			}
		})

		// Bottom batch completion summary bar (transitions cleanly at CueWorkerComplete = local 93).
		summaryBg := p.Background.Mix(p.Success.Mix(p.Background, 0.88), allDoneT)
		summaryBorder := p.Border.Mix(p.Success.Alpha(0.55), allDoneT)

		ui.Row(c).
			Margin(16, 0, 0, 0).
			FillWidth().
			Height(48).
			Padding(0, 16).
			Radius(theme.CardRadius).
			Background(summaryBg).
			Border(theme.BorderWidth, summaryBorder).
			AlignItems(ui.Center).
			Children(func() {
				if allDoneT > 0.5 {
					ui.Icon(c, IconSVG("circle-check")).Size(18, 18).TextColor(p.Success).Margin(0, 10, 0, 0).Shrink(0)
					ui.Text(c, "本轮 3 个并发任务已完成").
						Font(fonts.Sans).
						FontSize(16).
						FontWeight(600).
						TextColor(p.Success).
						SingleLine()
					ui.Spacer(c)
					MonoText(c, "节省 331.2 MB (72.6%)", 15, p.Success, fonts).SingleLine()
				} else {
					ui.Box(c).Size(8, 8).Radius(4).Background(p.AccentText).Margin(0, 10, 0, 0).Shrink(0)
					ui.Text(c, "3 条 FFmpeg 编码流水线同步执行中").
						Font(fonts.Sans).
						FontSize(16).
						FontWeight(500).
						TextColor(p.TextMuted).
						SingleLine()
					ui.Spacer(c)
					MonoText(c, fmt.Sprintf("总进度 %.0f%%", avgPct), 15, p.Text, fonts).SingleLine()
				}
			})
	}).Fill()
}

func renderScene3GPUCard(
	c *ui.Context,
	f render.Frame,
	p theme.Tokens,
	fonts FontStacks,
	gpuLockT float32,
) {
	type gpuRow struct {
		vendor  string
		encoder string
		desc    string
		active  bool
		start   float64
	}

	rows := []gpuRow{
		{
			vendor:  "NVIDIA",
			encoder: "h264_nvenc",
			desc:    "支持 H.264 / HEVC / AV1 硬编探测",
			active:  true,
			start:   26,
		},
		{
			vendor:  "Intel",
			encoder: "h264_qsv",
			desc:    "Quick Sync Video 核显与独显编码",
			active:  false,
			start:   32,
		},
		{
			vendor:  "AMD",
			encoder: "h264_amf",
			desc:    "Advanced Media Framework 硬件编码",
			active:  false,
			start:   38,
		},
	}

	SurfaceCard(c, p, 22, 24, func() {
		// Card header.
		ui.Row(c).FillWidth().AlignItems(ui.Center).Children(func() {
			ui.Icon(c, IconSVG("shield-check")).Size(20, 20).TextColor(p.AccentText).Margin(0, 10, 0, 0).Shrink(0)
			ui.Text(c, "GPU 硬件加速探测").
				Font(fonts.Sans).
				FontSize(21).
				FontWeight(600).
				TextColor(p.Text).
				SingleLine()
			ui.Spacer(c)
			statusLabel := "探测中…"
			statusBg := p.SurfaceHover
			statusFg := p.TextMuted
			if gpuLockT > 0.5 {
				statusLabel = "已启用"
				statusBg = p.Accent
				statusFg = p.OnAccent
			}
			ui.Box(c).
				Padding(4, 10).
				Radius(theme.ControlRadius).
				Background(statusBg).
				Children(func() {
					ui.Text(c, statusLabel).
						Font(fonts.Sans).
						FontSize(13).
						FontWeight(600).
						TextColor(statusFg).
						SingleLine()
				})
		})

		Hairline(c, p.Border, 412, 1).Margin(16, 0, 18, 0)

		// Three GPU vendor encoder rows.
		ui.Column(c).FillWidth().Gap(14).Grow(1).Children(func() {
			for _, r := range rows {
				op, dy := EnterOffset(f, r.start, 16, 10)
				bg := p.Background
				bd := p.Border
				bdW := theme.BorderWidth
				encCol := p.Text
				tagLabel := "支持自动探测"
				tagBg := p.SurfaceHover
				tagFg := p.TextMuted

				if r.active {
					bg = p.Background.Mix(p.Accent.Mix(p.Background, 0.78), gpuLockT)
					bd = p.Border.Mix(p.Accent, gpuLockT)
					if gpuLockT > 0.3 {
						bdW = 1.5
					}
					encCol = p.Text.Mix(p.AccentText, gpuLockT*0.45)
					if gpuLockT > 0.5 {
						tagLabel = "已探测并启用"
						tagBg = p.Accent
						tagFg = p.OnAccent
					} else {
						tagLabel = "探测中…"
					}
				}

				ui.Column(c).
					FillWidth().
					Padding(14, 16).
					Gap(6).
					Radius(theme.CardRadius).
					Background(bg).
					Border(bdW, bd).
					Opacity(op).
					Translate(0, dy).
					Children(func() {
						ui.Row(c).FillWidth().AlignItems(ui.Center).Children(func() {
							ui.Text(c, r.vendor).
								Font(fonts.Sans).
								FontSize(17).
								FontWeight(600).
								TextColor(p.Text).
								Margin(0, 10, 0, 0).
								SingleLine()
							MonoText(c, r.encoder, 15, encCol, fonts).SingleLine()
							ui.Spacer(c)
							ui.Box(c).
								Padding(3, 9).
								Radius(theme.ControlRadius).
								Background(tagBg).
								Shrink(0).
								Children(func() {
									ui.Text(c, tagLabel).
										Font(fonts.Sans).
										FontSize(12).
										FontWeight(600).
										TextColor(tagFg).
										SingleLine()
								})
						})
						ui.Text(c, r.desc).
							Font(fonts.Sans).
							FontSize(14).
							TextColor(p.TextMuted).
							SingleLine()
					})
			}
		})

		// CPU fallback note box at bottom of Stage 3.
		ui.Row(c).
			Margin(16, 0, 0, 0).
			FillWidth().
			Padding(13, 15).
			Gap(10).
			Radius(theme.CardRadius).
			Background(p.Background).
			Border(theme.BorderWidth, p.Border).
			AlignItems(ui.Start).
			Children(func() {
				ui.Icon(c, IconSVG("info")).Size(17, 17).TextColor(p.Ornament).Margin(2, 0, 0, 0).Shrink(0)
				BodyText(c, "未检测到硬编时自动回退 CPU 编码 (libx264 / libx265)，结果本地缓存。", 14, p.TextMuted, fonts)
			})
	}).Fill()
}

// renderPipelineConnector draws a restrained architectural connector between pipeline stages.
func renderPipelineConnector(c *ui.Context, p theme.Tokens, pulsePos, allDoneT, opacity float32) {
	lineCol := p.Border.Mix(p.Success.Alpha(0.65), allDoneT)
	dotCol := p.AccentText.Mix(p.Success, allDoneT)

	ui.Box(c).
		Width(60).
		Height(40).
		Shrink(0).
		Opacity(opacity).
		Children(func() {
			// Base horizontal hairline.
			ui.Box(c).Absolute().Top(19).Left(6).Right(14).Height(2).Background(lineCol)

			// Moving pulse tick while active.
			if allDoneT < 0.95 {
				dotX := 8 + pulsePos*30
				ui.Box(c).
					Absolute().
					Top(17).
					Left(dotX).
					Size(6, 6).
					Radius(3).
					Background(dotCol).
					Opacity(1 - allDoneT)
			}

			// End chevron pointer.
			ui.Icon(c, IconSVG("chevron-right")).
				Absolute().Top(10).Right(2).
				Size(20, 20).
				TextColor(dotCol)
		})
}
