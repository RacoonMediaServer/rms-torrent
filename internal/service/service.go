package service

import (
	"bytes"
	"context"
	"fmt"
	"git.rms.local/RacoonMediaServer/rms-media-discovery/pkg/client/client"
	"git.rms.local/RacoonMediaServer/rms-media-discovery/pkg/client/client/torrents"
	"github.com/go-openapi/strfmt"
	"sync"

	"git.rms.local/RacoonMediaServer/rms-shared/pkg/db"
	"git.rms.local/RacoonMediaServer/rms-shared/pkg/service/rms_torrent"
	"git.rms.local/RacoonMediaServer/rms-torrent/internal/torrent"
	"git.rms.local/RacoonMediaServer/rms-torrent/internal/types"
	"go-micro.dev/v4/logger"
	"google.golang.org/protobuf/types/known/emptypb"

	httptransport "github.com/go-openapi/runtime/client"
)

const remoteApiKey = "1bce398a-0957-453e-b772-da4b4140e0ba"
const discoveryEndpoint = "136.244.108.126"

type TorrentService struct {
	manager *torrent.Manager

	database *db.Database
	mutex    sync.Mutex
}

func NewService(database *db.Database, manager *torrent.Manager) *TorrentService {
	return &TorrentService{
		database: database,
		manager:  manager,
	}
}

func (service *TorrentService) ListTrackers(ctx context.Context, in *rms_torrent.ListTrackersRequest, out *rms_torrent.ListTrackersResponse) error {
	logger.Debug("ListTrackers() request")
	out.Trackers = []*rms_torrent.TrackerInfo{
		&rms_torrent.TrackerInfo{
			Id:            "rms-media-discovery",
			Name:          "Media Discovery",
			LoginRequired: false,
		},
	}
	return nil
}

func get[T any](v *T) T {
	if v == nil {
		return *new(T)
	}
	return *v
}

func (service *TorrentService) Search(ctx context.Context, in *rms_torrent.SearchRequest, out *rms_torrent.SearchResponse) error {
	logger.Debugf("Search('%+v') request", *in)

	tr := httptransport.New(discoveryEndpoint, "", client.DefaultSchemes)
	auth := httptransport.APIKeyAuth("X-Token", "header", remoteApiKey)
	cli := client.New(tr, strfmt.Default)

	typeHint := "movies"
	var limit int64
	limit = 10

	q := &torrents.SearchTorrentsParams{
		Limit:   &limit,
		Q:       in.Text,
		Type:    &typeHint,
		Context: ctx,
	}

	resp, err := cli.Torrents.SearchTorrents(q, auth)
	if err != nil {
		return err
	}
	list := resp.GetPayload().Results

	out.Results = make([]*rms_torrent.Torrent, len(list))
	for i, t := range list {
		result := rms_torrent.Torrent{
			Title: t.Title,
			Link:  get(t.Link),
			Size:  fmt.Sprintf("%d MB", t.Size),
			Peers: int32(t.Seeders),
		}
		out.Results[i] = &result
	}

	return nil
}

func (service *TorrentService) Download(ctx context.Context, in *rms_torrent.DownloadRequest, out *rms_torrent.DownloadResponse) error {
	logger.Debugf("Download('%+v') request", *in)

	tr := httptransport.New(discoveryEndpoint, "", client.DefaultSchemes)
	auth := httptransport.APIKeyAuth("X-Token", "header", remoteApiKey)
	cli := client.New(tr, strfmt.Default)

	req := &torrents.DownloadTorrentParams{
		Link:    in.TorrentLink,
		Context: ctx,
	}

	buf := bytes.NewBuffer([]byte{})

	_, err := cli.Torrents.DownloadTorrent(req, auth, buf)
	if err != nil {
		logger.Errorf("Download failed: %s", err)
		out.ErrorReason = err.Error()
		return err
	}

	_, err = service.manager.Download(buf.Bytes())
	if err != nil {
		logger.Errorf("Manager download error: %s", err)
		out.ErrorReason = err.Error()
		return nil
	}

	out.DownloadStarted = true

	return nil
}

func putError(err error, out *rms_torrent.SearchResponse) {
	e, ok := err.(types.Error)
	if !ok {
		out.Code = rms_torrent.SearchResponse_ERROR
		out.ErrorReason = err.Error()
		logger.Error(err)
		return
	}

	if e.Code == types.CaptchaRequired {
		out.Code = rms_torrent.SearchResponse_CAPTCHA_REQUIRED
		out.CaptchaURL = e.Captcha
	} else {
		out.Code = rms_torrent.SearchResponse_ERROR
	}

	out.ErrorReason = e.Error()
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
