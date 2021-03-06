package ui

import (
	"path/filepath"
	"strconv"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/Jacalz/wormhole-gui/internal/transport"
)

var (
	themes       = []string{"Adaptive (requires restart)", "Light", "Dark"}
	onOffOptions = []string{"On", "Off"}
)

// AppSettings contains settings specific to the application
type AppSettings struct {
	// Theme holds the current theme
	Theme string
}

type settings struct {
	themeSelect *widget.Select

	downloadPathButton *widget.Button
	overwriteFiles     *widget.RadioGroup
	notificationRadio  *widget.RadioGroup

	componentSlider     *widget.Slider
	componentLabel      *widget.Label
	appID               *widget.Entry
	rendezvousURL       *widget.Entry
	transitRelayAddress *widget.Entry

	client      *transport.Client
	appSettings *AppSettings
	window      fyne.Window
	app         fyne.App
}

func newSettings(a fyne.App, w fyne.Window, c *transport.Client, as *AppSettings) *settings {
	return &settings{app: a, window: w, client: c, appSettings: as}
}

func (s *settings) onThemeChanged(selected string) {
	s.app.Preferences().SetString("Theme", checkTheme(selected, s.app))
}

func (s *settings) onDownloadsPathChanged() {
	dialog.ShowFolderOpen(func(folder fyne.ListableURI, err error) {
		if err != nil {
			fyne.LogError("Error on selecting folder", err)
			dialog.ShowError(err, s.window)
			return
		} else if folder == nil {
			return
		}

		s.app.Preferences().SetString("DownloadPath", folder.String()[7:])
		s.client.DownloadPath = folder.String()[7:]
		s.downloadPathButton.SetText(folder.Name())
	}, s.window)
}

func (s *settings) onOverwriteFilesChanged(selected string) {
	s.client.OverwriteExisting = selected == "On"
	s.app.Preferences().SetString("OverwriteFiles", selected)
}

func (s *settings) onNotificationsChanged(selected string) {
	s.client.Notifications = selected == "On"
	s.app.Preferences().SetString("Notifications", selected)
}

func (s *settings) onComponentsChange(value float64) {
	s.client.PassPhraseComponentLength = int(value)
	s.app.Preferences().SetFloat("ComponentLength", value)
	s.componentLabel.SetText(strconv.Itoa(int(value)))
}

func (s *settings) onAppIDChanged(appID string) {
	s.client.AppID = appID
	s.app.Preferences().SetString("AppID", appID)
}

func (s *settings) onRendezvousURLChange(url string) {
	s.client.RendezvousURL = url
	s.app.Preferences().SetString("RendezvousURL", url)
}

func (s *settings) onTransitAdressChange(address string) {
	s.client.TransitRelayAddress = address
	s.app.Preferences().SetString("TransitRelayAddress", address)
}

func (s *settings) buildUI() *container.Scroll {
	s.themeSelect = &widget.Select{Options: themes, OnChanged: s.onThemeChanged, Selected: s.appSettings.Theme}

	s.client.DownloadPath = s.app.Preferences().StringWithFallback("DownloadPath", transport.UserDownloadsFolder())
	s.downloadPathButton = &widget.Button{Icon: theme.FolderOpenIcon(), OnTapped: s.onDownloadsPathChanged, Text: filepath.Base(s.client.DownloadPath)}

	s.overwriteFiles = &widget.RadioGroup{Options: onOffOptions, Horizontal: true, Required: true, OnChanged: s.onOverwriteFilesChanged}
	s.overwriteFiles.SetSelected(s.app.Preferences().StringWithFallback("OverwriteFiles", "Off"))

	s.notificationRadio = &widget.RadioGroup{Options: onOffOptions, Horizontal: true, Required: true, OnChanged: s.onNotificationsChanged}
	s.notificationRadio.SetSelected(s.app.Preferences().StringWithFallback("Notifications", onOffOptions[1]))

	s.componentSlider, s.componentLabel = &widget.Slider{Min: 2.0, Max: 6.0, Step: 1, OnChanged: s.onComponentsChange}, &widget.Label{}
	s.componentSlider.SetValue(s.app.Preferences().FloatWithFallback("ComponentLength", 2))

	s.appID = &widget.Entry{PlaceHolder: "lothar.com/wormhole/text-or-file-xfer", OnChanged: s.onAppIDChanged}
	s.appID.SetText(s.app.Preferences().String("AppID"))

	s.rendezvousURL = &widget.Entry{PlaceHolder: "ws://relay.magic-wormhole.io:4000/v1", OnChanged: s.onRendezvousURLChange}
	s.rendezvousURL.SetText(s.app.Preferences().String("RendezvousURL"))

	s.transitRelayAddress = &widget.Entry{PlaceHolder: "transit.magic-wormhole.io:4001", OnChanged: s.onTransitAdressChange}
	s.transitRelayAddress.SetText(s.app.Preferences().String("TransitRelayAddress"))

	interfaceContainer := container.NewGridWithColumns(2,
		newBoldLabel("Application Theme"), s.themeSelect,
	)

	dataContainer := container.NewGridWithColumns(2,
		newBoldLabel("Downloads Path"), s.downloadPathButton,
		newBoldLabel("Overwrite Files"), s.overwriteFiles,
		newBoldLabel("Notifications"), s.notificationRadio,
	)

	wormholeContainer := container.NewVBox(
		container.NewGridWithColumns(2, newBoldLabel("Passphrase Length"), container.NewBorder(nil, nil, nil, s.componentLabel, s.componentSlider)),
		&widget.Accordion{Items: []*widget.AccordionItem{
			{Title: "Advanced", Detail: container.NewGridWithColumns(2,
				newBoldLabel("AppID"), s.appID,
				newBoldLabel("Rendezvous URL"), s.rendezvousURL,
				newBoldLabel("Transit Relay Address"), s.transitRelayAddress,
			)},
		}},
	)

	return container.NewScroll(container.NewVBox(
		&widget.Card{Title: "User Interface", Content: interfaceContainer},
		&widget.Card{Title: "Data Handling", Content: dataContainer},
		&widget.Card{Title: "Wormhole Options", Content: wormholeContainer},
	))
}

func (s *settings) tabItem() *container.TabItem {
	return &container.TabItem{Text: "Settings", Icon: theme.SettingsIcon(), Content: s.buildUI()}
}

func checkTheme(themec string, a fyne.App) string {
	switch themec {
	case "Light":
		a.Settings().SetTheme(theme.LightTheme())
	case "Dark":
		a.Settings().SetTheme(theme.DarkTheme())
	}

	return themec
}

func newBoldLabel(text string) *widget.Label {
	return &widget.Label{Text: text, TextStyle: fyne.TextStyle{Bold: true}}
}
