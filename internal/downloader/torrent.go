package downloader

import (
	"bytes"
	"github.com/anacrolix/torrent"
	"github.com/anacrolix/torrent/metainfo"
	"golang.org/x/time/rate"
)

type torrentSession struct {
	cli   *torrent.Client
	t     *torrent.Torrent
	files []string
}

func newTorrentSession(settings Settings) (Downloader, error) {
	cfg := torrent.NewDefaultClientConfig()
	cfg.DataDir = settings.Destination
	if settings.DownloadLimit != 0 {
		cfg.DownloadRateLimiter = rate.NewLimiter(rate.Limit(settings.DownloadLimit), 1)
	}
	if settings.UploadLimit != 0 {
		cfg.UploadRateLimiter = rate.NewLimiter(rate.Limit(settings.UploadLimit), 1)
	}

	cli, err := torrent.NewClient(cfg)
	if err != nil {
		return nil, err
	}
	var t *torrent.Torrent
	var mi *metainfo.MetaInfo

	isMagnet := isMagnetLink(settings.Input)
	if !isMagnet {
		mi, err = metainfo.Load(bytes.NewReader(settings.Input))
		if err != nil {
			_ = cli.Close()
			return nil, err
		}
		t, err = cli.AddTorrent(mi)
	} else {
		t, err = cli.AddMagnet(string(settings.Input))
	}

	if err != nil {
		_ = cli.Close()
		return nil, err
	}

	<-t.GotInfo()

	s := torrentSession{
		cli: cli,
		t:   t,
	}

	return &s, nil
}

func (s *torrentSession) Start() {
	s.t.DownloadAll()
}

func (s *torrentSession) Files() []string {
	files := s.t.Files()
	result := make([]string, 0, len(files))
	for _, f := range files {
		result = append(result, f.Path())
	}

	return result
}

func (s *torrentSession) Stop() {
	s.t.CancelPieces(0, s.t.NumPieces())
}

func (s *torrentSession) Title() string {
	return s.t.Info().Name
}

func (s *torrentSession) Progress() float32 {
	completed := float64(s.t.BytesCompleted())
	left := float64(s.t.BytesMissing())
	return float32(completed/(completed+left)) * 100
}

func (s *torrentSession) IsComplete() bool {
	return s.t.Complete.Bool()
}

func (s *torrentSession) Close() {
	s.cli.Close()
}

func isMagnetLink(data []byte) bool {
	const magnetLinkSign = "magnet:"
	if len(data) < len(magnetLinkSign) {
		return false
	}
	return string(data[:len(magnetLinkSign)]) == magnetLinkSign
}
