package persistence

import (
	"context"
	"math"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/undangan-digital/api/domain/entities"
	"github.com/undangan-digital/api/domain/repositories"
	"github.com/undangan-digital/api/infra/models"
)

type rsvpRepository struct {
	db *gorm.DB
}

func NewRSVPRepository(db *gorm.DB) repositories.RSVPRepository {
	return &rsvpRepository{db: db}
}

func (r *rsvpRepository) Create(ctx context.Context, rsvp *entities.RSVP) error {
	model := &models.RSVPModel{
		ID:           rsvp.ID,
		InvitationID: rsvp.InvitationID,
		GuestName:    rsvp.GuestName,
		Status:       string(rsvp.Status),
		GuestCount:   rsvp.GuestCount,
		Message:      rsvp.Message,
	}
	if err := r.db.WithContext(ctx).Create(model).Error; err != nil {
		return err
	}
	// Update RSVP count on invitation
	r.db.WithContext(ctx).Model(&models.InvitationModel{}).
		Where("id = ?", rsvp.InvitationID).
		UpdateColumn("rsvp_count", gorm.Expr("rsvp_count + 1"))
	return nil
}

// Upsert BE-BUG-003: insert or update RSVP by (invitation_id, guest_name)
func (r *rsvpRepository) Upsert(ctx context.Context, invitationID uuid.UUID, guestName string, status entities.RSVPStatus, guestCount int, message string) (*entities.RSVP, error) {
	var existing models.RSVPModel
	err := r.db.WithContext(ctx).
		Where("invitation_id = ? AND guest_name = ?", invitationID, guestName).
		First(&existing).Error

	if err == nil {
		// Update existing record
		existing.Status = string(status)
		existing.GuestCount = guestCount
		existing.Message = message
		if err := r.db.WithContext(ctx).Save(&existing).Error; err != nil {
			return nil, err
		}
		return &entities.RSVP{
			ID:           existing.ID,
			InvitationID: existing.InvitationID,
			GuestName:    existing.GuestName,
			Status:       entities.RSVPStatus(existing.Status),
			GuestCount:   existing.GuestCount,
			Message:      existing.Message,
			CreatedAt:    existing.CreatedAt,
		}, nil
	}

	// Insert new record
	rsvp := entities.NewRSVP(invitationID, guestName, status, guestCount, message)
	model := &models.RSVPModel{
		ID:           rsvp.ID,
		InvitationID: rsvp.InvitationID,
		GuestName:    rsvp.GuestName,
		Status:       string(rsvp.Status),
		GuestCount:   rsvp.GuestCount,
		Message:      rsvp.Message,
	}
	if err := r.db.WithContext(ctx).Create(model).Error; err != nil {
		return nil, err
	}
	// Increment rsvp_count only on new insert
	r.db.WithContext(ctx).Model(&models.InvitationModel{}).
		Where("id = ?", invitationID).
		UpdateColumn("rsvp_count", gorm.Expr("rsvp_count + 1"))
	return rsvp, nil
}

func (r *rsvpRepository) ListByInvitationID(ctx context.Context, invitationID uuid.UUID, filter repositories.RSVPFilter) ([]*entities.RSVP, int64, error) {
	query := r.db.WithContext(ctx).Model(&models.RSVPModel{}).Where("invitation_id = ?", invitationID)

	if filter.Status != nil {
		query = query.Where("status = ?", string(*filter.Status))
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	page := filter.Page
	if page < 1 {
		page = 1
	}
	limit := filter.Limit
	if limit < 1 {
		limit = 50
	}
	offset := (page - 1) * limit

	var modelList []models.RSVPModel
	if err := query.Order("created_at DESC").Offset(offset).Limit(limit).Find(&modelList).Error; err != nil {
		return nil, 0, err
	}

	_ = math.Ceil(float64(total) / float64(limit)) // precompute if needed

	result := make([]*entities.RSVP, 0, len(modelList))
	for _, m := range modelList {
		result = append(result, &entities.RSVP{
			ID:           m.ID,
			InvitationID: m.InvitationID,
			GuestName:    m.GuestName,
			Status:       entities.RSVPStatus(m.Status),
			GuestCount:   m.GuestCount,
			Message:      m.Message,
			CreatedAt:    m.CreatedAt,
		})
	}
	return result, total, nil
}

func (r *rsvpRepository) GetSummary(ctx context.Context, invitationID uuid.UUID) (*entities.RSVPSummary, error) {
	type row struct {
		Status     string
		Count      int
		TotalGuests int
	}
	var rows []row
	err := r.db.WithContext(ctx).
		Model(&models.RSVPModel{}).
		Select("status, COUNT(*) as count, SUM(guest_count) as total_guests").
		Where("invitation_id = ?", invitationID).
		Group("status").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}

	summary := &entities.RSVPSummary{}
	for _, row := range rows {
		summary.TotalResponses += row.Count
		switch entities.RSVPStatus(row.Status) {
		case entities.RSVPStatusAttending:
			summary.Attending = row.Count
			summary.TotalGuests = row.TotalGuests
		case entities.RSVPStatusNotAttending:
			summary.NotAttending = row.Count
		case entities.RSVPStatusMaybe:
			summary.Maybe = row.Count
		}
	}
	return summary, nil
}
