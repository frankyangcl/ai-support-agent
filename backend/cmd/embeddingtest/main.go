package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/frankyangcl/ai-support-agent/backend/internal/embedding"
)

func main() {
	apiKey := os.Getenv("DASHSCOPE_API_KEY")
	baseURL := os.Getenv("BAILIAN_BASE_URL")

	if apiKey == "" {
		log.Fatal("DASHSCOPE_API_KEY is not set")
	}

	if baseURL == "" {
		log.Fatal("BAILIAN_BASE_URL is not set")
	}

	client := embedding.NewBailianClient(apiKey, baseURL)

	ctx, cancel := context.WithTimeout(
		context.Background(),
		30*time.Second,
	)
	defer cancel()

	vector, err := client.Embed(
		ctx,
		"Customers can request a refund within 30 days.",
	)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("dimensions: %d\n", len(vector))
	fmt.Printf("first 3 values: %v\n", vector[:3])
}
