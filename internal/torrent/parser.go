package torrent

func isMagnetLink(data []byte) bool {
	const magnetLinkSign = "magnet:"
	if len(data) < len(magnetLinkSign) {
		return false
	}
	return string(magnetLinkSign[:len(magnetLinkSign)]) == magnetLinkSign
}
