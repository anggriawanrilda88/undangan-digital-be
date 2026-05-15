package repositories

import (
	"context"

	"github.com/google/uuid"
	"github.com/undangan-digital/api/domain/entities"
)

// InvitationRepository defines the contract for invitation persistence
type InvitationRepository interface {
	// Create persists a new invitation
	Create(ctx context.Context, invitation *entities.Invitation) error

	// GetByID retrieves an invitation by UUID (owner access)
	GetByID(ctx context.Context, id uuid.UUID) (*entities.Invitation, error)

	// GetBySlug retrieves an invitation by slug (public access)
	GetBySlug(ctx context.Context, slug string) (*entities.Invitation, error)

	// ListByUserID retrieves all invitations for a given user
	ListByUserID(ctx context.Context, userID uuid.UUID) ([]*entities.Invitation, error)

	// Update persists changes to an existing invitation
	Update(ctx context.Context, invitation *entities.Invitation) error

	// Delete removes an invitation by UUID
	Delete(ctx context.Context, id uuid.UUID) error

	// IsSlugAvailable checks if a slug is not taken
	IsSlugAvailable(ctx context.Context, slug string) (bool, error)

	// CountPublishedByUser counts published invitations for a user
	CountPublishedByUser(ctx context.Context, userID uuid.UUID) (int64, error)
}
