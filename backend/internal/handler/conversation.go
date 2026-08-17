package handler

import (
	"database/sql"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/frankyangcl/ai-support-agent/backend/internal/repository"
	"github.com/gin-gonic/gin"
)

type ConversationHandler struct {
	Repo *repository.ConversationRepository
}

func NewConversationHandler(repo *repository.ConversationRepository) *ConversationHandler {
	return &ConversationHandler{Repo: repo}
}
func conversationID(c *gin.Context) (int64, bool) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		c.JSON(400, gin.H{"error": "invalid conversation id"})
		return 0, false
	}
	return id, true
}
func (h *ConversationHandler) Create(c *gin.Context) {
	owner, ok := subject(c)
	if !ok {
		return
	}
	var req struct {
		Title string `json:"title"`
	}
	_ = c.ShouldBindJSON(&req)
	title := strings.TrimSpace(req.Title)
	if title == "" {
		title = "New conversation"
	}
	if len([]rune(title)) > 120 {
		c.JSON(400, gin.H{"error": "title is too long"})
		return
	}
	conversation, err := h.Repo.Create(c.Request.Context(), owner, title)
	if err != nil {
		c.JSON(500, gin.H{"error": "failed to create conversation"})
		return
	}
	c.JSON(201, conversation)
}
func (h *ConversationHandler) List(c *gin.Context) {
	owner, ok := subject(c)
	if !ok {
		return
	}
	items, err := h.Repo.List(c.Request.Context(), owner)
	if err != nil {
		c.JSON(500, gin.H{"error": "failed to list conversations"})
		return
	}
	c.JSON(200, gin.H{"conversations": items})
}
func (h *ConversationHandler) Get(c *gin.Context) {
	owner, ok := subject(c)
	if !ok {
		return
	}
	id, ok := conversationID(c)
	if !ok {
		return
	}
	item, err := h.Repo.Get(c.Request.Context(), owner, id)
	if errors.Is(err, sql.ErrNoRows) {
		c.JSON(404, gin.H{"error": "conversation not found"})
		return
	}
	if err != nil {
		c.JSON(500, gin.H{"error": "failed to get conversation"})
		return
	}
	c.JSON(200, item)
}
func (h *ConversationHandler) Rename(c *gin.Context) {
	owner, ok := subject(c)
	if !ok {
		return
	}
	id, ok := conversationID(c)
	if !ok {
		return
	}
	var req struct {
		Title string `json:"title" binding:"required"`
	}
	if c.ShouldBindJSON(&req) != nil {
		c.JSON(400, gin.H{"error": "title is required"})
		return
	}
	title := strings.TrimSpace(req.Title)
	if title == "" || len([]rune(title)) > 120 {
		c.JSON(400, gin.H{"error": "invalid title"})
		return
	}
	item, err := h.Repo.Rename(c.Request.Context(), owner, id, title)
	if errors.Is(err, sql.ErrNoRows) {
		c.JSON(404, gin.H{"error": "conversation not found"})
		return
	}
	if err != nil {
		c.JSON(500, gin.H{"error": "failed to rename conversation"})
		return
	}
	c.JSON(200, item)
}
func (h *ConversationHandler) Delete(c *gin.Context) {
	owner, ok := subject(c)
	if !ok {
		return
	}
	id, ok := conversationID(c)
	if !ok {
		return
	}
	err := h.Repo.Delete(c.Request.Context(), owner, id)
	if errors.Is(err, sql.ErrNoRows) {
		c.JSON(404, gin.H{"error": "conversation not found"})
		return
	}
	if err != nil {
		c.JSON(500, gin.H{"error": "failed to delete conversation"})
		return
	}
	c.Status(http.StatusNoContent)
}
func (h *ConversationHandler) Messages(c *gin.Context) {
	owner, ok := subject(c)
	if !ok {
		return
	}
	id, ok := conversationID(c)
	if !ok {
		return
	}
	if _, err := h.Repo.Get(c.Request.Context(), owner, id); errors.Is(err, sql.ErrNoRows) {
		c.JSON(404, gin.H{"error": "conversation not found"})
		return
	} else if err != nil {
		c.JSON(500, gin.H{"error": "failed to get messages"})
		return
	}
	items, err := h.Repo.Messages(c.Request.Context(), owner, id, 1000)
	if err != nil {
		c.JSON(500, gin.H{"error": "failed to get messages"})
		return
	}
	c.JSON(200, gin.H{"messages": items})
}
