package main

import (
	"bufio"
	"context"
	"fmt"
	"github.com/micro/cli/v2"
	micro "github.com/micro/go-micro/v2"
	"log"
	"os"
	proto "racoondev.tk/gitea/racoon/rms-torrent/proto"
)

func main() {
	user := ""
	password := ""
	tracker := ""
	search := ""
	download := false

	service := micro.NewService(
		micro.Name("rtorrent.client"),
		micro.Flags(
			&cli.StringFlag{
				Name:        "user",
				Usage:       "user for login",
				Required:    false,
				Value:       "",
				DefaultText: "",
				Destination: &user,
			},
			&cli.StringFlag{
				Name:        "password",
				Usage:       "password for login",
				Required:    false,
				Value:       "",
				DefaultText: "",
				Destination: &password,
			},
			&cli.StringFlag{
				Name:        "tracker",
				Usage:       "site for search",
				Required:    false,
				Value:       "rutracker",
				DefaultText: "rutracker",
				Destination: &tracker,
			},
			&cli.StringFlag{
				Name:        "search",
				Usage:       "search text",
				Required:    true,
				Value:       "",
				DefaultText: "",
				Destination: &search,
			},
			&cli.BoolFlag{
				Name:        "download",
				Usage:       "download all torrents",
				Value:       false,
				DefaultText: "",
				Destination: &download,
			},
		),
	)
	service.Init()

	client := proto.NewRacoonTorrentService("rtorrent", service.Client())

	trackers, err := client.ListTrackers(context.TODO(), &proto.ListTrackersRequest{})
	if err != nil {
		log.Fatal(err)
	}

	for _, tracker := range trackers.Trackers {
		log.Printf("Tracker discovered: [%s] - '%s'", tracker.Id, tracker.Name)
	}

	request := proto.SearchRequest{
		Login:    user,
		Password: password,
		Text:     search,
		Tracker:  tracker,
	}

	response, err := client.Search(context.TODO(), &request)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	if response.Code == proto.SearchResponse_CAPTCHA_REQUIRED {
		fmt.Printf("Please enter code from captcha %s\n", response.CaptchaURL)
		reader := bufio.NewReader(os.Stdin)
		request.Captcha, _ = reader.ReadString('\n')
		response, err = client.Search(context.TODO(), &request)

		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	}

	// Print response
	fmt.Println(response)

	if response.Code == proto.SearchResponse_OK && download {
		for _, torrent := range response.Results {
			request := proto.DownloadRequest{
				SessionID:            response.SessionID,
				TorrentLink:          torrent.Link,
			}
			response, err := client.Download(context.TODO(), &request)
			if err != nil {
				log.Fatal(err)
			}

			if !response.DownloadStarted {
				log.Fatal(response.ErrorReason)
			} else {
				log.Printf("%s downloaded", torrent.Title)
			}
		}
	}
}
