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

func newTorrentSession(cli *torrent.Client, p *downloaderParameters) (Downloader, error) {
	var spec *torrent.TorrentSpec
	isMagnet := isMagnetLink(p.t.Content)
	if !isMagnet {
		mi, err := metainfo.Load(bytes.NewReader(p.t.Content))
		if err != nil {
			return nil, err
		}
		spec = torrent.TorrentSpecFromMetaInfo(mi)
	} else {
		var err error
		spec, err = torrent.TorrentSpecFromMagnetUri(string(p.t.Content))
		if err != nil {
			return nil, err
		}
	}

	opts := torrent.AddTorrentOpts{
		InfoHash:  spec.InfoHash,
		Storage:   storage.NewFile(path.Join(p.settings.DataDirectory, p.t.ID)),
		ChunkSize: spec.ChunkSize,
	}

	t, _ := cli.AddTorrentOpt(opts)
	if err := t.MergeSpec(spec); err != nil {
		t.Drop()
		return nil, err
	}
	<-t.GotInfo()

	t.AllowDataUpload()

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
	s.t.Info().TotalLength()

	return result
}

func (s *torrentSession) Stop() {
	s.t.CancelPieces(0, s.t.NumPieces())
}

func (s *torrentSession) Title() string {
	return s.t.Info().Name
}

func (s *torrentSession) Bytes() uint64 {
	return uint64(s.t.BytesCompleted())
}

func (s *torrentSession) RemainingBytes() uint64 {
	return uint64(s.t.BytesMissing())
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

func (s *torrentSession) SizeMB() uint64 {
	var size uint64
	for i := range s.t.Files() {
		size += uint64(s.t.Files()[i].Length())
	}
	return size / (1024. * 1024.)
}
