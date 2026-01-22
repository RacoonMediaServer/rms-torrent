package builtin

import (
	"go-micro.dev/v4/logger"
)

func (e *builtinEngine) pushToQueue(t *task) {
	e.queue = append(e.queue, t.id)
	logger.Infof("[%s] Downloading '%s' added to queue", t.id, t.t.Info().Name)
	if len(e.queue) == 1 {
		e.startNextTask()
	}
}

func (e *builtinEngine) startNextTask() {
	if len(e.queue) != 0 {
		e.tasks[e.queue[0]].Start()
	}
}

func (e *builtinEngine) removeFromQueue(id string) {
	for i, tid := range e.queue {
		if tid == id {
			e.queue = append(e.queue[:i], e.queue[i+1:]...)
			if i == 0 {
				e.startNextTask()
			}
			return
		}
	}
}
