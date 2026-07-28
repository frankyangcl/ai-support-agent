package handler

import (
	"net/http"

	"github.com/frankyangcl/ai-support-agent/backend/internal/service"
	"github.com/gin-gonic/gin"
)

type DocumentHandler struct {
	Service *service.DocumentService
}

type CreateDocumentRequest struct {
	Filename string `json:"filename" binding:"required"`
	Content  string `json:"content" binding:"required"`
}

func NewDocumentHandler(service *service.DocumentService) *DocumentHandler {
	return &DocumentHandler{
		Service: service,
	}
}

func (h *DocumentHandler) CreateDocument(c *gin.Context) {
	var req CreateDocumentRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	id, err := h.Service.CreateDocument(req.Filename, req.Content)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"id":       id,
		"filename": req.Filename,
	})
}

func (h *DocumentHandler) ListDocuments(c *gin.Context) {
	documents, err := h.Service.ListDocuments()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"documents": documents,
	})

	c.JSON(http.StatusOK, gin.H{
		"documents": documents,
	})
}
