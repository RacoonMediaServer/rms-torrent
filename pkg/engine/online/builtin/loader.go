package builtin

import (
	"github.com/RacoonMediaServer/rms-torrent/v4/pkg/engine"
)

type databaseLoader struct {
	db engine.TorrentDatabase
}

func (l databaseLoader) ListTorrents() (map[string][][]byte, error) {
	var result [][]byte

	torrents, err := l.db.Load()
	if err != nil {
		return nil, err
	}
	for _, t := range torrents {
		result = append(result, t.Content)
	}

	return map[string][][]byte{mainRoute: result}, nil
}
