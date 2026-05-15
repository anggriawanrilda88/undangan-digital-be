package entities

import (
	"time"

	"github.com/google/uuid"
)

// InvitationStatus represents the lifecycle state of an invitation
type InvitationStatus string

const (
	InvitationStatusDraft     InvitationStatus = "draft"
	InvitationStatusPublished InvitationStatus = "published"
	InvitationStatusArchived  InvitationStatus = "archived"
)

// Invitation is the core domain entity
type Invitation struct {
	ID          uuid.UUID        `json:"id"`
	UserID      uuid.UUID        `json:"userId"`
	Slug        string           `json:"slug"`
	Status      InvitationStatus `json:"status"`
	Config      InvitationConfig `json:"config"`
	Content     InvitationContent `json:"content"`
	RSVPCount   int              `json:"rsvpCount"`
	CreatedAt   time.Time        `json:"createdAt"`
	UpdatedAt   time.Time        `json:"updatedAt"`
	PublishedAt *time.Time       `json:"publishedAt,omitempty"`
}

// InvitationConfig holds template/visual settings
type InvitationConfig struct {
	TemplateID  string      `json:"templateId"`
	Colors      ColorConfig `json:"colors"`
	Fonts       FontConfig  `json:"fonts"`
	CouplePhoto string      `json:"couplePhoto,omitempty"`
	Music       MusicConfig `json:"music"`
}

type ColorConfig struct {
	Primary   string `json:"primary"`
	Secondary string `json:"secondary"`
}

type FontConfig struct {
	Heading string `json:"heading"`
	Body    string `json:"body"`
}

type MusicConfig struct {
	Enabled bool   `json:"enabled"`
	URL     string `json:"url,omitempty"`
}

// InvitationContent holds the actual wedding data filled by the couple
type InvitationContent struct {
	GroomName       string          `json:"groomName"`
	BrideName       string          `json:"brideName"`
	GroomParents    string          `json:"groomParents,omitempty"`
	BrideParents    string          `json:"brideParents,omitempty"`
	AkadDate        *time.Time      `json:"akadDate,omitempty"`
	AkadVenue       *VenueContent   `json:"akadVenue,omitempty"` // nullable — bisa beda tempat dari resepsi
	ReceptionDate   time.Time       `json:"receptionDate"`
	Venue           VenueContent    `json:"venue"`
	OpeningMessage  string          `json:"openingMessage,omitempty"`
	DigitalEnvelope *DigitalEnvelope `json:"digitalEnvelope,omitempty"`
}

type VenueContent struct {
	Name    string `json:"name"`
	Address string `json:"address"`
	MapsURL string `json:"mapsUrl,omitempty"`
}

type DigitalEnvelope struct {
	BankAccounts []BankAccount `json:"bankAccounts,omitempty"`
	QRISImageURL string        `json:"qrisImageUrl,omitempty"`
}

type BankAccount struct {
	BankName      string `json:"bankName"`
	AccountNumber string `json:"accountNumber"`
	AccountName   string `json:"accountName"`
}

func NewInvitation(userID uuid.UUID, slug string, config InvitationConfig, content InvitationContent) *Invitation {
	return &Invitation{
		ID:      uuid.New(),
		UserID:  userID,
		Slug:    slug,
		Status:  InvitationStatusDraft,
		Config:  config,
		Content: content,
	}
}

func (i *Invitation) Publish() {
	now := time.Now()
	i.Status = InvitationStatusPublished
	i.PublishedAt = &now
}

func (i *Invitation) Archive() {
	i.Status = InvitationStatusArchived
}

func (i *Invitation) IsPublic() bool {
	return i.Status == InvitationStatusPublished
}
