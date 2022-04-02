package main

import (
	"git.rms.local/RacoonMediaServer/rms-shared/pkg/db"
	"git.rms.local/RacoonMediaServer/rms-torrent/internal/accounts"
	tservice "git.rms.local/RacoonMediaServer/rms-torrent/internal/service"
	"git.rms.local/RacoonMediaServer/rms-torrent/internal/utils"
	proto "git.rms.local/RacoonMediaServer/rms-torrent/proto"
	"github.com/micro/cli/v2"
	micro "github.com/micro/go-micro/v2"
	"github.com/micro/go-micro/v2/logger"
	"os"
)

const version = "0.0.7"

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

	proto.RegisterRmsTorrentHandler(service.Server(), tservice.NewService(database))

	if err := service.Run(); err != nil {
		logger.Fatal(err)
		os.Exit(2)
	}
}
