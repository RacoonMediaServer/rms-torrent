package service

import (
	"context"
	rms_torrent "github.com/RacoonMediaServer/rms-packages/pkg/service/rms-torrent"
	"go-micro.dev/v4/logger"
	"google.golang.org/protobuf/types/known/emptypb"
)

type TorrentService struct {
	m DownloadManager
}

func (t TorrentService) Download(ctx context.Context, request *rms_torrent.DownloadRequest, response *rms_torrent.DownloadResponse) error {
	id, files, err := t.m.Download(request.What, request.Description, request.Faster)
	if err != nil {
		logger.Errorf("Download torrent failed: %s", err)
		return err
	}
	response.Files = files
	response.Id = id
	return nil
}

func (t TorrentService) GetTorrentInfo(ctx context.Context, request *rms_torrent.GetTorrentInfoRequest, info *rms_torrent.TorrentInfo) error {
	//TODO implement me
	panic("implement me")
}

func (t TorrentService) GetTorrents(ctx context.Context, request *rms_torrent.GetTorrentsRequest, response *rms_torrent.GetTorrentsResponse) error {
	response.Torrents = t.m.GetDownloads()
	return nil
}

func (t TorrentService) RemoveTorrent(ctx context.Context, request *rms_torrent.RemoveTorrentRequest, empty *emptypb.Empty) error {
	//TODO implement me
	panic("implement me")
}

func (t TorrentService) UpPriority(ctx context.Context, request *rms_torrent.UpPriorityRequest, empty *emptypb.Empty) error {
	//TODO implement me
	panic("implement me")
}

func (t TorrentService) GetSettings(ctx context.Context, empty *emptypb.Empty, settings *rms_torrent.TorrentSettings) error {
	//TODO implement me
	panic("implement me")
}

func (t TorrentService) SetSettings(ctx context.Context, settings *rms_torrent.TorrentSettings, empty *emptypb.Empty) error {
	//TODO implement me
	panic("implement me")
}

func NewService(m DownloadManager) *TorrentService {
	return &TorrentService{
		m: m,
	}
}
