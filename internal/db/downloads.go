package db

import "github.com/RacoonMediaServer/rms-torrent/internal/model"

func (d *Database) LoadDownloads() ([]model.Download, error) {
	var result []model.Download
	if err := d.conn.Find(&result).Error; err != nil {
		return nil, err
	}
	return result, nil
}

func (d *Database) AddDownload(record model.Download) error {
	return d.conn.Create(&record).Error
}

func (d *Database) UpdateDownload(record model.Download) error {
	return d.conn.Save(&record).Error
}

func (d *Database) RemoveDownload(id string) error {
	return d.conn.Model(&model.Download{}).Unscoped().Delete(&model.Download{ID: id}).Error
}
