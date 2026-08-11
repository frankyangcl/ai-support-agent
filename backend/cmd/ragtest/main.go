package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/frankyangcl/ai-support-agent/backend/internal/embedding"
	"github.com/frankyangcl/ai-support-agent/backend/internal/llm"
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

	dashscopeAPIKey := os.Getenv("DASHSCOPE_API_KEY")
	bailianBaseURL := os.Getenv("BAILIAN_BASE_URL")
	deepseekAPIKey := os.Getenv("DEEPSEEK_API_KEY")

	if dashscopeAPIKey == "" {
		log.Fatal("DASHSCOPE_API_KEY is not set")
	}

	if bailianBaseURL == "" {
		log.Fatal("BAILIAN_BASE_URL is not set")
	}

	if deepseekAPIKey == "" {
		log.Fatal("DEEPSEEK_API_KEY is not set")
	}

	db, err := sql.Open("pgx", databaseURL)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	chunkRepo := repository.NewChunkRepository(db)

	embeddingClient := embedding.NewBailianClient(
		dashscopeAPIKey,
		bailianBaseURL,
	)

	embeddingService := service.NewEmbeddingService(
		chunkRepo,
		embeddingClient,
	)

	deepseekClient := llm.NewDeepSeekClient(
		deepseekAPIKey,
	)

	ragService := service.NewRAGService(
		embeddingService,
		deepseekClient,
	)

	ctx, cancel := context.WithTimeout(
		context.Background(),
		60*time.Second,
	)
	defer cancel()

	result, err := ragService.Ask(
		ctx,
		"How long do customers have to request a refund?",
	)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("ANSWER")
	fmt.Println("------")
	fmt.Println(result.Answer)

	fmt.Println()
	fmt.Println("SOURCES")
	fmt.Println("-------")

	for _, source := range result.Sources {
		fmt.Printf(
			"document=%d chunk=%d distance=%.4f\n",
			source.DocumentID,
			source.ChunkIndex,
			source.Distance,
		)
	}
}
