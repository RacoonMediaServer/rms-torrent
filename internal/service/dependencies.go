package service

import (
	rms_torrent "github.com/RacoonMediaServer/rms-packages/pkg/service/rms-torrent"
	"github.com/RacoonMediaServer/rms-torrent/internal/model"
)

type Database interface {
	LoadSettings() (*rms_torrent.TorrentSettings, error)
	LoadTorrents() ([]*model.Torrent, error)
	AddTorrent(record *model.Torrent) error
	RemoveTorrent(id string) error
	SaveSettings(val *rms_torrent.TorrentSettings) error
	CompleteTorrent(id string) error
}
