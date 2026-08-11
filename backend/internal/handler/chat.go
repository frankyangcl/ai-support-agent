package handler

import (
	"net/http"
	"strings"

	"github.com/frankyangcl/ai-support-agent/backend/internal/service"
	"github.com/gin-gonic/gin"
)

type ChatHandler struct {
	RAGService *service.RAGService
}

type ChatRequest struct {
	Question string `json:"question" binding:"required"`
}

func NewChatHandler(
	ragService *service.RAGService,
) *ChatHandler {
	return &ChatHandler{
		RAGService: ragService,
	}
}

func (h *ChatHandler) Chat(c *gin.Context) {
	var req ChatRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "question is required",
		})
		return
	}

	req.Question = strings.TrimSpace(req.Question)

	if req.Question == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "question must not be empty",
		})
		return
	}

	result, err := h.RAGService.Ask(
		c.Request.Context(),
		req.Question,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	sources := make([]gin.H, 0, len(result.Sources))

	for _, source := range result.Sources {
		sources = append(sources, gin.H{
			"document_id": source.DocumentID,
			"chunk_index": source.ChunkIndex,
			"distance":    source.Distance,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"answer":  result.Answer,
		"sources": sources,
	})
}
