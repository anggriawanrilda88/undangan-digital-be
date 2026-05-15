package v1

import (
	"os"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/undangan-digital/api/app/usecase"
	"github.com/undangan-digital/api/infra/persistence"
	"github.com/undangan-digital/api/interface/rest/v1/invitation"
	"github.com/undangan-digital/api/interface/rest/v1/middleware"
	"github.com/undangan-digital/api/interface/rest/v1/rsvp"
)

func NewRouter(db *gorm.DB) *gin.Engine {
	if os.Getenv("ENV") == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()
	r.Use(gin.Logger())
	r.Use(gin.Recovery())
	r.Use(corsMiddleware())

	// ── Repositories
	invRepo := persistence.NewInvitationRepository(db)
	rsvpRepo := persistence.NewRSVPRepository(db)

	// ── Use Cases
	invUC := usecase.NewInvitationUseCase(invRepo)
	rsvpUC := usecase.NewRSVPUseCase(rsvpRepo, invRepo)

	// ── Handlers
	invHandler := invitation.NewHandler(invUC)
	rsvpHandler := rsvp.NewHandler(rsvpUC)

	api := r.Group("/api/v1")

	// Health check
	r.GET("/api/v1/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// ── Public routes (no auth) ──────────────────────────────────
	api.GET("/i/:slug", invHandler.GetPublicBySlug)
	api.POST("/invitations/:id/rsvp", rsvpHandler.Submit)

	// ── Auth middleware ──────────────────────────────────────────
	auth := api.Group("/")
	auth.Use(middleware.AuthMiddleware())
	{
		// Auth
		auth.GET("/auth/me", func(c *gin.Context) {
			userID, email, name := middleware.GetUserProfile(c)
			c.JSON(200, gin.H{
				"success": true,
				"data": gin.H{
					"id":    userID,
					"email": email,
					"name":  name,
				},
			})
		})

		// Invitations
		auth.GET("/invitations", invHandler.List)
		auth.POST("/invitations", invHandler.Create)
		auth.GET("/invitations/:id", invHandler.GetByID)
		auth.PUT("/invitations/:id", invHandler.Update)
		auth.DELETE("/invitations/:id", invHandler.Delete)
		auth.GET("/invitations/:id/rsvp", rsvpHandler.List)

		// Slug check — blocker US-04
		auth.GET("/slugs/check", invHandler.CheckSlug)
	}

	return r
}

func corsMiddleware() gin.HandlerFunc {
	allowedOrigins := map[string]bool{
		"https://undangan-digital.anggriawan.my.id": true,
		"http://localhost:3000":                    true, // local dev Next.js
	}

	// Allow additional origin from env (e.g. staging)
	if extra := os.Getenv("FRONTEND_URL"); extra != "" {
		allowedOrigins[extra] = true
	}

	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")
		if allowedOrigins[origin] {
			c.Header("Access-Control-Allow-Origin", origin)
		}
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Authorization, Content-Type")
		c.Header("Vary", "Origin")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	}
}
