package downloads

import "go-micro.dev/v4/logger"

func (m *Manager) pushToQueue(t *task) {
	m.queue = append(m.queue, t.id)
	logger.Infof("[%s] Downloading '%s' added to queue", t.id, t.d.Title())
	if len(m.queue) == 1 {
		t.Start()
	}
}

func (m *Manager) checkTaskIsComplete() {
	if len(m.queue) == 0 {
		return
	}

	t := m.tasks[m.queue[0]]
	if t.IsComplete() {
		m.queue = m.queue[1:]
		if len(m.queue) != 0 {
			t = m.tasks[m.queue[0]]
			t.Start()
		}
	}
}
