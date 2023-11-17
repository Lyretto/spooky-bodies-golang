package model

import (
	"time"

	"github.com/google/uuid"
)

type Level struct {
	ID           uuid.UUID   `gorm:"type:uuid;primary;default:gen_random_uuid()" json:"id"`
	UserID       uuid.UUID   `gorm:"type:uuid;not null" json:"-"`
	User         *User       `json:"user"`
	Name         string      `json:"name"`
	Content      string      `json:"content"`
	AuthorReplay string      `json:"replay"`
	Thumbnail    []uint8     `json:"image"`
	ValidationId uuid.UUID   `json:"-"`
	Validation   *Validation `json:"validation"`
	Version      uint        `json:"version"`
	Reports      uint        `json:"-"`
	Published    time.Time   `json:"published"`
	AuthorScore  int         `json:"score"`
}

func (l *Level) TableName() string {
	return "levels"
}
