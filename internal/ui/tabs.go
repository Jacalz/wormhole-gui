package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"github.com/Jacalz/wormhole-gui/internal/transport"
)

// Create will stitch together all ui components
func Create(app fyne.App, window fyne.Window) *container.AppTabs {
	bridge := transport.NewClient() // To make sure that it is configured correctly
	appSettings := &AppSettings{Theme: checkTheme(app.Preferences().StringWithFallback("Theme", "Adaptive (requires restart)"), app)}

	return &container.AppTabs{Items: []*container.TabItem{
		newSend(app, window, bridge, appSettings).tabItem(),
		newRecv(app, window, bridge, appSettings).tabItem(),
		newSettings(app, window, bridge, appSettings).tabItem(),
		newAbout().tabItem(),
	}}
}
