package downloads

import (
	"context"
	"go-micro.dev/v4/logger"
	"time"
)

func (m *Manager) startMonitor() {
	m.ctx, m.cancel = context.WithCancel(context.Background())
	m.wg.Add(1)
	go func() {
		defer m.wg.Done()
		m.monitor()
	}()
}

func (m *Manager) monitor() {

	for {
		select {
		case <-m.ctx.Done():
			return
		case <-time.After(checkCompleteInterval):
			m.mu.Lock()
			m.checkTaskStatus()
			m.mu.Unlock()
		}
	}
}

func (m *Manager) checkTaskStatus() {
	if len(m.queue) == 0 {
		return
	}

	t := m.tasks[m.queue[0]]
	if !t.CheckComplete() {
		t.CalcRemaining()
		if t.IsHang() {
			logger.Warnf("[%s] download task is hung ('%s')", t.t.ID, t.d.Title())
			if len(m.queue) > 1 {
				_ = m.upDownload(m.queue[1])
			}
		}
	} else {
		if m.OnDownloadComplete != nil {
			m.OnDownloadComplete(m.ctx, t.t)
		}
		m.queue = m.queue[1:]
		m.startNextTask()
	}
}

func (m *Manager) stopMonitor() {
	m.cancel()
	m.wg.Wait()
}
