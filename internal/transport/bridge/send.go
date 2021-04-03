package bridge

import (
	"context"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/widget"
	"github.com/Jacalz/wormhole-gui/internal/transport"
)

// SendItem is the item that is being sent.
type SendItem struct {
	URI      fyne.URI
	Progress *sendProgress
	Ctx      context.Context
	Cancel   context.CancelFunc
	Code     string
	Name     string
}

// SendList is a list of progress bars that track send progress.
type SendList struct {
	widget.List

	client *transport.Client

	Items []*SendItem
}

// Length returns the length of the data.
func (p *SendList) Length() int {
	return len(p.Items)
}

// CreateItem creates a new item in the list.
func (p *SendList) CreateItem() fyne.CanvasObject {
	return container.New(&listLayout{}, widget.NewFileIcon(nil), widget.NewLabel("Waiting for filename..."), newCodeDisplay(), newSendProgress())
}

// UpdateItem updates the data in the list.
func (p *SendList) UpdateItem(i int, item fyne.CanvasObject) {
	item.(*fyne.Container).Objects[0].(*widget.FileIcon).SetURI(p.Items[i].URI)
	item.(*fyne.Container).Objects[1].(*widget.Label).SetText(p.Items[i].Name)
	item.(*fyne.Container).Objects[2].(*fyne.Container).Objects[0].(*codeDisplay).SetText(p.Items[i].Code)
	p.Items[i].Progress = item.(*fyne.Container).Objects[3].(*sendProgress)
}

// RemoveItem removes the item at the specified index.
func (p *SendList) RemoveItem(i int) {
	copy(p.Items[i:], p.Items[i+1:])
	p.Items[p.Length()-1] = nil // Make sure that GC run on removed element
	p.Items = p.Items[:p.Length()-1]
	p.Refresh()
}

// OnSelected handles removing items and stopping send (in the future)
func (p *SendList) OnSelected(i int) {
	dialog.ShowConfirm("Remove from list", "Do you wish to remove the item from the list?", func(remove bool) {
		if remove {
			p.Items[i].Cancel()
			p.RemoveItem(i)
			p.Refresh()
		}
	}, fyne.CurrentApp().Driver().AllWindows()[0])

	p.Unselect(i)
}

// NewSendItem adds data about a new send to the list and then returns the channel to update the code.
func (p *SendList) NewSendItem(name string, uri fyne.URI) *SendItem {
	ctx, cancel := context.WithTimeout(context.Background(), p.client.Timeout)
	send := &SendItem{Name: name, Code: "Waiting for code...", URI: uri, Ctx: ctx, Cancel: cancel}
	p.Items = append(p.Items, send)
	p.Refresh()
	return send
}

// OnFileSelect is intended to be passed as callback to a FileOpen dialog.
func (p *SendList) OnFileSelect(file fyne.URIReadCloser, err error) {
	if err != nil {
		fyne.LogError("Error on selecting file to send", err)
		dialog.ShowError(err, fyne.CurrentApp().Driver().AllWindows()[0])
		return
	} else if file == nil {
		return
	}

	send := p.NewSendItem(file.URI().Name(), file.URI())

	go func() {
		defer func() {
			send.Cancel()
			if err = file.Close(); err != nil {
				fyne.LogError("Error on closing file", err)
			}
		}()

		code, result, err := p.client.NewFileSend(send.Ctx, file, send.Progress.update)
		if err != nil {
			fyne.LogError("Error on sending file", err)
			dialog.ShowError(err, fyne.CurrentApp().Driver().AllWindows()[0])
			return
		}

		send.Code = code
		p.Refresh()

		if res := <-result; res.Error != nil {
			fyne.LogError("Error on sending file", res.Error)
			dialog.ShowError(res.Error, fyne.CurrentApp().Driver().AllWindows()[0])
			p.client.ShowNotification("File send failed", "An error occurred when sending the file.")
		} else if res.OK {
			p.client.ShowNotification("File send completed", "The file was sent successfully.")
		}
	}()
}

// OnDirSelect is intended to be passed as callback to a FolderOpen dialog.
func (p *SendList) OnDirSelect(dir fyne.ListableURI, err error) {
	if err != nil {
		fyne.LogError("Error on selecting dir to send", err)
		dialog.ShowError(err, fyne.CurrentApp().Driver().AllWindows()[0])
		return
	} else if dir == nil {
		return
	}

	send := p.NewSendItem(dir.Name(), dir)

	go func() {
		defer send.Cancel()
		code, result, err := p.client.NewDirSend(send.Ctx, dir, send.Progress.update)
		if err != nil {
			fyne.LogError("Error on sending directory", err)
			dialog.ShowError(err, fyne.CurrentApp().Driver().AllWindows()[0])
			return
		}

		send.Code = code
		p.Refresh()

		if res := <-result; res.Error != nil {
			fyne.LogError("Error on sending directory", res.Error)
			dialog.ShowError(res.Error, fyne.CurrentApp().Driver().AllWindows()[0])
			p.client.ShowNotification("Directory send failed", "An error occurred when sending the directory.")
		} else if res.OK {
			p.client.ShowNotification("Directory send completed", "The directory was sent successfully.")
		}
	}()
}

// SendText sends new text.
func (p *SendList) SendText() {
	send := p.NewSendItem("Text Snippet", storage.NewFileURI("Text Snippet")) // The file URI is a hack to get the correct icon

	go func() {
		defer send.Cancel()
		if text := <-p.client.ShowTextSendWindow(); text != "" {
			code, result, err := p.client.NewTextSend(send.Ctx, text, send.Progress.update)
			if err != nil {
				fyne.LogError("Error on sending text", err)
				dialog.ShowError(err, fyne.CurrentApp().Driver().AllWindows()[0])
				return
			}

			send.Code = code
			p.Refresh()

			if res := <-result; res.Error != nil {
				fyne.LogError("Error on sending text", res.Error)
				dialog.ShowError(res.Error, fyne.CurrentApp().Driver().AllWindows()[0])
				p.client.ShowNotification("Text send failed", "An error occurred when sending the text.")
			} else if res.OK && p.client.Notifications {
				p.client.ShowNotification("Text send completed", "The text was sent successfully.")
			}
		} else {
			p.RemoveItem(p.Length() - 1)
		}
	}()
}

// NewSendList greates a list of progress bars.
func NewSendList(client *transport.Client) *SendList {
	p := &SendList{client: client}
	p.List.Length = p.Length
	p.List.CreateItem = p.CreateItem
	p.List.UpdateItem = p.UpdateItem
	p.List.OnSelected = p.OnSelected
	p.ExtendBaseWidget(p)

	return p
}
