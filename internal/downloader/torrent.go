package downloader

import (
	"bytes"
	"github.com/anacrolix/torrent"
	"github.com/anacrolix/torrent/metainfo"
	"github.com/anacrolix/torrent/storage"
	"path"
)

type torrentSession struct {
	t *torrent.Torrent
}

func newTorrentSession(cli *torrent.Client, input []byte, p *downloaderParameters) (Downloader, error) {
	var spec *torrent.TorrentSpec
	isMagnet := isMagnetLink(input)
	if !isMagnet {
		mi, err := metainfo.Load(bytes.NewReader(input))
		if err != nil {
			return nil, err
		}
		spec = torrent.TorrentSpecFromMetaInfo(mi)
	} else {
		var err error
		spec, err = torrent.TorrentSpecFromMagnetUri(string(input))
		if err != nil {
			return nil, err
		}
	}

	opts := torrent.AddTorrentOpts{
		InfoHash:  spec.InfoHash,
		Storage:   storage.NewFile(path.Join(p.settings.DataDirectory, p.subDirectory)),
		ChunkSize: spec.ChunkSize,
	}

	t, _ := cli.AddTorrentOpt(opts)
	<-t.GotInfo()

	return &torrentSession{t: t}, nil
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
	s.t.Drop()
}

func isMagnetLink(data []byte) bool {
	const magnetLinkSign = "magnet:"
	if len(data) < len(magnetLinkSign) {
		return false
	}
	return string(data[:len(magnetLinkSign)]) == magnetLinkSign
}
