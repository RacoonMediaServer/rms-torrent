package main

import (
	"context"
	"fmt"
	"git.rms.local/RacoonMediaServer/rms-shared/pkg/service/rms_torrent"
	"go-micro.dev/v4"
	"io"
	"os"
)
import "github.com/urfave/cli/v2"

func main() {
	var file string
	var magnet string
	service := micro.NewService(
		micro.Name("rms-torrent.client"),
		micro.Flags(
			&cli.StringFlag{
				Name:        "file",
				Usage:       "torrent file path",
				Required:    false,
				Destination: &file,
			},
			&cli.StringFlag{
				Name:        "magnet",
				Usage:       "magnet link to download",
				Required:    false,
				Destination: &magnet,
			},
		),
	)
	service.Init()

	client := rms_torrent.NewRmsTorrentService("rms-torrent", service.Client())
	content := []byte(magnet)

	if file != "" {
		f, err := os.Open(file)
		if err != nil {
			panic(err)
		}
		content, err = io.ReadAll(f)
		if err != nil {
			panic(err)
		}
		defer f.Close()
	}

	resp, err := client.Download(context.Background(), &rms_torrent.DownloadRequest{What: content})
	if err != nil {
		panic(err)
	}

	fmt.Printf("(%s) Files: %+v", resp.Id, resp.Files)
}
