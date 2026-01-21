package builtin

import (
	"fmt"
	"os"
	"path/filepath"
)

const mainRoute = "data"

type layout struct {
	baseDir            string
	pieceCompletionDir string
	cacheDir           string
	contentDir         string
	itemsDir           string
}

func newLayout(baseDir string) layout {
	return layout{
		baseDir:            baseDir,
		pieceCompletionDir: filepath.Join(baseDir, "piece-completion"),
		cacheDir:           filepath.Join(baseDir, "cache"),
		contentDir:         filepath.Join(baseDir, "content"),
		itemsDir:           filepath.Join(baseDir, "items"),
	}
}

func (l layout) makeLayout() error {
	if err := os.MkdirAll(l.pieceCompletionDir, 0744); err != nil {
		return fmt.Errorf("create piece completion directory failed: %w", err)
	}
	if err := os.MkdirAll(l.cacheDir, 0744); err != nil {
		return fmt.Errorf("create cache directory failed: %w", err)
	}
	if err := os.MkdirAll(l.contentDir, 0744); err != nil {
		return fmt.Errorf("create content directory failed: %w", err)
	}

	return nil
}
