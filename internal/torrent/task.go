package torrent

import (
	"context"
	"time"

	"git.rms.local/RacoonMediaServer/rms-shared/pkg/events"
	"github.com/cenkalti/rain/torrent"
	"go-micro.dev/v4/logger"
)

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
			logger.Debugf("%s: progress %f %% (%s, peers %d)", tLogName(t), (float64(stats.Bytes.Downloaded)/float64(stats.Bytes.Total))*100., stats.Status.String(), stats.Peers.Outgoing)
			if stats.Status == torrent.Stopped {
				return true
			}

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
