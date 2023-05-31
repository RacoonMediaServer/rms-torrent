package downloads

import (
	rms_torrent "github.com/RacoonMediaServer/rms-packages/pkg/service/rms-torrent"
	"github.com/RacoonMediaServer/rms-torrent/internal/downloader"
	"github.com/RacoonMediaServer/rms-torrent/internal/model"
	"go-micro.dev/v4/logger"
	"time"
)

type task struct {
	t      *model.Torrent
	d      downloader.Downloader
	status rms_torrent.Status

	lastBytes     uint64
	measures      []uint64
	measure       int
	remainingTime time.Duration
	hangCnt       int
}

const hungDecisionSeconds = 60

func (t *task) Start() {
	logger.Infof("[%s] Start downloading '%s'", t.t.ID, t.d.Title())
	t.d.Start()
	t.status = rms_torrent.Status_Downloading
	t.lastBytes = t.d.Bytes()
}

func (t *task) Stop() {
	logger.Infof("[%s] Stopping downloading '%s'", t.t.ID, t.d.Title())
	t.d.Stop()
	t.resetState()
}

func (t *task) CheckComplete() bool {
	if t.d.IsComplete() {
		logger.Infof("[%s] Task done '%s'", t.t.ID, t.d.Title())
		t.status = rms_torrent.Status_Done
		return true
	}
	return false
}

func (t *task) avgSpeed() float64 {
	var avg float64
	for _, v := range t.measures {
		avg += float64(v)
	}
	if len(t.measures) != 0 {
		avg = avg / float64(len(t.measures))
	}

	return avg
}

func (t *task) CalcRemaining() {
	const Smoothing float64 = 0.08
	const MaxMeasures = 20

	totalBytes := t.d.Bytes()
	bytes := totalBytes - t.lastBytes
	if bytes == 0 {
		t.hangCnt++
	} else {
		t.hangCnt = 0
	}
	if len(t.measures) >= MaxMeasures {
		t.measures[t.measure] = bytes
		t.measure++
		if t.measure >= MaxMeasures {
			t.measure = 0
		}
	} else {
		t.measures = append(t.measures, bytes)
	}
	speed := (Smoothing * float64(bytes)) + (1.-Smoothing)*t.avgSpeed()
	t.lastBytes = totalBytes
	t.remainingTime = time.Duration(float64(t.d.RemainingBytes())/speed) * time.Second
}

func (t *task) IsHang() bool {
	return t.hangCnt >= hungDecisionSeconds
}

func (t *task) progress() float32 {
	completed := float64(t.d.Bytes())
	left := float64(t.d.RemainingBytes())
	return float32(completed/(completed+left)) * 100
}

func (t *task) resetState() {
	t.lastBytes = 0
	t.measures = nil
	t.measure = 0
}

func (t *task) Info() *rms_torrent.TorrentInfo {
	return &rms_torrent.TorrentInfo{
		Id:            t.t.ID,
		Title:         t.d.Title(),
		Status:        t.status,
		Progress:      t.progress(),
		RemainingTime: int64(t.remainingTime),
		SizeMB:        t.d.SizeMB(),
	}
}

func (t *task) Close() {
	t.d.Close()
	t.resetState()
}
