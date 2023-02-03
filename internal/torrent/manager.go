package torrent

import (
	"bytes"
	"context"
	"fmt"
	"sync"
	"time"

	"git.rms.local/RacoonMediaServer/rms-shared/pkg/service/rms_torrent"
	"git.rms.local/RacoonMediaServer/rms-torrent/internal/utils"
	"github.com/cenkalti/rain/torrent"
	uuid "github.com/satori/go.uuid"
	"go-micro.dev/v4"
	"go-micro.dev/v4/logger"
)

type manager struct {
	session *torrent.Session

	taskCh chan *torrent.Torrent
	pub    micro.Event

	wg     sync.WaitGroup
	ctx    context.Context
	cancel context.CancelFunc
}

type Manager interface {
	Download(torrentFileContent []byte) (id string, err error)
	GetTorrentInfo(id string) (result rms_torrent.TorrentInfo, err error)
	GetTorrents(includeDoneTorrents bool) []*rms_torrent.TorrentInfo
	RemoveTorrent(id string) error
	Stop()
}

const maxDownloadTasks = 1000
const publishTimeout = 10 * time.Second

func tLogName(t *torrent.Torrent) string {
	return fmt.Sprintf("[%s:%s]", t.ID(), t.Stats().Name)
}

func New(settings utils.TorrentsSettings, pub micro.Event) (Manager, error) {
	var err error

	conf := torrent.DefaultConfig
	conf.DataDir = settings.Directory
	conf.Database = settings.Db
	conf.RPCEnabled = false
	conf.DataDirIncludesTorrentID = false
	conf.SpeedLimitDownload = int64(settings.MaxSpeed)
	conf.SpeedLimitUpload = int64(settings.MaxSpeed)

	m := &manager{
		taskCh: make(chan *torrent.Torrent, maxDownloadTasks),
		pub:    pub,
	}

	m.session, err = torrent.NewSession(conf)
	if err != nil {
		return m, fmt.Errorf("cannot run torrent client session: %w", err)
	}

	torrents := m.session.ListTorrents()
	for _, t := range torrents {
		stats := t.Stats()

		isComplete := stats.Pieces.Missing == 0
		if !isComplete && stats.Status != torrent.Stopped && stats.Status != torrent.Seeding {
			logger.Infof("%s: queued (status = %s)", tLogName(t), stats.Status.String())
			m.taskCh <- t
		}
	}

	for _, t := range torrents {
		stats := t.Stats()
		if stats.Status == torrent.Stopped {
			logger.Infof("%s: queued (status = %s)", tLogName(t), stats.Status.String())
			m.taskCh <- t
		}
	}

	m.ctx, m.cancel = context.WithCancel(context.TODO())

	m.wg.Add(1)
	go m.processTasks()

	return m, nil
}

func (m *manager) Download(torrentFileContent []byte) (id string, err error) {
	id = uuid.NewV4().String()
	t, err := m.session.AddTorrent(bytes.NewReader(torrentFileContent), &torrent.AddTorrentOptions{
		ID:      id,
		Stopped: true,
	})
	if err != nil {
		err = fmt.Errorf("cannot add torrent: %w", err)
		return
	}

	logger.Infof("%s: added to queue", tLogName(t))

	m.taskCh <- t

	return
}

func extractTorrentInfo(t *torrent.Torrent, out *rms_torrent.TorrentInfo) {
	stats := t.Stats()

	out.Id = t.ID()
	out.Title = t.Name()
	out.Progress = (float32(stats.Bytes.Downloaded) / float32(stats.Bytes.Total)) * 100.
	if stats.ETA != nil {
		out.Estimate = int64(*stats.ETA)
	}
	switch stats.Status {
	case torrent.Stopped:
		out.Status = rms_torrent.Status_Pending
	case torrent.Downloading:
		out.Status = rms_torrent.Status_Downloading
	default:
		out.Status = rms_torrent.Status_Done
	}
}

func (m *manager) GetTorrentInfo(id string) (result rms_torrent.TorrentInfo, err error) {
	t := m.session.GetTorrent(id)
	if t == nil {
		err = fmt.Errorf("torrent %s not found", id)
		return
	}

	extractTorrentInfo(t, &result)
	return
}

func (m *manager) GetTorrents(includeDoneTorrents bool) []*rms_torrent.TorrentInfo {
	torrents := m.session.ListTorrents()
	result := make([]*rms_torrent.TorrentInfo, 0, len(torrents))
	for _, t := range torrents {
		ti := rms_torrent.TorrentInfo{}
		extractTorrentInfo(t, &ti)

		if !includeDoneTorrents && ti.Status == rms_torrent.Status_Done {
			continue
		}
		result = append(result, &ti)
	}

	return result
}

func (m *manager) RemoveTorrent(id string) error {
	return m.session.RemoveTorrent(id)
}

func (m *manager) Stop() {
	logger.Info("Stopping...")
	m.cancel()
	m.wg.Wait()
	m.session.Close()
}
