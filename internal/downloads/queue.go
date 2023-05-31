package downloads

import (
	"go-micro.dev/v4/logger"
)

func (m *Manager) pushToQueue(t *task) {
	m.queue = append(m.queue, t.t.ID)
	logger.Infof("[%s] Downloading '%s' added to queue", t.t.ID, t.d.Title())
	if len(m.queue) == 1 {
		m.startNextTask()
	}
}

func (m *Manager) startNextTask() {
	if len(m.queue) != 0 {
		m.tasks[m.queue[0]].Start()
	}
}

func (m *Manager) removeFromQueue(id string) {
	for i, tid := range m.queue {
		if tid == id {
			m.queue = append(m.queue[:i], m.queue[i+1:]...)
			if i == 0 {
				m.startNextTask()
			}
			return
		}
	}
}
