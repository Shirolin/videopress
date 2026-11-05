package gui

import (
	"testing"
	"time"

	"github.com/ELY-Open-Source/MujicaUI/core"
	"github.com/ELY-Open-Source/MujicaUI/theme"
	"github.com/egoist/mygo/ui"

	"videopress/internal/engine"
	"videopress/internal/gif"
)

// newTestApp builds an app with no persisted preferences and no window.
func newTestApp(files ...string) *App {
	a := &App{
		pageScroll:     &ui.ScrollState{},
		settingsScroll: &ui.ScrollState{},
		settingsPage:   "general",
		gifTier:        gif.DefaultTier,
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
	for _, page := range []string{"general", "integration", "language", "debug"} {
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
