package downloads

import (
	rms_torrent "github.com/RacoonMediaServer/rms-packages/pkg/service/rms-torrent"
	"github.com/RacoonMediaServer/rms-torrent/internal/downloader"
)

type task struct {
	d      downloader.Downloader
	status rms_torrent.Status
}
