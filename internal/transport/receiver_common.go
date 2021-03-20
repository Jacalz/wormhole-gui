package transport

import (
	"bytes"
	"context"
	"io"
	"path/filepath"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/storage"
	"github.com/psanford/wormhole-william/wormhole"
)

// NewReceive runs a receive using wormhole-william and handles types accordingly.
func (c *Client) NewReceive(code string, pathname chan string) error {
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
		return c.receiveText(msg)
	}

	path := filepath.Join(c.DownloadPath, msg.Name)
	pathToSend = storage.NewFileURI(path).String()
	return c.receiveFileDir(msg, path)
}

func bail(msg *wormhole.IncomingMessage, err error) error {
	if msg == nil {
		return err
	} else if rerr := msg.Reject(); rerr != nil {
		return rerr
	}

	return err
}

func (c *Client) receiveText(msg *wormhole.IncomingMessage) error {
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
