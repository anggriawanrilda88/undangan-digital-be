package v1

import (
	"os"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/undangan-digital/api/app/usecase"
	"github.com/undangan-digital/api/infra/email"
	"github.com/undangan-digital/api/infra/persistence"
	"github.com/undangan-digital/api/infra/storage"
	"github.com/undangan-digital/api/interface/rest/v1/auth"
	"github.com/undangan-digital/api/interface/rest/v1/invitation"
	"github.com/undangan-digital/api/interface/rest/v1/middleware"
	"github.com/undangan-digital/api/interface/rest/v1/rsvp"
	"github.com/undangan-digital/api/interface/rest/v1/upload"
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
	userRepo := persistence.NewUserRepository(db)
	evRepo := persistence.NewEmailVerificationRepository(db)
	prRepo := persistence.NewPasswordResetRepository(db)
	invRepo := persistence.NewInvitationRepository(db)
	rsvpRepo := persistence.NewRSVPRepository(db)

	// ── Services
	emailSvc := email.NewService()

	// ── Use Cases
	authUC := usecase.NewAuthUseCase(userRepo, evRepo, prRepo, emailSvc)
	invUC := usecase.NewInvitationUseCase(invRepo)
	rsvpUC := usecase.NewRSVPUseCase(rsvpRepo, invRepo)

	// ── Handlers
	authHandler := auth.NewHandler(authUC)
	invHandler := invitation.NewHandler(invUC)
	rsvpHandler := rsvp.NewHandler(rsvpUC)

	// Upload handler (MinIO)
	minioClient, err := storage.NewMinIOClient()
	if err != nil {
		panic("failed to init MinIO: " + err.Error())
	}
	uploadHandler := upload.NewHandler(minioClient)

	api := r.Group("/api/v1")

	// Health check
	r.GET("/api/v1/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// ── Public routes (no auth) ──────────────────────────────────
	api.POST("/auth/register", authHandler.Register)
	api.POST("/auth/login", authHandler.Login)
	api.POST("/auth/verify-email", authHandler.VerifyEmail)
	api.POST("/auth/resend-otp", authHandler.ResendOTP)
	api.POST("/auth/forgot-password", authHandler.ForgotPassword)
	api.POST("/auth/reset-password", authHandler.ResetPassword)
	api.GET("/i/:slug", invHandler.GetPublicBySlug)
	api.POST("/invitations/:id/rsvp", rsvpHandler.Submit)

	// ── Auth middleware ──────────────────────────────────────────
	protected := api.Group("/")
	protected.Use(middleware.AuthMiddleware())
	{
		// Auth
		protected.GET("/auth/me", authHandler.Me)

		// Invitations
		protected.GET("/invitations", invHandler.List)
		protected.POST("/invitations", invHandler.Create)
		protected.GET("/invitations/:id", invHandler.GetByID)
		protected.PUT("/invitations/:id", invHandler.Update)
		protected.DELETE("/invitations/:id", invHandler.Delete)
		protected.GET("/invitations/:id/rsvp", rsvpHandler.List)

		// Slug check
		protected.GET("/slugs/check", invHandler.CheckSlug)

		// Upload
		protected.POST("/upload/image", uploadHandler.UploadImage)
	}

	return r
}

func corsMiddleware() gin.HandlerFunc {
	allowedOrigins := map[string]bool{
		"https://undangan-digital.anggriawan.my.id": true,
		"http://localhost:3000":                     true,
	}

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
