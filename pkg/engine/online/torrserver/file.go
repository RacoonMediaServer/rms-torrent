package torrserver

import "bytes"

type torrentFiles struct {
	reader *bytes.Reader
	name   string
}

func (t *torrentFiles) Read(p []byte) (n int, err error) {
	return t.reader.Read(p)
}

func (t *torrentFiles) Close() error {
	return nil
}

func (t *torrentFiles) Name() string {
	return t.name
}

func newTorrentFile(name string, content []byte) *torrentFiles {
	return &torrentFiles{
		reader: bytes.NewReader(content),
		name:   name,
	}
}
