package downloads

import (
	"context"
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
			m.checkTaskIsComplete()
			m.mu.Unlock()
		}
	}
}

func (m *Manager) stopMonitor() {
	m.cancel()
	m.wg.Wait()
}
