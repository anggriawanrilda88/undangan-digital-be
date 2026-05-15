package models

import (
	"time"

	"github.com/google/uuid"
)

// RSVPModel is the GORM model for the rsvps table
type RSVPModel struct {
	ID           uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	InvitationID uuid.UUID `gorm:"type:uuid;not null;index"`
	GuestName    string    `gorm:"type:varchar(100);not null"`
	Status       string    `gorm:"type:varchar(20);not null"`
	GuestCount   int       `gorm:"not null;default:1"`
	Message      string    `gorm:"type:text"`
	CreatedAt    time.Time `gorm:"index"`
}

func (RSVPModel) TableName() string {
	return "rsvps"
}
