package builtin

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	rms_torrent "github.com/RacoonMediaServer/rms-packages/pkg/service/rms-torrent"
	"github.com/RacoonMediaServer/rms-torrent/pkg/engine"
	torrentutils "github.com/RacoonMediaServer/rms-torrent/pkg/torrent-utils"
	"github.com/anacrolix/torrent"
	"github.com/anacrolix/torrent/metainfo"
	"github.com/anacrolix/torrent/storage"
	"go-micro.dev/v4/logger"
)

type Config struct {
	Directory string
}

type builtinEngine struct {
	cli        *torrent.Client
	dir        string
	db         engine.TorrentDatabase
	onComplete engine.CompleteAction

	mu    sync.RWMutex
	tasks map[string]*task
	queue []string

	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

const loadTorrentsTimeout = 2 * time.Minute

// Add implements engine.TorrentEngine.
func (e *builtinEngine) Add(ctx context.Context, category string, description string, forceLocation *string, content []byte) (engine.TorrentDescription, error) {
	result, err := e.addTorrent(ctx, content, false)
	if err != nil {
		return engine.TorrentDescription{}, fmt.Errorf("add torrent to the client failed: %s")
	}

	record := engine.TorrentRecord{
		TorrentDescription: *result,
		Content:            content,
	}

	if err = e.db.Add(record); err != nil {
		_ = e.removeTorrent(record.ID)
		return engine.TorrentDescription{}, fmt.Errorf("add new torrent record to db failed: %w", err)
	}

	return *result, nil
}

// Get implements engine.TorrentEngine.
func (e *builtinEngine) Get(ctx context.Context, id string) (*rms_torrent.TorrentInfo, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	t, ok := e.tasks[id]
	if !ok {
		return nil, errors.New("task not found")
	}
	return t.Info(), nil
}

// List implements engine.TorrentEngine.
func (e *builtinEngine) List(ctx context.Context, includeDone bool) ([]*rms_torrent.TorrentInfo, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	result := make([]*rms_torrent.TorrentInfo, 0, len(e.tasks))
	for _, t := range e.tasks {
		info := t.Info()
		if info.Status == rms_torrent.Status_Done && !includeDone {
			continue
		}
		result = append(result, t.Info())
	}

	return result, nil
}

// Remove implements engine.TorrentEngine.
func (e *builtinEngine) Remove(ctx context.Context, id string) error {
	// TODO: fix possible issues
	if err := e.db.Del(id); err != nil {
		return fmt.Errorf("cannot remove torrent %s from db: %s", id, err)
	}
	return e.removeTorrent(id)
}

// Stop implements engine.TorrentEngine.
func (e *builtinEngine) Stop() error {
	e.mu.Lock()
	defer e.mu.Unlock()

	e.stopMonitor()

	for _, t := range e.tasks {
		t.Close()
	}

	e.cli.Close()
	return nil
}

// UpPriority implements engine.TorrentEngine.
func (e *builtinEngine) UpPriority(ctx context.Context, id string) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.upTorrentUnsafe(id)
}

func NewTorrentEngine(cfg Config, db engine.TorrentDatabase, onComplete engine.CompleteAction) (engine.TorrentEngine, error) {
	e := builtinEngine{
		dir:        cfg.Directory,
		tasks:      map[string]*task{},
		db:         db,
		onComplete: onComplete,
	}

	tConf := torrent.NewDefaultClientConfig()
	tConf.DataDir = cfg.Directory

	cli, err := torrent.NewClient(tConf)
	if err != nil {
		return nil, fmt.Errorf("create torrent client failed: %w", err)
	}
	e.cli = cli

	torrents, err := db.Load()
	if err != nil {
		return nil, fmt.Errorf("Load saved torrents failed: %w", err)
	}

	tCtx, cancel := context.WithTimeout(context.Background(), loadTorrentsTimeout)
	defer cancel()

	for _, t := range torrents {
		_, err := e.addTorrent(tCtx, t.Content, t.Complete)
		if err != nil {
			logger.Warnf("Load torrent failed: %s", err)
		}
	}

	e.startMonitor()

	return &e, nil
}

func (e *builtinEngine) addTorrent(ctx context.Context, content []byte, complete bool) (*engine.TorrentDescription, error) {
	var spec *torrent.TorrentSpec
	isMagnet := torrentutils.IsMagnetLink(content)
	if !isMagnet {
		mi, err := metainfo.Load(bytes.NewReader(content))
		if err != nil {
			return nil, err
		}
		spec = torrent.TorrentSpecFromMetaInfo(mi)
	} else {
		var err error
		spec, err = torrent.TorrentSpecFromMagnetUri(string(content))
		if err != nil {
			return nil, err
		}
	}

	loc := filepath.Join(e.dir, spec.InfoHash.HexString())
	opts := torrent.AddTorrentOpts{
		InfoHash:  spec.InfoHash,
		Storage:   storage.NewFile(loc),
		ChunkSize: spec.ChunkSize,
	}

	t, _ := e.cli.AddTorrentOpt(opts)
	if err := t.MergeSpec(spec); err != nil {
		t.Drop()
		return nil, err
	}
	select {
	case <-ctx.Done():
		t.Drop()
		return nil, ctx.Err()
	case <-t.GotInfo():
	}

	t.AllowDataUpload()

	files := t.Files()
	fileList := make([]string, 0, len(files))
	for _, f := range files {
		fileList = append(fileList, f.Path())
	}

	id := t.InfoHash().HexString()
	tTask := &task{
		id: id,
		t:  t,
	}
	if complete {
		tTask.status = rms_torrent.Status_Done
	}

	e.mu.Lock()
	e.tasks[id] = tTask
	if !complete {
		e.pushToQueue(tTask)
	}
	e.mu.Unlock()

	return &engine.TorrentDescription{
		ID:       id,
		Title:    t.Name(),
		Location: loc,
		Files:    fileList,
	}, nil
}

func (e *builtinEngine) removeTorrent(id string) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	t, ok := e.tasks[id]
	if !ok {
		return errors.New("torrent not found")
	}

	delete(e.tasks, id)
	e.removeFromQueue(id)

	t.Stop()
	t.t.Drop()

	_ = os.RemoveAll(filepath.Join(e.dir, id))
	return nil
}

func (q *builtinEngine) upTorrentUnsafe(id string) error {
	found := -1
	for i := 0; i < len(q.queue) && found == -1; i++ {
		if q.queue[i] == id {
			found = i
		}
	}
	if found < 0 {
		return errors.New("invalid torrent id")
	}

	if found == 0 {
		return nil
	}

	q.tasks[q.queue[0]].Stop()
	newQueue := make([]string, 0, len(q.queue))
	newQueue = append(newQueue, id)
	newQueue = append(newQueue, q.queue[:found]...)
	newQueue = append(newQueue, q.queue[found+1:]...)
	q.queue = newQueue

	q.startNextTask()
	return nil
}
