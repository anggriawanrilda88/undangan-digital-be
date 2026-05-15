package entities

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID            uuid.UUID
	Email         string
	Password      string // hashed
	Name          string
	EmailVerified bool
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

type EmailVerification struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	Email     string
	OTP       string
	ExpiresAt time.Time
	Used      bool
	CreatedAt time.Time
}

type PasswordReset struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	Token     string
	ExpiresAt time.Time
	Used      bool
	CreatedAt time.Time
}
