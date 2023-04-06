package downloader

import (
	"fmt"
	"github.com/anacrolix/torrent"
	"golang.org/x/time/rate"
)

// Factory is Downloader factory
type Factory struct {
	settings FactorySettings
	cli      *torrent.Client
}

type downloaderParameters struct {
	settings     FactorySettings
	subDirectory string
	noRateLimit  bool
}

// FactorySettings are parameters for constructing Downloader's
type FactorySettings struct {
	DataDirectory string
	UploadLimit   uint64
	DownloadLimit uint64
}

func NewFactory(settings FactorySettings) (*Factory, error) {
	f := Factory{settings: settings}
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

	return &f, nil
}

func (f Factory) New(subDirectory string, noRateLimit bool, content []byte) (Downloader, error) {
	p := downloaderParameters{
		settings:     f.settings,
		subDirectory: subDirectory,
		noRateLimit:  noRateLimit,
	}
	return newTorrentSession(f.cli, content, &p)
}

func (f Factory) Close() {
	f.cli.Close()
}
