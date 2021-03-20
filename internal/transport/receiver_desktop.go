// +build !mobile

package transport

import (
	"io"
	"io/ioutil"
	"os"

	"fyne.io/fyne/v2"
	"github.com/Jacalz/wormhole-gui/internal/transport/zip"
	"github.com/psanford/wormhole-william/wormhole"
)

func (c *Client) receiveFileDir(msg *wormhole.IncomingMessage, path string) (err error) {
	if !c.OverwriteExisting {
		if _, err := os.Stat(path); err == nil || os.IsExist(err) {
			fyne.LogError("Settings prevent overwriting existing files and folders", err)
			return bail(msg, os.ErrExist)
		}
	}

	if msg.Type == wormhole.TransferFile {
		file, err := os.Create(path)
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

	return
}
