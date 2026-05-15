package dto

import (
	"time"

	"github.com/google/uuid"
	"github.com/undangan-digital/api/domain/entities"
)

// ── Request DTOs ──────────────────────────────────────────────────────────────

type CreateInvitationRequest struct {
	Slug    string                        `json:"slug" binding:"required,min=3,max=60,alphanum_dash"`
	Config  entities.InvitationConfig     `json:"config"`
	Content CreateInvitationContentRequest `json:"content" binding:"required"`
}

type CreateInvitationContentRequest struct {
	GroomName      string     `json:"groomName" binding:"required"`
	BrideName      string     `json:"brideName" binding:"required"`
	GroomParents   string     `json:"groomParents"`
	BrideParents   string     `json:"brideParents"`
	AkadDate       *time.Time `json:"akadDate"`
	AkadVenue      *VenueDTO  `json:"akadVenue"`
	ReceptionDate  time.Time  `json:"receptionDate" binding:"required"`
	Venue          VenueDTO   `json:"venue" binding:"required"`
	OpeningMessage string     `json:"openingMessage"`
	DigitalEnvelope *entities.DigitalEnvelope `json:"digitalEnvelope"`
}

type VenueDTO struct {
	Name    string `json:"name" binding:"required"`
	Address string `json:"address" binding:"required"`
	MapsURL string `json:"mapsUrl"`
}

type UpdateInvitationRequest struct {
	Slug    *string                        `json:"slug"`
	Status  *entities.InvitationStatus     `json:"status"`
	Config  *entities.InvitationConfig     `json:"config"`
	Content *CreateInvitationContentRequest `json:"content"`
}

// ── Response DTOs ─────────────────────────────────────────────────────────────

type InvitationResponse struct {
	ID          uuid.UUID                   `json:"id"`
	UserID      uuid.UUID                   `json:"userId"`
	Slug        string                      `json:"slug"`
	Status      entities.InvitationStatus   `json:"status"`
	Config      entities.InvitationConfig   `json:"config"`
	Content     entities.InvitationContent  `json:"content"`
	RSVPCount   int                         `json:"rsvpCount"`
	CreatedAt   time.Time                   `json:"createdAt"`
	UpdatedAt   time.Time                   `json:"updatedAt"`
	PublishedAt *time.Time                  `json:"publishedAt,omitempty"`
}

type InvitationSummaryResponse struct {
	ID            uuid.UUID                 `json:"id"`
	Slug          string                    `json:"slug"`
	Status        entities.InvitationStatus `json:"status"`
	GroomName     string                    `json:"groomName"`
	BrideName     string                    `json:"brideName"`
	ReceptionDate time.Time                 `json:"receptionDate"`
	RSVPCount     int                       `json:"rsvpCount"`
	UpdatedAt     time.Time                 `json:"updatedAt"`
}

type PublicInvitationResponse struct {
	Slug    string                     `json:"slug"`
	Config  entities.InvitationConfig  `json:"config"`
	Content entities.InvitationContent `json:"content"`
}

type SlugCheckResponse struct {
	Available bool   `json:"available"`
	Slug      string `json:"slug"`
}

func ToInvitationResponse(inv *entities.Invitation) *InvitationResponse {
	return &InvitationResponse{
		ID:          inv.ID,
		UserID:      inv.UserID,
		Slug:        inv.Slug,
		Status:      inv.Status,
		Config:      inv.Config,
		Content:     inv.Content,
		RSVPCount:   inv.RSVPCount,
		CreatedAt:   inv.CreatedAt,
		UpdatedAt:   inv.UpdatedAt,
		PublishedAt: inv.PublishedAt,
	}
}

func ToInvitationSummaryResponse(inv *entities.Invitation) *InvitationSummaryResponse {
	return &InvitationSummaryResponse{
		ID:            inv.ID,
		Slug:          inv.Slug,
		Status:        inv.Status,
		GroomName:     inv.Content.GroomName,
		BrideName:     inv.Content.BrideName,
		ReceptionDate: inv.Content.ReceptionDate,
		RSVPCount:     inv.RSVPCount,
		UpdatedAt:     inv.UpdatedAt,
	}
}

func ToPublicInvitationResponse(inv *entities.Invitation) *PublicInvitationResponse {
	return &PublicInvitationResponse{
		Slug:    inv.Slug,
		Config:  inv.Config,
		Content: inv.Content,
	}
}
