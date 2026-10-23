package gui

import (
	"github.com/egoist/mygo/ui"

	"github.com/ELY-Open-Source/MujicaUI/core"
	"github.com/ELY-Open-Source/MujicaUI/feedback"
	"github.com/ELY-Open-Source/MujicaUI/icons"
	"github.com/ELY-Open-Source/MujicaUI/layout"
	"github.com/ELY-Open-Source/MujicaUI/navigation"
	"github.com/ELY-Open-Source/MujicaUI/theme"
)

// toastKind tones a toast message.
type toastKind int

const (
	toastInfo toastKind = iota
	toastSuccess
	toastError
)

// setToast shows a message in the corner.
func (a *App) setToast(msg string, kind toastKind) {
	a.toast = msg
	a.toastKind = kind
}

// View builds the window's interface for one frame.
func (a *App) View(c *ui.Context) {
	core.Use(c, core.Settings{Mode: core.System, Locale: a.uiLocale()})

	if a.toast != "" && a.win != nil {
		sev := core.SeverityInfo
		switch a.toastKind {
		case toastSuccess:
			sev = core.SeveritySuccess
		case toastError:
			sev = core.SeverityError
		}
		feedback.ShowToast(c, feedback.ToastMessage{ID: a.toast, Severity: sev, Message: a.toast})
		a.toast = ""
	}

	ui.Column(c).Fill().Children(func() {
		a.titleBar(c)
		// Grow, not Fill: a 100% height here starts below the title bar and
		// would run that far past the window's bottom edge.
		ui.Box(c).FillWidth().Grow(1).MinHeight(0).Padding(16, 20).Children(func() {
			switch {
			case a.ffmpegError != "":
				a.setupView(c)
			case a.page == 1:
				a.settingsView(c)
			default:
				a.workspaceView(c)
			}
		})
	})
	feedback.Toaster(c)
}

// titleBar draws the brand, the navigation and the version badge.
func (a *App) titleBar(c *ui.Context) {
	layout.TitleBar(c, "Videopress", layout.TitleBarOptions{
		Leading: func() {
			ui.Icon(c, ic("play")).Size(16, 16).TextColor(core.Tokens(c).AccentText)
		},
		Trailing: func() {
			selected := a.page
			seg := []navigation.Segment{
				{Label: a.t("nav.workspace"), Icon: ic("layout-grid")},
				{Label: a.t("nav.settings"), Icon: ic("settings")},
			}
			navigation.SegmentedControl(c, &selected, seg, navigation.SegmentedControlOptions{Label: a.t("nav.workspace")})
			if selected != a.page && !a.isCompressing {
				a.page = selected
			}
		},
	})
}

// sectionLabel is a small capitalised caption above a control group.
func (a *App) sectionLabel(c *ui.Context, key string) ui.Element {
	return ui.Text(c, a.t(key)).FontSize(core.FontSize(c, theme.CaptionSize)).
		FontWeight(600).TextColor(core.Tokens(c).TextMuted)
}

// h3 is a serif group heading inside a panel.
func h3(c *ui.Context, s string) ui.Element {
	return ui.Text(c, s).Font(theme.SerifFont).FontSize(core.FontSize(c, theme.H3Size)).FontWeight(600)
}

// ic resolves a bundled icon, falling back to a neutral glyph so a missing name
// never takes the window down.
func ic(name string) *ui.SVG {
	if svg, ok := icons.Get(name); ok {
		return svg
	}
	svg, _ := icons.Get("circle")
	return svg
}
