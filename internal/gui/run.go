package gui

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"

	"github.com/ELY-Open-Source/MujicaUI/core"
	"github.com/ELY-Open-Source/MujicaUI/theme"

	"videopress/internal/cli"
	"videopress/internal/ffmpeg"
	"videopress/internal/sendto"
	"videopress/resources"

	// Register MyGO's own Simplified Chinese texts for built-in controls.
	_ "github.com/egoist/mygo/locales/zh-Hans"
)

// Run opens the application window and blocks until it quits.
func Run(app *App) error {
	// 右键菜单按选中项逐个启动进程；只保留首个实例，后续启动把文件转交给
	// 它再退出，避免每个文件各弹一个窗口。
	if !mygo.App.RequestSingleInstanceLock() {
		return nil
	}
	mygo.App.OnSecondInstance(app.enqueueForwarded)

	mygo.App.SetLanguages("en", "zh-Hans")
	mygo.App.SetLocale(app.prefs.Language)

	mygo.App.WhenReady(func() {
		app.locale = mygo.App.Locale()
		if v := mygo.App.Version(); v != "" && v != "0.0.0" {
			cli.Version = v
		}
		app.showAdvanced = app.prefs.ShowAdvanced
		ffmpeg.EnableDebugLog = app.prefs.EnableDebugLog
		app.detectFFmpeg()
		app.detectGPU()

		// 后台对已安装的系统右键菜单和快捷方式执行路径自愈
		go func() {
			if err := sendto.RefreshIfStale(app.executablePath, app.lang()); err != nil && app.prefs.EnableDebugLog {
				if dir, dirErr := os.UserCacheDir(); dirErr == nil {
					logFile := filepath.Join(dir, "videopress_debug.log")
					msg := fmt.Sprintf("[%s] 系统快捷集成自愈失败: %v\n", time.Now().Format("2006-01-02 15:04:05"), err)
					f, fErr := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
					if fErr == nil {
						_, _ = f.WriteString(msg)
						_ = f.Close()
					}
				}
			}
		}()

		win := mygo.NewWindow(mygo.WindowOptions{
			Title:         "Videopress",
			Width:         900,
			Height:        820,
			MinWidth:      760,
			MinHeight:     600,
			StateKey:      "main",
			TitleBarStyle: mygo.TitleBarHidden,
			// The window shows before its first frame paints; without this it
			// flashes white. These are MujicaUI's own page colors.
			BackgroundColor: windowBackground(app.themeMode()),
			Content:         ui.View(app.View),
		})
		_ = win.SetIcon(resources.IconPNG)
		app.Attach(win)
	})
	return mygo.App.Run()
}

// windowBackground is the page color shown before the first frame, pinned to
// the chosen appearance so a manual theme never flashes the system one.
func windowBackground(mode core.Mode) string {
	light := colorHex(theme.Light().Background)
	dark := colorHex(theme.Dark().Background)
	switch mode {
	case core.Light:
		return light
	case core.Dark:
		return dark
	}
	return "light-dark(" + light + ", " + dark + ")"
}

func colorHex(c ui.Color) string {
	return fmt.Sprintf("#%02X%02X%02X", c.R, c.G, c.B)
}

// enqueueForwarded adds the files handed over by a later launch, then brings
// the window forward. MyGO calls it on the main thread.
func (a *App) enqueueForwarded(args []string, workingDir string) {
	files := make([]string, 0, len(args))
	for _, arg := range args {
		if !filepath.IsAbs(arg) {
			arg = filepath.Join(workingDir, arg)
		}
		files = append(files, arg)
	}
	a.update(func() { a.addFiles(files) })
	if a.win != nil {
		a.win.Show()
		a.win.Focus()
	}
}

// uiLocale maps the app's language to MujicaUI's.
func (a *App) uiLocale() core.Locale {
	if a.locale == "zh-Hans" {
		return core.ZhCN
	}
	return core.EnUS
}
