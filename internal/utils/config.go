package utils

import (
	"racoondev.tk/gitea/racoon/rms-shared/pkg/configuration"
)

type ProxySettings struct {
	Enabled bool
	URL     string
}

type Configuration struct {
	Database  configuration.Database
	Directory string
	Proxy     ProxySettings
}

var config Configuration

func Config() Configuration {
	return config
}

func LoadConfig(filePath string) error  {
	return configuration.LoadConfiguration(filePath, &config)
}

func GetProxyURL() string {
	if !config.Proxy.Enabled {
		return ""
	}

	return config.Proxy.URL
}
