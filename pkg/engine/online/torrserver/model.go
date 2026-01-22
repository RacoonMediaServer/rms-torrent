package torrserver

import (
	rms_torrent "github.com/RacoonMediaServer/rms-packages/pkg/service/rms-torrent"
	"github.com/RacoonMediaServer/rms-torrent/v4/pkg/torrserver/client/helpers"
)

func convertTorrentInfo(ti *helpers.TorrentInfo) *rms_torrent.TorrentInfo {
	return &rms_torrent.TorrentInfo{
		Id:       ti.Hash,
		Title:    ti.Title,
		Status:   rms_torrent.Status_Done,
		Progress: 100,
		SizeMB:   ti.Size / (1024. * 1024.),
	}
}
