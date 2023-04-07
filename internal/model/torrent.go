package model

type Torrent struct {
	ID          string `gorm:"primaryKey"`
	Content     []byte
	Description string
	Complete    bool
	Fast        bool
}
