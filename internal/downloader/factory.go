package downloader

import (
	"github.com/RacoonMediaServer/rms-torrent/internal/config"
	"github.com/RacoonMediaServer/rms-torrent/internal/model"
)

// Factory is Downloader factory
type Factory interface {
	New(t *model.Torrent) (Downloader, error)
	Close()
}

// FactorySettings are parameters for constructing Downloader's
type FactorySettings struct {
	Fuse          config.Fuse
	DataDirectory string
	UploadLimit   uint64
	DownloadLimit uint64
}

func NewFactory(settings FactorySettings) (Factory, error) {
	if !settings.Fuse.Enabled {
		return newOfflineDownloaderFactory(settings)
	}

	return newOnlineDownloaderFactory(settings)
}
