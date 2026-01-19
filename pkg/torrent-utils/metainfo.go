package torrentutils

import (
	"bytes"

	"github.com/anacrolix/torrent/metainfo"
)

func IsMagnetLink(torrentContent []byte) bool {
	const magnetLinkSign = "magnet:"
	if len(torrentContent) < len(magnetLinkSign) {
		return false
	}
	return string(torrentContent[:len(magnetLinkSign)]) == magnetLinkSign
}

func GetTorrentInfoHash(torrentContent []byte) (string, error) {
	if !IsMagnetLink(torrentContent) {
		mi, err := metainfo.Load(bytes.NewReader(torrentContent))
		if err != nil {
			return "", err
		}
		return mi.HashInfoBytes().HexString(), nil
	} else {
		m, err := metainfo.ParseMagnetV2Uri(string(torrentContent))
		if err != nil {
			return "", err
		}
		return m.InfoHash.Value.HexString(), nil
	}
}
