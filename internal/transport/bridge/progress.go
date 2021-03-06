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

// newSendProgress creates a new fyne progress bar and update function for wormhole send.
func newSendProgress() *sendProgress {
	p := &sendProgress{}
	p.ExtendBaseWidget(p)
	p.update = wormhole.WithProgress(p.updateProgress)

	return p
}

type recvProgress struct {
	widget.ProgressBarInfinite
	done       *widget.ProgressBar
	statusText string
}

func (r *recvProgress) status() string {
	return r.statusText
}

func (r *recvProgress) finished() {
	r.Hide()
	r.done.Show()
}

func (r *recvProgress) completed() {
	r.done.Value = 1.0
	r.finished()
}

func (r *recvProgress) failed() {
	r.done.Value = 0.0
	r.finished()
}

func (r *recvProgress) setStatus(stat string) {
	switch stat {
	case "Failed":
		r.failed()
	case "Completed":
		r.completed()
	}

	r.statusText = stat
}

func newRecvProgress() *fyne.Container {
	r := &recvProgress{done: &widget.ProgressBar{}}
	r.done.TextFormatter = r.status
	r.done.Hide()
	r.ExtendBaseWidget(r)

	return container.NewMax(r, r.done)
}
