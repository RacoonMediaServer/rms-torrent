package trackers

import (
	"git.rms.local/RacoonMediaServer/rms-shared/pkg/service/rms_torrent"
	"git.rms.local/RacoonMediaServer/rms-torrent/internal/trackers/rutor"
	"git.rms.local/RacoonMediaServer/rms-torrent/internal/trackers/rutracker"
	"git.rms.local/RacoonMediaServer/rms-torrent/internal/types"
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

func ListTrackers() []*rms_torrent.TrackerInfo {
	list := make([]*rms_torrent.TrackerInfo, 0)
	for k, v := range trackers {
		info := rms_torrent.TrackerInfo{
			Id:            k,
			Name:          v.Name,
			LoginRequired: v.LoginRequired,
		}

		list = append(list, &info)
	}

	return list
}
