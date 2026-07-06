package router

import (
"database/sql"

"github.com/frankyangcl/ai-support-agent/backend/internal/handler"
"github.com/gin-gonic/gin"
)

func Setup(db *sql.DB) *gin.Engine {
r := gin.Default()

healthHandler := handler.NewHealthHandler(db)

r.GET("/health", healthHandler.Health)
r.GET("/health/db", healthHandler.DatabaseHealth)

return r
}
