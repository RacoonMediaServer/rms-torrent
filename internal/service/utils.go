package service

import (
	rms_torrent "github.com/RacoonMediaServer/rms-packages/pkg/service/rms-torrent"
	"github.com/RacoonMediaServer/rms-torrent/internal/config"
	"github.com/RacoonMediaServer/rms-torrent/internal/downloader"
)

func newFactory(settings *rms_torrent.TorrentSettings) (downloader.Factory, error) {
	cfg := config.Config()
	return downloader.NewFactory(downloader.FactorySettings{
		Fuse:          cfg.Fuse,
		DataDirectory: cfg.Directory,
		UploadLimit:   settings.UploadLimit,
		DownloadLimit: settings.DownloadLimit,
	})
}
