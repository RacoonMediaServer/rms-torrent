package builtin

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"time"

	tConfig "github.com/RacoonMediaServer/distribyted/config"
	"github.com/RacoonMediaServer/distribyted/fuse"
	"github.com/RacoonMediaServer/distribyted/torrent"
	rms_torrent "github.com/RacoonMediaServer/rms-packages/pkg/service/rms-torrent"
	"github.com/RacoonMediaServer/rms-torrent/v4/pkg/engine"
	"github.com/anacrolix/missinggo/v2/filecache"
	aTorrent "github.com/anacrolix/torrent"
	"github.com/anacrolix/torrent/metainfo"
	"github.com/anacrolix/torrent/storage"
)

type Config struct {
	Directory   string
	Limit       uint
	AddTimeout  uint `json:"add-timeout"`
	ReadTimeout uint `json:"read-timeout"`
	TTL         uint
}

type bultinEngine struct {
	layout    layout
	fuse      *fuse.Handler
	fileStore *torrent.FileItemStore
	cli       *aTorrent.Client
	service   *torrent.Service
	db        engine.TorrentDatabase
}

// Add implements engine.TorrentEngine.
func (e *bultinEngine) Add(ctx context.Context, category string, description string, forceLocation *string, content []byte) (engine.TorrentDescription, error) {
	title, hash, err := e.service.Add(mainRoute, content)
	if err != nil {
		return engine.TorrentDescription{}, err
	}

	result := engine.TorrentDescription{
		ID:       hash,
		Title:    title,
		Location: filepath.Join(e.layout.contentDir, mainRoute),
	}

	record := engine.TorrentRecord{
		TorrentDescription: result,
		Content:            content,
	}

	if err = e.db.Add(record); err != nil {
		_ = e.service.RemoveFromHash(mainRoute, hash)
	}

	return result, err
}

// Get implements engine.TorrentEngine.
func (e *bultinEngine) Get(ctx context.Context, id string) (*rms_torrent.TorrentInfo, error) {
	t, ok := e.cli.Torrent(metainfo.NewHashFromHex(id))
	if !ok {
		return nil, errors.New("not found")
	}
	return convertTorrentInfo(t), nil
}

// List implements engine.TorrentEngine.
func (e *bultinEngine) List(ctx context.Context, includeDone bool) ([]*rms_torrent.TorrentInfo, error) {
	torrents := e.cli.Torrents()
	result := make([]*rms_torrent.TorrentInfo, 0, len(torrents))
	for _, t := range torrents {
		result = append(result, convertTorrentInfo(t))
	}
	return result, nil
}

// Remove implements engine.TorrentEngine.
func (e *bultinEngine) Remove(ctx context.Context, id string) error {
	_ = e.service.RemoveFromHash(mainRoute, id)
	return e.db.Del(id)
}

// Stop implements engine.TorrentEngine.
func (e *bultinEngine) Stop() error {
	_ = e.fileStore.Close()
	e.cli.Close()
	e.fuse.Unmount()
	return nil
}

// UpPriority implements engine.TorrentEngine.
func (e *bultinEngine) UpPriority(ctx context.Context, id string) error {
	return errors.ErrUnsupported
}

func NewEngine(cfg Config, db engine.TorrentDatabase) (engine.TorrentEngine, error) {
	if db == nil {
		db = &engine.VoidDatabase{}
	}

	absPath, err := filepath.Abs(cfg.Directory)
	if err != nil {
		return nil, err
	}

	e := bultinEngine{
		layout: newLayout(absPath),
	}

	if err := e.layout.makeLayout(); err != nil {
		return nil, err
	}

	fileCache, err := filecache.NewCache(e.layout.cacheDir)
	if err != nil {
		return nil, fmt.Errorf("create cache failed: %w", err)
	}
	fileCache.SetCapacity(int64(cfg.Limit) * 1024 * 1024 * 1024)

	torrentStorage := storage.NewResourcePieces(fileCache.AsResourceProvider())

	fileStorage, err := torrent.NewFileItemStore(e.layout.itemsDir, time.Duration(cfg.TTL)*time.Hour)
	if err != nil {
		return nil, fmt.Errorf("create file store failed: %w", err)
	}

	id, err := torrent.GetOrCreatePeerID(filepath.Join(e.layout.baseDir, "ID"))
	if err != nil {
		return nil, fmt.Errorf("create ID failed: %w", err)
	}

	conf := tConfig.TorrentGlobal{
		ReadTimeout:     int(cfg.ReadTimeout),
		AddTimeout:      int(cfg.AddTimeout),
		GlobalCacheSize: -1,
		MetadataFolder:  e.layout.baseDir,
	}

	cli, err := torrent.NewClient(torrentStorage, fileStorage, &conf, id)
	if err != nil {
		return nil, fmt.Errorf("start torrent client failed: %w", err)
	}

	stats := torrent.NewStats()

	loaders := []torrent.DatabaseLoader{&databaseLoader{db}}
	service := torrent.NewService(loaders, stats, cli, conf.AddTimeout, conf.ReadTimeout, true)

	fss, err := service.Load()
	if err != nil {
		return nil, fmt.Errorf("load torrents failed: %w", err)
	}

	mh := fuse.NewHandler(true, e.layout.contentDir)
	if err = mh.Mount(fss); err != nil {
		return nil, fmt.Errorf("mount fuse directory: %w", err)
	}

	e.fuse = mh
	e.fileStore = fileStorage
	e.cli = cli
	e.service = service
	e.db = db

	return &e, nil
}
