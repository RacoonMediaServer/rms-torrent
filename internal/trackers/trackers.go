package trackers

import (
	"git.rms.local/RacoonMediaServer/rms-torrent/internal/trackers/rutor"
	"git.rms.local/RacoonMediaServer/rms-torrent/internal/trackers/rutracker"
	"git.rms.local/RacoonMediaServer/rms-torrent/internal/types"
	proto "git.rms.local/RacoonMediaServer/rms-torrent/proto"
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
	list := make([]*proto.TrackerInfo, 0)
	for k, v := range trackers {
		info := proto.TrackerInfo{
			Id:            k,
			Name:          v.Name,
			LoginRequired: v.LoginRequired,
		}

		list = append(list, &info)
	}

	return list
}
