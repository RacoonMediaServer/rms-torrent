package qbittorrent

import (
	"strings"
	"time"

	rms_torrent "github.com/RacoonMediaServer/rms-packages/pkg/service/rms-torrent"
	"github.com/RacoonMediaServer/rms-torrent/v4/pkg/qbittorrent"
)

func convertTorrentInfo(src *qbittorrent.TorrentInfo) *rms_torrent.TorrentInfo {
	return &rms_torrent.TorrentInfo{
		Id:            src.Hash,
		Title:         src.Name,
		Progress:      src.Progress * 100.,
		RemainingTime: int64(time.Second * time.Duration(src.Eta)),
		SizeMB:        uint64(float32(src.Size) / (1024. * 1024.)),
		Status:        convertTorrentStatus(src.State),
		Location:      src.ContentPath,
	}
}

func convertTorrentStatus(state string) rms_torrent.Status {
	if strings.HasSuffix(state, "UP") || state == "uploading" {
		return rms_torrent.Status_Done
	}
	if strings.HasSuffix(state, "DL") {
		return rms_torrent.Status_Pending
	}

	if state == "allocating" || state == "downloading" {
		return rms_torrent.Status_Downloading
	}

	if state == "error" || state == "missingFiles" {
		return rms_torrent.Status_Failed
	}
	return rms_torrent.Status_Pending
}
