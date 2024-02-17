package main

import (
	"github.com/RacoonMediaServer/rms-packages/pkg/pubsub"
	rms_torrent "github.com/RacoonMediaServer/rms-packages/pkg/service/rms-torrent"
	"github.com/RacoonMediaServer/rms-torrent/internal/config"
	"github.com/RacoonMediaServer/rms-torrent/internal/db"
	tservice "github.com/RacoonMediaServer/rms-torrent/internal/service"
	"github.com/urfave/cli/v2"
	micro "go-micro.dev/v4"
	"go-micro.dev/v4/logger"

	// Plugins
	_ "github.com/go-micro/plugins/v4/registry/etcd"
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

	tService := tservice.NewService(database, pubsub.NewPublisher(service))

	if err = tService.Initialize(); err != nil {
		logger.Fatalf("Initialize service failed: %s", err)
	}

	if err = rms_torrent.RegisterRmsTorrentHandler(service.Server(), tService); err != nil {
		logger.Fatalf("Cannot initialize service handler: %s", err)
	}

	if err = service.Run(); err != nil {
		logger.Fatal(err)
	}
}
