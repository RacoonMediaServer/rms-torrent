package service

import (
	"context"
	"errors"

	rms_torrent "github.com/RacoonMediaServer/rms-packages/pkg/service/rms-torrent"
	"github.com/RacoonMediaServer/rms-torrent/v4/internal/config"
	"github.com/RacoonMediaServer/rms-torrent/v4/pkg/engine"
	"go-micro.dev/v4/logger"
	"google.golang.org/protobuf/types/known/emptypb"
)

type Service struct {
	e engine.TorrentEngine
}

// Download implements rms_torrent.RmsTorrentHandler.
func (s *Service) Download(ctx context.Context, req *rms_torrent.DownloadRequest, resp *rms_torrent.DownloadResponse) error {
	added, err := s.e.Add(ctx, req.Category, req.Description, req.ForceLocation, req.What)
	if err != nil {
		logger.Errorf("Add torrent failed: %s", err)
		return err
	}

	resp.Id = added.ID
	resp.Files = added.Files
	resp.Title = added.Title
	resp.Location = added.Location

	logger.Infof("Torrent '%s' [%s, %d files] added (loc: %s)", added.Title, added.ID, len(added.Files), added.Location)
	return nil
}

// GetSettings implements rms_torrent.RmsTorrentHandler.
func (s *Service) GetSettings(context.Context, *emptypb.Empty, *rms_torrent.TorrentSettings) error {
	// TODO: implement
	return errors.ErrUnsupported
}

// GetTorrentInfo implements rms_torrent.RmsTorrentHandler.
func (s *Service) GetTorrentInfo(ctx context.Context, req *rms_torrent.GetTorrentInfoRequest, resp *rms_torrent.TorrentInfo) error {
	result, err := s.e.Get(ctx, req.Id)
	if err != nil {
		logger.Errorf("Get torrent info of %s failed: %s", req.Id, err)
		return err
	}

	*resp = *result
	return nil
}

// GetTorrents implements rms_torrent.RmsTorrentHandler.
func (s *Service) GetTorrents(ctx context.Context, req *rms_torrent.GetTorrentsRequest, resp *rms_torrent.GetTorrentsResponse) error {
	result, err := s.e.List(ctx, req.IncludeDoneTorrents)
	if err != nil {
		logger.Errorf("Get torrents failed: %s", err)
		return err
	}

	resp.Torrents = result
	return nil
}

// RemoveTorrent implements rms_torrent.RmsTorrentHandler.
func (s *Service) RemoveTorrent(ctx context.Context, req *rms_torrent.RemoveTorrentRequest, empty *emptypb.Empty) error {
	if err := s.e.Remove(ctx, req.Id); err != nil {
		logger.Errorf("Remove torrent %s failed: %s", req.Id, err)
		return err
	}
	logger.Infof("Torrent %s removed", req.Id)
	return nil
}

// SetSettings implements rms_torrent.RmsTorrentHandler.
func (s *Service) SetSettings(context.Context, *rms_torrent.TorrentSettings, *emptypb.Empty) error {
	// TODO: implement
	return errors.ErrUnsupported
}

// UpPriority implements rms_torrent.RmsTorrentHandler.
func (s *Service) UpPriority(ctx context.Context, req *rms_torrent.UpPriorityRequest, empty *emptypb.Empty) error {
	return s.e.UpPriority(ctx, req.Id)
}

func NewService(cfg config.Configuration, e engine.TorrentEngine) *Service {
	return &Service{e: e}
}
