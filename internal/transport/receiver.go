package transport

import (
	"bytes"
	"context"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"

	"fyne.io/fyne/v2"
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

	path := filepath.Join(c.DownloadPath, msg.Name)
	pathToSend = storage.NewFileURI(path).String()

	if !c.OverwriteExisting {
		if _, err := os.Stat(path); err == nil || os.IsExist(err) {
			fyne.LogError("Settings prevent overwriting existing files and folders", err)
			return bail(msg, os.ErrExist)
		}
	}

	if msg.Type == wormhole.TransferFile {
		file, err := storage.Writer(storage.NewFileURI(path))
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

	if fyne.CurrentDevice().IsMobile() {
		file, err := storage.Writer(storage.NewFileURI(path + ".zip"))
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

	err = zip.Extract(tmp, n, path)
	if err != nil {
		fyne.LogError("Error on unzipping contents", err)
		return err
	}

	return nil
}
