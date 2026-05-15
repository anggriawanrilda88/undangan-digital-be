package middleware

import (
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

func jwtExpiry() time.Time {
	return time.Now().Add(24 * time.Hour)
}

func jwtNow() time.Time {
	return time.Now()
}

// GenerateJWT creates a signed JWT for the given userID
func GenerateJWT(userID uuid.UUID) (string, error) {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = os.Getenv("SUPABASE_JWT_SECRET")
	}
	if secret == "" {
		return "", fmt.Errorf("JWT secret not configured")
	}
	claims := jwt.MapClaims{
		"sub": userID.String(),
		"iat": jwtNow().Unix(),
		"exp": jwtExpiry().Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}
