package builtin

import (
	"time"

	rms_torrent "github.com/RacoonMediaServer/rms-packages/pkg/service/rms-torrent"
	"github.com/anacrolix/torrent"
	"go-micro.dev/v4/logger"
)

type task struct {
	id     string
	t      *torrent.Torrent
	status rms_torrent.Status

	lastBytes     uint64
	measures      []uint64
	measure       int
	remainingTime time.Duration
	hangCnt       int
}

const hungDecisionSeconds = 60

func (t *task) Start() {
	logger.Infof("[%s] Start downloading '%s'", t.id, t.t.Info().Name)
	t.t.DownloadAll()
	t.status = rms_torrent.Status_Downloading
	t.lastBytes = uint64(t.t.BytesCompleted())
}

func (t *task) Stop() {
	logger.Infof("[%s] Stopping downloading '%s'", t.id, t.t.Info().Name)
	t.t.CancelPieces(0, t.t.NumPieces())
	t.resetState()
}

func (t *task) CheckComplete() bool {
	if t.t.Complete().Bool() {
		logger.Infof("[%s] Task done '%s'", t.id, t.t.Info().Name)
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

	totalBytes := uint64(t.t.BytesCompleted())
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
	t.remainingTime = time.Duration(float64(t.t.BytesMissing())/speed) * time.Second
}

func (t *task) IsHang() bool {
	return t.hangCnt >= hungDecisionSeconds
}

func (t *task) progress() float32 {
	if t.t.BytesMissing() == 0 && t.t.BytesCompleted() == 0 {
		return 100
	}
	completed := float64(t.t.BytesCompleted())
	left := float64(t.t.BytesMissing())
	return float32(completed/(completed+left)) * 100
}

func (t *task) resetState() {
	t.lastBytes = 0
	t.measures = nil
	t.measure = 0
}

func (t *task) Info() *rms_torrent.TorrentInfo {
	return &rms_torrent.TorrentInfo{
		Id:            t.id,
		Title:         t.t.Info().Name,
		Status:        t.status,
		Progress:      t.progress(),
		RemainingTime: int64(t.remainingTime),
		SizeMB:        uint64(t.t.Info().Length / (1024. * 1024.)),
	}
}

func (t *task) Close() {
	t.t.Drop()
	t.resetState()
}
