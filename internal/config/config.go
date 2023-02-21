package config

import "github.com/RacoonMediaServer/rms-packages/pkg/configuration"

type TorrentsSettings struct {
	Directory string
	MaxSpeed  uint
	Db        string
}

type Configuration struct {
	Torrents TorrentsSettings
}

var config Configuration

func Config() Configuration {
	return config
}

func Load(filePath string) error {
	return configuration.Load(filePath, &config)
}
