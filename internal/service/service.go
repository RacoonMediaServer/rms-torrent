package service

import (
	"context"
	"encoding/base64"
	"path"
	"sort"
	"sync"

	"git.rms.local/RacoonMediaServer/rms-shared/pkg/db"
	"git.rms.local/RacoonMediaServer/rms-shared/pkg/service/rms_torrent"
	"git.rms.local/RacoonMediaServer/rms-torrent/internal/accounts"
	"git.rms.local/RacoonMediaServer/rms-torrent/internal/trackers"
	"git.rms.local/RacoonMediaServer/rms-torrent/internal/types"
	"git.rms.local/RacoonMediaServer/rms-torrent/internal/utils"
	uuid "github.com/satori/go.uuid"
	"go-micro.dev/v4/logger"
	"google.golang.org/protobuf/types/known/emptypb"
)

type TorrentService struct {
	sessions map[string]types.SearchSession
	database *db.Database
	mutex    sync.Mutex
}

func NewService(database *db.Database) *TorrentService {
	return &TorrentService{
		sessions: make(map[string]types.SearchSession),
		database: database,
	}
}

func (service *TorrentService) ListTrackers(ctx context.Context, in *rms_torrent.ListTrackersRequest, out *rms_torrent.ListTrackersResponse) error {
	logger.Debug("ListTrackers() request")
	out.Trackers = trackers.ListTrackers()
	return nil
}

func (service *TorrentService) Search(ctx context.Context, in *rms_torrent.SearchRequest, out *rms_torrent.SearchResponse) error {
	logger.Debugf("Search('%+v') request", *in)

	login, password := accounts.Get(in.Tracker)

	service.mutex.Lock()
	defer service.mutex.Unlock()

	id := base64.StdEncoding.EncodeToString([]byte(in.Tracker + ":" + login + ":" + password))
	session, err := service.getSession(id, login, password, in.Tracker)

	if err != nil {
		putError(err, out)
		return nil
	}

	if in.Captcha != "" {
		session.SetCaptchaText(in.Captcha)
	}

	torrents, err := session.Search(in.Text)
	if err != nil {
		putError(err, out)
		return nil
	}

	sort.Slice(torrents, func(i, j int) bool {
		return torrents[i].Peers > torrents[j].Peers
	})

	out.Results = make([]*rms_torrent.Torrent, len(torrents))
	for i, t := range torrents {
		result := rms_torrent.Torrent{
			Title: t.Title,
			Link:  t.DownloadLink,
			Size:  t.Size,
			Peers: int32(t.Peers),
		}
		out.Results[i] = &result
	}

	out.SessionID = id

	return nil
}

func (service *TorrentService) Download(ctx context.Context, in *rms_torrent.DownloadRequest, out *rms_torrent.DownloadResponse) error {
	logger.Debugf("Download('%+v') request", *in)

	service.mutex.Lock()
	defer service.mutex.Unlock()

	out.DownloadStarted = false

	session, ok := service.sessions[in.SessionID]
	if !ok {
		logger.Errorf("Unknown session: %s", in.SessionID)
		out.ErrorReason = "Unknown session"
		return nil
	}

	file := path.Join(utils.Config().Directory, uuid.NewV4().String()+".torrent")

	if err := session.Download(in.TorrentLink, file); err != nil {
		out.ErrorReason = err.Error()
		return nil
	}

	logger.Infof("Torrent downloaded: %s", file)

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

func (service *TorrentService) getSession(id, user, password, tracker string) (types.SearchSession, error) {
	var err error
	session, ok := service.sessions[id]
	if !ok {
		session, err = trackers.NewSession(tracker)
		if session != nil {
			session.Setup(types.SessionSettings{
				User:      user,
				Password:  password,
				UserAgent: "RacoonMediaServer",
				ProxyURL:  utils.GetProxyURL(),
			})
			service.sessions[id] = session
		}
	}

	return session, err
}

func (service *TorrentService) RefreshSettings(ctx context.Context, in *emptypb.Empty, out *emptypb.Empty) error {
	if err := accounts.Load(service.database); err != nil {
		logger.Errorf("Update torrent accounts failed: %+v", err)
	}
	return nil
}
