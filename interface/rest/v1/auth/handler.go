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
	Password string `json:"password" binding:"required,min=8"`
	Name     string `json:"name" binding:"required"`
}

type loginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
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
		response.InternalError(c, err.Error())
		return
	}

	c.JSON(http.StatusCreated, gin.H{
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

func (h *Handler) Me(c *gin.Context) {
	userID, _ := middleware.GetUserID(c)
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    gin.H{"id": userID},
	})
}
