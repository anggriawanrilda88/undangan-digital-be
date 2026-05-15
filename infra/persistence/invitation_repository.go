package persistence

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/datatypes"

	"github.com/undangan-digital/api/domain/entities"
	domainerrors "github.com/undangan-digital/api/domain/errors"
	"github.com/undangan-digital/api/domain/repositories"
	"github.com/undangan-digital/api/infra/models"
)

type invitationRepository struct {
	db *gorm.DB
}

func NewInvitationRepository(db *gorm.DB) repositories.InvitationRepository {
	return &invitationRepository{db: db}
}

func (r *invitationRepository) Create(ctx context.Context, inv *entities.Invitation) error {
	model, err := toInvitationModel(inv)
	if err != nil {
		return err
	}
	return r.db.WithContext(ctx).Create(model).Error
}

func (r *invitationRepository) GetByID(ctx context.Context, id uuid.UUID) (*entities.Invitation, error) {
	var model models.InvitationModel
	err := r.db.WithContext(ctx).First(&model, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domainerrors.ErrInvitationNotFound
		}
		return nil, err
	}
	return toInvitationEntity(&model)
}

func (r *invitationRepository) GetBySlug(ctx context.Context, slug string) (*entities.Invitation, error) {
	var model models.InvitationModel
	err := r.db.WithContext(ctx).First(&model, "slug = ?", slug).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domainerrors.ErrInvitationNotFound
		}
		return nil, err
	}
	return toInvitationEntity(&model)
}

func (r *invitationRepository) ListByUserID(ctx context.Context, userID uuid.UUID) ([]*entities.Invitation, error) {
	var modelList []models.InvitationModel
	if err := r.db.WithContext(ctx).Where("user_id = ?", userID).Order("updated_at DESC").Find(&modelList).Error; err != nil {
		return nil, err
	}
	result := make([]*entities.Invitation, 0, len(modelList))
	for _, m := range modelList {
		inv, err := toInvitationEntity(&m)
		if err != nil {
			return nil, err
		}
		result = append(result, inv)
	}
	return result, nil
}

func (r *invitationRepository) Update(ctx context.Context, inv *entities.Invitation) error {
	model, err := toInvitationModel(inv)
	if err != nil {
		return err
	}
	return r.db.WithContext(ctx).Save(model).Error
}

func (r *invitationRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&models.InvitationModel{}, "id = ?", id).Error
}

func (r *invitationRepository) IsSlugAvailable(ctx context.Context, slug string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&models.InvitationModel{}).Where("slug = ?", slug).Count(&count).Error
	return count == 0, err
}

// ── Mappers ──────────────────────────────────────────────────────────────────

func toInvitationModel(inv *entities.Invitation) (*models.InvitationModel, error) {
	configJSON, err := json.Marshal(inv.Config)
	if err != nil {
		return nil, err
	}
	contentJSON, err := json.Marshal(inv.Content)
	if err != nil {
		return nil, err
	}
	return &models.InvitationModel{
		ID:          inv.ID,
		UserID:      inv.UserID,
		Slug:        inv.Slug,
		Status:      string(inv.Status),
		Config:      datatypes.JSON(configJSON),
		Content:     datatypes.JSON(contentJSON),
		RSVPCount:   inv.RSVPCount,
		CreatedAt:   inv.CreatedAt,
		UpdatedAt:   inv.UpdatedAt,
		PublishedAt: inv.PublishedAt,
	}, nil
}

func toInvitationEntity(m *models.InvitationModel) (*entities.Invitation, error) {
	var config entities.InvitationConfig
	if err := json.Unmarshal(m.Config, &config); err != nil {
		return nil, err
	}
	var content entities.InvitationContent
	if err := json.Unmarshal(m.Content, &content); err != nil {
		return nil, err
	}
	return &entities.Invitation{
		ID:          m.ID,
		UserID:      m.UserID,
		Slug:        m.Slug,
		Status:      entities.InvitationStatus(m.Status),
		Config:      config,
		Content:     content,
		RSVPCount:   m.RSVPCount,
		CreatedAt:   m.CreatedAt,
		UpdatedAt:   m.UpdatedAt,
		PublishedAt: m.PublishedAt,
	}, nil
}
