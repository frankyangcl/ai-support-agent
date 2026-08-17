package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/frankyangcl/ai-support-agent/backend/internal/repository"
)

type RAGResult struct {
	Answer  string
	Sources []repository.ChunkSearchResult
}
type HistoryMessage struct {
	Role    string
	Content string
}

type RAGService struct {
	EmbeddingService ContextSearcher
	LLMClient        ChatClient
}

type ContextSearcher interface {
	Search(context.Context, string, string, int) ([]repository.ChunkSearchResult, error)
}

type ChatClient interface {
	Chat(context.Context, string, string) (string, error)
	StreamChat(context.Context, string, string, func(string) error) error
}

func NewRAGService(
	embeddingService ContextSearcher,
	llmClient ChatClient,
) *RAGService {
	return &RAGService{
		EmbeddingService: embeddingService,
		LLMClient:        llmClient,
	}
}

func (s *RAGService) Ask(
	ctx context.Context,
	ownerSub string,
	question string,
) (*RAGResult, error) {
	return s.AskWithHistory(ctx, ownerSub, question, nil)
}
func (s *RAGService) AskWithHistory(ctx context.Context, ownerSub, question string, history []HistoryMessage) (*RAGResult, error) {
	question = strings.TrimSpace(question)
	if question == "" {
		return nil, fmt.Errorf("question must not be empty")
	}

	results, systemPrompt, userPrompt, err := s.prepare(ctx, ownerSub, question, history)
	if err != nil {
		return nil, err
	}

	if len(results) == 0 {
		return &RAGResult{
			Answer:  "I couldn't find enough relevant information in the knowledge base to answer that question.",
			Sources: nil,
		}, nil
	}

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

func (s *RAGService) Stream(
	ctx context.Context,
	ownerSub string,
	question string,
	onStart func([]repository.ChunkSearchResult) error,
	onDelta func(string) error,
) ([]repository.ChunkSearchResult, error) {
	return s.StreamWithHistory(ctx, ownerSub, question, nil, onStart, onDelta)
}
func (s *RAGService) StreamWithHistory(ctx context.Context, ownerSub, question string, history []HistoryMessage, onStart func([]repository.ChunkSearchResult) error, onDelta func(string) error) ([]repository.ChunkSearchResult, error) {
	results, systemPrompt, userPrompt, err := s.prepare(ctx, ownerSub, question, history)
	if err != nil {
		return nil, err
	}
	if err := onStart(results); err != nil {
		return results, err
	}
	if len(results) == 0 {
		return results, onDelta("I couldn't find enough relevant information in the knowledge base to answer that question.")
	}
	if err := s.LLMClient.StreamChat(ctx, systemPrompt, userPrompt, onDelta); err != nil {
		return results, fmt.Errorf("generate streaming answer: %w", err)
	}
	return results, nil
}

func (s *RAGService) prepare(ctx context.Context, ownerSub, question string, history []HistoryMessage) ([]repository.ChunkSearchResult, string, string, error) {
	question = strings.TrimSpace(question)
	if question == "" {
		return nil, "", "", fmt.Errorf("question must not be empty")
	}
	results, err := s.EmbeddingService.Search(ctx, ownerSub, question, 8)
	if err != nil {
		return nil, "", "", fmt.Errorf("retrieve context: %w", err)
	}
	if len(results) > 3 {
		results = results[:3]
	}
	if len(results) == 0 {
		return results, "", "", nil
	}
	var contextBuilder strings.Builder
	for i, result := range results {
		fmt.Fprintf(&contextBuilder, "[Source %d - Document %d - Chunk %d]\n%s\n\n", i+1, result.DocumentID, result.ChunkIndex, result.Content)
	}
	systemPrompt := `You are a customer support assistant.

Answer the user's question using only the provided knowledge base context.

Rules:
- Do not invent information.
- If the answer cannot be found in the context, say that the knowledge base does not contain enough information.
- Be concise and direct.
- Cite the relevant source using [Source N].`
	var historyBuilder strings.Builder
	for _, message := range history {
		fmt.Fprintf(&historyBuilder, "%s: %s\n", strings.Title(message.Role), message.Content)
	}
	userPrompt := fmt.Sprintf("Knowledge base context:\n\n%s\nConversation history:\n%s\nCurrent user question:\n%s", contextBuilder.String(), historyBuilder.String(), question)
	return results, systemPrompt, userPrompt, nil
}
