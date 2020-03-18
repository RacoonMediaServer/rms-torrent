package main

import (
	"github.com/micro/cli/v2"
	micro "github.com/micro/go-micro/v2"
	"github.com/micro/go-micro/v2/config"
	"github.com/micro/go-micro/v2/config/source/env"
	"github.com/micro/go-micro/v2/logger"
	tservice "racoondev.tk/gitea/racoon/rtorrent/internal/service"
	proto "racoondev.tk/gitea/racoon/rtorrent/proto"
)

const version = "0.0.1"

func main() {
	useDebug := false

	service := micro.NewService(
		micro.Name("rtorrent"),
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

	service.Init()

	if useDebug {
		logger.Init(logger.WithLevel(logger.DebugLevel))
	}

	err := config.Load(
		env.NewSource(env.WithStrippedPrefix("RTORRENT")),
	)

	if err != nil {
		logger.Fatal(err)
	}

	logger.Infof("Configuration: %+v", config.Map())

	directory := config.Get("directory").String("/tmp")

	proto.RegisterRacoonTorrentHandler(service.Server(), tservice.NewService(directory))

	if err := service.Run(); err != nil {
		logger.Fatal(err)
	}
}
