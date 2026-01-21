package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"time"

	rms_torrent "github.com/RacoonMediaServer/rms-packages/pkg/service/rms-torrent"
	"github.com/urfave/cli/v2"
	"go-micro.dev/v4"
	"go-micro.dev/v4/client"
)

func main() {
	var command string
	var item string
	var serverName string
	service := micro.NewService(
		micro.Name("rms-torrent.client"),
		micro.Flags(
			&cli.StringFlag{
				Name:        "command",
				Usage:       "Must be one of: download, list, remove, up, apply",
				Required:    true,
				Destination: &command,
			},
			&cli.StringFlag{
				Name:        "item",
				Usage:       "item - id or path to torrent file or magnet link",
				Required:    false,
				Destination: &item,
			},
			&cli.StringFlag{
				Name:        "serverName",
				Usage:       "serverName - id of server",
				Required:    false,
				DefaultText: "rms-torrent",
				Destination: &serverName,
			},
		),
	)
	service.Init()

	client := rms_torrent.NewRmsTorrentService(serverName, service.Client())

	switch command {
	case "download":
		if err := download(client, item); err != nil {
			panic(err)
		}
	case "list":
		if err := list(client); err != nil {
			panic(err)
		}
	case "remove":
		if err := remove(client, item); err != nil {
			panic(err)
		}
	case "up":
		if err := up(client, item); err != nil {
			panic(err)
		}
	case "apply":
		if err := apply(client, item); err != nil {
			panic(err)
		}
	default:
		panic("unknown command: " + command)
	}
}

func download(cli rms_torrent.RmsTorrentService, file string) error {
	content := []byte(file)

	if !strings.HasPrefix(file, "magnet") {
		f, err := os.Open(file)
		if err != nil {
			return err
		}
		content, err = io.ReadAll(f)
		if err != nil {
			return err
		}
		defer f.Close()
	}

	req := rms_torrent.DownloadRequest{
		What:        content,
		Description: "Test download",
		Faster:      false,
	}
	resp, err := cli.Download(context.Background(), &req, client.WithRequestTimeout(40*time.Second))
	if err != nil {
		return err
	}

	fmt.Printf("(%s) Download files: %+v\n", resp.Id, resp.Files)
	return nil
}

func list(cli rms_torrent.RmsTorrentService) error {
	result, err := cli.GetTorrents(context.Background(), &rms_torrent.GetTorrentsRequest{IncludeDoneTorrents: true})
	if err != nil {
		return err
	}
	for _, t := range result.Torrents {
		fmt.Println(t.Id, t.Title, t.Status, t.Progress, time.Duration(t.RemainingTime), t.SizeMB)
	}
	return nil
}

func remove(cli rms_torrent.RmsTorrentService, id string) error {
	_, err := cli.RemoveTorrent(context.Background(), &rms_torrent.RemoveTorrentRequest{Id: id})
	return err
}

func up(cli rms_torrent.RmsTorrentService, id string) error {
	_, err := cli.UpPriority(context.Background(), &rms_torrent.UpPriorityRequest{Id: id})
	return err
}

func apply(cli rms_torrent.RmsTorrentService, downloadLimit string) error {
	limit, err := strconv.ParseUint(downloadLimit, 10, 64)
	if err != nil {
		return err
	}
	settings := &rms_torrent.TorrentSettings{
		DownloadLimit: limit,
		UploadLimit:   0,
	}
	_, err = cli.SetSettings(context.Background(), settings)
	return err
}
