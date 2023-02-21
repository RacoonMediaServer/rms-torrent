package torrent

import (
	"container/list"
	"context"
	"errors"
	"github.com/RacoonMediaServer/rms-packages/pkg/events"
	"github.com/cenkalti/rain/torrent"
	"go-micro.dev/v4"
	"go-micro.dev/v4/logger"
	"sync"
	"time"
)

const publishTimeout = 10 * time.Second

type torrentQueue struct {
	mu     sync.Mutex
	q      *list.List
	byId   map[string]*list.Element
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
	pub    micro.Event

	task       *torrent.Torrent
	taskCancel context.CancelFunc
}

func newTorrentQueue(ctx context.Context, pub micro.Event) *torrentQueue {
	ctx, cancel := context.WithCancel(ctx)

	return &torrentQueue{
		q:      list.New(),
		byId:   map[string]*list.Element{},
		ctx:    ctx,
		cancel: cancel,
		pub:    pub,
	}
}

func (q *torrentQueue) Push(t *torrent.Torrent) {
	q.mu.Lock()
	defer q.mu.Unlock()

	q.byId[t.ID()] = q.q.PushBack(t)

	if q.task == nil {
		q.startNextTask()
	}
}

func (q *torrentQueue) Remove(id string) {
	q.mu.Lock()
	defer q.mu.Unlock()

	if v, ok := q.byId[id]; ok {
		q.q.Remove(v)
		delete(q.byId, id)
		return
	}

	if q.task != nil && q.task.ID() == id {
		q.taskCancel()
		q.startNextTask()
	}
}

func (q *torrentQueue) Up(id string) error {
	q.mu.Lock()
	defer q.mu.Unlock()

	if q.task != nil && q.task.ID() == id {
		return nil
	}

	v, ok := q.byId[id]
	if !ok {
		return errors.New("torrent not found")
	}

	if q.task != nil {
		q.taskCancel()
		_ = q.task.Stop()
		q.byId[id] = q.q.PushFront(q.task)
		q.task = nil
	}

	q.q.MoveToFront(v)
	q.startNextTask()

	return nil
}

func (q *torrentQueue) startNextTask() {
	for q.q.Len() != 0 {
		e := q.q.Front()
		t := e.Value.(*torrent.Torrent)

		q.q.Remove(e)
		delete(q.byId, t.ID())

		if err := t.Start(); err != nil {
			logger.Errorf("%s: start download failed: %s", tLogName(t), err)
			continue
		}

		logger.Infof("%s: download started", tLogName(t))

		var ctx context.Context

		q.task = t
		ctx, q.taskCancel = context.WithCancel(q.ctx)
		q.wg.Add(1)
		go q.processTask(ctx, t)

		return
	}
}

func (q *torrentQueue) processTask(ctx context.Context, t *torrent.Torrent) {
	defer q.wg.Done()
	for {
		select {
		case <-time.After(5 * time.Second):
			stats := t.Stats()
			logger.Debugf("%s: progress %f %% (%s, peers %d)", tLogName(t), (float64(stats.Bytes.Downloaded)/float64(stats.Bytes.Total))*100., stats.Status.String(), stats.Peers.Outgoing)
			if stats.Status == torrent.Stopped {
				return
			}

		case <-t.NotifyComplete():
			q.completeTask(t)
			return
		case <-ctx.Done():
			return
		}
	}
}

func (q *torrentQueue) completeTask(t *torrent.Torrent) {
	logger.Infof("%s: download complete", tLogName(t))
	id := t.ID()
	name := t.Name()

	q.publish(&events.Notification{
		Kind:      events.Notification_DownloadComplete,
		TorrentID: &id,
		ItemTitle: &name,
	})

	q.mu.Lock()
	defer q.mu.Unlock()

	q.task = nil
	q.startNextTask()
}

func (q *torrentQueue) publish(event *events.Notification) {
	ctx, cancel := context.WithTimeout(q.ctx, publishTimeout)
	defer cancel()

	if err := q.pub.Publish(ctx, event); err != nil {
		logger.Warnf("Publish notification failed: %s", err)
	}
}

func (q *torrentQueue) Stop() {
	q.cancel()
	q.wg.Wait()
}
