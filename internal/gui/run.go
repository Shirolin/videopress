package gui

import (
	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"

	"github.com/ELY-Open-Source/MujicaUI/core"

	"videopress/internal/cli"
	"videopress/internal/ffmpeg"

	// Register MyGO's own Simplified Chinese texts for built-in controls.
	_ "github.com/egoist/mygo/locales/zh-Hans"
)

// Run opens the application window and blocks until it quits.
func Run(app *App) error {
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

		win := mygo.NewWindow(mygo.WindowOptions{
			Title:         "Videopress",
			Width:         900,
			Height:        820,
			MinWidth:      760,
			MinHeight:     600,
			StateKey:      "main",
			TitleBarStyle: mygo.TitleBarHidden,
			Content:       ui.View(app.View),
		})
		app.Attach(win)
	})
	return mygo.App.Run()
}

// uiLocale maps the app's language to MujicaUI's.
func (a *App) uiLocale() core.Locale {
	if a.locale == "zh-Hans" {
		return core.ZhCN
	}
	return core.EnUS
}
