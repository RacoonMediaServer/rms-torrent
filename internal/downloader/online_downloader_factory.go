package downloader

import (
	"fmt"
	"github.com/RacoonMediaServer/distribyted/fuse"
	"github.com/RacoonMediaServer/distribyted/torrent"
	"github.com/RacoonMediaServer/rms-torrent/internal/model"
	"github.com/anacrolix/missinggo/v2/filecache"
	"github.com/anacrolix/torrent/storage"
	"os"
	"path/filepath"
	"time"

	aTorrent "github.com/anacrolix/torrent"

	tConfig "github.com/RacoonMediaServer/distribyted/config"
)

const (
	cacheTTL    = 24 * time.Hour
	readTimeout = 60
	addTimeout  = 60
)

type onlineDownloaderFactory struct {
	settings FactorySettings

	fuse      *fuse.Handler
	fileStore *torrent.FileItemStore
	cli       *aTorrent.Client
	service   *torrent.Service
}

func newOnlineDownloaderFactory(settings FactorySettings) (Factory, error) {
	pieceCompletionDir := filepath.Join(settings.Fuse.CacheDirectory, "piece-completion")
	if err := os.MkdirAll(pieceCompletionDir, 0744); err != nil {
		return nil, fmt.Errorf("create piece completion directory failed: %w", err)
	}
	fcache, err := filecache.NewCache(filepath.Join(settings.Fuse.CacheDirectory, "cache"))
	if err != nil {
		return nil, fmt.Errorf("create cache failed: %w", err)
	}
	if settings.Fuse.Limit != 0 {
		fcache.SetCapacity(int64(settings.Fuse.Limit) * 1024 * 1024 * 1024)
	}

	torrentStorage := storage.NewResourcePieces(fcache.AsResourceProvider())

	fileStore, err := torrent.NewFileItemStore(filepath.Join(settings.Fuse.CacheDirectory, "items"), cacheTTL)
	if err != nil {
		return nil, fmt.Errorf("create file store failed: %w", err)
	}

	id, err := torrent.GetOrCreatePeerID(filepath.Join(settings.Fuse.CacheDirectory, "ID"))
	if err != nil {
		return nil, fmt.Errorf("create ID failed: %w", err)
	}

	conf := tConfig.TorrentGlobal{
		ReadTimeout:     readTimeout,
		AddTimeout:      addTimeout,
		GlobalCacheSize: -1,
		MetadataFolder:  settings.DataDirectory,
		DisableIPv6:     false,
	}

	cli, err := torrent.NewClient(torrentStorage, fileStore, &conf, id)
	if err != nil {
		return nil, fmt.Errorf("start torrent client failed: %w", err)
	}

	stats := torrent.NewStats()

	loaders := []torrent.DatabaseLoader{} // TODO
	service := torrent.NewService(loaders, stats, cli, conf.AddTimeout, conf.ReadTimeout)

	fss, err := service.Load()
	if err != nil {
		return nil, fmt.Errorf("load torrents failed: %w", err)
	}

	fh := fuse.NewHandler(true, settings.DataDirectory)
	if err = fh.Mount(fss); err != nil {
		return nil, fmt.Errorf("mount fuse directory: %w", err)
	}

	f := onlineDownloaderFactory{
		settings:  settings,
		fuse:      fh,
		fileStore: fileStore,
		cli:       cli,
		service:   service,
	}
	return &f, nil
}

func (f onlineDownloaderFactory) New(t *model.Torrent) (Downloader, error) {
	title, err := f.service.Add(t.ID, t.Content)
	if err != nil {
		return nil, err
	}

	d := onlineDownloader{
		id:    t.ID,
		title: title,
		dir:   f.settings.DataDirectory,
	}

	return &d, nil
}

func (f onlineDownloaderFactory) Close() {
	_ = f.fileStore.Close()
	f.cli.Close()
	f.fuse.Unmount()
}
