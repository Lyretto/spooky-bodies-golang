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
	Content      string      `json:"content"`
	Thumbnail    []uint8     `json:"image"`
	ValidationId uuid.UUID   `json:"validationId"`
	Validation   *Validation `json:"-"`
	Version      uint        `json:"-"`
	Reports      uint        `json:"-"`
	Updated      time.Time   `json:"-"`
}

func (l *Level) TableName() string {
	return "levels"
}
