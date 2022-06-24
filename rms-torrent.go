package main

import (
	"os"

	"git.rms.local/RacoonMediaServer/rms-shared/pkg/db"
	"git.rms.local/RacoonMediaServer/rms-shared/pkg/service/rms_torrent"
	"git.rms.local/RacoonMediaServer/rms-torrent/internal/accounts"
	tservice "git.rms.local/RacoonMediaServer/rms-torrent/internal/service"
	"git.rms.local/RacoonMediaServer/rms-torrent/internal/utils"
	"github.com/urfave/cli/v2"
	micro "go-micro.dev/v4"
	"go-micro.dev/v4/logger"
)

const version = "1.0.0"

func main() {
	useDebug := false

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
			return utils.LoadConfig(configFile)
		}),
	)

	if useDebug {
		logger.Init(logger.WithLevel(logger.DebugLevel))
	}

	database, err := db.Connect(utils.Config().Database)
	if err != nil {
		logger.Fatal(err)
		os.Exit(1)
	}

	if err := accounts.Load(database); err != nil {
		logger.Errorf("Load torrent accounts failed: %+v", err)
	}

	rms_torrent.RegisterRmsTorrentHandler(service.Server(), tservice.NewService(database))

	if err := service.Run(); err != nil {
		logger.Fatal(err)
		os.Exit(2)
	}
}
