package main

import (
	rms_torrent "github.com/RacoonMediaServer/rms-packages/pkg/service/rms-torrent"
	"github.com/RacoonMediaServer/rms-torrent/internal/config"
	"github.com/RacoonMediaServer/rms-torrent/internal/db"
	"github.com/RacoonMediaServer/rms-torrent/internal/downloader"
	"github.com/RacoonMediaServer/rms-torrent/internal/downloads"
	tservice "github.com/RacoonMediaServer/rms-torrent/internal/service"
	"github.com/urfave/cli/v2"
	micro "go-micro.dev/v4"
	"go-micro.dev/v4/logger"
)

const Version = "0.0.0"

func main() {
	logger.Infof("rms-torrent v%s", Version)
	useDebug := false

	service := micro.NewService(
		micro.Name("rms-torrent"),
		micro.Version(Version),
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
			return config.Load(configFile)
		}),
	)

	if useDebug {
		_ = logger.Init(logger.WithLevel(logger.DebugLevel))
	}

	cfg := config.Config()
	database, err := db.Connect(cfg.Database)
	if err != nil {
		logger.Fatalf("Connect to database failed: %s", err)
	}

	downloaderFactory, err := downloader.NewFactory(downloader.FactorySettings{
		DataDirectory: cfg.Directory,
	})
	if err != nil {
		logger.Fatalf("Create downloader factory failed: %s", err)
	}
	manager := downloads.NewManager(downloaderFactory, database)

	if err := rms_torrent.RegisterRmsTorrentHandler(service.Server(), tservice.NewService(manager)); err != nil {
		logger.Fatalf("Cannot initialize service handler: %s", err)
	}

	if err := service.Run(); err != nil {
		logger.Fatal(err)
	}
}
