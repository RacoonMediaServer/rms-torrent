package db

import "github.com/RacoonMediaServer/rms-torrent/internal/model"

func (d *Database) LoadTorrents() ([]*model.Torrent, error) {
	var result []*model.Torrent
	if err := d.conn.Find(&result).Error; err != nil {
		return nil, err
	}
	return result, nil
}

func (d *Database) AddTorrent(record *model.Torrent) error {
	return d.conn.Create(record).Error
}

func (d *Database) CompleteTorrent(id string) error {
	return d.conn.Model(&model.Torrent{ID: id}).Update("complete", true).Error
}
func (d *Database) UpdateTorrent(record *model.Torrent) error {
	return d.conn.Save(record).Error
}

func (d *Database) RemoveTorrent(id string) error {
	return d.conn.Model(&model.Torrent{}).Unscoped().Delete(&model.Torrent{ID: id}).Error
}
