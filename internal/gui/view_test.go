package gui

import (
	"testing"
	"time"

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
