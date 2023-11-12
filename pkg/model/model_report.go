package model

import "github.com/google/uuid"

type Report struct {
	ID      uuid.UUID `gorm:"type:uuid;primary;default:gen_random_uuid()" json:"id"`
	UserID  uuid.UUID `gorm:"type:uuid;not null" json:"userId"`
	User    *User     `json:"-"`
	LevelID uuid.UUID `gorm:"type:uuid;not null" json:"levelId"`
	Level   *Level    `json:"-"`
}

func (r *Report) TableName() string {
	return "reports"
}
