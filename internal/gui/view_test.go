package gui

import (
	"strings"
	"testing"
	"time"

	"github.com/ELY-Open-Source/MujicaUI/core"
	"github.com/ELY-Open-Source/MujicaUI/theme"
	"github.com/egoist/mygo/ui"

	"videopress/internal/engine"
)

// newTestApp builds an app with no persisted preferences and no window.
func newTestApp(files ...string) *App {
	a := &App{
		pageScroll:     &ui.ScrollState{},
		settingsScroll: &ui.ScrollState{},
		settingsPage:   "general",
		gifFormats:     map[string]bool{"gif": true, "apng": true, "webp": true},
		integration:    map[string]bool{},
		prefs:          defaultPrefs(),
	}
	a.addFiles(files)
	return a
}

// TestWorkspaceFlow walks a file through the queue: added, compressing,
// finished, and counted in the summary bar.
func TestWorkspaceFlow(t *testing.T) {
	a := newTestApp(`D:\clips\demo.mp4`, `D:\clips\notes.txt`)
	if len(a.queue) != 1 || a.queue[0].Name != "demo.mp4" {
		t.Fatalf("queue = %+v, want only demo.mp4", a.queue)
	}

	a.locale = "zh-Hans"
	tester := ui.NewTester(a.View, 900, 820)
	if !tester.HasText("demo.mp4") {
		t.Fatalf("queue row missing; texts = %v", tester.Texts())
	}

	a.applyProgress(engine.ProgressEvent{Path: `D:\clips\demo.mp4`, Percent: 42})
	if a.queue[0].Status != "compressing" || a.queue[0].Percent != 42 {
		t.Fatalf("after progress: %+v", a.queue[0])
	}

	a.applyReports([]engine.JobReport{{
		InputName: "demo.mp4", InputPath: `D:\clips\demo.mp4`,
		Status: "success", SourceSize: 10 << 20, TargetSize: 4 << 20,
		Duration: 3 * time.Second,
	}})
	item := a.queue[0]
	if item.Status != "success" || item.Percent != 100 || item.Size != 10<<20 || item.TargetSize != 4<<20 {
		t.Fatalf("after report: %+v", item)
	}

	_, done, saved, original := a.stats()
	if done != 1 || saved != 6<<20 || original != 10<<20 {
		t.Fatalf("stats = done %d saved %d original %d", done, saved, original)
	}

	tester.Frame()
	if !tester.HasText("4.0 MB") || !tester.HasText("60.0%") {
		t.Fatalf("result row missing; texts = %v", tester.Texts())
	}
}

// TestRecompressedWarningFollowsLanguage: the warning under a row that looks
// already compressed is looked up per frame, so switching language rewrites it.
func TestRecompressedWarningFollowsLanguage(t *testing.T) {
	a := newTestApp(`D:\clips\demo.compressed.mp4`)
	if !a.queue[0].Warning || a.queue[0].Error != "" {
		t.Fatalf("expected a warning row with no stored error, got %+v", a.queue[0])
	}

	a.locale = "en"
	tester := ui.NewTester(a.View, 900, 820)
	if !tester.HasText("This video may already be compressed; compressing again will hurt quality.") {
		t.Fatalf("English warning missing; texts = %v", tester.Texts())
	}

	a.locale = "zh-Hans"
	tester.Frame()
	if !tester.HasText("该视频可能已经被压缩过，二次压缩会损害画质。") {
		t.Fatalf("Chinese warning missing; texts = %v", tester.Texts())
	}
}

// TestSettingsPageRenders guards the settings layout against a panic or an
// empty page.
func TestSettingsPageRenders(t *testing.T) {
	a := newTestApp()
	a.locale = "zh-Hans"
	a.page = 1
	for _, page := range []string{"general", "appearance", "integration", "language", "debug", "about"} {
		a.settingsPage = page
		tester := ui.NewTester(a.View, 900, 820)
		if len(tester.Texts()) == 0 {
			t.Fatalf("settings page %q rendered nothing", page)
		}
	}
}

// TestSettingsSwitchCentered verifies that the switch thumb is centered
// vertically and horizontally inside the track's 1px border in both states.
func TestSettingsSwitchCentered(t *testing.T) {
	on, off := true, false
	var onBox, offBox ui.Rect
	tt := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{Mode: core.Dark})
		ui.Column(c).Padding(16).Gap(12).AlignItems(ui.Start).Children(func() {
			onBox = settingsSwitch(c, &on, false).Bounds()
			offBox = settingsSwitch(c, &off, false).Bounds()
		})
	}, 120, 120)
	tt.SetPreferences(ui.Preferences{ReduceMotion: true})
	tt.Frame()

	img := tt.Image()
	k := theme.Dark()
	// Track sits vertically centered inside the switch row (height 36 -> track Y = box.Y + (36-20)/2 = box.Y + 8).
	checkThumb := func(name string, box ui.Rect, wantX float32, wantThumb ui.Color) {
		trackX := int(box.X)
		trackY := int(box.Y + (box.H-switchH)/2)
		midX := trackX + int(theme.BorderWidth+wantX+switchK/2)
		// Top gap (2px inside border: trackY+1 and trackY+2) and bottom gap (trackY+17 and trackY+18) must not be thumb color.
		topGap := img.RGBAAt(midX, trackY+2)
		botGap := img.RGBAAt(midX, trackY+int(switchH)-3)
		topThumb := img.RGBAAt(midX, trackY+3)
		botThumb := img.RGBAAt(midX, trackY+int(switchH)-4)
		if topGap.R == wantThumb.R && topGap.G == wantThumb.G && topGap.B == wantThumb.B {
			t.Fatalf("%s: top gap at y=%d painted with thumb color %v", name, trackY+2, topGap)
		}
		if botGap.R == wantThumb.R && botGap.G == wantThumb.G && botGap.B == wantThumb.B {
			t.Fatalf("%s: bottom gap at y=%d painted with thumb color %v", name, trackY+int(switchH)-3, botGap)
		}
		if topThumb.R != wantThumb.R || botThumb.R != wantThumb.R {
			t.Fatalf("%s: thumb vertical span mismatch: top=%v bot=%v want=%v", name, topThumb, botThumb, wantThumb)
		}
	}
	checkThumb("on", onBox, 18, k.OnAccent)
	checkThumb("off", offBox, 2, k.ControlBorder)
}

// TestGifDurationNote reads the probed source length and flags windows that
// run past it.
func TestGifDurationNote(t *testing.T) {
	a := newTestApp(`D:\clips\a.mp4`, `D:\clips\b.mp4`)
	a.ffmpegPath = "ffmpeg"
	a.locale = "zh-Hans"
	a.prefs.GifStart, a.prefs.GifDuration = 0, 5
	a.gifDurations = map[string]int64{`D:\clips\a.mp4`: 42_000}

	note, warn := a.gifDurationNote()
	if note != "源视频时长 0:42" || warn {
		t.Fatalf("note = %q warn=%v, want plain 0:42", note, warn)
	}

	a.gifDurations[`D:\clips\b.mp4`] = 90_000
	note, warn = a.gifDurationNote()
	if note != "源视频时长 0:42–1:30" || warn {
		t.Fatalf("note = %q warn=%v, want a range", note, warn)
	}

	a.prefs.GifStart, a.prefs.GifDuration = 40, 5
	if note, warn = a.gifDurationNote(); !warn || !strings.Contains(note, "结束点 45s 超出最短源视频长度 0:42") {
		t.Fatalf("overflow note = %q warn=%v", note, warn)
	}

	a.prefs.GifStart = 50
	if note, warn = a.gifDurationNote(); !warn || !strings.Contains(note, "起点 50s 已超过最短源视频长度 0:42") {
		t.Fatalf("start-beyond note = %q warn=%v", note, warn)
	}

	// An unreadable source is not fatal; it is called out beside the range.
	a.prefs.GifStart, a.prefs.GifDuration = 0, 5
	a.gifDurations[`D:\clips\c.mp4`] = 0
	a.addFiles([]string{`D:\clips\c.mp4`})
	if note, warn = a.gifDurationNote(); warn || !strings.Contains(note, "部分文件无法读取时长") {
		t.Fatalf("partial note = %q warn=%v", note, warn)
	}
	a.gifProbing = map[string]bool{`D:\clips\b.mp4`: true}
	if note, warn = a.gifDurationNote(); warn || !strings.Contains(note, "其余文件读取中") {
		t.Fatalf("pending note = %q warn=%v", note, warn)
	}

	a.gifProbing = map[string]bool{`D:\clips\a.mp4`: true}
	a.gifDurations = map[string]int64{}
	if note, warn = a.gifDurationNote(); note != "正在读取源视频时长…" || warn {
		t.Fatalf("probing note = %q warn=%v", note, warn)
	}

	a.queue = nil
	if note, warn = a.gifDurationNote(); note != "" || warn {
		t.Fatalf("empty-queue note = %q warn=%v, want none", note, warn)
	}
}

func TestGifCustomName(t *testing.T) {
	a := newTestApp()
	a.prefs.GifWidth, a.prefs.GifFPS, a.prefs.GifStart, a.prefs.GifDuration = 360, 15, 40, 5
	if got := a.gifCustomName(); got != "custom-360px-15fps-40-45s" {
		t.Fatalf("gifCustomName = %q", got)
	}
}

// TestThemeMode maps the saved appearance preference to a MujicaUI mode.
func TestThemeMode(t *testing.T) {
	a := newTestApp()
	cases := map[string]core.Mode{
		"":       core.System,
		"system": core.System,
		"light":  core.Light,
		"dark":   core.Dark,
	}
	for pref, want := range cases {
		a.prefs.Theme = pref
		if got := a.themeMode(); got != want {
			t.Fatalf("themeMode(%q) = %v, want %v", pref, got, want)
		}
	}
}

// TestCancelledReportReturnsRowToQueue: cancelling one file puts its row back
// in the waiting state instead of marking it failed.
func TestCancelledReportReturnsRowToQueue(t *testing.T) {
	a := newTestApp(`D:\clips\a.mp4`)
	a.queue[0].Status = "compressing"
	a.queue[0].Percent = 40

	a.applyReports([]engine.JobReport{{InputName: "a.mp4", InputPath: `D:\clips\a.mp4`, Status: "cancelled"}})

	item := a.queue[0]
	if item.Status != "waiting" || item.Percent != 0 || item.Error != "" || item.Size != 0 {
		t.Fatalf("after cancel: %+v", item)
	}
}

// TestCancelledProgressRequeuesRow: the engine's cancel event must not leave
// the row showing a failure while the rest of the batch keeps running.
func TestCancelledProgressRequeuesRow(t *testing.T) {
	a := newTestApp(`D:\clips\a.mp4`)
	a.queue[0].Status = "compressing"
	a.queue[0].Percent = 40
	a.queue[0].Speed, a.queue[0].ETA = 2.5, 30000

	a.applyProgress(engine.ProgressEvent{
		Path: `D:\clips\a.mp4`, Cancelled: true, Done: true, Error: "任务已取消",
	})

	item := a.queue[0]
	if item.Status != "waiting" || item.Percent != 0 || item.Speed != 0 || item.ETA != 0 || item.Error != "" {
		t.Fatalf("after cancel event: %+v", item)
	}
}

func TestTitleBarAppLogo(t *testing.T) {
	if appLogo == nil {
		t.Fatal("expected appLogo to be initialized from embedded resources")
	}
	w, h := appLogo.Size()
	if w <= 0 || h <= 0 {
		t.Fatalf("invalid appLogo size: %dx%d", w, h)
	}

	a := newTestApp()
	a.locale = "zh-Hans"
	tester := ui.NewTester(a.View, 300, 60)
	tester.SetPreferences(ui.Preferences{ReduceMotion: true})
	tester.Frame()
}

// TestPendingAndStats checks which rows a batch picks up and how the summary
// counts skipped rows.
func TestPendingAndStats(t *testing.T) {
	a := newTestApp(`D:\clips\a.mp4`, `D:\clips\b.mp4`, `D:\clips\c.mp4`, `D:\clips\d.mp4`)
	a.queue[0].Status = "success"
	a.queue[0].Size, a.queue[0].TargetSize = 100, 10
	a.queue[1].Status = "skipped"
	a.queue[2].Status = "failed"

	pending := a.pendingPaths()
	if len(pending) != 2 || pending[0] != `D:\clips\c.mp4` || pending[1] != `D:\clips\d.mp4` {
		t.Fatalf("pendingPaths = %v, want the failed and waiting rows", pending)
	}

	total, done, saved, original := a.stats()
	if total != 4 || done != 2 || saved != 90 || original != 100 {
		t.Fatalf("stats = total %d done %d saved %d original %d, want 4 2 90 100", total, done, saved, original)
	}
	if failed := a.failedCount(); failed != 1 {
		t.Fatalf("failedCount = %d, want 1", failed)
	}
	if skipped := a.skippedCount(); skipped != 1 {
		t.Fatalf("skippedCount = %d, want 1", skipped)
	}
	// success + skipped + failed are done, the last row is still waiting.
	if p := a.batchProgress(); p != 0.75 {
		t.Fatalf("batchProgress = %.2f, want 0.75", p)
	}

	// Rows added after the run started must not dilute its progress.
	a.compressBatch = map[string]bool{}
	a.compressBatchDone = map[string]bool{}
	for _, it := range a.queue {
		a.compressBatch[it.Path] = true
	}
	a.addFiles([]string{`D:\clips\e.mp4`})
	if p := a.batchProgress(); p != 0.75 {
		t.Fatalf("batchProgress after adding a row = %.2f, want 0.75", p)
	}
}

// TestBatchProgressCountsCancelledRows: a row stopped on its own leaves the
// batch as soon as its cancel event arrives, rather than pinning progress at 0
// until the whole run ends.
func TestBatchProgressCountsCancelledRows(t *testing.T) {
	a := newTestApp(`D:\clips\a.mp4`, `D:\clips\b.mp4`)
	key := `D:\clips\a.mp4`
	a.compressBatch = map[string]bool{key: true, `D:\clips\b.mp4`: true}
	a.compressBatchDone = map[string]bool{}
	a.queue[0].Status = "compressing"
	a.queue[1].Status = "compressing"

	a.applyProgress(engine.ProgressEvent{Path: key, Cancelled: true, Done: true, Error: "任务已取消"})

	if p := a.batchProgress(); p != 0.5 {
		t.Fatalf("batchProgress = %.2f, want 0.5 (cancelled row handled, other still at 0)", p)
	}
	if a.queue[0].Status != "waiting" {
		t.Fatalf("cancelled row status = %q, want waiting", a.queue[0].Status)
	}
}
