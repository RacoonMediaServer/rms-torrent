package main

import (
	"context"
	"fmt"
	micro "github.com/micro/go-micro/v2"
	"os"
	proto "racoondev.tk/gitea/racoon/rtorrent/proto"
)

func main() {
	service := micro.NewService(micro.Name("rtorrent.client"))
	service.Init()

	client := proto.NewRacoonTorrentService("rtorrent", service.Client())

	response, err := client.Search(context.TODO(), &proto.SearchRequest{Text: "South Park"})
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// Print response
	fmt.Println(response)
}
