package service

import (
	"context"
	"encoding/base64"
	"racoondev.tk/gitea/racoon/rtorrent/internal/trackers"
	"racoondev.tk/gitea/racoon/rtorrent/internal/types"
	proto "racoondev.tk/gitea/racoon/rtorrent/proto"
	"sync"
)

type TorrentService struct {
	sessions map[string]types.SearchSession
	mutex sync.Mutex
}

func NewService() *TorrentService {
	return &TorrentService{
		sessions: make(map[string]types.SearchSession),
	}
}

func putError(err error, out *proto.SearchResponse) {
	e, ok := err.(types.Error)
	if !ok {
		out.Code = proto.SearchResponse_ERROR
		out.ErrorReason = err.Error()
		return
	}

	if e.Code == types.CaptchaRequired {
		out.Code = proto.SearchResponse_CAPTCHA_REQUIRED
		out.CaptchaURL = e.Captcha
	} else {
		out.Code = proto.SearchResponse_ERROR
	}

	out.ErrorReason = e.Error()
}

func (service *TorrentService) getSession(id, user, password, tracker string) (types.SearchSession, error) {
	service.mutex.Lock()
	defer service.mutex.Unlock()

	var err error
	session, ok := service.sessions[id]
	if !ok {
		session, err = trackers.NewSession(tracker)
		if session != nil {
			session.Setup(types.SessionSettings{
				User:      user,
				Password:  password,
				UserAgent: "RacoonMediaServer",
			})
		}
	} else {
		delete(service.sessions, id)
	}

	return session, err
}

func (service *TorrentService) putSession(id string, session types.SearchSession) {
	service.mutex.Lock()
	defer service.mutex.Unlock()

	service.sessions[id] = session
}

func (service *TorrentService) Search(ctx context.Context, in *proto.SearchRequest, out *proto.SearchResponse) error {
	id := base64.StdEncoding.EncodeToString([]byte(in.Tracker+":"+in.Login+":"+in.Password))
	session, err := service.getSession(id, in.Login, in.Password, in.Tracker)

	if err != nil {
		putError(err, out)
		return nil
	}

	defer service.putSession(id, session)

	if in.Captcha != "" {
		session.SetCaptchaText(in.Captcha)
	}

	torrents, err := session.Search(in.Text, 10)
	if err != nil {
		putError(err, out)
		return nil
	}

	out.Results = make([]*proto.Torrent, len(torrents))
	for i, t := range torrents {
		result := proto.Torrent{
			Title:                t.Title,
			Link:                 t.DownloadLink,
			Size:                 t.Size,
			Quality:              "",
		}
		out.Results[i] = &result
	}

	out.SessionID = id

	return nil
}

func (service *TorrentService) Download(ctx context.Context, in *proto.DownloadRequest, out *proto.DownloadResponse) error {
	return nil
}
