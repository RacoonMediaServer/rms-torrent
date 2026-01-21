package main

import (
	"context"
	"os"

	"github.com/RacoonMediaServer/rms-packages/pkg/events"
	"github.com/RacoonMediaServer/rms-packages/pkg/pubsub"
	rms_torrent "github.com/RacoonMediaServer/rms-packages/pkg/service/rms-torrent"
	"github.com/RacoonMediaServer/rms-torrent/internal/config"
	"github.com/RacoonMediaServer/rms-torrent/internal/db"
	tservice "github.com/RacoonMediaServer/rms-torrent/internal/service"
	"github.com/RacoonMediaServer/rms-torrent/internal/tfactory"
	"github.com/urfave/cli/v2"
	micro "go-micro.dev/v4"
	"go-micro.dev/v4/logger"

	// Plugins
	_ "github.com/go-micro/plugins/v4/registry/etcd"
)

const Version = "0.0.0"

func main() {
	logger.Infof("rms-torrent v%s", Version)
	useDebug := false
	onlineMode := os.Getenv("ONLINE_MODE") != ""
	var err error

	publicName := "rms-torrent"
	if onlineMode {
		publicName += "-online"
	}

	service := micro.NewService(
		micro.Name(publicName),
		micro.Version(Version),
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
			return config.Load(configFile)
		}),
	)

	if useDebug {
		_ = logger.Init(logger.WithLevel(logger.DebugLevel))
	}

	cfg := config.Config()

	var database *db.Database
	if isDatabaseEnabled(cfg, onlineMode) {
		database, err = db.Connect(cfg.Database)
		if err != nil {
			logger.Fatalf("Connect to database failed: %s", err)
		}
	}

	tEngine, err := tfactory.CreateEngine(onlineMode, database, cfg, func(event *events.Notification) error {
		pub := pubsub.NewPublisher(service)
		return pub.Publish(context.Background(), event)
	})
	if err != nil {
		logger.Fatalf("Create torrent engine failed: %s", err)
	}

	tService := tservice.NewService(cfg, tEngine)

	if err = rms_torrent.RegisterRmsTorrentHandler(service.Server(), tService); err != nil {
		logger.Fatalf("Cannot initialize service handler: %s", err)
	}

	if err := service.Run(); err != nil {
		logger.Fatal(err)
	}
}

func isDatabaseEnabled(cfg config.Configuration, onlineMode bool) bool {
	if onlineMode {
		return cfg.Online.Driver != "builtin"
	}
	return cfg.Offline.Driver != "builtin"
}
