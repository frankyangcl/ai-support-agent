package handler

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/frankyangcl/ai-support-agent/backend/internal/auth"
	"github.com/frankyangcl/ai-support-agent/backend/internal/repository"
	"github.com/frankyangcl/ai-support-agent/backend/internal/service"
	"github.com/gin-gonic/gin"
)

type ChatHandler struct {
	RAGService    ChatService
	Conversations *repository.ConversationRepository
	HistoryLimit  int
}

type ChatService interface {
	Ask(context.Context, string, string) (*service.RAGResult, error)
	Stream(context.Context, string, string, func([]repository.ChunkSearchResult) error, func(string) error) ([]repository.ChunkSearchResult, error)
	AskWithHistory(context.Context, string, string, []service.HistoryMessage) (*service.RAGResult, error)
	StreamWithHistory(context.Context, string, string, []service.HistoryMessage, func([]repository.ChunkSearchResult) error, func(string) error) ([]repository.ChunkSearchResult, error)
}

type ChatRequest struct {
	Question       string `json:"question" binding:"required"`
	ConversationID *int64 `json:"conversation_id,omitempty"`
}

func NewChatHandler(
	ragService ChatService, options ...any,
) *ChatHandler {
	h := &ChatHandler{RAGService: ragService, HistoryLimit: 20}
	if len(options) > 0 {
		h.Conversations, _ = options[0].(*repository.ConversationRepository)
	}
	if len(options) > 1 {
		if limit, ok := options[1].(int); ok {
			h.HistoryLimit = limit
		}
	}
	return h
}

func (h *ChatHandler) Stream(c *gin.Context) {
	var req ChatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "question is required"})
		return
	}
	req.Question = strings.TrimSpace(req.Question)
	if req.Question == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "question must not be empty"})
		return
	}

	started := false
	emit := func(event string, data any) error {
		payload, err := json.Marshal(data)
		if err != nil {
			return err
		}
		if _, err := fmt.Fprintf(c.Writer, "event: %s\ndata: %s\n\n", event, payload); err != nil {
			return err
		}
		c.Writer.Flush()
		return nil
	}

	ownerSub, err := auth.Subject(c.Request)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	if h.Conversations == nil {
		h.streamLegacy(c, ownerSub, req.Question, emit, &started)
		return
	}
	turn, err := h.Conversations.BeginTurn(c.Request.Context(), ownerSub, req.ConversationID, req.Question, h.HistoryLimit)
	if errors.Is(err, repository.ErrGenerationConflict) {
		c.JSON(http.StatusConflict, gin.H{"error": "generation already active"})
		return
	}
	if errors.Is(err, sql.ErrNoRows) {
		c.JSON(http.StatusNotFound, gin.H{"error": "conversation not found"})
		return
	}
	if err != nil {
		c.JSON(500, gin.H{"error": "failed to start chat"})
		return
	}
	history := historyMessages(turn.History)
	var answer strings.Builder
	sources, err := h.RAGService.StreamWithHistory(c.Request.Context(), ownerSub, req.Question, history,
		func(_ []repository.ChunkSearchResult) error {
			c.Header("Content-Type", "text/event-stream")
			c.Header("Cache-Control", "no-cache")
			c.Header("X-Accel-Buffering", "no")
			c.Status(http.StatusOK)
			started = true
			return emit("start", gin.H{"conversation_id": turn.ConversationID, "user_message_id": turn.UserMessageID, "assistant_message_id": turn.AssistantMessageID})
		},
		func(text string) error { answer.WriteString(text); return emit("delta", gin.H{"text": text}) },
	)
	if err != nil {
		if errors.Is(err, context.Canceled) || c.Request.Context().Err() != nil {
			h.finishTerminal(turn.AssistantMessageID, "cancelled", answer.String(), nil)
			return
		}
		h.finishTerminal(turn.AssistantMessageID, "failed", answer.String(), nil)
		if !started {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to start chat stream"})
			return
		}
		_ = emit("error", gin.H{"error": "stream interrupted"})
		return
	}
	if err = h.finishTerminal(turn.AssistantMessageID, "completed", answer.String(), sourceResponses(sources)); err != nil {
		_ = emit("error", gin.H{"error": "stream persistence failed"})
		return
	}
	if err := emit("citations", gin.H{"citations": sourceResponses(sources)}); err != nil {
		return
	}
	_ = emit("done", gin.H{})
}

func sourceResponses(sources []repository.ChunkSearchResult) []gin.H {
	result := make([]gin.H, 0, len(sources))
	for _, source := range sources {
		preview := []rune(source.Content)
		if len(preview) > 200 {
			preview = preview[:200]
		}
		result = append(result, gin.H{"document_id": source.DocumentID, "filename": source.Filename, "chunk_index": source.ChunkIndex, "distance": source.Distance, "preview": string(preview)})
	}
	return result
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

	ownerSub, err := auth.Subject(c.Request)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	if h.Conversations == nil {
		result, e := h.RAGService.Ask(c.Request.Context(), ownerSub, req.Question)
		if e != nil {
			c.JSON(500, gin.H{"error": e.Error()})
			return
		}
		c.JSON(200, gin.H{"answer": result.Answer, "sources": sourceResponses(result.Sources)})
		return
	}
	turn, err := h.Conversations.BeginTurn(c.Request.Context(), ownerSub, req.ConversationID, req.Question, h.HistoryLimit)
	if errors.Is(err, repository.ErrGenerationConflict) {
		c.JSON(409, gin.H{"error": "generation already active"})
		return
	}
	if errors.Is(err, sql.ErrNoRows) {
		c.JSON(404, gin.H{"error": "conversation not found"})
		return
	}
	if err != nil {
		c.JSON(500, gin.H{"error": "failed to start chat"})
		return
	}
	result, err := h.RAGService.AskWithHistory(
		c.Request.Context(),
		ownerSub,
		req.Question,
		historyMessages(turn.History),
	)
	if err != nil {
		h.finishTerminal(turn.AssistantMessageID, "failed", "", nil)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "chat generation failed",
		})
		return
	}
	if err = h.finishTerminal(turn.AssistantMessageID, "completed", result.Answer, sourceResponses(result.Sources)); err != nil {
		c.JSON(500, gin.H{"error": "failed to save answer"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"answer":          result.Answer,
		"sources":         sourceResponses(result.Sources),
		"conversation_id": turn.ConversationID, "user_message_id": turn.UserMessageID, "assistant_message_id": turn.AssistantMessageID,
	})
}

func historyMessages(messages []repository.Message) []service.HistoryMessage {
	out := make([]service.HistoryMessage, 0, len(messages))
	for _, m := range messages {
		out = append(out, service.HistoryMessage{Role: m.Role, Content: m.Content})
	}
	return out
}
func (h *ChatHandler) finishTerminal(id int64, status, content string, citations any) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return h.Conversations.FinishAssistant(ctx, id, status, content, citations)
}

func (h *ChatHandler) streamLegacy(c *gin.Context, owner, question string, emit func(string, any) error, started *bool) {
	sources, err := h.RAGService.Stream(c.Request.Context(), owner, question, func(_ []repository.ChunkSearchResult) error {
		c.Header("Content-Type", "text/event-stream")
		c.Header("Cache-Control", "no-cache")
		c.Header("X-Accel-Buffering", "no")
		c.Status(200)
		*started = true
		return emit("start", gin.H{})
	}, func(text string) error { return emit("delta", gin.H{"text": text}) })
	if err != nil {
		if errors.Is(err, context.Canceled) || c.Request.Context().Err() != nil {
			return
		}
		if !*started {
			c.JSON(500, gin.H{"error": "failed to start chat stream"})
			return
		}
		_ = emit("error", gin.H{"error": "stream interrupted"})
		return
	}
	if emit("citations", gin.H{"citations": sourceResponses(sources)}) != nil {
		return
	}
	_ = emit("done", gin.H{})
}
