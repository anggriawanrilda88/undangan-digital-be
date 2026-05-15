package usecase

import (
	"context"
	"regexp"
	"strings"

	"github.com/google/uuid"
	"github.com/undangan-digital/api/app/dto"
	"github.com/undangan-digital/api/domain/entities"
	domainerrors "github.com/undangan-digital/api/domain/errors"
	"github.com/undangan-digital/api/domain/repositories"
)

// sanitizeInput strips HTML tags and trims whitespace
var htmlTagRegex = regexp.MustCompile(`<[^>]*>`)

func sanitizeInput(s string) string {
	s = htmlTagRegex.ReplaceAllString(s, "")
	return strings.TrimSpace(s)
}

type RSVPUseCase struct {
	rsvpRepo repositories.RSVPRepository
	invRepo  repositories.InvitationRepository
}

func NewRSVPUseCase(rsvpRepo repositories.RSVPRepository, invRepo repositories.InvitationRepository) *RSVPUseCase {
	return &RSVPUseCase{rsvpRepo: rsvpRepo, invRepo: invRepo}
}

func (uc *RSVPUseCase) Submit(ctx context.Context, invitationID uuid.UUID, req dto.CreateRSVPRequest) (*entities.RSVP, error) {
	// Verify invitation exists and is published
	inv, err := uc.invRepo.GetByID(ctx, invitationID)
	if err != nil {
		return nil, err
	}
	if !inv.IsPublic() {
		return nil, domainerrors.ErrInvitationNotFound
	}

	// BE-BUG-002: not_attending harus selalu 0 guest
	guestCount := req.GuestCount
	if req.Status == entities.RSVPStatusNotAttending {
		guestCount = 0
	} else if guestCount == 0 {
		// attending/maybe tanpa explicit guestCount → default 1
		guestCount = 1
	}

	// BE-BUG-004: sanitize input dari HTML/XSS
	guestName := sanitizeInput(req.GuestName)
	message := sanitizeInput(req.Message)

	// BE-BUG-003: upsert — kalau sudah ada (invitation_id, guest_name), update
	rsvp, err := uc.rsvpRepo.Upsert(ctx, invitationID, guestName, req.Status, guestCount, message)
	if err != nil {
		return nil, err
	}
	return rsvp, nil
}

func (uc *RSVPUseCase) List(ctx context.Context, userID, invitationID uuid.UUID, req dto.ListRSVPRequest) ([]*entities.RSVP, *entities.RSVPSummary, int64, error) {
	// Verify ownership
	inv, err := uc.invRepo.GetByID(ctx, invitationID)
	if err != nil {
		return nil, nil, 0, err
	}
	if inv.UserID != userID {
		return nil, nil, 0, domainerrors.ErrInvitationForbidden
	}

	filter := repositories.RSVPFilter{
		Page:  req.Page,
		Limit: req.Limit,
	}
	if req.Status != "" {
		status := entities.RSVPStatus(req.Status)
		filter.Status = &status
	}

	rsvps, total, err := uc.rsvpRepo.ListByInvitationID(ctx, invitationID, filter)
	if err != nil {
		return nil, nil, 0, err
	}

	summary, err := uc.rsvpRepo.GetSummary(ctx, invitationID)
	if err != nil {
		return nil, nil, 0, err
	}

	return rsvps, summary, total, nil
}
