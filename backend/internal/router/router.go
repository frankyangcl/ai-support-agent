package router

import (
	"database/sql"

	"github.com/frankyangcl/ai-support-agent/backend/internal/chunker"
	"github.com/frankyangcl/ai-support-agent/backend/internal/handler"
	"github.com/frankyangcl/ai-support-agent/backend/internal/parser"
	"github.com/frankyangcl/ai-support-agent/backend/internal/repository"
	"github.com/frankyangcl/ai-support-agent/backend/internal/service"
	"github.com/gin-gonic/gin"

	"github.com/frankyangcl/ai-support-agent/backend/internal/config"
	"github.com/frankyangcl/ai-support-agent/backend/internal/embedding"
	"github.com/frankyangcl/ai-support-agent/backend/internal/llm"
	"github.com/gin-contrib/cors"
)

func Setup(
	db *sql.DB,
	cfg config.Config,
) *gin.Engine {
	r := gin.Default()

	r.Use(cors.New(cors.Config{
		AllowOrigins: []string{
			"http://localhost:3000",
		},
		AllowMethods: []string{
			"GET",
			"POST",
			"PUT",
			"PATCH",
			"DELETE",
			"OPTIONS",
		},
		AllowHeaders: []string{
			"Origin",
			"Content-Type",
			"Authorization",
		},
	}))

	healthHandler := handler.NewHealthHandler(db)
	documentRepo := repository.NewDocumentRepository(db)
	chunkRepo := repository.NewChunkRepository(db)
	pdfParser := parser.NewPDFParser()
	textChunker := chunker.NewTextChunker()

	embeddingClient := embedding.NewBailianClient(
		cfg.DashScopeAPIKey,
		cfg.BailianBaseURL,
	)

	embeddingService := service.NewEmbeddingService(
		chunkRepo,
		embeddingClient,
	)

	deepSeekClient := llm.NewDeepSeekClient(
		cfg.DeepSeekAPIKey,
	)

	ragService := service.NewRAGService(
		embeddingService,
		deepSeekClient,
	)

	chatHandler := handler.NewChatHandler(ragService)

	documentService := service.NewDocumentService(
		documentRepo,
		chunkRepo,
		pdfParser,
		textChunker,
		embeddingService,
	)
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
		api.GET(
			"/documents/:id",
			documentHandler.GetDocument,
		)
		api.POST("/chat", chatHandler.Chat)
	}

	return r
}
