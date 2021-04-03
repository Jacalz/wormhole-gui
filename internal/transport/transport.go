// Package transport handles sending and receiving using wormhole-william
package transport

import (
	"os"
	"path/filepath"
	"time"

	"fyne.io/fyne/v2"
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

	// Timeout holds the amount of time after which sends and receives should cancel.
	Timeout time.Duration

	// DownloadPath holds the download path used for saving received files.
	DownloadPath string
}

// ShowNotification sends a notification if c.Notifications is true.
func (c *Client) ShowNotification(title, content string) {
	if c.Notifications {
		fyne.CurrentApp().SendNotification(&fyne.Notification{Title: title, Content: content})
	}
}

// NewClient returns a new client for sending and receiving using wormhole-william
func NewClient() *Client {
	return &Client{display: createTextWindow(), Timeout: time.Minute}
}

// UserDownloadsFolder returns the downloads folder corresponding to the current user.
func UserDownloadsFolder() string {
	dir, err := os.UserHomeDir()
	if err != nil {
		fyne.LogError("Could not get home dir", err)
	}

	return filepath.Join(dir, "Downloads")
}
