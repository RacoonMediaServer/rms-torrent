package downloader

import "github.com/RacoonMediaServer/rms-torrent/internal/model"

type downloaderParameters struct {
	settings FactorySettings
	t        *model.Torrent
}
