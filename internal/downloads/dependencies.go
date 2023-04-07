package downloads

import "github.com/RacoonMediaServer/rms-torrent/internal/downloader"

type DownloaderFactory interface {
	New(subDirectory string, noRateLimit bool, content []byte) (downloader.Downloader, error)
	Close()
}
