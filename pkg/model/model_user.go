package model

import (
	"time"

	"github.com/google/uuid"
)

type PlatformType = string

const PlatformNone = PlatformType("none")
const PlatformSteam = PlatformType("steam")
const PlatformNintendo = PlatformType("nintendo")

type User struct {
	ID             uuid.UUID    `gorm:"type:uuid;primary;default:gen_random_uuid()" json:"id"`
	PlatformType   PlatformType `gorm:"type:string" json:"platformType"`
	PlatformUserID string       `gorm:"index:idx_platform_id_unique,unique" json:"platformUserId"`
	PlatformName   string       `json:"platformName"`
	IsMod          bool         `gorm:"default:false" json:"-"`
	CreatedAt      time.Time    `json:"createdAt"`
}

func (u *User) TableName() string {
	return "users"
}
