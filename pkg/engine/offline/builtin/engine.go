package builtin

import (
	"context"
	"errors"

	rms_torrent "github.com/RacoonMediaServer/rms-packages/pkg/service/rms-torrent"
	"github.com/RacoonMediaServer/rms-torrent/pkg/engine"
)

type Config struct {
}

type builtinEngine struct {
}

// Add implements engine.TorrentEngine.
func (b *builtinEngine) Add(ctx context.Context, category string, description string, forceLocation *string, content []byte) (engine.TorrentDescription, error) {
	return engine.TorrentDescription{}, errors.ErrUnsupported
}

// Get implements engine.TorrentEngine.
func (b *builtinEngine) Get(ctx context.Context, id string) (*rms_torrent.TorrentInfo, error) {
	return nil, errors.ErrUnsupported
}

// List implements engine.TorrentEngine.
func (b *builtinEngine) List(ctx context.Context, includeDone bool) ([]*rms_torrent.TorrentInfo, error) {
	return nil, errors.ErrUnsupported
}

// Remove implements engine.TorrentEngine.
func (b *builtinEngine) Remove(ctx context.Context, id string) error {
	return errors.ErrUnsupported
}

// Stop implements engine.TorrentEngine.
func (b *builtinEngine) Stop() error {
	return errors.ErrUnsupported
}

// UpPriority implements engine.TorrentEngine.
func (b *builtinEngine) UpPriority(ctx context.Context, id string) error {
	return errors.ErrUnsupported
}

func NewTorrentEngine(cfg Config, db engine.TorrentDatabase, onComplete engine.CompleteAction) (engine.TorrentEngine, error) {
	return &builtinEngine{}, nil
}
