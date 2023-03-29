package downloads

import "github.com/anacrolix/torrent"

func (m *Manager) newClient(noDownloadLimit bool) (*torrent.Client, error) {
	return torrent.NewClient(nil)
}
