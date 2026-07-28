package handler

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

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

const maxUploadSize = 10 << 20 // 10 MB

func (h *DocumentHandler) UploadDocument(c *gin.Context) {
	fileHeader, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "file is required",
		})
		return
	}

	if fileHeader.Size > maxUploadSize {
		c.JSON(http.StatusRequestEntityTooLarge, gin.H{
			"error": "file size must not exceed 10 MB",
		})
		return
	}

	extension := strings.ToLower(filepath.Ext(fileHeader.Filename))
	if extension != ".pdf" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "only PDF files are allowed",
		})
		return
	}

	uploadDir := "uploads"

	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to create upload directory",
		})
		return
	}

	storedFilename := fmt.Sprintf(
		"%d_%s",
		time.Now().UnixNano(),
		filepath.Base(fileHeader.Filename),
	)

	destination := filepath.Join(uploadDir, storedFilename)

	if err := c.SaveUploadedFile(fileHeader, destination); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to save uploaded file",
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"filename":         fileHeader.Filename,
		"stored_filename":  storedFilename,
		"size":             fileHeader.Size,
		"content_type":     fileHeader.Header.Get("Content-Type"),
		"storage_location": destination,
	})
}
