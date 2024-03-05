package downloader

import (
	"bytes"
	"github.com/anacrolix/torrent"
	"github.com/anacrolix/torrent/metainfo"
	"github.com/anacrolix/torrent/storage"
	"path"
)

type offlineDownloader struct {
	t *torrent.Torrent
}

func newOfflineDownloader(cli *torrent.Client, p *downloaderParameters) (Downloader, error) {
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

	return &offlineDownloader{t: t}, nil
}

func (d *offlineDownloader) Start() {
	d.t.DownloadAll()
}

func (d *offlineDownloader) Files() []string {
	files := d.t.Files()
	result := make([]string, 0, len(files))
	for _, f := range files {
		result = append(result, f.Path())
	}
	d.t.Info().TotalLength()

	return result
}

func (d *offlineDownloader) Stop() {
	d.t.CancelPieces(0, d.t.NumPieces())
}

func (d *offlineDownloader) Title() string {
	return d.t.Info().Name
}

func (d *offlineDownloader) Bytes() uint64 {
	return uint64(d.t.BytesCompleted())
}

func (d *offlineDownloader) RemainingBytes() uint64 {
	return uint64(d.t.BytesMissing())
}

func (d *offlineDownloader) IsComplete() bool {
	return d.t.Complete.Bool()
}

func (d *offlineDownloader) Close() {
	d.t.Drop()
}

func isMagnetLink(data []byte) bool {
	const magnetLinkSign = "magnet:"
	if len(data) < len(magnetLinkSign) {
		return false
	}
	return string(data[:len(magnetLinkSign)]) == magnetLinkSign
}

func (d *offlineDownloader) SizeMB() uint64 {
	var size uint64
	for i := range d.t.Files() {
		size += uint64(d.t.Files()[i].Length())
	}
	return size / (1024. * 1024.)
}
