package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/frankyangcl/ai-support-agent/backend/internal/llm"
	"github.com/frankyangcl/ai-support-agent/backend/internal/repository"
)

type RAGResult struct {
	Answer  string
	Sources []repository.ChunkSearchResult
}

type RAGService struct {
	EmbeddingService *EmbeddingService
	LLMClient        *llm.DeepSeekClient
}

func NewRAGService(
	embeddingService *EmbeddingService,
	llmClient *llm.DeepSeekClient,
) *RAGService {
	return &RAGService{
		EmbeddingService: embeddingService,
		LLMClient:        llmClient,
	}
}

func (s *RAGService) Ask(
	ctx context.Context,
	question string,
) (*RAGResult, error) {
	question = strings.TrimSpace(question)
	if question == "" {
		return nil, fmt.Errorf("question must not be empty")
	}

	results, err := s.EmbeddingService.Search(
		ctx,
		question,
		8,
	)
	if len(results) > 3 {
		results = results[:3]
	}
	if err != nil {
		return nil, fmt.Errorf("retrieve context: %w", err)
	}

	if len(results) == 0 {
		return &RAGResult{
			Answer:  "I couldn't find enough relevant information in the knowledge base to answer that question.",
			Sources: nil,
		}, nil
	}

	var contextBuilder strings.Builder

	for i, result := range results {
		fmt.Fprintf(
			&contextBuilder,
			"[Source %d - Document %d - Chunk %d]\n%s\n\n",
			i+1,
			result.DocumentID,
			result.ChunkIndex,
			result.Content,
		)
	}

	systemPrompt := `You are a customer support assistant.

Answer the user's question using only the provided knowledge base context.

Rules:
- Do not invent information.
- If the answer cannot be found in the context, say that the knowledge base does not contain enough information.
- Be concise and direct.
- Cite the relevant source using [Source N].`

	userPrompt := fmt.Sprintf(
		"Knowledge base context:\n\n%s\nUser question:\n%s",
		contextBuilder.String(),
		question,
	)

	answer, err := s.LLMClient.Chat(
		ctx,
		systemPrompt,
		userPrompt,
	)
	if err != nil {
		return nil, fmt.Errorf("generate answer: %w", err)
	}

	return &RAGResult{
		Answer:  answer,
		Sources: results,
	}, nil
}
