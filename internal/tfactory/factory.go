package tfactory

import (
	"fmt"

	"github.com/RacoonMediaServer/rms-torrent/internal/config"
	"github.com/RacoonMediaServer/rms-torrent/pkg/engine"
	offline_builtin "github.com/RacoonMediaServer/rms-torrent/pkg/engine/offline/builtin"
	"github.com/RacoonMediaServer/rms-torrent/pkg/engine/offline/qbittorrent"
	online_builtin "github.com/RacoonMediaServer/rms-torrent/pkg/engine/online/builtin"
	"github.com/RacoonMediaServer/rms-torrent/pkg/engine/online/torrserver"
)

func CreateEngine(onlineMode bool, cfg config.Configuration, completeAction engine.CompleteAction) (result engine.TorrentEngine, err error) {
	if onlineMode {
		return CreateOnlineEngine(cfg.Online)
	}
	return CreateOfflineEngine(cfg.Offline, completeAction)
}

func CreateOfflineEngine(cfg config.OfflineEngine, completeAction engine.CompleteAction) (result engine.TorrentEngine, err error) {
	switch cfg.Driver {
	case "qbittorrent":
		result, err = qbittorrent.NewTorrentEngine(cfg.Qbittorrent, completeAction)
	case "builtin":
		result, err = offline_builtin.NewTorrentEngine(cfg.Builtin, &engine.VoidDatabase{}, completeAction)
	default:
		err = fmt.Errorf("unknown driver: %s", cfg.Driver)
	}

	return
}

func CreateOnlineEngine(cfg config.OnlineEngine) (result engine.TorrentEngine, err error) {
	switch cfg.Driver {
	case "torrserver":
		result, err = torrserver.NewEngine(cfg.TorrServer)
	case "builtin":
		result, err = online_builtin.NewEngine(cfg.Builtin, &engine.VoidDatabase{})
	default:
		err = fmt.Errorf("unknown driver: %s", cfg.Driver)
	}

	return
}
