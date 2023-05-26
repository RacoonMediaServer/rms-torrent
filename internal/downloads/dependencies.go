package downloads

import (
	"github.com/RacoonMediaServer/rms-torrent/internal/downloader"
	"github.com/RacoonMediaServer/rms-torrent/internal/model"
)

type DownloaderFactory interface {
	New(t *model.Torrent) (downloader.Downloader, error)
	Close()
}
