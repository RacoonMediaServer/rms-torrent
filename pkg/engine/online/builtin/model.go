package builtin

import (
	"path/filepath"

	rms_torrent "github.com/RacoonMediaServer/rms-packages/pkg/service/rms-torrent"
	"github.com/anacrolix/torrent"
)

func (e *bultinEngine) convertTorrentInfo(t *torrent.Torrent) *rms_torrent.TorrentInfo {
	title := t.Info().Name
	return &rms_torrent.TorrentInfo{
		Id:            t.InfoHash().HexString(),
		Title:         title,
		Status:        rms_torrent.Status_Done,
		Progress:      100,
		RemainingTime: 0,
		SizeMB:        uint64(t.Length()) / (1024. * 1024.),
		Location:      filepath.Join(e.layout.contentDir, mainRoute, title),
	}
}
