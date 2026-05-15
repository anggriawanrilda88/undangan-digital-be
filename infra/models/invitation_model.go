package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
)

// InvitationModel is the GORM model for the invitations table
type InvitationModel struct {
	ID          uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	UserID      uuid.UUID      `gorm:"type:uuid;not null;index"`
	Slug        string         `gorm:"type:varchar(60);uniqueIndex;not null"`
	Status      string         `gorm:"type:varchar(20);not null;default:'draft'"`
	Config      datatypes.JSON `gorm:"type:jsonb;not null;default:'{}'"`
	Content     datatypes.JSON `gorm:"type:jsonb;not null;default:'{}'"`
	RSVPCount   int            `gorm:"default:0"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
	PublishedAt *time.Time `gorm:"index"`
}

func (InvitationModel) TableName() string {
	return "invitations"
}
