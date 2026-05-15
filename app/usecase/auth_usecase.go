package usecase

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"github.com/undangan-digital/api/domain/entities"
	domainerrors "github.com/undangan-digital/api/domain/errors"
	"github.com/undangan-digital/api/domain/repositories"
	"github.com/undangan-digital/api/interface/rest/v1/middleware"
)

type AuthUseCase struct {
	userRepo repositories.UserRepository
}

func NewAuthUseCase(userRepo repositories.UserRepository) *AuthUseCase {
	return &AuthUseCase{userRepo: userRepo}
}

type RegisterInput struct {
	Email    string
	Password string
	Name     string
}

type AuthResult struct {
	Token string
	User  *entities.User
}

func (uc *AuthUseCase) Register(ctx context.Context, input RegisterInput) (*AuthResult, error) {
	// Check duplicate email
	_, err := uc.userRepo.FindByEmail(ctx, input.Email)
	if err == nil {
		return nil, domainerrors.ErrEmailTaken
	}
	if !errors.Is(err, domainerrors.ErrNotFound) {
		return nil, err
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := &entities.User{
		ID:       uuid.New(),
		Email:    input.Email,
		Password: string(hashed),
		Name:     input.Name,
	}
	if err := uc.userRepo.Create(ctx, user); err != nil {
		return nil, err
	}

	token, err := middleware.GenerateJWT(user.ID)
	if err != nil {
		return nil, err
	}

	return &AuthResult{Token: token, User: user}, nil
}

type LoginInput struct {
	Email    string
	Password string
}

func (uc *AuthUseCase) Login(ctx context.Context, input LoginInput) (*AuthResult, error) {
	user, err := uc.userRepo.FindByEmail(ctx, input.Email)
	if errors.Is(err, domainerrors.ErrNotFound) {
		return nil, domainerrors.ErrInvalidCredential
	}
	if err != nil {
		return nil, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(input.Password)); err != nil {
		return nil, domainerrors.ErrInvalidCredential
	}

	token, err := middleware.GenerateJWT(user.ID)
	if err != nil {
		return nil, err
	}

	return &AuthResult{Token: token, User: user}, nil
}
