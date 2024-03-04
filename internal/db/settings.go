package db

import (
	rms_torrent "github.com/RacoonMediaServer/rms-packages/pkg/service/rms-torrent"
)

type torrentSettings struct {
	ID       uint                        `gorm:"primaryKey"`
	Settings rms_torrent.TorrentSettings `gorm:"embedded"`
}

func (d *Database) LoadSettings() (*rms_torrent.TorrentSettings, error) {
	var record torrentSettings
	if err := d.conn.Model(&torrentSettings{}).FirstOrCreate(&record, &record).Error; err != nil {
		return nil, err
	}
	return &record.Settings, nil
}

func (d *Database) SaveSettings(val *rms_torrent.TorrentSettings) error {
	return d.conn.Save(&torrentSettings{ID: 1, Settings: *val}).Error
}
