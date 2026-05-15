package persistence

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/undangan-digital/api/domain/entities"
	domainerrors "github.com/undangan-digital/api/domain/errors"
	"github.com/undangan-digital/api/domain/repositories"
	"github.com/undangan-digital/api/infra/models"
)

type userRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) repositories.UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) Create(ctx context.Context, user *entities.User) error {
	m := &models.UserModel{
		ID:            user.ID,
		Email:         user.Email,
		Password:      user.Password,
		Name:          user.Name,
		EmailVerified: user.EmailVerified,
	}
	if err := r.db.WithContext(ctx).Create(m).Error; err != nil {
		return err
	}
	user.ID = m.ID
	user.CreatedAt = m.CreatedAt
	user.UpdatedAt = m.UpdatedAt
	return nil
}

func (r *userRepository) FindByEmail(ctx context.Context, email string) (*entities.User, error) {
	var m models.UserModel
	err := r.db.WithContext(ctx).Where("email = ?", email).First(&m).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, domainerrors.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return toUserEntity(&m), nil
}

func (r *userRepository) FindByID(ctx context.Context, id uuid.UUID) (*entities.User, error) {
	var m models.UserModel
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&m).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, domainerrors.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return toUserEntity(&m), nil
}

func (r *userRepository) UpdateEmailVerified(ctx context.Context, userID uuid.UUID, verified bool) error {
	return r.db.WithContext(ctx).Model(&models.UserModel{}).
		Where("id = ?", userID).
		Update("email_verified", verified).Error
}

func (r *userRepository) UpdatePassword(ctx context.Context, userID uuid.UUID, hashedPassword string) error {
	return r.db.WithContext(ctx).Model(&models.UserModel{}).
		Where("id = ?", userID).
		Update("password", hashedPassword).Error
}

func toUserEntity(m *models.UserModel) *entities.User {
	return &entities.User{
		ID:            m.ID,
		Email:         m.Email,
		Password:      m.Password,
		Name:          m.Name,
		EmailVerified: m.EmailVerified,
		CreatedAt:     m.CreatedAt,
		UpdatedAt:     m.UpdatedAt,
	}
}

// ── EmailVerification Repository ──────────────────────────────────────────────

type emailVerificationRepository struct {
	db *gorm.DB
}

func NewEmailVerificationRepository(db *gorm.DB) repositories.EmailVerificationRepository {
	return &emailVerificationRepository{db: db}
}

func (r *emailVerificationRepository) Create(ctx context.Context, ev *entities.EmailVerification) error {
	m := &models.EmailVerificationModel{
		ID:        ev.ID,
		UserID:    ev.UserID,
		Email:     ev.Email,
		OTP:       ev.OTP,
		ExpiresAt: ev.ExpiresAt,
		Used:      false,
	}
	return r.db.WithContext(ctx).Create(m).Error
}

func (r *emailVerificationRepository) FindLatestByEmail(ctx context.Context, email string) (*entities.EmailVerification, error) {
	var m models.EmailVerificationModel
	err := r.db.WithContext(ctx).
		Where("email = ? AND used = false", email).
		Order("created_at DESC").
		First(&m).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, domainerrors.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &entities.EmailVerification{
		ID:        m.ID,
		UserID:    m.UserID,
		Email:     m.Email,
		OTP:       m.OTP,
		ExpiresAt: m.ExpiresAt,
		Used:      m.Used,
		CreatedAt: m.CreatedAt,
	}, nil
}

func (r *emailVerificationRepository) MarkUsed(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Model(&models.EmailVerificationModel{}).
		Where("id = ?", id).Update("used", true).Error
}

func (r *emailVerificationRepository) CountRecentByEmail(ctx context.Context, email string, since time.Time) (int, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&models.EmailVerificationModel{}).
		Where("email = ? AND created_at >= ?", email, since).
		Count(&count).Error
	return int(count), err
}

// ── PasswordReset Repository ──────────────────────────────────────────────────

type passwordResetRepository struct {
	db *gorm.DB
}

func NewPasswordResetRepository(db *gorm.DB) repositories.PasswordResetRepository {
	return &passwordResetRepository{db: db}
}

func (r *passwordResetRepository) Create(ctx context.Context, pr *entities.PasswordReset) error {
	m := &models.PasswordResetModel{
		ID:        pr.ID,
		UserID:    pr.UserID,
		Token:     pr.Token,
		ExpiresAt: pr.ExpiresAt,
		Used:      false,
	}
	return r.db.WithContext(ctx).Create(m).Error
}

func (r *passwordResetRepository) FindByToken(ctx context.Context, token string) (*entities.PasswordReset, error) {
	var m models.PasswordResetModel
	err := r.db.WithContext(ctx).Where("token = ?", token).First(&m).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, domainerrors.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &entities.PasswordReset{
		ID:        m.ID,
		UserID:    m.UserID,
		Token:     m.Token,
		ExpiresAt: m.ExpiresAt,
		Used:      m.Used,
		CreatedAt: m.CreatedAt,
	}, nil
}

func (r *passwordResetRepository) MarkUsed(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Model(&models.PasswordResetModel{}).
		Where("id = ?", id).Update("used", true).Error
}
