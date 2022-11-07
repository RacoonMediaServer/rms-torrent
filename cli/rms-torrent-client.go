package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"

	"git.rms.local/RacoonMediaServer/rms-shared/pkg/service/rms_torrent"
	"github.com/urfave/cli/v2"
	micro "go-micro.dev/v4"
)

func main() {
	tracker := ""
	search := ""
	download := false

	service := micro.NewService(
		micro.Name("rms-torrent.client"),
		micro.Flags(
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

	client := rms_torrent.NewRmsTorrentService("rms-torrent", service.Client())

	trackers, err := client.ListTrackers(context.Background(), &rms_torrent.ListTrackersRequest{})
	if err != nil {
		log.Fatal(err)
	}

	for _, tracker := range trackers.Trackers {
		log.Printf("Tracker discovered: [%s] - '%s'", tracker.Id, tracker.Name)
	}

	request := rms_torrent.SearchRequest{
		Text:    search,
		Tracker: tracker,
	}

	response, err := client.Search(context.Background(), &request)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	if response.Code == rms_torrent.SearchResponse_CAPTCHA_REQUIRED {
		fmt.Printf("Please enter code from captcha %s\n", response.CaptchaURL)
		reader := bufio.NewReader(os.Stdin)
		request.Captcha, _ = reader.ReadString('\n')
		response, err = client.Search(context.Background(), &request)

		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	}

	// Print response
	fmt.Println(response)

	if response.Code == rms_torrent.SearchResponse_OK && download {
		for _, torrent := range response.Results {
			request := rms_torrent.DownloadRequest{
				SessionID:   response.SessionID,
				TorrentLink: torrent.Link,
			}
			response, err := client.Download(context.Background(), &request)
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
