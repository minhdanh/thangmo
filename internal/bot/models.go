package bot

import (
	"github.com/jinzhu/gorm"
)

type RSSLink struct {
	gorm.Model
	Url     string `gorm:"unique;not null"`
	AddedBy int    `gorm:"not null"`
}

type HNRegistration struct {
	gorm.Model
	UserID   int `gorm:"unique;not null"`
	MinScore int `gorm:"default:0"`
}

type RSSRegistration struct {
	RSSLinkID int `gorm:"primary_key"`
	UserID    int `gorm:"primary_key"`
	Alias     string
}
