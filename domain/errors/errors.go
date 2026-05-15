package errors

import "errors"

// Domain errors — gunakan ini di usecase, jangan expose error DB/infra langsung
var (
	// Invitation errors
	ErrInvitationNotFound  = errors.New("invitation not found")
	ErrInvitationForbidden = errors.New("you do not have access to this invitation")
	ErrSlugTaken           = errors.New("slug is already taken")
	ErrInvitationNotPublic = errors.New("invitation is not published")

	// Generic errors
	ErrNotFound = errors.New("not found")

	// RSVP errors
	ErrRSVPNotFound = errors.New("rsvp not found")

	// Auth errors
	ErrUnauthorized      = errors.New("unauthorized")
	ErrInvalidToken      = errors.New("invalid or expired token")
	ErrEmailTaken        = errors.New("email already registered")
	ErrInvalidCredential = errors.New("invalid email or password")

	// Validation errors
	ErrInvalidSlug = errors.New("slug must be 3–60 lowercase alphanumeric characters and hyphens")
)
