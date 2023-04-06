package db

import (
	rms_torrent "github.com/RacoonMediaServer/rms-packages/pkg/service/rms-torrent"
)

type settings struct {
	ID       uint                        `gorm:"primaryKey"`
	Settings rms_torrent.TorrentSettings `gorm:"embedded"`
}

func (d *Database) LoadSettings() (*rms_torrent.TorrentSettings, error) {
	var record settings
	if err := d.conn.Model(&settings{}).FirstOrCreate(&record, &record).Error; err != nil {
		return nil, err
	}
	return &record.Settings, nil
}

func (d *Database) SaveSettings(val *rms_torrent.TorrentSettings) error {
	return d.conn.Save(&settings{ID: 1, Settings: *val}).Error
}
