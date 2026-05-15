package rsvp

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/undangan-digital/api/app/dto"
	"github.com/undangan-digital/api/app/usecase"
	"github.com/undangan-digital/api/interface/rest/v1/middleware"
	"github.com/undangan-digital/api/interface/rest/v1/response"
	"math"
)

type Handler struct {
	uc *usecase.RSVPUseCase
}

func NewHandler(uc *usecase.RSVPUseCase) *Handler {
	return &Handler{uc: uc}
}

// Submit godoc
// POST /api/v1/invitations/:id/rsvp (public)
func (h *Handler) Submit(c *gin.Context) {
	invID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.ValidationError(c, "Invalid invitation ID")
		return
	}
	var req dto.CreateRSVPRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err.Error())
		return
	}
	rsvp, err := h.uc.Submit(c.Request.Context(), invID, req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Created(c, dto.ToRSVPResponse(rsvp))
}

// List godoc
// GET /api/v1/invitations/:id/rsvp (authenticated)
func (h *Handler) List(c *gin.Context) {
	userID, _ := middleware.GetUserID(c)
	invID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.ValidationError(c, "Invalid invitation ID")
		return
	}
	var req dto.ListRSVPRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.ValidationError(c, err.Error())
		return
	}

	rsvps, summary, total, err := h.uc.List(c.Request.Context(), userID, invID, req)
	if err != nil {
		response.Error(c, err)
		return
	}

	rsvpResponses := make([]*dto.RSVPResponse, 0, len(rsvps))
	for _, r := range rsvps {
		rsvpResponses = append(rsvpResponses, dto.ToRSVPResponse(r))
	}

	limit := req.Limit
	if limit == 0 {
		limit = 50
	}

	response.OK(c, dto.RSVPListResponse{
		RSVPs:   rsvpResponses,
		Summary: summary,
		Pagination: &dto.PaginationResponse{
			Page:       req.Page,
			Limit:      limit,
			Total:      total,
			TotalPages: int(math.Ceil(float64(total) / float64(limit))),
		},
	})
}
