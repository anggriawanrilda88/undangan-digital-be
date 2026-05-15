package usecase

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"os"
	"regexp"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"github.com/undangan-digital/api/domain/entities"
	domainerrors "github.com/undangan-digital/api/domain/errors"
	"github.com/undangan-digital/api/domain/repositories"
	"github.com/undangan-digital/api/infra/email"
	"github.com/undangan-digital/api/interface/rest/v1/middleware"
)

type AuthUseCase struct {
	userRepo    repositories.UserRepository
	evRepo      repositories.EmailVerificationRepository
	prRepo      repositories.PasswordResetRepository
	emailSvc    *email.Service
}

func NewAuthUseCase(
	userRepo repositories.UserRepository,
	evRepo repositories.EmailVerificationRepository,
	prRepo repositories.PasswordResetRepository,
	emailSvc *email.Service,
) *AuthUseCase {
	return &AuthUseCase{userRepo: userRepo, evRepo: evRepo, prRepo: prRepo, emailSvc: emailSvc}
}

// ── BE-S02-1: Password validation ────────────────────────────────────────────

var (
	hasUpperRegex = regexp.MustCompile(`[A-Z]`)
	hasLowerRegex = regexp.MustCompile(`[a-z]`)
	hasDigitRegex = regexp.MustCompile(`[0-9]`)
)

func validatePasswordStrength(password string) error {
	if len(password) < 8 || !hasUpperRegex.MatchString(password) ||
		!hasLowerRegex.MatchString(password) || !hasDigitRegex.MatchString(password) {
		return domainerrors.ErrWeakPassword
	}
	return nil
}

// ── Register ─────────────────────────────────────────────────────────────────

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
	// BE-S02-1: password strength
	if err := validatePasswordStrength(input.Password); err != nil {
		return nil, err
	}

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
		ID:            uuid.New(),
		Email:         input.Email,
		Password:      string(hashed),
		Name:          input.Name,
		EmailVerified: false,
	}
	if err := uc.userRepo.Create(ctx, user); err != nil {
		return nil, err
	}

	// Send OTP
	if err := uc.sendOTP(ctx, user); err != nil {
		// Non-fatal: user tetap terbuat, OTP bisa di-resend
		fmt.Printf("[WARN] Gagal kirim OTP ke %s: %v\n", user.Email, err)
	}

	// Return tanpa token — user harus verify dulu
	return &AuthResult{Token: "", User: user}, nil
}

// ── Login ────────────────────────────────────────────────────────────────────

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

	// BE-S02-2: block login jika belum verified
	if !user.EmailVerified {
		return nil, domainerrors.ErrEmailNotVerified
	}

	token, err := middleware.GenerateJWT(user.ID, user.Email, user.Name)
	if err != nil {
		return nil, err
	}

	return &AuthResult{Token: token, User: user}, nil
}

// ── BE-S02-2: OTP Verification ───────────────────────────────────────────────

func (uc *AuthUseCase) VerifyEmail(ctx context.Context, email, otp string) (*AuthResult, error) {
	user, err := uc.userRepo.FindByEmail(ctx, email)
	if errors.Is(err, domainerrors.ErrNotFound) {
		return nil, domainerrors.ErrInvalidOTP
	}
	if err != nil {
		return nil, err
	}

	if user.EmailVerified {
		return nil, domainerrors.ErrAlreadyVerified
	}

	ev, err := uc.evRepo.FindLatestByEmail(ctx, email)
	if errors.Is(err, domainerrors.ErrNotFound) {
		return nil, domainerrors.ErrInvalidOTP
	}
	if err != nil {
		return nil, err
	}

	if time.Now().After(ev.ExpiresAt) {
		return nil, domainerrors.ErrOTPExpired
	}
	if ev.OTP != otp {
		return nil, domainerrors.ErrInvalidOTP
	}

	// Mark OTP used + verify user
	if err := uc.evRepo.MarkUsed(ctx, ev.ID); err != nil {
		return nil, err
	}
	if err := uc.userRepo.UpdateEmailVerified(ctx, user.ID, true); err != nil {
		return nil, err
	}
	user.EmailVerified = true

	token, err := middleware.GenerateJWT(user.ID, user.Email, user.Name)
	if err != nil {
		return nil, err
	}
	return &AuthResult{Token: token, User: user}, nil
}

func (uc *AuthUseCase) ResendOTP(ctx context.Context, emailAddr string) error {
	user, err := uc.userRepo.FindByEmail(ctx, emailAddr)
	if errors.Is(err, domainerrors.ErrNotFound) {
		// Jangan expose apakah email terdaftar
		return nil
	}
	if err != nil {
		return err
	}
	if user.EmailVerified {
		return domainerrors.ErrAlreadyVerified
	}

	// Rate limit: max 3x per 10 menit
	since := time.Now().Add(-10 * time.Minute)
	count, err := uc.evRepo.CountRecentByEmail(ctx, emailAddr, since)
	if err != nil {
		return err
	}
	if count >= 3 {
		return domainerrors.ErrOTPRateLimited
	}

	return uc.sendOTP(ctx, user)
}

func (uc *AuthUseCase) sendOTP(ctx context.Context, user *entities.User) error {
	otp, err := generateOTP()
	if err != nil {
		return err
	}

	ev := &entities.EmailVerification{
		ID:        uuid.New(),
		UserID:    user.ID,
		Email:     user.Email,
		OTP:       otp,
		ExpiresAt: time.Now().Add(10 * time.Minute),
	}
	if err := uc.evRepo.Create(ctx, ev); err != nil {
		return err
	}
	return uc.emailSvc.SendOTP(user.Email, user.Name, otp)
}

func generateOTP() (string, error) {
	n, err := rand.Int(rand.Reader, big.NewInt(1000000))
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%06d", n.Int64()), nil
}

// ── BE-S02-3: Forgot / Reset Password ────────────────────────────────────────

func (uc *AuthUseCase) ForgotPassword(ctx context.Context, emailAddr string) error {
	user, err := uc.userRepo.FindByEmail(ctx, emailAddr)
	if errors.Is(err, domainerrors.ErrNotFound) {
		// Selalu return nil — jangan expose apakah email terdaftar
		return nil
	}
	if err != nil {
		return err
	}

	token, err := generateResetToken()
	if err != nil {
		return err
	}

	pr := &entities.PasswordReset{
		ID:        uuid.New(),
		UserID:    user.ID,
		Token:     token,
		ExpiresAt: time.Now().Add(1 * time.Hour),
	}
	if err := uc.prRepo.Create(ctx, pr); err != nil {
		return err
	}

	appBaseURL := getEnvOrDefault("APP_BASE_URL", "https://undangan-digital.anggriawan.my.id")
	resetLink := fmt.Sprintf("%s/auth/reset-password?token=%s", appBaseURL, token)
	return uc.emailSvc.SendPasswordReset(user.Email, user.Name, resetLink)
}

func (uc *AuthUseCase) ResetPassword(ctx context.Context, token, newPassword string) error {
	if err := validatePasswordStrength(newPassword); err != nil {
		return err
	}

	pr, err := uc.prRepo.FindByToken(ctx, token)
	if errors.Is(err, domainerrors.ErrNotFound) {
		return domainerrors.ErrInvalidToken
	}
	if err != nil {
		return err
	}

	if pr.Used {
		return domainerrors.ErrInvalidToken
	}
	if time.Now().After(pr.ExpiresAt) {
		return domainerrors.ErrTokenExpired
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	if err := uc.userRepo.UpdatePassword(ctx, pr.UserID, string(hashed)); err != nil {
		return err
	}
	return uc.prRepo.MarkUsed(ctx, pr.ID)
}

func generateResetToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func getEnvOrDefault(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
