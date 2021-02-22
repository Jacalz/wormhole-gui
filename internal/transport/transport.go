// Package transport handles sending and receiving using wormhole-william
package transport

import (
	"os"
	"path/filepath"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/storage"
	"github.com/psanford/wormhole-william/wormhole"
)

// Client defines the client for handling sending and receiving using wormhole-william
type Client struct {
	wormhole.Client

	// Save a reference to the window to avoid creating a new one when sending and receiving text
	display *textDisplay

	// Notification holds the settings value for if we have notifications enabled or not.
	Notifications bool

	// OverwriteExisting holds the settings value for if we should overwrite already existing files.
	OverwriteExisting bool

	// DownloadPath holds the download path used for saving receieved files.
	DownloadPath fyne.URI
}

// ShowNotification sends a notification if c.Notifications is true.
func (c *Client) ShowNotification(title, content string) {
	if c.Notifications {
		fyne.CurrentApp().SendNotification(&fyne.Notification{Title: title, Content: content})
	}
}

// NewClient returns a new client for sending and receiving using wormhole-william
func NewClient() *Client {
	return &Client{display: createTextWindow()}
}

// UserDownloadsFolder returns the downloads folder corresponding to the current user.
func UserDownloadsFolder() fyne.URI {
	dir, err := os.UserHomeDir()
	if err != nil {
		fyne.LogError("Could not get home dir", err)
	}

	return storage.NewFileURI(filepath.Join(dir, "Downloads"))
}
