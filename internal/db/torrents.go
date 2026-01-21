package db

import (
	"github.com/RacoonMediaServer/rms-torrent/pkg/engine"
)

// backward compatibility
type torrentModel struct {
	ID          string `gorm:"primaryKey"`
	Content     []byte
	Description string
	Complete    bool
	Fast        bool
}

func (d *Database) Load() ([]engine.TorrentRecord, error) {
	var torrents []*torrentModel
	if err := d.conn.Find(&torrents).Error; err != nil {
		return nil, err
	}
	result := make([]engine.TorrentRecord, 0, len(torrents))
	for _, t := range torrents {
		record := engine.TorrentRecord{
			TorrentDescription: engine.TorrentDescription{
				ID:    t.ID,
				Title: t.Description,
			},
			Content: t.Content,
		}
		result = append(result, record)
	}

	return result, nil
}

func (d *Database) Add(record engine.TorrentRecord) error {
	t := torrentModel{
		ID:          record.ID,
		Description: record.Title,
		Content:     record.Content,
	}
	return d.conn.Create(&t).Error
}

func (d *Database) Complete(id string) error {
	return d.conn.Model(&torrentModel{ID: id}).Update("complete", true).Error
}

func (d *Database) Del(id string) error {
	return d.conn.Model(&torrentModel{}).Unscoped().Delete(&torrentModel{ID: id}).Error
}
