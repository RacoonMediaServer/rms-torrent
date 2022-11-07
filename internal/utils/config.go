package utils

import (
	"git.rms.local/RacoonMediaServer/rms-shared/pkg/configuration"
)

type ProxySettings struct {
	Enabled bool
	URL     string
}

type TorrentsSettings struct {
	Directory string
	MaxSpeed  uint
	Db        string
}

type Configuration struct {
	Database configuration.Database
	Proxy    ProxySettings
	Torrents TorrentsSettings
}

var config Configuration

func Config() Configuration {
	return config
}

func LoadConfig(filePath string) error {
	return configuration.LoadConfiguration(filePath, &config)
}

func GetProxyURL() string {
	if !config.Proxy.Enabled {
		return ""
	}

	return config.Proxy.URL
}
