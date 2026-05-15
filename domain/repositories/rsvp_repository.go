package repositories

import (
	"context"

	"github.com/google/uuid"
	"github.com/undangan-digital/api/domain/entities"
)

// RSVPFilter optional filters for listing RSVPs
type RSVPFilter struct {
	Status *entities.RSVPStatus
	Page   int
	Limit  int
}

// RSVPRepository defines the contract for RSVP persistence
type RSVPRepository interface {
	// Create persists a new RSVP
	Create(ctx context.Context, rsvp *entities.RSVP) error

	// Upsert inserts or updates RSVP by (invitation_id, guest_name)
	Upsert(ctx context.Context, invitationID uuid.UUID, guestName string, status entities.RSVPStatus, guestCount int, message string) (*entities.RSVP, error)

	// ListByInvitationID retrieves RSVPs for an invitation with optional filters
	ListByInvitationID(ctx context.Context, invitationID uuid.UUID, filter RSVPFilter) ([]*entities.RSVP, int64, error)

	// GetSummary returns aggregated RSVP statistics for an invitation
	GetSummary(ctx context.Context, invitationID uuid.UUID) (*entities.RSVPSummary, error)
}
