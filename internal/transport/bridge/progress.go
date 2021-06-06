package bridge

import (
	"sync"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"github.com/psanford/wormhole-william/wormhole"
)

// sendProgress is contains a widget for displaying wormhole send progress.
type sendProgress struct {
	widget.ProgressBar

	// Update is the SendOption that should be passed to the wormhole client.
	update wormhole.SendOption
	once   sync.Once
}

// UpdateProgress is the function that runs when updating the progress.
func (p *sendProgress) updateProgress(sent int64, total int64) {
	p.once.Do(func() { p.Max = float64(total) })
	p.SetValue(float64(sent))
}

func (p *sendProgress) setStatus(status string) {
	if status != "" {
		p.TextFormatter = func() string { return status }
		p.Refresh()
	}
}

// newSendProgress creates a new fyne progress bar and update function for wormhole send.
func newSendProgress() *sendProgress {
	p := &sendProgress{}
	p.ExtendBaseWidget(p)
	p.update = wormhole.WithProgress(p.updateProgress)

	return p
}

type recvProgress struct {
	widget.ProgressBarInfinite
	done *widget.ProgressBar
}

func (r *recvProgress) setStatus(status string) {
	switch status {
	case "":
		return
	case "Failed":
		r.done.Value = 0.0
	case "Completed":
		r.done.Value = 1.0
	}

	r.done.TextFormatter = func() string { return status }
	r.Hide()
	r.done.Show()
}

func newRecvProgress() *fyne.Container {
	r := &recvProgress{done: &widget.ProgressBar{}}
	r.done.Hide()
	r.ExtendBaseWidget(r)

	return container.NewMax(r, r.done)
}
