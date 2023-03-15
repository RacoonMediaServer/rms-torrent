package torrent

import (
	"bytes"
	"context"
	"fmt"
	"github.com/RacoonMediaServer/rms-packages/pkg/events"
	rms_torrent "github.com/RacoonMediaServer/rms-packages/pkg/service/rms-torrent"
	"github.com/RacoonMediaServer/rms-torrent/internal/config"
	"github.com/cenkalti/rain/torrent"
	uuid "github.com/satori/go.uuid"
	"go-micro.dev/v4"
	"go-micro.dev/v4/logger"
)

type Manager struct {
	session *torrent.Session
	q       *torrentQueue
	pub     micro.Event
}

func tLogName(t *torrent.Torrent) string {
	return fmt.Sprintf("[%s:%s]", t.ID(), t.Stats().Name)
}

func New(settings config.TorrentsSettings, pub micro.Event) (*Manager, error) {
	var err error

	conf := torrent.DefaultConfig
	conf.DataDir = settings.Directory
	conf.Database = settings.Db
	conf.RPCEnabled = false
	conf.DataDirIncludesTorrentID = true
	conf.SpeedLimitDownload = int64(settings.MaxSpeed)
	conf.SpeedLimitUpload = int64(settings.MaxSpeed)

	m := &Manager{pub: pub}

	m.session, err = torrent.NewSession(conf)
	if err != nil {
		return m, fmt.Errorf("cannot run torrent client session: %w", err)
	}

	m.q = newTorrentQueue(context.Background(), pub)

	torrents := m.session.ListTorrents()
	for _, t := range torrents {
		stats := t.Stats()

		isComplete := stats.Pieces.Missing == 0
		if !isComplete && stats.Status != torrent.Stopped && stats.Status != torrent.Seeding {
			logger.Infof("%s: queued (status = %s)", tLogName(t), stats.Status.String())
			m.q.Push(t)
		}
	}

	for _, t := range torrents {
		stats := t.Stats()
		if stats.Status == torrent.Stopped {
			logger.Infof("%s: queued (status = %s)", tLogName(t), stats.Status.String())
			m.q.Push(t)
		}
	}

	return m, nil
}

func (m *Manager) Download(content []byte) (id string, files []string, err error) {
	id = uuid.NewV4().String()
	var t *torrent.Torrent

	opts := &torrent.AddTorrentOptions{
		ID:      id,
		Stopped: true,
	}

	isMagnet := isMagnetLink(content)

	if isMagnet {
		t, err = m.session.AddURI(string(content), opts)
	} else {
		t, err = m.session.AddTorrent(bytes.NewReader(content), opts)
	}

	if err != nil {
		err = fmt.Errorf("cannot add torrent: %w", err)
		return
	}

	if isMagnet {
		if err = t.Start(); err != nil {
			err = fmt.Errorf("cannot retrieve data by magnet link: %w", err)
			_ = m.session.RemoveTorrent(id)
			return
		}
		<-t.NotifyMetadata()
		_ = t.Stop()
		content, err = t.Torrent()
		if err != nil {
			err = fmt.Errorf("cannot retrieve torrent data by magnet link: %w", err)
			_ = m.session.RemoveTorrent(id)
			return
		}
	}

	files, err = getTorrentFiles(t.Name(), content)
	if err != nil {
		_ = m.session.RemoveTorrent(id)
		err = fmt.Errorf("get torrent files failed: %w", err)
		return
	}

	if len(files) == 0 {
		files = []string{t.Name()}
	}

	logger.Infof("%s: added to queue", tLogName(t))

	m.q.Push(t)

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

func (m *Manager) GetTorrentInfo(id string) (result rms_torrent.TorrentInfo, err error) {
	t := m.session.GetTorrent(id)
	if t == nil {
		err = fmt.Errorf("torrent %s not found", id)
		return
	}

	extractTorrentInfo(t, &result)
	return
}

func (m *Manager) GetTorrents(includeDoneTorrents bool) []*rms_torrent.TorrentInfo {
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

func (m *Manager) RemoveTorrent(ctx context.Context, id string) error {
	m.q.Remove(id)
	if err := m.session.RemoveTorrent(id); err != nil {
		return err
	}

	event := events.Notification{
		Kind:      events.Notification_TorrentRemoved,
		TorrentID: &id,
	}

	ctx, cancel := context.WithTimeout(ctx, publishTimeout)
	defer cancel()

	if err := m.pub.Publish(ctx, &event); err != nil {
		logger.Warnf("Publish notification failed: %s", err)
	}

	return nil
}

func (m *Manager) Up(id string) error {
	return m.q.Up(id)
}

func (m *Manager) Stop() {
	logger.Info("Stopping...")
	m.q.Stop()
	m.session.Close()
}
