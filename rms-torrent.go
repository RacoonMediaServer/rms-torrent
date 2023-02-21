package main

import (
	"github.com/RacoonMediaServer/rms-packages/pkg/pubsub"
	rms_torrent "github.com/RacoonMediaServer/rms-packages/pkg/service/rms-torrent"
	"github.com/RacoonMediaServer/rms-torrent/internal/config"
	tservice "github.com/RacoonMediaServer/rms-torrent/internal/service"
	"github.com/RacoonMediaServer/rms-torrent/internal/torrent"
	"github.com/urfave/cli/v2"
	micro "go-micro.dev/v4"
	"go-micro.dev/v4/logger"
)

const Version = "0.0.0"

func main() {
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

	manager, err := torrent.New(config.Config().Torrents, pubsub.NewPublisher(service))
	if err != nil {
		logger.Fatalf("cannot start manager: %s", err)
	}

	if err = rms_torrent.RegisterRmsTorrentHandler(service.Server(), tservice.NewService(manager)); err != nil {
		logger.Fatalf("Cannot initialize service handler: %s", err)
	}

	if err := service.Run(); err != nil {
		logger.Fatal(err)
	}

	manager.Stop()
}
