// +build mobile

package transport

import (
	"io"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"
	"github.com/psanford/wormhole-william/wormhole"
)

func (c *Client) receiveFileDir(msg *wormhole.IncomingMessage, path string) (err error) {
	filename := msg.Name
	if msg.Type == wormhole.TransferDirectory {
		filename += ".zip"
	}

	w := fyne.CurrentApp().Driver().AllWindows()[0]
	save := dialog.NewFileSave(func(file fyne.URIWriteCloser, serr error) {
		if serr != nil {
			fyne.LogError("Could not create the file", serr)
			dialog.ShowError(serr, w)
			return
		} else if file == nil {
			return
		}

		defer func() {
			if cerr := file.Close(); cerr != nil {
				fyne.LogError("Error on closing file", err)
				err = cerr
			}
		}()

		if _, serr = io.Copy(file, msg); err != nil {
			fyne.LogError("Error on copying contents to file", err)
			dialog.ShowError(err, w)
			err = serr
			return
		}
	}, w)
	save.SetFileName(filename)
	save.Show()

	return
}
