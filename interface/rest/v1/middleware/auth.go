package middleware

import (
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

const ContextKeyUserID = "userID"
const ContextKeyEmail = "email"
const ContextKeyName = "name"

// AuthMiddleware verifies Supabase JWT and injects userID into context
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error": gin.H{
					"code":    "UNAUTHORIZED",
					"message": "Authorization header is required",
				},
			})
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error": gin.H{
					"code":    "UNAUTHORIZED",
					"message": "Invalid authorization format. Use: Bearer <token>",
				},
			})
			return
		}

		tokenStr := parts[1]
		userID, err := verifySupabaseJWT(tokenStr)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error": gin.H{
					"code":    "INVALID_TOKEN",
					"message": "Token is invalid or expired",
				},
			})
			return
		}

		c.Set(ContextKeyUserID, userID)

		// Inject email & name from JWT claims (BE-BUG-001)
		if tokenStr != "" {
			if email, name, err2 := extractEmailName(tokenStr); err2 == nil {
				c.Set(ContextKeyEmail, email)
				c.Set(ContextKeyName, name)
			}
		}
		c.Next()
	}
}

func verifySupabaseJWT(tokenStr string) (uuid.UUID, error) {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = os.Getenv("SUPABASE_JWT_SECRET")
	}
	if secret == "" {
		return uuid.Nil, fmt.Errorf("SUPABASE_JWT_SECRET not configured")
	}

	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secret), nil
	})
	if err != nil || !token.Valid {
		return uuid.Nil, fmt.Errorf("invalid token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return uuid.Nil, fmt.Errorf("invalid claims")
	}

	sub, ok := claims["sub"].(string)
	if !ok || sub == "" {
		return uuid.Nil, fmt.Errorf("missing sub claim")
	}

	userID, err := uuid.Parse(sub)
	if err != nil {
		return uuid.Nil, fmt.Errorf("invalid user id in token")
	}

	return userID, nil
}

// GetUserID extracts userID from gin context (use in handlers after AuthMiddleware)
func GetUserID(c *gin.Context) (uuid.UUID, bool) {
	val, exists := c.Get(ContextKeyUserID)
	if !exists {
		return uuid.Nil, false
	}
	userID, ok := val.(uuid.UUID)
	return userID, ok
}

// GetUserProfile extracts id, email, name from gin context
func GetUserProfile(c *gin.Context) (uuid.UUID, string, string) {
	userID, _ := GetUserID(c)
	email, _ := c.Get(ContextKeyEmail)
	name, _ := c.Get(ContextKeyName)
	emailStr, _ := email.(string)
	nameStr, _ := name.(string)
	return userID, emailStr, nameStr
}

// extractEmailName parses JWT claims for email and name without re-verifying signature
func extractEmailName(tokenStr string) (string, string, error) {
	parser := jwt.NewParser(jwt.WithoutClaimsValidation())
	token, _, err := parser.ParseUnverified(tokenStr, jwt.MapClaims{})
	if err != nil {
		return "", "", err
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", "", fmt.Errorf("invalid claims")
	}
	email, _ := claims["email"].(string)
	// Supabase stores name in user_metadata.full_name or raw_user_meta_data
	name := ""
	if meta, ok := claims["user_metadata"].(map[string]interface{}); ok {
		name, _ = meta["full_name"].(string)
		if name == "" {
			name, _ = meta["name"].(string)
		}
	}
	return email, name, nil
}

// verifyJWT is an alias for verifySupabaseJWT (used by optional_auth.go)
func verifyJWT(tokenStr string) (uuid.UUID, error) {
	return verifySupabaseJWT(tokenStr)
}
