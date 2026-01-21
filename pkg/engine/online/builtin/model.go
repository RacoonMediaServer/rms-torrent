package builtin

import (
	rms_torrent "github.com/RacoonMediaServer/rms-packages/pkg/service/rms-torrent"
	"github.com/anacrolix/torrent"
)

func convertTorrentInfo(t *torrent.Torrent) *rms_torrent.TorrentInfo {
	return &rms_torrent.TorrentInfo{
		Id:            t.InfoHash().HexString(),
		Title:         t.Info().Name,
		Status:        rms_torrent.Status_Done,
		Progress:      100,
		RemainingTime: 0,
		SizeMB:        uint64(t.Length()) / (1024. * 1024.),
	}
}
