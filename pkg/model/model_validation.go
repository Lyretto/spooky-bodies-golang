package model

import "github.com/google/uuid"

type ResultType = string

const ResultOk = ResultType("ok")
const ResultContentSuspect = ResultType("content-suspect")
const ResultNameSuspect = ResultType("name-Suspect")
const ResultContentTooComplex = ResultType("content-complex")

type Validation struct {
	ID           uuid.UUID  `gorm:"type:uuid;primary;default:gen_random_uuid()" json:"id"`
	ValidatorID  uuid.UUID  `gorm:"type:uuid" json:"validatorUserId"`
	User         *User      `json:"-"`
	LevelID      uuid.UUID  `gorm:"type:uuid" json:"levelId"`
	Level        *Level     `json:"-"`
	LevelVersion uint       `json:"version"`
	Result       ResultType `gorm:"type:string" json:"result"`
}

func (v *Validation) TableName() string {
	return "validations"
}
