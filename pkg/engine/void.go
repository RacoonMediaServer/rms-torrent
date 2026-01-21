package engine

type VoidDatabase struct{}

// Add implements TorrentDatabase.
func (v *VoidDatabase) Add(t TorrentRecord) error {
	return nil
}

// Del implements TorrentDatabase.
func (v *VoidDatabase) Del(id string) error {
	return nil
}

// Load implements TorrentDatabase.
func (v *VoidDatabase) Load() ([]TorrentRecord, error) {
	return []TorrentRecord{}, nil
}
