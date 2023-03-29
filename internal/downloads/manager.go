package downloads

import (
	"context"
	rms_torrent "github.com/RacoonMediaServer/rms-packages/pkg/service/rms-torrent"
	"github.com/RacoonMediaServer/rms-torrent/internal/config"
	"github.com/RacoonMediaServer/rms-torrent/internal/downloader"
	uuid "github.com/satori/go.uuid"
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
}

func NewManager() *Manager {
	m := &Manager{
		tasks: map[string]*task{},
	}

	m.ctx, m.cancel = context.WithCancel(context.Background())
	m.wg.Add(1)
	go func() {
		defer m.wg.Done()
		m.monitor()
	}()

	return m
}

func (m *Manager) Download(content []byte, description string, faster bool) (id string, files []string, err error) {
	var d downloader.Downloader
	id = uuid.NewV4().String()

	d, err = downloader.New(downloader.Settings{
		Input:       content,
		Destination: path.Join(config.Config().Directory, id),
	})
	if err != nil {
		return
	}

	files = d.Files()

	t := &task{
		id: id,
		d:  d,
	}

	m.mu.Lock()
	m.tasks[id] = t
	m.pushToQueue(t)
	m.mu.Unlock()

	return
}

func (m *Manager) GetDownloads() []*rms_torrent.TorrentInfo {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make([]*rms_torrent.TorrentInfo, 0, len(m.tasks))
	for id, t := range m.tasks {
		ti := rms_torrent.TorrentInfo{
			Id:       id,
			Title:    t.d.Title(),
			Progress: t.d.Progress(),
			Status:   t.status,
		}
		result = append(result, &ti)
	}

	return result
}

func (m *Manager) monitor() {

	for {
		select {
		case <-m.ctx.Done():
			return
		case <-time.After(checkCompleteInterval):
			m.mu.Lock()
			m.checkTaskIsComplete()
			m.mu.Unlock()
		}
	}
}

func (m *Manager) Stop() {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, t := range m.tasks {
		t.d.Close()
	}
	m.cancel()
	m.wg.Wait()
}
