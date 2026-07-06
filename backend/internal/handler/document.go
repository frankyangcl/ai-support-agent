package handler

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
)

type DocumentHandler struct {
	DB *sql.DB
}

type CreateDocumentRequest struct {
	Filename string `json:"filename" binding:"required"`
	Content  string `json:"content" binding:"required"`
}

func NewDocumentHandler(db *sql.DB) *DocumentHandler {
	return &DocumentHandler{
		DB: db,
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

	var id int

	err := h.DB.QueryRow(
		`INSERT INTO documents (filename, content)
		 VALUES ($1, $2)
		 RETURNING id`,
		req.Filename,
		req.Content,
	).Scan(&id)

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