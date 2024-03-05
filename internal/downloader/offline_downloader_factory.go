package downloader

import (
	"fmt"
	"github.com/RacoonMediaServer/rms-torrent/internal/model"
	"github.com/anacrolix/torrent"
	"golang.org/x/time/rate"
)

type torrentFactory struct {
	settings FactorySettings
	cli      *torrent.Client
	fastCli  *torrent.Client
}

func newOfflineDownloaderFactory(settings FactorySettings) (Factory, error) {
	f := torrentFactory{settings: settings}
	cfg := torrent.NewDefaultClientConfig()

	cfg.DataDir = settings.DataDirectory
	if settings.DownloadLimit != 0 {
		cfg.DownloadRateLimiter = rate.NewLimiter(rate.Limit(settings.DownloadLimit), 1)
	}
	if settings.UploadLimit != 0 {
		cfg.UploadRateLimiter = rate.NewLimiter(rate.Limit(settings.UploadLimit), 1)
	}

	cli, err := torrent.NewClient(cfg)
	if err != nil {
		return nil, fmt.Errorf("create torrent client failed: %w", err)
	}
	f.cli = cli

	cfg = torrent.NewDefaultClientConfig()
	cfg.DataDir = settings.DataDirectory
	if settings.UploadLimit != 0 {
		cfg.UploadRateLimiter = rate.NewLimiter(rate.Limit(settings.UploadLimit), 1)
	}
	cfg.ListenPort += 1
	cli, err = torrent.NewClient(cfg)
	if err != nil {
		f.cli.Close()
		return nil, fmt.Errorf("create torrent client failed: %w", err)
	}
	f.fastCli = cli

	return &f, nil
}

func (f torrentFactory) New(t *model.Torrent) (Downloader, error) {
	p := downloaderParameters{
		settings: f.settings,
		t:        t,
	}
	if t.Fast && !t.Complete {
		return newOfflineDownloader(f.fastCli, &p)
	}
	return newOfflineDownloader(f.cli, &p)
}

func (f torrentFactory) Close() {
	f.cli.Close()
	f.fastCli.Close()
}
