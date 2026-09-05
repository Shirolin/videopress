package gif

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestTierByName(t *testing.T) {
	for _, name := range []string{"smooth", "balanced", "hd"} {
		tier, err := TierByName(name)
		if err != nil {
			t.Fatalf("TierByName(%s) error: %v", name, err)
		}
		if tier.Name != name {
			t.Errorf("TierByName(%s) = %s, want %s", name, tier.Name, name)
		}
	}
	if _, err := TierByName("bogus"); err == nil {
		t.Error("TierByName(bogus) should error")
	}
}

func TestTierSizeLimits(t *testing.T) {
	smooth, _ := TierByName("smooth")
	balanced, _ := TierByName("balanced")
	hd, _ := TierByName("hd")

	if smooth.MaxSizeMB != 1 || smooth.MaxWidth != 240 {
		t.Errorf("smooth limits wrong: %+v", smooth)
	}
	if balanced.MaxSizeMB != 2 || balanced.MaxWidth != 360 {
		t.Errorf("balanced limits wrong: %+v", balanced)
	}
	if hd.MaxSizeMB != 5 || hd.MaxWidth != 480 {
		t.Errorf("hd limits wrong: %+v", hd)
	}
	if smooth.FPS != 15 || balanced.FPS != 15 || hd.FPS != 15 {
		t.Error("all tiers should be 15 fps")
	}
}

func TestGIFArgs(t *testing.T) {
	tier, _ := TierByName("balanced")
	args := BuildGIFArgs("in.mp4", "out.gif", tier)
	s := strings.Join(args, " ")
	for _, want := range []string{"-t", "00:00:05", "fps=15", "palettegen=128", "paletteuse"} {
		if !strings.Contains(s, want) {
			t.Errorf("GIF args missing %q: %s", want, s)
		}
	}
}

func TestAPNGArgs(t *testing.T) {
	tier, _ := TierByName("balanced")
	args := BuildAPNGArgs("in.mp4", "out.apng", tier, 8)
	s := strings.Join(args, " ")
	for _, want := range []string{"-t", "00:00:05", "fps=8", "-plays", "0"} {
		if !strings.Contains(s, want) {
			t.Errorf("APNG args missing %q: %s", want, s)
		}
	}
}

func TestWebPArgs(t *testing.T) {
	tier, _ := TierByName("balanced")
	args := BuildWebPArgs("in.mp4", "out.webp", tier)
	s := strings.Join(args, " ")
	for _, want := range []string{"-t", "00:00:05", "fps=15", "libwebp_anim", "-lossless", "0"} {
		if !strings.Contains(s, want) {
			t.Errorf("WebP args missing %q: %s", want, s)
		}
	}
}

func TestOutputPathFor(t *testing.T) {
	p := outputPathFor("in.mp4", "balanced", FormatGIF, "")
	want := filepath.Join("gif_export", "in.balanced.gif")
	if !strings.HasSuffix(p, want) {
		t.Errorf("output path wrong: %s", p)
	}
	if !strings.HasSuffix(p, ".gif") {
		t.Errorf("output path should end .gif: %s", p)
	}
}

func TestFormatExt(t *testing.T) {
	cases := map[Format]string{
		FormatGIF:  ".gif",
		FormatAPNG: ".apng",
		FormatWebP: ".webp",
	}
	for f, want := range cases {
		if f.Ext() != want {
			t.Errorf("Ext(%s) = %s, want %s", f, f.Ext(), want)
		}
	}
}

func TestAdaptiveAPNGFPS(t *testing.T) {
	const mib = 1024 * 1024
	cases := []struct {
		name     string
		budget   int64
		probe    int64
		duration float64
		max      int
		want     int
	}{
		{"smooth预算压到3fps", 1 * mib, 1 * mib, 5, 15, 3},
		{"balanced预算给到6fps", 2 * mib, 1 * mib, 5, 15, 6},
		{"预算充足钳到上限", 100 * mib, 1 * mib, 5, 15, 15},
		{"极小预算保底2fps", 100 * 1024, 1 * mib, 5, 15, 2},
		{"探测为空回退上限", 1 * mib, 0, 5, 15, 15},
		{"时长非法回退上限", 1 * mib, 1 * mib, 0, 15, 15},
		{"上限非法回1", 1 * mib, 1 * mib, 5, 0, 1},
	}
	for _, c := range cases {
		if got := adaptiveAPNGFPS(c.budget, c.probe, c.duration, c.max); got != c.want {
			t.Errorf("%s: adaptiveAPNGFPS = %d, want %d", c.name, got, c.want)
		}
	}
}

func TestTierDurationSec(t *testing.T) {
	tier, _ := TierByName("balanced")
	if got := tierDurationSec(tier); got != 5 {
		t.Errorf("tierDurationSec(balanced) = %v, want 5", got)
	}
	bad := Tier{MaxDuration: "bogus"}
	if got := tierDurationSec(bad); got != 5 {
		t.Errorf("tierDurationSec(bogus) = %v, want fallback 5", got)
	}
}

func TestRunRejectsUnknownFormat(t *testing.T) {
	eng := New(Dependencies{})
	_, err := eng.Run(context.Background(), ExportRequest{
		Files:   []string{"clip.mp4"},
		Tier:    "balanced",
		Formats: []Format{Format("bogus")},
		Force:   true,
	})
	if err == nil || !strings.Contains(err.Error(), "unknown format") {
		t.Fatalf("Run(bogus format) should error with unknown format, got %v", err)
	}
}

func TestRunAPNGFallsBackToFullFPSWithMock(t *testing.T) {
	var calls [][]string
	eng := New(Dependencies{
		ResolveBinary: func(dir string) (string, error) { return "ffmpeg", nil },
		RunCommand: func(ctx context.Context, name string, args []string) error {
			calls = append(calls, args)
			return nil
		},
		MkdirAll:   func(path string, perm os.FileMode) error { return nil },
	})
	results, err := eng.Run(context.Background(), ExportRequest{
		Files:   []string{"clip.mp4"},
		Tier:    "balanced",
		Formats: []Format{FormatAPNG},
		Force:   true,
	})
	if err != nil {
		t.Fatalf("Run(APNG mock) error: %v", err)
	}
	if len(results) != 1 || results[0].Status != "success" {
		t.Fatalf("Run(APNG mock) unexpected results: %+v", results)
	}
	// mock 不落盘探测产物，回退全帧率：探测 1 次 + 编码 1 次。
	if len(calls) != 2 {
		t.Fatalf("expected probe + encode calls, got %d", len(calls))
	}
	joined := strings.Join(calls[1], " ")
	for _, want := range []string{"fps=15", "-t", "00:00:05"} {
		if !strings.Contains(joined, want) {
			t.Errorf("encode args missing %q: %s", want, joined)
		}
	}
}

func TestRunDedupesRepeatedFormats(t *testing.T) {
	var calls [][]string
	eng := New(Dependencies{
		ResolveBinary: func(dir string) (string, error) { return "ffmpeg", nil },
		RunCommand: func(ctx context.Context, name string, args []string) error {
			calls = append(calls, args)
			return nil
		},
		MkdirAll:   func(path string, perm os.FileMode) error { return nil },
		PathExists: func(path string) bool { return false },
	})
	results, err := eng.Run(context.Background(), ExportRequest{
		Files:   []string{"clip.mp4"},
		Tier:    "balanced",
		Formats: []Format{FormatGIF, FormatGIF, FormatWebP},
		Force:   true,
	})
	if err != nil {
		t.Fatalf("Run error: %v", err)
	}
	if len(results) != 2 {
		t.Fatalf("expected deduped 2 results, got %d", len(results))
	}
	if len(calls) != 2 {
		t.Fatalf("expected deduped 2 ffmpeg calls, got %d", len(calls))
	}
}

func TestRunMarksOverBudget(t *testing.T) {
	dir := t.TempDir()
	eng := New(Dependencies{
		ResolveBinary: func(dir string) (string, error) { return "ffmpeg", nil },
		RunCommand: func(ctx context.Context, name string, args []string) error {
			out := args[len(args)-1]
			return os.WriteFile(out, make([]byte, 3<<20), 0644)
		},
		MkdirAll:   os.MkdirAll,
		PathExists: func(path string) bool { return false },
	})
	results, err := eng.Run(context.Background(), ExportRequest{
		Files:     []string{"clip.mp4"},
		Tier:      "balanced",
		Formats:   []Format{FormatGIF},
		OutputDir: dir,
		Force:     true,
	})
	if err != nil {
		t.Fatalf("Run error: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	r := results[0]
	if r.Status != "success" {
		t.Fatalf("expected success, got %+v", r)
	}
	if r.Size != 3<<20 {
		t.Errorf("expected size 3MiB, got %d", r.Size)
	}
	if !r.OverBudget {
		t.Error("3MiB balanced output should be marked OverBudget")
	}
}

func TestSkippedMarksOverBudget(t *testing.T) {
	dir := t.TempDir()
	// OutputDir 直接作为输出目录（不拼 gif_export），预置同名产物。
	if err := os.WriteFile(filepath.Join(dir, "clip.balanced.gif"), make([]byte, 3<<20), 0644); err != nil {
		t.Fatal(err)
	}
	eng := New(Dependencies{
		ResolveBinary: func(dir string) (string, error) { return "ffmpeg", nil },
		RunCommand: func(ctx context.Context, name string, args []string) error {
			t.Error("skipped output must not invoke ffmpeg")
			return nil
		},
		MkdirAll:   os.MkdirAll,
		PathExists: func(path string) bool { return true },
	})
	results, err := eng.Run(context.Background(), ExportRequest{
		Files:     []string{"clip.mp4"},
		Tier:      "balanced",
		Formats:   []Format{FormatGIF},
		OutputDir: dir,
		Force:     false,
	})
	if err != nil {
		t.Fatalf("Run error: %v", err)
	}
	if len(results) != 1 || results[0].Status != "skipped" {
		t.Fatalf("expected 1 skipped result, got %+v", results)
	}
	if !results[0].OverBudget {
		t.Error("3MiB skipped output should be marked OverBudget")
	}
}