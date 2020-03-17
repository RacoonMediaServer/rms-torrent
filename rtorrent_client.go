package main

import (
	"bufio"
	"context"
	"fmt"
	"github.com/micro/cli/v2"
	micro "github.com/micro/go-micro/v2"
	"os"
	proto "racoondev.tk/gitea/racoon/rtorrent/proto"
)

func main() {
	user := ""
	password := ""
	tracker := ""
	search := ""

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
		),
	)
	service.Init()

	client := proto.NewRacoonTorrentService("rtorrent", service.Client())

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
}
