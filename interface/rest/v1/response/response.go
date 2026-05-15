package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
	domainerrors "github.com/undangan-digital/api/domain/errors"
	"errors"
)

func OK(c *gin.Context, data any) {
	c.JSON(http.StatusOK, gin.H{"success": true, "data": data})
}

func Created(c *gin.Context, data any) {
	c.JSON(http.StatusCreated, gin.H{"success": true, "data": data})
}

func NoContent(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"success": true, "data": nil})
}

func Error(c *gin.Context, err error) {
	code, httpStatus := mapDomainError(err)
	c.JSON(httpStatus, gin.H{
		"success": false,
		"error": gin.H{
			"code":    code,
			"message": err.Error(),
		},
	})
}

func ValidationError(c *gin.Context, message string) {
	c.JSON(http.StatusBadRequest, gin.H{
		"success": false,
		"error": gin.H{
			"code":    "VALIDATION_ERROR",
			"message": message,
		},
	})
}

func BadRequest(c *gin.Context, message string) {
	c.JSON(http.StatusBadRequest, gin.H{
		"success": false,
		"error": gin.H{
			"code":    "VALIDATION_ERROR",
			"message": message,
		},
	})
}

func Conflict(c *gin.Context, message string) {
	c.JSON(http.StatusConflict, gin.H{
		"success": false,
		"error": gin.H{
			"code":    "CONFLICT",
			"message": message,
		},
	})
}

func InternalError(c *gin.Context, message string) {
	c.JSON(http.StatusInternalServerError, gin.H{
		"success": false,
		"error": gin.H{
			"code":    "INTERNAL_ERROR",
			"message": message,
		},
	})
}

func mapDomainError(err error) (string, int) {
	switch {
	case errors.Is(err, domainerrors.ErrInvitationNotFound),
		errors.Is(err, domainerrors.ErrRSVPNotFound):
		return "NOT_FOUND", http.StatusNotFound

	case errors.Is(err, domainerrors.ErrInvitationForbidden):
		return "FORBIDDEN", http.StatusForbidden

	case errors.Is(err, domainerrors.ErrSlugTaken):
		return "SLUG_TAKEN", http.StatusConflict

	case errors.Is(err, domainerrors.ErrUnauthorized),
		errors.Is(err, domainerrors.ErrInvalidToken):
		return "UNAUTHORIZED", http.StatusUnauthorized

	case errors.Is(err, domainerrors.ErrInvalidSlug):
		return "VALIDATION_ERROR", http.StatusBadRequest

	default:
		return "INTERNAL_ERROR", http.StatusInternalServerError
	}
}
