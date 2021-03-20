// +build mobile

package transport

// NewFileSend takes the chosen file and sends it using wormhole-william.
func (c *Client) NewFileSend(file fyne.URIReadCloser, progress wormhole.SendOption) (string, chan wormhole.SendResult, error) {
	seeker, err := createTemp(file)
	if err != nil {
		fyne.LogError("Could not create temporary file", err)
		return "", nil, err
	}

	return c.SendFile(context.Background(), file.URI().Name(), seeker, progress)
}

// NewDirSend is no-op on mobile due to relying on a desktop file-system.
func (c *Client) NewDirSend(dir fyne.ListableURI, progress wormhole.SendOption) (string, chan wormhole.SendResult, error) {
}

// NewTextSend takes a text input and sends the text using wormhole-william.
func (c *Client) NewTextSend(text string, progress wormhole.SendOption) (string, chan wormhole.SendResult, error) {
	return c.SendText(context.Background(), text, progress)
}

func createTemp(opened fyne.URIReadCloser) (seeker io.ReadSeekCloser, err error) {
	uri, err := storage.Child(fyne.CurrentApp().Storage().RootURI(), opened.URI().Name())
	if err != nil {
		fyne.LogError("Could not create file uri", err)
		return nil, err
	}

	write, err := storage.Writer(uri)
	if err != nil {
		fyne.LogError("Could not open file writer", err)
		return nil, err
	}

	defer func() {
		if cerr := write.Close(); cerr != nil {
			err = cerr
		}
	}()

	size, err := io.Copy(write, opened)
	if err != nil {
		fyne.LogError("Could not copy file contents", err)
		return nil, err
	}

	read, err := storage.Reader(write.URI())
	if err != nil {
		fyne.LogError("Could not open file reader", err)
		return nil, err
	}

	seeker = &sizeSeekCloser{read, size}
	return
}

type sizeSeekCloser struct {
	reader io.ReadCloser
	size   int64
}

// Seek fakes the size check done by wormhole-william. It does not use anything else.
func (s *sizeSeekCloser) Seek(offset int64, whence int) (int64, error) {
	return s.size, nil
}

func (s *sizeSeekCloser) Read(p []byte) (n int, err error) {
	return s.reader.Read(p)
}

func (s *sizeSeekCloser) Close() error {
	return s.reader.Close()
}
