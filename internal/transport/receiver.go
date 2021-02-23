package transport

import (
	"bytes"
	"context"
	"io"
	"io/ioutil"
	"os"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
	"github.com/Jacalz/wormhole-gui/internal/transport/zip"
	"github.com/psanford/wormhole-william/wormhole"
)

func bail(msg *wormhole.IncomingMessage, err error) error {
	if msg == nil {
		return err
	} else if rerr := msg.Reject(); rerr != nil {
		return rerr
	}

	return err
}

// NewReceive runs a receive using wormhole-william and handles types accordingly.
func (c *Client) NewReceive(code string, pathname chan string) (err error) {
	// We want to always send a URI, even on fail, in order to not block goroutines
	pathToSend := "text"
	defer func() {
		pathname <- pathToSend
	}()

	msg, err := c.Receive(context.Background(), code)
	if err != nil {
		fyne.LogError("Error on receiving data", err)
		return bail(msg, err)
	}

	if msg.Type == wormhole.TransferText {
		text := &bytes.Buffer{}
		text.Grow(int(msg.TransferBytes64))

		_, err := io.Copy(text, msg)
		if err != nil {
			fyne.LogError("Could not copy the received text", err)
			return err
		}

		c.showTextReceiveWindow(text)
		return nil
	}

	if fyne.CurrentDevice().IsMobile() {
		var merr error = nil
		save := dialog.NewFileSave(func(file fyne.URIWriteCloser, err error) {
			if err != nil {
				fyne.LogError("Could not open the file", err)
				merr = err
				return
			} else if file == nil {
				return
			}

			defer func() {
				if cerr := file.Close(); cerr != nil {
					fyne.LogError("Error on closing file", err)
					merr = cerr
				}
			}()

			_, err = io.Copy(file, msg)
			if err != nil {
				fyne.LogError("Error on copying contents to file", err)
				merr = err
				return
			}
		}, fyne.CurrentApp().Driver().AllWindows()[0])

		filename := msg.Name
		if msg.Type == wormhole.TransferDirectory {
			filename += ".zip"
		}
		save.SetFileName(filename)

		save.Show()
		return merr
	}

	child, err := storage.Child(c.DownloadPath, msg.Name)
	if err != nil {
		fyne.LogError("Could not create a child uri", err)
		return bail(msg, err)
	}

	pathToSend = child.String()

	if exists, err := storage.Exists(child); !c.OverwriteExisting && (exists || err != nil) {
		fyne.LogError("Settings prevent overwriting existing files and folders", err)
		return bail(msg, os.ErrExist)
	}

	if msg.Type == wormhole.TransferFile {
		file, err := storage.Writer(child)
		if err != nil {
			fyne.LogError("Error on creating file", err)
			return bail(msg, err)
		}

		defer func() {
			if cerr := file.Close(); cerr != nil {
				fyne.LogError("Error on closing file", err)
				err = cerr
			}
		}()

		_, err = io.Copy(file, msg)
		if err != nil {
			fyne.LogError("Error on copying contents to file", err)
			return err
		}

		return nil
	}

	tmp, err := ioutil.TempFile("", msg.Name+"-*.zip.tmp")
	if err != nil {
		fyne.LogError("Error on creating tempfile", err)
		return bail(msg, err)
	}

	defer func() {
		if cerr := tmp.Close(); cerr != nil {
			fyne.LogError("Error on closing file", err)
			err = cerr
		}

		if rerr := os.Remove(tmp.Name()); rerr != nil {
			fyne.LogError("Error on removing temp file", err)
			err = rerr
		}
	}()

	n, err := io.Copy(tmp, msg)
	if err != nil {
		fyne.LogError("Error on copying contents to file", err)
		return err
	}

	err = zip.Extract(tmp, n, child.Path())
	if err != nil {
		fyne.LogError("Error on unzipping contents", err)
		return err
	}

	return nil
}
