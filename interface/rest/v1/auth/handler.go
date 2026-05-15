package auth

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	domainerrors "github.com/undangan-digital/api/domain/errors"
	"github.com/undangan-digital/api/app/usecase"
	"github.com/undangan-digital/api/interface/rest/v1/middleware"
	"github.com/undangan-digital/api/interface/rest/v1/response"
)

type Handler struct {
	authUC *usecase.AuthUseCase
}

func NewHandler(authUC *usecase.AuthUseCase) *Handler {
	return &Handler{authUC: authUC}
}

type registerRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
	Name     string `json:"name" binding:"required"`
}

type loginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type verifyEmailRequest struct {
	Email string `json:"email" binding:"required,email"`
	OTP   string `json:"otp" binding:"required,len=6"`
}

type resendOTPRequest struct {
	Email string `json:"email" binding:"required,email"`
}

type forgotPasswordRequest struct {
	Email string `json:"email" binding:"required,email"`
}

type resetPasswordRequest struct {
	Token       string `json:"token" binding:"required"`
	NewPassword string `json:"newPassword" binding:"required"`
}

func (h *Handler) Register(c *gin.Context) {
	var req registerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	result, err := h.authUC.Register(c.Request.Context(), usecase.RegisterInput{
		Email:    req.Email,
		Password: req.Password,
		Name:     req.Name,
	})
	if err != nil {
		if errors.Is(err, domainerrors.ErrEmailTaken) {
			response.Conflict(c, err.Error())
			return
		}
		if errors.Is(err, domainerrors.ErrWeakPassword) {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error":   gin.H{"code": "WEAK_PASSWORD", "message": err.Error()},
			})
			return
		}
		response.InternalError(c, err.Error())
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"data": gin.H{
			"message": "Registrasi berhasil! Cek email kamu untuk kode verifikasi.",
			"user": gin.H{
				"id":    result.User.ID,
				"email": result.User.Email,
				"name":  result.User.Name,
			},
		},
	})
}

func (h *Handler) Login(c *gin.Context) {
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	result, err := h.authUC.Login(c.Request.Context(), usecase.LoginInput{
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		if errors.Is(err, domainerrors.ErrInvalidCredential) {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error":   gin.H{"code": "INVALID_CREDENTIAL", "message": err.Error()},
			})
			return
		}
		if errors.Is(err, domainerrors.ErrEmailNotVerified) {
			c.JSON(http.StatusForbidden, gin.H{
				"success": false,
				"error":   gin.H{"code": "EMAIL_NOT_VERIFIED", "message": err.Error()},
			})
			return
		}
		response.InternalError(c, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"token": result.Token,
			"user": gin.H{
				"id":    result.User.ID,
				"email": result.User.Email,
				"name":  result.User.Name,
			},
		},
	})
}

func (h *Handler) VerifyEmail(c *gin.Context) {
	var req verifyEmailRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	result, err := h.authUC.VerifyEmail(c.Request.Context(), req.Email, req.OTP)
	if err != nil {
		if errors.Is(err, domainerrors.ErrInvalidOTP) {
			c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": gin.H{"code": "INVALID_OTP", "message": err.Error()}})
			return
		}
		if errors.Is(err, domainerrors.ErrOTPExpired) {
			c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": gin.H{"code": "OTP_EXPIRED", "message": err.Error()}})
			return
		}
		if errors.Is(err, domainerrors.ErrAlreadyVerified) {
			c.JSON(http.StatusConflict, gin.H{"success": false, "error": gin.H{"code": "ALREADY_VERIFIED", "message": err.Error()}})
			return
		}
		response.InternalError(c, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"token": result.Token,
			"user": gin.H{
				"id":    result.User.ID,
				"email": result.User.Email,
				"name":  result.User.Name,
			},
		},
	})
}

func (h *Handler) ResendOTP(c *gin.Context) {
	var req resendOTPRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if err := h.authUC.ResendOTP(c.Request.Context(), req.Email); err != nil {
		if errors.Is(err, domainerrors.ErrOTPRateLimited) {
			c.JSON(http.StatusTooManyRequests, gin.H{"success": false, "error": gin.H{"code": "RATE_LIMITED", "message": err.Error()}})
			return
		}
		if errors.Is(err, domainerrors.ErrAlreadyVerified) {
			c.JSON(http.StatusConflict, gin.H{"success": false, "error": gin.H{"code": "ALREADY_VERIFIED", "message": err.Error()}})
			return
		}
		response.InternalError(c, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": gin.H{"message": "OTP berhasil dikirim"}})
}

func (h *Handler) ForgotPassword(c *gin.Context) {
	var req forgotPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	// Non-fatal — selalu 200
	_ = h.authUC.ForgotPassword(c.Request.Context(), req.Email)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    gin.H{"message": "Jika email terdaftar, link reset akan dikirim"},
	})
}

func (h *Handler) ResetPassword(c *gin.Context) {
	var req resetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if err := h.authUC.ResetPassword(c.Request.Context(), req.Token, req.NewPassword); err != nil {
		if errors.Is(err, domainerrors.ErrInvalidToken) {
			c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": gin.H{"code": "INVALID_TOKEN", "message": err.Error()}})
			return
		}
		if errors.Is(err, domainerrors.ErrTokenExpired) {
			c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": gin.H{"code": "TOKEN_EXPIRED", "message": err.Error()}})
			return
		}
		if errors.Is(err, domainerrors.ErrWeakPassword) {
			c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": gin.H{"code": "WEAK_PASSWORD", "message": err.Error()}})
			return
		}
		response.InternalError(c, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": gin.H{"message": "Password berhasil diubah"}})
}

func (h *Handler) Me(c *gin.Context) {
	userID, email, name := middleware.GetUserProfile(c)
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    gin.H{"id": userID, "email": email, "name": name},
	})
}
