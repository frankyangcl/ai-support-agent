package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/frankyangcl/ai-support-agent/backend/internal/llm"
)

func main() {
	apiKey := os.Getenv("DEEPSEEK_API_KEY")
	if apiKey == "" {
		log.Fatal("DEEPSEEK_API_KEY is not set")
	}

	client := llm.NewDeepSeekClient(apiKey)

	ctx, cancel := context.WithTimeout(
		context.Background(),
		30*time.Second,
	)
	defer cancel()

	answer, err := client.Chat(
		ctx,
		"You are a concise assistant.",
		"Reply with exactly: API works",
	)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(answer)
}
