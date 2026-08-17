package handler

import (
	"database/sql"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type HealthHandler struct {
	DB *sql.DB
}

func NewHealthHandler(db *sql.DB) *HealthHandler {
	return &HealthHandler{
		DB: db,
	}
}

func (h *HealthHandler) Health(c *gin.Context) {
	c.JSON(200, gin.H{
		"status": "ok",
	})
}

func (h *HealthHandler) DatabaseHealth(c *gin.Context) {
	if err := h.DB.Ping(); err != nil {
		c.JSON(500, gin.H{
			"database": "error",
			"error":    err.Error(),
		})
		return
	}

	c.JSON(200, gin.H{
		"database": "ok",
	})
}

func (h *DocumentHandler) GetDocument(c *gin.Context) {

	id, err := strconv.Atoi(c.Param("id"))

	if err != nil {

		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid id",
		})

		return
	}

	doc, chunks, err := h.Service.GetDocument(id)

	if err != nil {

		c.JSON(http.StatusNotFound, gin.H{
			"error": err.Error(),
		})

		return
	}

	type ChunkResponse struct {
		Index          int    `json:"index"`
		CharacterCount int    `json:"character_count"`
		Preview        string `json:"preview"`
	}

	resp := make([]ChunkResponse, 0, len(chunks))

	for _, c := range chunks {

		preview := []rune(c.Content)

		if len(preview) > 120 {
			preview = preview[:120]
		}

		resp = append(resp, ChunkResponse{
			Index:          c.ChunkIndex,
			CharacterCount: c.CharacterCount,
			Preview:        string(preview),
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"document": gin.H{
			"id":              doc.ID,
			"filename":        doc.Filename,
			"created_at":      doc.CreatedAt,
			"character_count": len([]rune(doc.Content)),
		},
		"chunks": resp,
	})
}
