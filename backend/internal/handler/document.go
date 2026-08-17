package handler

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"errors"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/frankyangcl/ai-support-agent/backend/internal/auth"
	"github.com/frankyangcl/ai-support-agent/backend/internal/service"
	"github.com/gin-gonic/gin"
)

type DocumentHandler struct {
	Service        *service.DocumentService
	MaxUploadBytes int64
}
type CreateDocumentRequest struct {
	Filename string `json:"filename" binding:"required"`
	Content  string `json:"content" binding:"required"`
}

func NewDocumentHandler(s *service.DocumentService, maxMB int64) *DocumentHandler {
	return &DocumentHandler{s, maxMB << 20}
}
func subject(c *gin.Context) (string, bool) {
	sub, err := auth.Subject(c.Request)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return "", false
	}
	return sub, true
}

func (h *DocumentHandler) CreateDocument(c *gin.Context) {
	owner, ok := subject(c)
	if !ok {
		return
	}
	var req CreateDocumentRequest
	if c.ShouldBindJSON(&req) != nil {
		c.JSON(400, gin.H{"error": "invalid document"})
		return
	}
	id, err := h.Service.CreateDocument(owner, req.Filename, req.Content)
	if err != nil {
		c.JSON(500, gin.H{"error": "failed to create document"})
		return
	}
	c.JSON(201, gin.H{"id": id, "filename": req.Filename, "status": "ready"})
}
func (h *DocumentHandler) ListDocuments(c *gin.Context) {
	owner, ok := subject(c)
	if !ok {
		return
	}
	docs, err := h.Service.ListDocuments(owner)
	if err != nil {
		c.JSON(500, gin.H{"error": "failed to list documents"})
		return
	}
	c.JSON(200, gin.H{"documents": docs})
}
func (h *DocumentHandler) GetDocument(c *gin.Context) {
	owner, ok := subject(c)
	if !ok {
		return
	}
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(400, gin.H{"error": "invalid document id"})
		return
	}
	doc, err := h.Service.GetDocument(owner, id)
	if errors.Is(err, sql.ErrNoRows) {
		c.JSON(404, gin.H{"error": "document not found"})
		return
	}
	if err != nil {
		c.JSON(500, gin.H{"error": "failed to get document"})
		return
	}
	c.JSON(200, doc.Document)
}

func (h *DocumentHandler) UploadDocument(c *gin.Context) {
	owner, ok := subject(c)
	if !ok {
		return
	}
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, h.MaxUploadBytes+(1<<20))
	header, err := c.FormFile("file")
	if err != nil {
		if strings.Contains(err.Error(), "request body too large") {
			c.JSON(413, gin.H{"error": "file is too large"})
		} else {
			c.JSON(400, gin.H{"error": "file is required"})
		}
		return
	}
	if header.Size <= 0 {
		c.JSON(400, gin.H{"error": "file must not be empty"})
		return
	}
	if header.Size > h.MaxUploadBytes {
		c.JSON(413, gin.H{"error": "file is too large"})
		return
	}
	name := header.Filename
	normalized := strings.ReplaceAll(name, "\\", "/")
	if name == "" || filepath.Base(normalized) != normalized || filepath.IsAbs(normalized) || strings.ToLower(filepath.Ext(normalized)) != ".pdf" {
		c.JSON(400, gin.H{"error": "invalid PDF filename"})
		return
	}
	file, err := header.Open()
	if err != nil {
		c.JSON(400, gin.H{"error": "invalid upload"})
		return
	}
	prefix := make([]byte, 5)
	n, readErr := io.ReadFull(file, prefix)
	file.Close()
	if readErr != nil && readErr != io.ErrUnexpectedEOF {
		c.JSON(400, gin.H{"error": "invalid PDF"})
		return
	}
	if n < 5 || string(prefix) != "%PDF-" {
		c.JSON(400, gin.H{"error": "file is not a valid PDF"})
		return
	}
	mime := header.Header.Get("Content-Type")
	if mime != "" && mime != "application/pdf" && mime != "application/octet-stream" {
		c.JSON(400, gin.H{"error": "invalid PDF content type"})
		return
	}
	if err = os.MkdirAll(h.Service.UploadDir, 0755); err != nil {
		c.JSON(500, gin.H{"error": "failed to store upload"})
		return
	}
	random := make([]byte, 16)
	if _, err = rand.Read(random); err != nil {
		c.JSON(500, gin.H{"error": "failed to store upload"})
		return
	}
	storage := hex.EncodeToString(random) + ".pdf"
	path := filepath.Join(h.Service.UploadDir, storage)
	if err = c.SaveUploadedFile(header, path); err != nil {
		c.JSON(500, gin.H{"error": "failed to store upload"})
		return
	}
	doc, processErr := h.Service.CreateDocumentFromPDF(c.Request.Context(), owner, name, path, storage, "application/pdf", header.Size)
	if errors.Is(processErr, service.ErrDuplicateDocument) {
		_ = os.Remove(path)
		c.JSON(409, gin.H{"error": "document already exists"})
		return
	}
	if doc == nil {
		_ = os.Remove(path)
		c.JSON(500, gin.H{"error": "failed to create document"})
		return
	}
	response := gin.H{"id": doc.ID, "filename": doc.Filename, "status": doc.Status, "created_at": doc.CreatedAt, "updated_at": doc.UpdatedAt, "file_size": doc.FileSize, "mime_type": doc.MimeType, "chunk_count": doc.ChunkCount}
	if processErr != nil {
		response["error"] = "Processing failed"
	}
	c.JSON(201, response)
}

func (h *DocumentHandler) RetryDocument(c *gin.Context) {
	owner, ok := subject(c)
	if !ok {
		return
	}
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(400, gin.H{"error": "invalid document id"})
		return
	}
	doc, err := h.Service.Retry(c.Request.Context(), owner, id)
	if errors.Is(err, sql.ErrNoRows) {
		c.JSON(404, gin.H{"error": "document not found"})
		return
	}
	if errors.Is(err, service.ErrInvalidState) {
		c.JSON(409, gin.H{"error": "document is not failed"})
		return
	}
	if doc == nil {
		c.JSON(500, gin.H{"error": "retry failed"})
		return
	}
	response := gin.H{"id": doc.ID, "filename": doc.Filename, "status": doc.Status, "updated_at": doc.UpdatedAt, "chunk_count": doc.ChunkCount}
	if err != nil {
		response["error"] = "Processing failed"
	}
	c.JSON(200, response)
}
func (h *DocumentHandler) DeleteDocument(c *gin.Context) {
	owner, ok := subject(c)
	if !ok {
		return
	}
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(400, gin.H{"error": "invalid document id"})
		return
	}
	err = h.Service.DeleteDocument(owner, id)
	if errors.Is(err, sql.ErrNoRows) {
		c.JSON(404, gin.H{"error": "document not found"})
		return
	}
	if errors.Is(err, service.ErrInvalidState) {
		c.JSON(409, gin.H{"error": "document is processing"})
		return
	}
	if err != nil {
		c.JSON(500, gin.H{"error": "failed to delete document"})
		return
	}
	c.JSON(200, gin.H{"status": "deleted", "id": id})
}
