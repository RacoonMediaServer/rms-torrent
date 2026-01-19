package torrserver

import (
	"context"
	"fmt"
	"net/url"

	rms_torrent "github.com/RacoonMediaServer/rms-packages/pkg/service/rms-torrent"
	"github.com/RacoonMediaServer/rms-torrent/pkg/engine"
	"github.com/RacoonMediaServer/rms-torrent/pkg/torrserver/client"
	"github.com/RacoonMediaServer/rms-torrent/pkg/torrserver/client/api"
	httptransport "github.com/go-openapi/runtime/client"
	"github.com/go-openapi/strfmt"
)

type Config struct {
	URL string
}

type torrserverEngine struct {
	cli *client.TorrserverClient
}

// Add implements engine.TorrentEngine.
func (t *torrserverEngine) Add(ctx context.Context, category string, description string, forceLocation *string, content []byte) (engine.TorrentDescription, error) {
	save := "true"

	resp, err := t.cli.API.PostTorrentUpload(&api.PostTorrentUploadParams{
		Category: &category,
		Save:     &save,
		Context:  ctx,
		File:     newTorrentFile("download.torrent", content),
	})
	if err != nil {
		return engine.TorrentDescription{}, fmt.Errorf("upload torrent failed: %w", err)
	}

	result := engine.TorrentDescription{
		ID:    resp.Payload.Hash,
		Title: resp.Payload.Title,
	}
	return result, nil
}

// Get implements engine.TorrentEngine.
func (t *torrserverEngine) Get(ctx context.Context, id string) (*rms_torrent.TorrentInfo, error) {
	panic("unimplemented")
}

// List implements engine.TorrentEngine.
func (t *torrserverEngine) List(ctx context.Context, includeDone bool) ([]*rms_torrent.TorrentInfo, error) {
	panic("unimplemented")
}

// Remove implements engine.TorrentEngine.
func (t *torrserverEngine) Remove(ctx context.Context, id string) error {
	panic("unimplemented")
}

// Stop implements engine.TorrentEngine.
func (t *torrserverEngine) Stop() error {
	panic("unimplemented")
}

// UpPriority implements engine.TorrentEngine.
func (t *torrserverEngine) UpPriority(ctx context.Context, id string) error {
	panic("unimplemented")
}

func NewEngine(config Config) (engine.TorrentEngine, error) {
	u, err := url.Parse(config.URL)
	if err != nil {
		return nil, err
	}
	tr := httptransport.New(u.Host, u.RawPath, []string{u.Scheme})
	cli := client.New(tr, strfmt.Default)
	return &torrserverEngine{cli: cli}, nil
}
