package downloads

import (
	"context"
	"errors"
	rms_torrent "github.com/RacoonMediaServer/rms-packages/pkg/service/rms-torrent"
	"github.com/RacoonMediaServer/rms-torrent/internal/config"
	"github.com/RacoonMediaServer/rms-torrent/internal/downloader"
	uuid "github.com/satori/go.uuid"
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
	for _, t := range m.tasks {
		result = append(result, t.Info())
	}

	return result
}

func (m *Manager) GetDownloadInfo(id string) (*rms_torrent.TorrentInfo, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	t, ok := m.tasks[id]
	if !ok {
		return nil, errors.New("task not found")
	}
	return t.Info(), nil
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

func (m *Manager) RemoveDownload(id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	t, ok := m.tasks[id]
	if !ok {
		return errors.New("download not found")
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

func (m *Manager) Stop() {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, t := range m.tasks {
		t.d.Close()
	}
	m.cancel()
	m.wg.Wait()
}
