package db

import (
	"github.com/RacoonMediaServer/rms-packages/pkg/configuration"
	"github.com/RacoonMediaServer/rms-torrent/internal/model"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Database struct {
	conn *gorm.DB
}

func Connect(config configuration.Database) (*Database, error) {
	db, err := gorm.Open(postgres.Open(config.GetConnectionString()))
	if err != nil {
		return nil, err
	}

	if err = db.AutoMigrate(&torrentSettings{}); err != nil {
		return nil, err
	}
	if err = db.AutoMigrate(&model.Torrent{}); err != nil {
		return nil, err
	}
	return &Database{conn: db}, nil
}
