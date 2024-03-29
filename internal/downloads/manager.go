package downloads

import (
	"context"
	"errors"
	"fmt"
	"github.com/RacoonMediaServer/rms-packages/pkg/events"
	rms_torrent "github.com/RacoonMediaServer/rms-packages/pkg/service/rms-torrent"
	"github.com/RacoonMediaServer/rms-torrent/internal/config"
	"github.com/RacoonMediaServer/rms-torrent/internal/downloader"
	"github.com/RacoonMediaServer/rms-torrent/internal/model"
	"go-micro.dev/v4/logger"
	"os"
	"path"
	"sync"
	"time"
)

const checkCompleteInterval = 1 * time.Second

type Manager struct {
	mu    sync.RWMutex
	tasks map[string]*task
	queue []string

	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup

	f downloader.Factory

	OnDownloadComplete func(ctx context.Context, torrentTitle string, t *model.Torrent)
	OnMalfunction      func(text string, code events.Malfunction_Code)
}

func NewManager(f downloader.Factory) *Manager {
	m := &Manager{
		tasks: map[string]*task{},
		f:     f,
	}

	m.startMonitor()

	return m
}

func (m *Manager) Download(record *model.Torrent) (files []string, err error) {
	var d downloader.Downloader

	d, err = m.f.New(record)
	if err != nil {
		return
	}

	files = d.Files()

	t := &task{
		t: record,
		d: d,
	}
	if record.Complete {
		t.status = rms_torrent.Status_Done
	}

	m.mu.Lock()
	m.tasks[record.ID] = t
	if !record.Complete {
		m.pushToQueue(t)
	}
	m.mu.Unlock()

	return
}

func (m *Manager) GetTorrents(includeDone bool) []*rms_torrent.TorrentInfo {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make([]*rms_torrent.TorrentInfo, 0, len(m.tasks))
	for _, t := range m.tasks {
		info := t.Info()
		if info.Status == rms_torrent.Status_Done && !includeDone {
			continue
		}
		result = append(result, t.Info())
	}

	return result
}

func (m *Manager) GetTorrentInfo(id string) (*rms_torrent.TorrentInfo, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	t, ok := m.tasks[id]
	if !ok {
		return nil, errors.New("task not found")
	}
	return t.Info(), nil
}

func (m *Manager) RemoveTorrent(id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	t, ok := m.tasks[id]
	if !ok {
		return errors.New("torrent not found")
	}

	delete(m.tasks, id)
	m.removeFromQueue(id)

	t.Stop()
	t.d.Close()

	_ = os.RemoveAll(path.Join(config.Config().Directory, id))
	return nil
}

func (m *Manager) UpDownload(id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.upDownload(id)
}

func (m *Manager) upDownload(id string) error {
	found := -1
	for i := 0; i < len(m.queue) && found == -1; i++ {
		if m.queue[i] == id {
			found = i
		}
	}
	if found < 0 {
		return errors.New("invalid torrent id")
	}

	if found == 0 {
		return nil
	}

	m.tasks[m.queue[0]].Stop()
	newQueue := make([]string, 0, len(m.queue))
	newQueue = append(newQueue, id)
	newQueue = append(newQueue, m.queue[:found]...)
	newQueue = append(newQueue, m.queue[found+1:]...)
	m.queue = newQueue

	m.startNextTask()
	return nil
}

func (m *Manager) Reset(f downloader.Factory) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.f = f
	var err error
	for _, t := range m.tasks {
		t.d, err = m.f.New(t.t)
		if err != nil {
			errorString := fmt.Sprintf("[%s] Cannot restart downloading task '%s': %s", t.t.ID, t.t.Description, err)
			logger.Error(errorString)
			if m.OnMalfunction != nil {
				m.OnMalfunction(errorString, events.Malfunction_ActionFailed)
			}
		}
	}

	m.startNextTask()
	m.startMonitor()
}

func (m *Manager) Stop() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.stopMonitor()

	for _, t := range m.tasks {
		t.Close()
	}

	m.f.Close()
}
