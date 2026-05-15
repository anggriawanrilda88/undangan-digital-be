package entities

import (
	"time"

	"github.com/google/uuid"
)

// RSVPStatus represents a guest's attendance response
type RSVPStatus string

const (
	RSVPStatusAttending    RSVPStatus = "attending"
	RSVPStatusNotAttending RSVPStatus = "not_attending"
	RSVPStatusMaybe        RSVPStatus = "maybe"
)

// RSVP is the core domain entity for guest responses
type RSVP struct {
	ID           uuid.UUID  `json:"id"`
	InvitationID uuid.UUID  `json:"invitationId"`
	GuestName    string     `json:"guestName"`
	Status       RSVPStatus `json:"status"`
	GuestCount   int        `json:"guestCount"`
	Message      string     `json:"message,omitempty"`
	CreatedAt    time.Time  `json:"createdAt"`
}

// RSVPSummary provides aggregated RSVP statistics
type RSVPSummary struct {
	TotalResponses int `json:"totalResponses"`
	Attending      int `json:"attending"`
	NotAttending   int `json:"notAttending"`
	Maybe          int `json:"maybe"`
	TotalGuests    int `json:"totalGuests"`
}

func NewRSVP(invitationID uuid.UUID, guestName string, status RSVPStatus, guestCount int, message string) *RSVP {
	return &RSVP{
		ID:           uuid.New(),
		InvitationID: invitationID,
		GuestName:    guestName,
		Status:       status,
		GuestCount:   guestCount,
		Message:      message,
	}
}
