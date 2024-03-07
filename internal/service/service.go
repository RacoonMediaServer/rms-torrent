package service

import (
	"context"
	"fmt"
	"github.com/RacoonMediaServer/rms-packages/pkg/events"
	"github.com/RacoonMediaServer/rms-packages/pkg/misc"
	rms_torrent "github.com/RacoonMediaServer/rms-packages/pkg/service/rms-torrent"
	"github.com/RacoonMediaServer/rms-torrent/internal/config"
	"github.com/RacoonMediaServer/rms-torrent/internal/downloads"
	"github.com/RacoonMediaServer/rms-torrent/internal/model"
	uuid "github.com/satori/go.uuid"
	"go-micro.dev/v4"
	"go-micro.dev/v4/logger"
	"google.golang.org/protobuf/types/known/emptypb"
	"time"
)

type TorrentService struct {
	db  Database
	m   *downloads.Manager
	pub micro.Event
}

const publishTimeout = 10 * time.Second

func NewService(db Database, pub micro.Event) *TorrentService {
	return &TorrentService{
		db:  db,
		pub: pub,
	}
}

func (t *TorrentService) Initialize() error {
	settings, err := t.db.LoadSettings()
	if err != nil {
		return fmt.Errorf("load settings failed: %w", err)
	}

	f, err := newFactory(settings)
	if err != nil {
		return fmt.Errorf("create downloader factory failed: %w", err)
	}

	torrents, err := t.db.LoadTorrents()
	if err != nil {
		return fmt.Errorf("load torrents failed: %w", err)
	}

	t.m = downloads.NewManager(f)
	t.m.OnDownloadComplete = func(ctx context.Context, torrentTitle string, tm *model.Torrent) {
		if err := t.db.CompleteTorrent(tm.ID); err != nil {
			logger.Warnf("Update torrent status %s in database failed: %s", tm.ID, err)
		}
		description := torrentTitle
		if tm.Description != "" {
			description = fmt.Sprintf("%s (%s)", tm.Description, torrentTitle)
		}
		if !config.Config().Fuse.Enabled {
			t.publish(ctx, &events.Notification{
				Kind:      events.Notification_DownloadComplete,
				TorrentID: &tm.ID,
				ItemTitle: &description,
			})
		}
	}
	t.m.OnMalfunction = func(text string, code events.Malfunction_Code) {
		t.publish(context.Background(), &events.Malfunction{
			Timestamp:  time.Now().Unix(),
			Error:      text,
			System:     events.Malfunction_Media,
			Code:       code,
			StackTrace: misc.GetStackTrace(),
		})
	}

	for _, torrent := range torrents {
		_, err = t.m.Download(torrent)
		if err != nil {
			logger.Errorf("Resume torrent downloading '%s' failed: %s", torrent.Description, err)
		}
	}
	return nil
}

func (t *TorrentService) Download(ctx context.Context, request *rms_torrent.DownloadRequest, response *rms_torrent.DownloadResponse) error {
	record := model.Torrent{
		ID:          uuid.NewV4().String(),
		Content:     request.What,
		Fast:        request.Faster,
		Description: request.Description,
	}

	files, err := t.m.Download(&record)
	if err != nil {
		logger.Errorf("Download torrent failed: %s", err)
		return err
	}

	if err = t.db.AddTorrent(&record); err != nil {
		_ = t.m.RemoveTorrent(record.ID)
		logger.Errorf("Register torrent to database failed: %s", record.ID)
		return err
	}

	response.Files = files
	response.Id = record.ID
	return nil
}

func (t *TorrentService) GetTorrentInfo(ctx context.Context, request *rms_torrent.GetTorrentInfoRequest, info *rms_torrent.TorrentInfo) error {
	i, err := t.m.GetTorrentInfo(request.Id)
	if err != nil {
		logger.Errorf("Get info about '%s' failed: %s", request.Id, err)
		return err
	}
	*info = *i
	return nil
}

func (t *TorrentService) GetTorrents(ctx context.Context, request *rms_torrent.GetTorrentsRequest, response *rms_torrent.GetTorrentsResponse) error {
	response.Torrents = t.m.GetTorrents(request.IncludeDoneTorrents)
	return nil
}

func (t *TorrentService) RemoveTorrent(ctx context.Context, request *rms_torrent.RemoveTorrentRequest, empty *emptypb.Empty) error {
	if err := t.db.RemoveTorrent(request.Id); err != nil {
		logger.Errorf("Remove torrent %s from database failed: %s", request.Id, err)
		return err
	}
	if err := t.m.RemoveTorrent(request.Id); err != nil {
		logger.Errorf("Remove torrent %s failed: %s", request.Id, err)
		return err
	}
	return nil
}

func (t *TorrentService) UpPriority(ctx context.Context, request *rms_torrent.UpPriorityRequest, empty *emptypb.Empty) error {
	if err := t.m.UpDownload(request.Id); err != nil {
		logger.Errorf("Up priority for %s failed: %s", request.Id, err)
		return err
	}
	return nil
}

func (t *TorrentService) GetSettings(ctx context.Context, empty *emptypb.Empty, settings *rms_torrent.TorrentSettings) error {
	stored, err := t.db.LoadSettings()
	if err != nil {
		logger.Errorf("Load settings failed: %s", err)
		return err
	}

	settings.UploadLimit = stored.UploadLimit
	settings.DownloadLimit = stored.DownloadLimit

	return nil
}

func (t *TorrentService) SetSettings(ctx context.Context, settings *rms_torrent.TorrentSettings, empty *emptypb.Empty) error {
	if err := t.db.SaveSettings(settings); err != nil {
		logger.Errorf("Save settings failed: %s", err)
		return err
	}

	t.m.Stop()

	f, err := newFactory(settings)
	if err != nil {
		logger.Fatalf("Cannot recreate downloader factory: %s", err)
		return err
	}

	t.m.Reset(f)
	return nil
}

func (t *TorrentService) publish(ctx context.Context, event interface{}) {
	ctx, cancel := context.WithTimeout(ctx, publishTimeout)
	defer cancel()

	if err := t.pub.Publish(ctx, event); err != nil {
		logger.Warnf("Publish notification failed: %s", err)
	}
}
