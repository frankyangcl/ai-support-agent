package handler

import (
"database/sql"

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
