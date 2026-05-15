package models

import (
	"time"

	"github.com/google/uuid"
)

// UserModel is the GORM model for the users table
type UserModel struct {
	ID            uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	Email         string    `gorm:"type:varchar(255);uniqueIndex;not null"`
	Password      string    `gorm:"type:varchar(255);not null"`
	Name          string    `gorm:"type:varchar(100)"`
	EmailVerified bool      `gorm:"not null;default:false"`
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

func (UserModel) TableName() string {
	return "users"
}

// EmailVerificationModel is the GORM model for the email_verifications table
type EmailVerificationModel struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	UserID    uuid.UUID `gorm:"type:uuid;not null;index"`
	Email     string    `gorm:"type:varchar(255);not null;index"`
	OTP       string    `gorm:"type:varchar(6);not null"`
	ExpiresAt time.Time `gorm:"not null"`
	Used      bool      `gorm:"not null;default:false"`
	CreatedAt time.Time
}

func (EmailVerificationModel) TableName() string {
	return "email_verifications"
}

// PasswordResetModel is the GORM model for the password_resets table
type PasswordResetModel struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	UserID    uuid.UUID `gorm:"type:uuid;not null"`
	Token     string    `gorm:"type:varchar(64);uniqueIndex;not null"`
	ExpiresAt time.Time `gorm:"not null"`
	Used      bool      `gorm:"not null;default:false"`
	CreatedAt time.Time
}

func (PasswordResetModel) TableName() string {
	return "password_resets"
}
