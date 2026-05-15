package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
)

// OptionalAuthMiddleware tries to verify JWT if present, but does NOT abort if missing/invalid.
// Use for public endpoints that need to know caller identity when available.
func OptionalAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.Next()
			return
		}
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			c.Next()
			return
		}
		userID, err := verifyJWT(parts[1])
		if err != nil {
			c.Next()
			return
		}
		c.Set(ContextKeyUserID, userID)
		c.Next()
	}
}
