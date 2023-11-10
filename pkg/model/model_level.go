package model

import "github.com/google/uuid"

type Level struct {
	ID      uuid.UUID `gorm:"type:uuid;primary;default:gen_random_uuid()" json:"id"`
	UserID  uuid.UUID `gorm:"not null" json:"userId"`
	User    *User     `json:"-"`
	Content string    `json:"content"`
}

func (l *Level) TableName() string {
	return "levels"
}
