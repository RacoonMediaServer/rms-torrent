package main

import (
	"github.com/micro/cli/v2"
	micro "github.com/micro/go-micro/v2"
	"github.com/micro/go-micro/v2/logger"
	"os"
	"racoondev.tk/gitea/racoon/rms-shared/pkg/configuration"
	"racoondev.tk/gitea/racoon/rms-shared/pkg/db"
	"racoondev.tk/gitea/racoon/rms-torrent/internal/accounts"
	tservice "racoondev.tk/gitea/racoon/rms-torrent/internal/service"
	proto "racoondev.tk/gitea/racoon/rms-torrent/proto"
)

const version = "0.0.2"

type Configuration struct {
	Database  configuration.Database
	Directory string
}

func main() {
	useDebug := false
	config := Configuration{}

	service := micro.NewService(
		micro.Name("rms-torrent"),
		micro.Version(version),
		micro.Flags(
			&cli.BoolFlag{
				Name:        "verbose",
				Aliases:     []string{"debug"},
				Usage:       "debug log level",
				Value:       false,
				Destination: &useDebug,
			},
		),
	)

	service.Init(
		micro.Action(func(context *cli.Context) error {
			configFile := "/etc/rms/rms-torrent.json"
			if context.IsSet("config") {
				configFile = context.String("config")
			}
			return configuration.LoadConfiguration(configFile, &config)
		}),
	)

	if useDebug {
		logger.Init(logger.WithLevel(logger.DebugLevel))
	}

	database, err := db.Connect(config.Database)
	if err != nil {
		logger.Fatal(err)
		os.Exit(1)
	}

	if err := accounts.Load(database); err != nil {
		logger.Errorf("Load torrent accounts failed: %+v", err)
	}

	proto.RegisterRmsTorrentHandler(service.Server(), tservice.NewService(database,config.Directory))

	if err := service.Run(); err != nil {
		logger.Fatal(err)
		os.Exit(2)
	}
}
