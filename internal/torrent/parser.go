package torrent

import (
	"bytes"
	"github.com/jackpal/bencode-go"
	"path"
)

type fileDict struct {
	Path []string
}
type infoDict struct {
	Files []fileDict
}
type torrentMetadata struct {
	Info infoDict
}

func isMagnetLink(data []byte) bool {
	const magnetLinkSign = "magnet:"
	if len(data) < len(magnetLinkSign) {
		return false
	}
	return string(data[:len(magnetLinkSign)]) == magnetLinkSign
}

func getTorrentFiles(baseDir string, data []byte) (files []string, err error) {
	var m torrentMetadata
	err = bencode.Unmarshal(bytes.NewReader(data), &m)

	if err == nil {
		for _, f := range m.Info.Files {
			total := baseDir
			for _, p := range f.Path {
				total = path.Join(total, p)
			}
			files = append(files, total)
		}
	}
	return
}
