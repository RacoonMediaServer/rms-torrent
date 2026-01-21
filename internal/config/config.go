package config

import (
	"github.com/RacoonMediaServer/rms-packages/pkg/configuration"
	offline_builtin "github.com/RacoonMediaServer/rms-torrent/pkg/engine/offline/builtin"
	"github.com/RacoonMediaServer/rms-torrent/pkg/engine/offline/qbittorrent"
	online_builtin "github.com/RacoonMediaServer/rms-torrent/pkg/engine/online/builtin"
	"github.com/RacoonMediaServer/rms-torrent/pkg/engine/online/torrserver"
)

type OfflineEngine struct {
	Driver      string
	Qbittorrent qbittorrent.Config
	Builtin     offline_builtin.Config
}

type OnlineEngine struct {
	Driver     string
	TorrServer torrserver.Config
	Builtin    online_builtin.Config
}

type Configuration struct {
	Offline  OfflineEngine
	Online   OnlineEngine
	Database configuration.Database
}

var config Configuration

func Config() Configuration {
	return config
}

func Load(filePath string) error {
	return configuration.Load(filePath, &config)
}
