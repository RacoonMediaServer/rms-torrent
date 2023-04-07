package db

import (
	"github.com/RacoonMediaServer/rms-torrent/internal/model"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type Database struct {
	conn *gorm.DB
}

func Connect(path string) (*Database, error) {
	db, err := gorm.Open(sqlite.Open(path))
	if err != nil {
		return nil, err
	}

	if err = db.AutoMigrate(&settings{}); err != nil {
		return nil, err
	}
	if err = db.AutoMigrate(&model.Torrent{}); err != nil {
		return nil, err
	}
	return &Database{conn: db}, nil
}
