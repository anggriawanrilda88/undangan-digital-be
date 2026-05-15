package middleware

import (
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

func jwtExpiry() time.Time {
	return time.Now().Add(7 * 24 * time.Hour)
}

func jwtNow() time.Time {
	return time.Now()
}

// GenerateJWT creates a signed JWT for the given userID with email and name claims
func GenerateJWT(userID uuid.UUID, email, name string) (string, error) {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = os.Getenv("SUPABASE_JWT_SECRET")
	}
	if secret == "" {
		return "", fmt.Errorf("JWT secret not configured")
	}
	claims := jwt.MapClaims{
		"sub":   userID.String(),
		"email": email,
		"user_metadata": map[string]interface{}{
			"full_name": name,
		},
		"iat": jwtNow().Unix(),
		"exp": jwtExpiry().Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}
