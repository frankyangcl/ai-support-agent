package router

import (
	"database/sql"

	"github.com/frankyangcl/ai-support-agent/backend/internal/handler"
	"github.com/frankyangcl/ai-support-agent/backend/internal/parser"
	"github.com/frankyangcl/ai-support-agent/backend/internal/repository"
	"github.com/frankyangcl/ai-support-agent/backend/internal/service"
	"github.com/gin-gonic/gin"
)

func Setup(db *sql.DB) *gin.Engine {
	r := gin.Default()

	healthHandler := handler.NewHealthHandler(db)
	documentRepo := repository.NewDocumentRepository(db)
	pdfParser := parser.NewPDFParser()
	documentService := service.NewDocumentService(documentRepo, pdfParser)
	documentHandler := handler.NewDocumentHandler(documentService)

	r.GET("/health", healthHandler.Health)
	r.GET("/health/db", healthHandler.DatabaseHealth)

	api := r.Group("/api")
	{
		api.POST("/documents", documentHandler.CreateDocument)
		api.GET("/documents", documentHandler.ListDocuments)
		api.POST("/documents/upload", documentHandler.UploadDocument)
		api.GET(
			"/documents/uploaded/:filename/text",
			documentHandler.ExtractDocumentText,
		)
	}

	return r
}
