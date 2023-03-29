package downloads

func isMagnetLink(data []byte) bool {
	const magnetLinkSign = "magnet:"
	if len(data) < len(magnetLinkSign) {
		return false
	}
	return string(data[:len(magnetLinkSign)]) == magnetLinkSign
}
