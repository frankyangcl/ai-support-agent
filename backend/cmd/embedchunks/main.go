package main

import (
	"context"
	"database/sql"
	"log"
	"os"
	"time"

	"github.com/frankyangcl/ai-support-agent/backend/internal/embedding"
	"github.com/frankyangcl/ai-support-agent/backend/internal/repository"
	"github.com/frankyangcl/ai-support-agent/backend/internal/service"
	_ "github.com/jackc/pgx/v5/stdlib"
)

func main() {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		databaseURL =
			"postgres://postgres:postgres@localhost:5433/ai_support_agent?sslmode=disable"
	}

	apiKey := os.Getenv("DASHSCOPE_API_KEY")
	baseURL := os.Getenv("BAILIAN_BASE_URL")

	if apiKey == "" {
		log.Fatal("DASHSCOPE_API_KEY is not set")
	}

	if baseURL == "" {
		log.Fatal("BAILIAN_BASE_URL is not set")
	}

	db, err := sql.Open("pgx", databaseURL)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	chunkRepo := repository.NewChunkRepository(db)
	client := embedding.NewBailianClient(apiKey, baseURL)

	embeddingService := service.NewEmbeddingService(
		chunkRepo,
		client,
	)

	ctx, cancel := context.WithTimeout(
		context.Background(),
		2*time.Minute,
	)
	defer cancel()

	processed, err := embeddingService.ProcessPendingChunks(ctx)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("processed %d chunks\n", processed)
}
