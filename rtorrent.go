package main

import (
	micro "github.com/micro/go-micro/v2"
	"github.com/micro/go-micro/v2/config"
	"github.com/micro/go-micro/v2/config/source/env"
	"log"
	tservice "racoondev.tk/gitea/racoon/rtorrent/internal/service"
	proto "racoondev.tk/gitea/racoon/rtorrent/proto"
)

const version = "0.0.1"

func main() {
	service := micro.NewService(
		micro.Name("rtorrent"),
		micro.Version(version),
	)

	service.Init()

	err := config.Load(
		env.NewSource(env.WithStrippedPrefix("RTORRENT")),
	)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Configuration: %+v\n", config.Map())

	proto.RegisterRacoonTorrentHandler(service.Server(), tservice.NewService())

	if err := service.Run(); err != nil {
		log.Fatal(err)
	}
}
