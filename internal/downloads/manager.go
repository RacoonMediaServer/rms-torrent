package downloads

import (
	rms_torrent "github.com/RacoonMediaServer/rms-packages/pkg/service/rms-torrent"
	"github.com/RacoonMediaServer/rms-torrent/internal/config"
	"github.com/RacoonMediaServer/rms-torrent/internal/downloader"
	uuid "github.com/satori/go.uuid"
	"path"
	"sync"
)

type Manager struct {
	mu    sync.RWMutex
	tasks map[string]*task
}

func NewManager() *Manager {
	return &Manager{
		tasks: map[string]*task{},
	}
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

	d.Start()
	files = d.Files()

	t := &task{d: d}

	m.mu.Lock()
	m.tasks[id] = t
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
