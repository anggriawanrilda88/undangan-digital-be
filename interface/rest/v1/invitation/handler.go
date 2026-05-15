package invitation

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/undangan-digital/api/app/dto"
	"github.com/undangan-digital/api/app/usecase"
	"github.com/undangan-digital/api/interface/rest/v1/middleware"
	"github.com/undangan-digital/api/interface/rest/v1/response"
)

type Handler struct {
	uc *usecase.InvitationUseCase
}

func NewHandler(uc *usecase.InvitationUseCase) *Handler {
	return &Handler{uc: uc}
}

// List godoc
// GET /api/v1/invitations
func (h *Handler) List(c *gin.Context) {
	userID, _ := middleware.GetUserID(c)
	invitations, err := h.uc.ListByUser(c.Request.Context(), userID)
	if err != nil {
		response.Error(c, err)
		return
	}
	summaries := make([]*dto.InvitationSummaryResponse, 0, len(invitations))
	for _, inv := range invitations {
		summaries = append(summaries, dto.ToInvitationSummaryResponse(inv))
	}
	response.OK(c, summaries)
}

// Create godoc
// POST /api/v1/invitations
func (h *Handler) Create(c *gin.Context) {
	userID, _ := middleware.GetUserID(c)
	var req dto.CreateInvitationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err.Error())
		return
	}
	inv, err := h.uc.Create(c.Request.Context(), userID, req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Created(c, dto.ToInvitationResponse(inv))
}

// GetByID godoc
// GET /api/v1/invitations/:id
func (h *Handler) GetByID(c *gin.Context) {
	userID, _ := middleware.GetUserID(c)
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.ValidationError(c, "Invalid invitation ID")
		return
	}
	inv, err := h.uc.GetByID(c.Request.Context(), userID, id)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, dto.ToInvitationResponse(inv))
}

// Update godoc
// PUT /api/v1/invitations/:id
func (h *Handler) Update(c *gin.Context) {
	userID, _ := middleware.GetUserID(c)
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.ValidationError(c, "Invalid invitation ID")
		return
	}
	var req dto.UpdateInvitationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err.Error())
		return
	}
	inv, err := h.uc.Update(c.Request.Context(), userID, id, req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, dto.ToInvitationResponse(inv))
}

// Delete godoc
// DELETE /api/v1/invitations/:id
func (h *Handler) Delete(c *gin.Context) {
	userID, _ := middleware.GetUserID(c)
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.ValidationError(c, "Invalid invitation ID")
		return
	}
	if err := h.uc.Delete(c.Request.Context(), userID, id); err != nil {
		response.Error(c, err)
		return
	}
	response.NoContent(c)
}

// CheckSlug godoc
// GET /api/v1/slugs/check?slug=xxx
func (h *Handler) CheckSlug(c *gin.Context) {
	slug := c.Query("slug")
	if slug == "" {
		response.ValidationError(c, "slug query parameter is required")
		return
	}
	available, err := h.uc.CheckSlug(c.Request.Context(), slug)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, dto.SlugCheckResponse{Available: available, Slug: slug})
}

// GetPublicBySlug godoc
// GET /api/v1/i/:slug (public)
func (h *Handler) GetPublicBySlug(c *gin.Context) {
	slug := c.Param("slug")
	inv, err := h.uc.GetPublicBySlug(c.Request.Context(), slug)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   gin.H{"code": "NOT_FOUND", "message": "Undangan tidak ditemukan"},
		})
		return
	}
	response.OK(c, dto.ToPublicInvitationResponse(inv))
}
