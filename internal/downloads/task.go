package downloads

import (
	rms_torrent "github.com/RacoonMediaServer/rms-packages/pkg/service/rms-torrent"
	"github.com/RacoonMediaServer/rms-torrent/internal/downloader"
	"github.com/RacoonMediaServer/rms-torrent/internal/model"
	"go-micro.dev/v4/logger"
)

type task struct {
	t      *model.Torrent
	d      downloader.Downloader
	status rms_torrent.Status
}

func (t *task) Start() {
	logger.Infof("[%s] Start downloading '%s'", t.t.ID, t.d.Title())
	t.d.Start()
	t.status = rms_torrent.Status_Downloading
}

func (t *task) Stop() {
	logger.Infof("[%s] Stopping downloading '%s'", t.t.ID, t.d.Title())
	t.d.Stop()
	t.status = rms_torrent.Status_Pending
}

func (t *task) CheckComplete() bool {
	if t.d.IsComplete() {
		logger.Infof("[%s] Task done '%s'", t.t.ID, t.d.Title())
		t.status = rms_torrent.Status_Done
		return true
	}
	return false
}

func (t *task) Info() *rms_torrent.TorrentInfo {
	return &rms_torrent.TorrentInfo{
		Id:       t.t.ID,
		Title:    t.d.Title(),
		Status:   t.status,
		Progress: t.d.Progress(),
		Estimate: 0,
		SizeMB:   t.d.SizeMB(),
	}
}
