package tfactory

import (
	"fmt"

	"github.com/RacoonMediaServer/rms-torrent/internal/config"
	"github.com/RacoonMediaServer/rms-torrent/internal/db"
	"github.com/RacoonMediaServer/rms-torrent/pkg/engine"
	offline_builtin "github.com/RacoonMediaServer/rms-torrent/pkg/engine/offline/builtin"
	"github.com/RacoonMediaServer/rms-torrent/pkg/engine/offline/qbittorrent"
	online_builtin "github.com/RacoonMediaServer/rms-torrent/pkg/engine/online/builtin"
	"github.com/RacoonMediaServer/rms-torrent/pkg/engine/online/torrserver"
)

func CreateEngine(onlineMode bool, dbase *db.Database, cfg config.Configuration, completeAction engine.CompleteAction) (result engine.TorrentEngine, err error) {
	var tdb engine.TorrentDatabase = &engine.VoidDatabase{}
	if dbase != nil {
		tdb = dbase
	}
	if onlineMode {
		return createOnlineEngine(cfg.Online, tdb)
	}
	return createOfflineEngine(cfg.Offline, tdb, completeAction)
}

func createOfflineEngine(cfg config.OfflineEngine, dbase engine.TorrentDatabase, completeAction engine.CompleteAction) (result engine.TorrentEngine, err error) {
	switch cfg.Driver {
	case "qbittorrent":
		result, err = qbittorrent.NewTorrentEngine(cfg.Qbittorrent, completeAction)
	case "builtin":
		result, err = offline_builtin.NewTorrentEngine(cfg.Builtin, dbase, completeAction)
	default:
		err = fmt.Errorf("unknown driver: %s", cfg.Driver)
	}

	return
}

func createOnlineEngine(cfg config.OnlineEngine, dbase engine.TorrentDatabase) (result engine.TorrentEngine, err error) {
	switch cfg.Driver {
	case "torrserver":
		result, err = torrserver.NewEngine(cfg.TorrServer)
	case "builtin":
		result, err = online_builtin.NewEngine(cfg.Builtin, dbase)
	default:
		err = fmt.Errorf("unknown driver: %s", cfg.Driver)
	}

	return
}
