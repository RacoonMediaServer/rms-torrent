package downloader

import (
	"io/fs"
	"path/filepath"
)

type onlineDownloader struct {
	id    string
	title string
	dir   string
}

func (d *onlineDownloader) Start() {
}

func (d *onlineDownloader) Files() []string {
	torrentDir := filepath.Join(d.dir, mainRoute, d.title)
	var files []string
	_ = filepath.WalkDir(torrentDir, func(path string, e fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if !e.IsDir() {
			relPath, _ := filepath.Rel(torrentDir, path)
			files = append(files, relPath)
		}
		return nil
	})
	return files
}

func (d *onlineDownloader) Title() string {
	return d.title
}

func (d *onlineDownloader) Stop() {
}

func (d *onlineDownloader) Bytes() uint64 {
	return 0
}

func (d *onlineDownloader) RemainingBytes() uint64 {
	return 0
}

func (d *onlineDownloader) IsComplete() bool {
	return true
}

func (d *onlineDownloader) SizeMB() uint64 {
	return d.Bytes() / (1024 * 1024)
}

func (d *onlineDownloader) Close() {
}
