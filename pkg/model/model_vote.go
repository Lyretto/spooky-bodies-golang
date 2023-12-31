package model

import "github.com/google/uuid"

type VoteType = string

const VoteLike = VoteType("like")
const VoteDislike = VoteType("dislike")

type Vote struct {
	ID      uuid.UUID `gorm:"type:uuid;primary;default:gen_random_uuid()" json:"id"`
	UserID  uuid.UUID `gorm:"type:uuid;not null" json:"userId"`
	User    *User     `json:"-"`
	LevelID uuid.UUID `gorm:"type:uuid;not null" json:"levelId"`
	Level   *Level    `json:"-"`
	Type    VoteType  `gorm:"type:string" json:"type"`
}

func (v *Vote) TableName() string {
	return "votes"
}
