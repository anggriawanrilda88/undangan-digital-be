package upload

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/undangan-digital/api/infra/storage"
)

type Handler struct {
	minio *storage.MinIOClient
}

func NewHandler(minio *storage.MinIOClient) *Handler {
	return &Handler{minio: minio}
}

// UploadImage handles POST /api/v1/upload/image (multipart/form-data, field: "file")
// Response: { success: true, data: { url: "..." } }
func (h *Handler) UploadImage(c *gin.Context) {
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   gin.H{"code": "BAD_REQUEST", "message": "field 'file' is required"},
		})
		return
	}
	defer file.Close()

	ct := header.Header.Get("Content-Type")
	allowed := map[string]bool{
		"image/jpeg": true,
		"image/png":  true,
		"image/webp": true,
		"image/gif":  true,
	}
	if !allowed[ct] {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   gin.H{"code": "INVALID_FILE_TYPE", "message": "only jpeg, png, webp, gif allowed"},
		})
		return
	}

	if header.Size > 5*1024*1024 {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   gin.H{"code": "FILE_TOO_LARGE", "message": "max file size is 5MB"},
		})
		return
	}

	publicURL, err := h.minio.UploadFile(c.Request.Context(), file, header)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   gin.H{"code": "UPLOAD_FAILED", "message": err.Error()},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    gin.H{"url": publicURL},
	})
}

type presignRequest struct {
	Filename    string `json:"filename" binding:"required"`
	ContentType string `json:"contentType" binding:"required"`
}

// PresignUpload handles POST /api/v1/upload/presign
func (h *Handler) PresignUpload(c *gin.Context) {
	var req presignRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   gin.H{"code": "BAD_REQUEST", "message": err.Error()},
		})
		return
	}

	allowed := map[string]bool{
		"image/jpeg": true,
		"image/png":  true,
		"image/webp": true,
		"image/gif":  true,
	}
	if !allowed[req.ContentType] {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   gin.H{"code": "INVALID_FILE_TYPE", "message": "only jpeg, png, webp, gif allowed"},
		})
		return
	}

	result, err := h.minio.PresignPut(c.Request.Context(), req.Filename, req.ContentType)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   gin.H{"code": "PRESIGN_FAILED", "message": err.Error()},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"uploadUrl": result.UploadURL,
			"publicUrl": result.PublicURL,
		},
	})
}
