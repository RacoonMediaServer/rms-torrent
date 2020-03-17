package trackers

import (
	"racoondev.tk/gitea/racoon/rtorrent/internal/trackers/rutor"
	"racoondev.tk/gitea/racoon/rtorrent/internal/trackers/rutracker"
	"racoondev.tk/gitea/racoon/rtorrent/internal/types"
)

func NewSession(tracker string) (types.SearchSession, error) {
	switch tracker {
	case "rutracker":
		return new(rutracker.SearchSession), nil
	case "rutor":
		return new(rutor.SearchSession), nil
	default:
		return nil, types.Error{Code: types.UnknownTracker}
	}
}
