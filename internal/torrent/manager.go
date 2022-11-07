package torrent

import (
	"bytes"
	"context"
	"fmt"
	"sync"
	"time"

	"git.rms.local/RacoonMediaServer/rms-shared/pkg/events"
	"git.rms.local/RacoonMediaServer/rms-torrent/internal/utils"
	"github.com/cenkalti/rain/torrent"
	uuid "github.com/satori/go.uuid"
	"go-micro.dev/v4"
	"go-micro.dev/v4/logger"
)

type Manager struct {
	session *torrent.Session

	taskCh chan *torrent.Torrent
	pub    micro.Event

	wg     sync.WaitGroup
	ctx    context.Context
	cancel context.CancelFunc
}

const maxDownloadTasks = 1000
const publishTimeout = 10 * time.Second

func tLogName(t *torrent.Torrent) string {
	return fmt.Sprintf("[%s:%s]", t.ID(), t.Stats().Name)
}

func NewManager(settings utils.TorrentsSettings, pub micro.Event) (*Manager, error) {
	var err error

	conf := torrent.DefaultConfig
	conf.DataDir = settings.Directory
	conf.Database = settings.Db
	conf.RPCEnabled = false
	conf.DataDirIncludesTorrentID = false
	conf.SpeedLimitDownload = int64(settings.MaxSpeed)
	conf.SpeedLimitUpload = int64(settings.MaxSpeed)

	m := &Manager{
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

func (m *Manager) Download(torrentFileContent []byte) (id string, err error) {
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

func (m *Manager) processTasks() {
	defer m.wg.Done()

	for {
		select {
		case t := <-m.taskCh:
			if !m.startTask(t) {
				return
			}
		case <-m.ctx.Done():
			return
		}
	}
}

func (m *Manager) startTask(t *torrent.Torrent) bool {
	if err := t.Start(); err != nil {
		logger.Errorf("%s: start download failed: %s", tLogName(t), err)
		return true
	}
	logger.Infof("%s: download started", tLogName(t))

	for {
		select {
		case <-time.After(5 * time.Second):
			stats := t.Stats()
			logger.Infof("%s: progress %f %% (%s, peers %d)", tLogName(t), (float64(stats.Bytes.Downloaded)/float64(stats.Bytes.Total))*100., stats.Status.String(), stats.Peers.Outgoing)
		case <-t.NotifyComplete():
			m.completeTask(t)
			return true
		case <-m.ctx.Done():
			return false
		}
	}
}

func (m *Manager) completeTask(t *torrent.Torrent) {
	logger.Infof("%s: download complete", tLogName(t))
	m.publish(&events.Notification{
		Kind: events.Notification_DownloadComplete,
		Detailed: map[string]string{
			"id":   t.ID(),
			"item": t.Name(),
		},
	})
}

func (m *Manager) publish(event *events.Notification) {
	ctx, cancel := context.WithTimeout(m.ctx, publishTimeout)
	defer cancel()

	if err := m.pub.Publish(ctx, event); err != nil {
		logger.Warnf("Publish notification failed: %s", err)
	}
}

func (m *Manager) Stop() {
	logger.Info("Stopping...")
	m.cancel()
	m.wg.Wait()
	m.session.Close()
}
