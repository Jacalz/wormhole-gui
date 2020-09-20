package widgets

import (
	"time"

	"fyne.io/fyne"
	"fyne.io/fyne/layout"
	"fyne.io/fyne/theme"
	"fyne.io/fyne/widget"
)

// NewBoldLabel returns a new label with bold text.
func NewBoldLabel(text string) *widget.Label {
	return widget.NewLabelWithStyle(text, fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
}

// CodeDisplay is a label extended to copy the code with a menu popup on rightclick.
type CodeDisplay struct {
	widget.Label
	button *widget.Button
}

func (c *CodeDisplay) copyOnPress() {
	clipboard := fyne.CurrentApp().Driver().AllWindows()[0].Clipboard()
	clipboard.SetContent(c.Text)

	c.button.SetIcon(theme.ConfirmIcon())
	time.Sleep(500 * time.Millisecond)
	c.button.SetIcon(theme.ContentCopyIcon())
}

func (c *CodeDisplay) waitForCode(code chan string) {
	go func(code chan string) { // Get the channel out of the scope to avoid stalling render thread
		c.SetText(<-code)
	}(code)
}

func newCodeDisplay() *fyne.Container {
	c := &CodeDisplay{button: widget.NewButtonWithIcon("", theme.ContentCopyIcon(), nil)}
	c.ExtendBaseWidget(c)

	c.SetText("Waiting for code...")
	c.button.OnTapped = c.copyOnPress

	return fyne.NewContainerWithLayout(layout.NewHBoxLayout(), c, c.button)
}