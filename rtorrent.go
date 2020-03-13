package main

import (
	"fmt"
	micro "github.com/micro/go-micro/v2"
	"racoondev.tk/gitea/racoon/rtorrent/internal"
	proto "racoondev.tk/gitea/racoon/rtorrent/proto"
)

const version = "0.0.1"


func main() {
	service := micro.NewService(
		micro.Name("rtorrent"),
		micro.Version(version),
	)

	service.Init()

	proto.RegisterRacoonTorrentHandler(service.Server(), new(internal.TorrentService))

	if err := service.Run(); err != nil {
		fmt.Println(err)
	}
}
