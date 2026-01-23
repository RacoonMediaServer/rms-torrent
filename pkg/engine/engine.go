package engine

import (
	"context"

	"github.com/RacoonMediaServer/rms-packages/pkg/events"
	rms_torrent "github.com/RacoonMediaServer/rms-packages/pkg/service/rms-torrent"
)

type TorrentDescription struct {
	ID       string
	Title    string
	Files    []string
	Location string
}

type EventAction func(events.Notification_Kind, *rms_torrent.TorrentInfo) error

type TorrentEngine interface {
	Add(ctx context.Context, category, description string, forceLocation *string, content []byte) (TorrentDescription, error)
	Get(ctx context.Context, id string) (*rms_torrent.TorrentInfo, error)
	List(ctx context.Context, includeDone bool) ([]*rms_torrent.TorrentInfo, error)
	Remove(ctx context.Context, id string) error
	UpPriority(ctx context.Context, id string) error
	Stop() error
}
