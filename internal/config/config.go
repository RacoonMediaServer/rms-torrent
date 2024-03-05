package config

import "github.com/RacoonMediaServer/rms-packages/pkg/configuration"

type Fuse struct {
	Enabled        bool
	CacheDirectory string `json:"cache-directory"`
	Limit          uint
}

type Configuration struct {
	Directory string
	Database  configuration.Database
	Fuse      Fuse
}

var config Configuration

func Config() Configuration {
	return config
}

func Load(filePath string) error {
	return configuration.Load(filePath, &config)
}
