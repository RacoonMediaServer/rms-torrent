package service

import (
	"context"
	"git.rms.local/RacoonMediaServer/rms-shared/pkg/service/rms_torrent"
	"git.rms.local/RacoonMediaServer/rms-torrent/internal/torrent"
	"go-micro.dev/v4/logger"
	"google.golang.org/protobuf/types/known/emptypb"
)

type TorrentService struct {
	manager torrent.Manager
}

func (service *TorrentService) UpPriority(ctx context.Context, request *rms_torrent.UpPriorityRequest, empty *emptypb.Empty) error {
	//TODO implement me
	panic("implement me")
}

func NewService(manager torrent.Manager) *TorrentService {
	return &TorrentService{
		manager: manager,
	}
}

func (service *TorrentService) Download(ctx context.Context, in *rms_torrent.DownloadRequest, out *rms_torrent.DownloadResponse) error {
	logger.Debugf("Download('%+v') request", *in)

	id, files, err := service.manager.Download(in.What)
	if err != nil {
		logger.Errorf("manager download error: %s", err)
		return err
	}

	out.Id = id
	out.Files = files

	return nil
}

func (service *TorrentService) RefreshSettings(ctx context.Context, in *emptypb.Empty, out *emptypb.Empty) error {
	return nil
}

func (service *TorrentService) GetTorrentInfo(ctx context.Context, in *rms_torrent.GetTorrentInfoRequest, out *rms_torrent.TorrentInfo) error {
	result, err := service.manager.GetTorrentInfo(in.Id)
	if err != nil {
		logger.Warnf("Cannot get torrent info: %s", err)
		return err
	}
	*out = result
	return nil
}

func (service *TorrentService) GetTorrents(ctx context.Context, in *rms_torrent.GetTorrentsRequest, out *rms_torrent.GetTorrentsResponse) error {
	out.Torrents = service.manager.GetTorrents(in.IncludeDoneTorrents)
	return nil
}

func (service *TorrentService) RemoveTorrent(ctx context.Context, in *rms_torrent.RemoveTorrentRequest, out *emptypb.Empty) error {
	if err := service.manager.RemoveTorrent(in.Id); err != nil {
		logger.Warnf("cannot remove torrent: %s", err)
		return err
	}
	return nil
}
