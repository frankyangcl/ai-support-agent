package main

import (
	"context"
	"database/sql"
	"fmt"
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

	if apiKey == "" || baseURL == "" {
		log.Fatal("embedding environment variables are missing")
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
		30*time.Second,
	)
	defer cancel()

	results, err := embeddingService.Search(
		ctx,
		"How long do customers have to request a refund?",
		3,
	)
	if err != nil {
		log.Fatal(err)
	}

	for _, result := range results {
		fmt.Printf(
			"chunk=%d distance=%.4f\n%s\n\n",
			result.ChunkIndex,
			result.Distance,
			result.Content,
		)
	}
}
