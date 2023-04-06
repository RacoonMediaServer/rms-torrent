package model

type Download struct {
	ID       string `gorm:"primaryKey"`
	Content  []byte
	Complete bool
	Fast     bool
	Order    uint
}
