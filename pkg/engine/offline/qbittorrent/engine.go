package qbittorrent

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"os"
	"time"

	rms_torrent "github.com/RacoonMediaServer/rms-packages/pkg/service/rms-torrent"
	"github.com/RacoonMediaServer/rms-torrent/v4/pkg/engine"
	"github.com/RacoonMediaServer/rms-torrent/v4/pkg/qbittorrent"
	torrentutils "github.com/RacoonMediaServer/rms-torrent/v4/pkg/torrent-utils"
	"go-micro.dev/v4/logger"
)

type Config struct {
	URL      string
	User     string
	Password string
	Fifo     string
	Script   string
}

type qbittorrentEngine struct {
	config Config
	w      *fifoWatcher
}

//go:embed notify.sh
var notifyScript string

const getTitleRetries = 20
const getTitleRetryInterval = 1 * time.Second

// Add implements engine.TorrentEngine.
func (q *qbittorrentEngine) Add(ctx context.Context, category string, description string, forceLocation *string, content []byte) (engine.TorrentDescription, error) {
	resp := engine.TorrentDescription{}
	cli := &qbittorrent.Client{URL: q.config.URL}
	if err := cli.Login(q.config.User, q.config.Password); err != nil {
		return resp, fmt.Errorf("login failed: %w", err)
	}

	hash, err := torrentutils.GetTorrentInfoHash(content)
	if err != nil {
		return resp, fmt.Errorf("get torrent hash failed: %w", err)
	}
	resp.ID = hash

	if err = cli.AddTorrent(ctx, category, description, forceLocation, content); err != nil {
		return resp, fmt.Errorf("add torrent failed: %w", err)
	}

	title := ""
	for i := 0; i < getTitleRetries; i++ {
		info, err := cli.GetTorrent(ctx, hash)
		if err == nil && len(info) != 0 && info[0].Name != "" {
			title = info[0].Name
			break
		}
		select {
		case <-ctx.Done():
			return resp, ctx.Err()
		case <-time.After(getTitleRetryInterval):
		}
	}

	if title == "" {
		return resp, errors.New("get added torrent info failed")
	}
	resp.Title = title

	files, err := cli.GetFiles(ctx, hash)
	if err != nil {
		logger.Warnf("Get files failed: %s", err)
		return resp, nil
	}
	resp.Files = make([]string, len(files))
	for i := range files {
		resp.Files[i] = files[i].Name
	}

	return resp, nil
}

// Get implements engine.TorrentEngine.
func (q *qbittorrentEngine) Get(ctx context.Context, id string) (*rms_torrent.TorrentInfo, error) {
	cli := &qbittorrent.Client{URL: q.config.URL}
	if err := cli.Login(q.config.User, q.config.Password); err != nil {
		return nil, fmt.Errorf("login failed: %w", err)
	}

	torrents, err := cli.GetTorrent(ctx, id)
	if err != nil {
		logger.Errorf("Get torrent failed: %s", err)
		return &rms_torrent.TorrentInfo{}, err
	}
	if len(torrents) != 1 {
		return &rms_torrent.TorrentInfo{}, errors.New("not found")
	}

	return convertTorrentInfo(&torrents[0]), nil
}

// List implements engine.TorrentEngine.
func (q *qbittorrentEngine) List(ctx context.Context, includeDone bool) ([]*rms_torrent.TorrentInfo, error) {
	cli := &qbittorrent.Client{URL: q.config.URL}
	if err := cli.Login(q.config.User, q.config.Password); err != nil {
		return nil, fmt.Errorf("login failed: %w", err)
	}

	all, err := cli.GetTorrents(ctx)
	if err != nil {
		return nil, fmt.Errorf("get torrents failed: %w", err)
	}

	result := make([]*rms_torrent.TorrentInfo, 0, len(all))
	for _, t := range all {
		converted := convertTorrentInfo(&t)
		if !includeDone && converted.Status == rms_torrent.Status_Done {
			continue
		}
		result = append(result, converted)
	}
	return result, nil
}

// Remove implements engine.TorrentEngine.
func (q *qbittorrentEngine) Remove(ctx context.Context, id string) error {
	cli := &qbittorrent.Client{URL: q.config.URL}
	if err := cli.Login(q.config.User, q.config.Password); err != nil {
		return fmt.Errorf("login failed: %w", err)
	}

	return cli.RemoveTorrent(ctx, id)
}

// Stop implements engine.TorrentEngine.
func (q *qbittorrentEngine) Stop() error {
	if q.w != nil {
		return q.w.stop()
	}
	return nil
}

// UpPriority implements engine.TorrentEngine.
func (q *qbittorrentEngine) UpPriority(ctx context.Context, id string) error {
	cli := &qbittorrent.Client{URL: q.config.URL}
	if err := cli.Login(q.config.User, q.config.Password); err != nil {
		return fmt.Errorf("login failed: %w", err)
	}

	return cli.IncPriority(ctx, id)
}

func NewTorrentEngine(config Config, completeAction engine.CompleteAction) (engine.TorrentEngine, error) {
	cli := &qbittorrent.Client{URL: config.URL}
	if err := cli.Login(config.User, config.Password); err != nil {
		return nil, fmt.Errorf("login failed: %w", err)
	}

	if err := os.WriteFile(config.Script, []byte(notifyScript), 0777); err != nil {
		return nil, fmt.Errorf("write script failed: %w", err)
	}

	p := qbittorrent.Preferences{
		AutorunEnabled:     true,
		AutorunProgram:     fmt.Sprintf(`%s %s %s`, config.Script, config.Fifo, notificationPattern),
		AutoTmmEnabled:     true,
		QueueMode:          true,
		MaxActiveDownloads: 1,
	}
	if err := cli.SetPreferences(context.Background(), p); err != nil {
		return nil, fmt.Errorf("set notification catcher failed: %w", err)
	}

	var w *fifoWatcher
	if completeAction != nil {
		w = &fifoWatcher{}
		if err := w.startFifoWatcher(config.Fifo, completeAction); err != nil {
			return nil, fmt.Errorf("start watching on fifo failed: %w", err)
		}
	}

	return &qbittorrentEngine{config: config, w: w}, nil
}
