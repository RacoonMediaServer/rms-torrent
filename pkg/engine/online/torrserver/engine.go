package torrserver

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"path/filepath"

	rms_torrent "github.com/RacoonMediaServer/rms-packages/pkg/service/rms-torrent"
	"github.com/RacoonMediaServer/rms-torrent/v4/pkg/engine"
	"github.com/RacoonMediaServer/rms-torrent/v4/pkg/torrserver/client"
	"github.com/RacoonMediaServer/rms-torrent/v4/pkg/torrserver/client/api"
	"github.com/RacoonMediaServer/rms-torrent/v4/pkg/torrserver/client/helpers"
	httptransport "github.com/go-openapi/runtime/client"
	"github.com/go-openapi/strfmt"
)

type Config struct {
	Location string
	URL      string
}

type torrserverEngine struct {
	cli *client.TorrserverClient
	js  *helpers.JsClient
	loc string
}

// Add implements engine.TorrentEngine.
func (t *torrserverEngine) Add(ctx context.Context, category string, description string, forceLocation *string, content []byte) (engine.TorrentDescription, error) {
	save := "true"

	// TODO: remove this fix of compatibility with rms-library
	category = fixCategory(category)

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
		ID:       resp.Payload.Hash,
		Title:    resp.Payload.Title,
		Location: filepath.Join(t.loc, category),
	}

	for _, f := range resp.Payload.FileStats {
		result.Files = append(result.Files, f.Path)
	}

	if len(result.Files) == 1 {
		result.Location = filepath.Join(result.Location, result.Title)
	}

	return result, nil
}

// Get implements engine.TorrentEngine.
func (t *torrserverEngine) Get(ctx context.Context, id string) (*rms_torrent.TorrentInfo, error) {
	resp, err := t.js.GetTorrent(ctx, id)
	if err != nil {
		return nil, err
	}
	return convertTorrentInfo(resp), nil
}

// List implements engine.TorrentEngine.
func (t *torrserverEngine) List(ctx context.Context, includeDone bool) ([]*rms_torrent.TorrentInfo, error) {
	if !includeDone {
		return []*rms_torrent.TorrentInfo{}, nil
	}

	resp, err := t.js.GetTorrents(ctx)
	if err != nil {
		return nil, fmt.Errorf("fetch torrents failed: %w", err)
	}
	result := make([]*rms_torrent.TorrentInfo, 0, len(resp))
	for _, ti := range resp {
		result = append(result, convertTorrentInfo(&ti))
	}

	return result, nil
}

// Remove implements engine.TorrentEngine.
func (t *torrserverEngine) Remove(ctx context.Context, id string) error {
	return t.js.RemoveTorrent(ctx, id)
}

// Stop implements engine.TorrentEngine.
func (t *torrserverEngine) Stop() error {
	return nil
}

// UpPriority implements engine.TorrentEngine.
func (t *torrserverEngine) UpPriority(ctx context.Context, id string) error {
	return errors.ErrUnsupported
}

func NewEngine(config Config) (engine.TorrentEngine, error) {
	u, err := url.Parse(config.URL)
	if err != nil {
		return nil, err
	}
	tr := httptransport.New(u.Host, u.RawPath, []string{u.Scheme})
	cli := client.New(tr, strfmt.Default)
	js := helpers.JsClient{URL: config.URL}
	return &torrserverEngine{cli: cli, js: &js, loc: config.Location}, nil
}

func fixCategory(category string) string {
	switch category {
	case "rms_movies":
		return "movie"
	case "rms_tv":
		return "tv"
	case "rms_music":
		return "music"
	case "movie":
		return "movie"
	case "tv":
		return "tv"
	case "music":
		return "music"
	}
	return "other"
}
