package desktop

import (
	"videopress/frontend"
	"videopress/internal/gui"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/windows"
)

// Run starts the Wails GUI application.
func Run(app *gui.App) error {
	return wails.Run(&options.App{
		Title:           "Videopress",
		Width:           850,
		Height:          620,
		MinWidth:        720,
		MinHeight:       520,
		Frameless:       true,
		CSSDragProperty: "--wails-draggable",
		CSSDragValue:    "drag",
		AssetServer: &assetserver.Options{
			Assets: frontend.Dist,
		},
		BackgroundColour: &options.RGBA{R: 22, G: 20, B: 16, A: 1},
		Windows: &windows.Options{
			Theme: windows.Dark,
		},
		OnStartup: app.Startup,
		Bind: []interface{}{
			app,
		},
	})
}
