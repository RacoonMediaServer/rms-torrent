package config

import "github.com/RacoonMediaServer/rms-packages/pkg/configuration"

type Configuration struct {
	Directory string
	Database  configuration.Database
}

var config Configuration

func Config() Configuration {
	return config
}

func Load(filePath string) error {
	return configuration.Load(filePath, &config)
}
