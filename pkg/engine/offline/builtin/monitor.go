package builtin

import (
	"context"
	"time"

	"github.com/RacoonMediaServer/rms-packages/pkg/events"
	"go-micro.dev/v4/logger"
)

const checkCompleteInterval = 1 * time.Second

func (e *builtinEngine) startMonitor() {
	e.ctx, e.cancel = context.WithCancel(context.Background())
	e.wg.Add(1)
	go func() {
		defer e.wg.Done()
		e.monitor()
	}()
}

func (e *builtinEngine) monitor() {

	for {
		select {
		case <-e.ctx.Done():
			return
		case <-time.After(checkCompleteInterval):
			e.mu.Lock()
			e.checkTaskStatusUnsafe()
			e.mu.Unlock()
		}
	}
}

func (e *builtinEngine) checkTaskStatusUnsafe() {
	if len(e.queue) == 0 {
		return
	}

	t := e.tasks[e.queue[0]]
	if !t.CheckComplete() {
		t.CalcRemaining()
		if t.IsHang() {
			logger.Warnf("[%s] download task is hung ('%s')", t.id, t.t.Info().Name)
			if len(e.queue) > 1 {
				_ = e.upTorrentUnsafe(e.queue[1])
			}
		}
	} else {
		if e.onComplete != nil {
			if err := e.onComplete(events.Notification_DownloadComplete, t.Info()); err != nil {
				logger.Warnf("Complete handler failed: %s", err)
			}
			if err := e.db.Complete(t.id); err != nil {
				logger.Warnf("Write complete flag to database failed: %s", err)
			}
		}
		e.queue = e.queue[1:]
		e.startNextTask()
	}
}

func (e *builtinEngine) stopMonitor() {
	e.cancel()
	e.wg.Wait()
}
