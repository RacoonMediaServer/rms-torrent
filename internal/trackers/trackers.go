package trackers

import (
	"racoondev.tk/gitea/racoon/rms-torrent/internal/trackers/rutor"
	"racoondev.tk/gitea/racoon/rms-torrent/internal/trackers/rutracker"
	"racoondev.tk/gitea/racoon/rms-torrent/internal/types"
	proto "racoondev.tk/gitea/racoon/rms-torrent/proto"
)

type tracker struct {
	Name          string
	Factory       func() types.SearchSession
	LoginRequired bool
}

var (
	trackers = map[string]*tracker{
		"rutracker": &tracker{
			Name: "RuTracker.org",
			Factory: func() types.SearchSession {
				return new(rutracker.SearchSession)
			},
			LoginRequired: true,
		},
		"rutor": &tracker{
			Name: "RUTOR.org",
			Factory: func() types.SearchSession {
				return new(rutor.SearchSession)
			},
			LoginRequired: false,
		},
	}
)

func NewSession(trackerID string) (types.SearchSession, error) {
	tracker, ok := trackers[trackerID]
	if !ok {
		return nil, types.Raise(types.UnknownTracker)
	}

	return tracker.Factory(), nil
}

func ListTrackers() []*proto.TrackerInfo {
	list := make([]*proto.TrackerInfo,0)
	for k,v := range trackers {
		info := proto.TrackerInfo{
			Id:            k,
			Name:          v.Name,
			LoginRequired: v.LoginRequired,
		}

		list = append(list, &info)
	}

	return list
}
