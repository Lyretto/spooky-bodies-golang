package model

import (
	"time"

	"github.com/google/uuid"
)

type Level struct {
	ID           uuid.UUID   `gorm:"type:uuid;primary;default:gen_random_uuid()" json:"id"`
	UserID       uuid.UUID   `gorm:"type:uuid;not null" json:"userId"`
	User         *User       `json:"-"`
	Name         string      `gorm:"not null" json:"name"`
	Content      string      `gorm:"not null" json:"content"`
	AuthorReplay string      `gorm:"not null" json:"replay"`
	Thumbnail    []uint8     `json:"image"`
	ValidationId uuid.UUID   `json:"validationId"`
	Validation   *Validation `json:"-"`
	Version      uint        `json:"version"`
	Reports      uint        `json:"-"`
	Published    time.Time   `json:"published"`
	AuthorScore  int         `json:"score"`
}

func (l *Level) TableName() string {
	return "levels"
}
