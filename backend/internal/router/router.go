package router

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"

	"github.com/frankyangcl/ai-support-agent/backend/internal/auth"
	"github.com/frankyangcl/ai-support-agent/backend/internal/chunker"
	"github.com/frankyangcl/ai-support-agent/backend/internal/handler"
	"github.com/frankyangcl/ai-support-agent/backend/internal/httpmw"
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
) (*gin.Engine, error) {
	gin.SetMode(cfg.GinMode)
	r := gin.New()
	if err := r.SetTrustedProxies(cfg.TrustedProxies); err != nil {
		return nil, fmt.Errorf("configure trusted proxies: %w", err)
	}
	r.Use(httpmw.RequestID(), httpmw.SecurityHeaders(), httpmw.AccessLog(slog.Default()), httpmw.Recovery(slog.Default()), httpmw.BodyLimit(cfg.JSONBodyLimit))
	if len(cfg.CORSAllowedOrigins) > 0 {
		r.Use(cors.New(cors.Config{
			AllowOrigins: cfg.CORSAllowedOrigins,
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
	}

	healthHandler := handler.NewHealthHandler(db)
	documentRepo := repository.NewDocumentRepository(db)
	conversationRepo := repository.NewConversationRepository(db)
	if err := conversationRepo.RecoverProcessing(context.Background()); err != nil {
		return nil, fmt.Errorf("recover interrupted conversations: %w", err)
	}
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

	chatHandler := handler.NewChatHandler(ragService, conversationRepo, cfg.ChatHistoryLimit)
	conversationHandler := handler.NewConversationHandler(conversationRepo)

	documentService := service.NewDocumentService(
		documentRepo,
		chunkRepo,
		pdfParser,
		textChunker,
		embeddingService,
	)
	documentHandler := handler.NewDocumentHandler(documentService, cfg.MaxUploadSizeMB)
	authMiddleware, err := auth.New(cfg.Auth0Domain, cfg.Auth0Audience)
	if err != nil {
		return nil, fmt.Errorf("configure authentication: %w", err)
	}

	r.GET("/health", healthHandler.Health)
	r.GET("/health/db", healthHandler.DatabaseHealth)
	r.GET("/ready", healthHandler.DatabaseHealth)

	api := r.Group("/api", authMiddleware)
	costLimiter := httpmw.NewLimiter(cfg.RateLimitRequests, cfg.RateLimitWindow)
	{
		api.GET("/me", handler.Me)
		api.GET("/conversations", conversationHandler.List)
		api.POST("/conversations", conversationHandler.Create)
		api.GET("/conversations/:id", conversationHandler.Get)
		api.PATCH("/conversations/:id", conversationHandler.Rename)
		api.DELETE("/conversations/:id", conversationHandler.Delete)
		api.GET("/conversations/:id/messages", conversationHandler.Messages)
		api.POST("/documents", documentHandler.CreateDocument)
		api.GET("/documents", documentHandler.ListDocuments)
		api.POST("/documents/upload", costLimiter.Middleware(), documentHandler.UploadDocument)
		api.GET(
			"/documents/:id",
			documentHandler.GetDocument,
		)
		api.POST("/chat", costLimiter.Middleware(), chatHandler.Chat)
		api.POST("/chat/stream", costLimiter.Middleware(), chatHandler.Stream)
		api.POST("/documents/:id/retry", costLimiter.Middleware(), documentHandler.RetryDocument)
		api.DELETE(
			"/documents/:id",
			documentHandler.DeleteDocument,
		)
	}

	return r, nil
}
