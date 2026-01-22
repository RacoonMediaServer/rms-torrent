package engine

type TorrentRecord struct {
	TorrentDescription
	Content  []byte
	Complete bool
}

type TorrentDatabase interface {
	Add(t TorrentRecord) error
	Load() ([]TorrentRecord, error)
	Complete(id string) error
	Del(id string) error
}
