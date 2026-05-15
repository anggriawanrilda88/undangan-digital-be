package usecase

import (
	"context"
	"regexp"

	"github.com/google/uuid"
	"github.com/undangan-digital/api/app/dto"
	"github.com/undangan-digital/api/domain/entities"
	domainerrors "github.com/undangan-digital/api/domain/errors"
	"github.com/undangan-digital/api/domain/repositories"
)

var slugRegex = regexp.MustCompile(`^[a-z0-9-]+$`)

type InvitationUseCase struct {
	invRepo repositories.InvitationRepository
}

func NewInvitationUseCase(invRepo repositories.InvitationRepository) *InvitationUseCase {
	return &InvitationUseCase{invRepo: invRepo}
}

func (uc *InvitationUseCase) Create(ctx context.Context, userID uuid.UUID, req dto.CreateInvitationRequest) (*entities.Invitation, error) {
	if !slugRegex.MatchString(req.Slug) {
		return nil, domainerrors.ErrInvalidSlug
	}

	available, err := uc.invRepo.IsSlugAvailable(ctx, req.Slug)
	if err != nil {
		return nil, err
	}
	if !available {
		return nil, domainerrors.ErrSlugTaken
	}

	content := entities.InvitationContent{
		GroomName:       req.Content.GroomName,
		BrideName:       req.Content.BrideName,
		GroomParents:    req.Content.GroomParents,
		BrideParents:    req.Content.BrideParents,
		AkadDate:        req.Content.AkadDate,
		AkadVenue:       toVenueContentPtr(req.Content.AkadVenue),
		ReceptionDate:   req.Content.ReceptionDate,
		Venue:           entities.VenueContent(req.Content.Venue),
		OpeningMessage:  req.Content.OpeningMessage,
		DigitalEnvelope: req.Content.DigitalEnvelope,
	}

	inv := entities.NewInvitation(userID, req.Slug, req.Config, content)
	if err := uc.invRepo.Create(ctx, inv); err != nil {
		return nil, err
	}
	return inv, nil
}

func (uc *InvitationUseCase) GetByID(ctx context.Context, userID, invID uuid.UUID) (*entities.Invitation, error) {
	inv, err := uc.invRepo.GetByID(ctx, invID)
	if err != nil {
		return nil, err
	}
	if inv.UserID != userID {
		return nil, domainerrors.ErrInvitationForbidden
	}
	return inv, nil
}

func (uc *InvitationUseCase) GetPublicBySlug(ctx context.Context, slug string) (*entities.Invitation, error) {
	inv, err := uc.invRepo.GetBySlug(ctx, slug)
	if err != nil {
		return nil, err
	}
	if !inv.IsPublic() {
		return nil, domainerrors.ErrInvitationNotFound
	}
	return inv, nil
}

func (uc *InvitationUseCase) ListByUser(ctx context.Context, userID uuid.UUID) ([]*entities.Invitation, error) {
	return uc.invRepo.ListByUserID(ctx, userID)
}

func (uc *InvitationUseCase) Update(ctx context.Context, userID, invID uuid.UUID, req dto.UpdateInvitationRequest) (*entities.Invitation, error) {
	inv, err := uc.invRepo.GetByID(ctx, invID)
	if err != nil {
		return nil, err
	}
	if inv.UserID != userID {
		return nil, domainerrors.ErrInvitationForbidden
	}

	if req.Slug != nil && *req.Slug != inv.Slug {
		if !slugRegex.MatchString(*req.Slug) {
			return nil, domainerrors.ErrInvalidSlug
		}
		available, err := uc.invRepo.IsSlugAvailable(ctx, *req.Slug)
		if err != nil {
			return nil, err
		}
		if !available {
			return nil, domainerrors.ErrSlugTaken
		}
		inv.Slug = *req.Slug
	}

	if req.Config != nil {
		inv.Config = *req.Config
	}
	if req.Content != nil {
		inv.Content = entities.InvitationContent{
			GroomName:       req.Content.GroomName,
			BrideName:       req.Content.BrideName,
			GroomParents:    req.Content.GroomParents,
			BrideParents:    req.Content.BrideParents,
			AkadDate:        req.Content.AkadDate,
			AkadVenue:       toVenueContentPtr(req.Content.AkadVenue),
			ReceptionDate:   req.Content.ReceptionDate,
			Venue:           entities.VenueContent(req.Content.Venue),
			OpeningMessage:  req.Content.OpeningMessage,
			DigitalEnvelope: req.Content.DigitalEnvelope,
		}
	}
	if req.Status != nil {
		switch *req.Status {
		case entities.InvitationStatusPublished:
			// BE-S02-5: enforce max 1 published per user
			count, err := uc.invRepo.CountPublishedByUser(ctx, userID)
			if err != nil {
				return nil, err
			}
			if count > 0 && inv.Status != entities.InvitationStatusPublished {
				return nil, domainerrors.ErrAlreadyHasPublished
			}
			inv.Publish()
		case entities.InvitationStatusArchived:
			inv.Archive()
		default:
			inv.Status = *req.Status
		}
	}
	// BE-S02-4: status tidak diubah jika field status tidak dikirim

	if err := uc.invRepo.Update(ctx, inv); err != nil {
		return nil, err
	}
	return inv, nil
}

func (uc *InvitationUseCase) Delete(ctx context.Context, userID, invID uuid.UUID) error {
	inv, err := uc.invRepo.GetByID(ctx, invID)
	if err != nil {
		return err
	}
	if inv.UserID != userID {
		return domainerrors.ErrInvitationForbidden
	}
	return uc.invRepo.Delete(ctx, invID)
}

func (uc *InvitationUseCase) CheckSlug(ctx context.Context, slug string) (bool, error) {
	if !slugRegex.MatchString(slug) {
		return false, domainerrors.ErrInvalidSlug
	}
	return uc.invRepo.IsSlugAvailable(ctx, slug)
}

func toVenueContentPtr(v *dto.VenueDTO) *entities.VenueContent {
	if v == nil {
		return nil
	}
	vc := entities.VenueContent(*v)
	return &vc
}
