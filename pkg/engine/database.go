package engine

type TorrentRecord struct {
	TorrentDescription
	Content []byte
}

type TorrentDatabase interface {
	Add(t TorrentRecord) error
	Load() ([]TorrentRecord, error)
	Del(id string) error
}
