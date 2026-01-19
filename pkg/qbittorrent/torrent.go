package qbittorrent

type TorrentInfo struct {
	Hash     string
	Eta      int64
	Name     string
	Progress float32
	Size     uint64
	State    string
}
