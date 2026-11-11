package models

import (
	"github.com/nbd-wtf/go-nostr"
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
	Created int64      `gorm:"index;not null"`
	EventId string     `gorm:"index"`
	Pubkey  string     `gorm:"index;not null"`
	Kind    uint32     `gorm:"index;not null"`
	Content string     `gorm:"not null"`
	Sig     string     `gorm:"not null"`
	Tags    nostr.Tags `gorm:"type:jsonb;default:'[]';not null;serializer:json;index:idx_events_tags,type:gin"`
}
