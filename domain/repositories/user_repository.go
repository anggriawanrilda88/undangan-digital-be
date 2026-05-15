package repositories

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/undangan-digital/api/domain/entities"
)

type UserRepository interface {
	Create(ctx context.Context, user *entities.User) error
	FindByEmail(ctx context.Context, email string) (*entities.User, error)
	FindByID(ctx context.Context, id uuid.UUID) (*entities.User, error)
	UpdateEmailVerified(ctx context.Context, userID uuid.UUID, verified bool) error
	UpdatePassword(ctx context.Context, userID uuid.UUID, hashedPassword string) error
}

type EmailVerificationRepository interface {
	Create(ctx context.Context, ev *entities.EmailVerification) error
	FindLatestByEmail(ctx context.Context, email string) (*entities.EmailVerification, error)
	MarkUsed(ctx context.Context, id uuid.UUID) error
	CountRecentByEmail(ctx context.Context, email string, since time.Time) (int, error)
}

type PasswordResetRepository interface {
	Create(ctx context.Context, pr *entities.PasswordReset) error
	FindByToken(ctx context.Context, token string) (*entities.PasswordReset, error)
	MarkUsed(ctx context.Context, id uuid.UUID) error
}
