package models

import (
	"github.com/jackc/pgx/pgtype"
	"gorm.io/gorm"
)

/*
AModel defintion of a Model used for later inheritance
*/
type AModel struct {
	*gorm.Model
}

/*
Event Type definition of the type object in the database
*/
type Event struct {
	AModel
	Id        string       `gorm:"unique;index"`
	Pubkey    string       `gorm:"index;not null"`
	Kind      uint32       `gorm:"index;not null"`
	Content   string       `gorm:"not null"`
	Sig       string       `gorm:"not null"`
	Tag       string       `gorm:"index"`
	TagValues pgtype.JSONB `gorm:"type:jsonb;default:'[]';not null"`
}
