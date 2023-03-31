package service

import rms_torrent "github.com/RacoonMediaServer/rms-packages/pkg/service/rms-torrent"

type DownloadManager interface {
	Download(content []byte, description string, faster bool) (string, []string, error)
	GetDownloads() []*rms_torrent.TorrentInfo
	GetDownloadInfo(id string) (*rms_torrent.TorrentInfo, error)
	RemoveDownload(id string) error
	UpDownload(id string) error
}
