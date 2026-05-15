package dto

import (
	"time"

	"github.com/google/uuid"
	"github.com/undangan-digital/api/domain/entities"
)

// ── Request DTOs ──────────────────────────────────────────────────────────────

type CreateRSVPRequest struct {
	GuestName  string             `json:"guestName" binding:"required,min=2,max=100"`
	Status     entities.RSVPStatus `json:"status" binding:"required,oneof=attending not_attending maybe"`
	GuestCount int                `json:"guestCount" binding:"min=0,max=10"`
	Message    string             `json:"message" binding:"max=500"`
}

type ListRSVPRequest struct {
	Status string `form:"status" binding:"omitempty,oneof=attending not_attending maybe"`
	Page   int    `form:"page,default=1" binding:"min=1"`
	Limit  int    `form:"limit,default=50" binding:"min=1,max=100"`
}

// ── Response DTOs ─────────────────────────────────────────────────────────────

type RSVPResponse struct {
	ID           uuid.UUID          `json:"id"`
	InvitationID uuid.UUID          `json:"invitationId"`
	GuestName    string             `json:"guestName"`
	Status       entities.RSVPStatus `json:"status"`
	GuestCount   int                `json:"guestCount"`
	Message      string             `json:"message,omitempty"`
	CreatedAt    time.Time          `json:"createdAt"`
}

type RSVPListResponse struct {
	RSVPs      []*RSVPResponse        `json:"rsvps"`
	Summary    *entities.RSVPSummary  `json:"summary"`
	Pagination *PaginationResponse    `json:"pagination"`
}

type PaginationResponse struct {
	Page       int   `json:"page"`
	Limit      int   `json:"limit"`
	Total      int64 `json:"total"`
	TotalPages int   `json:"totalPages"`
}

func ToRSVPResponse(rsvp *entities.RSVP) *RSVPResponse {
	return &RSVPResponse{
		ID:           rsvp.ID,
		InvitationID: rsvp.InvitationID,
		GuestName:    rsvp.GuestName,
		Status:       rsvp.Status,
		GuestCount:   rsvp.GuestCount,
		Message:      rsvp.Message,
		CreatedAt:    rsvp.CreatedAt,
	}
}
