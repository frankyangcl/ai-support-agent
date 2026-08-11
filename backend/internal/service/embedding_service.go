package service

import (
	"context"
	"fmt"

	"github.com/frankyangcl/ai-support-agent/backend/internal/embedding"
	"github.com/frankyangcl/ai-support-agent/backend/internal/repository"
)

type EmbeddingService struct {
	ChunkRepo *repository.ChunkRepository
	Client    *embedding.BailianClient
}

func NewEmbeddingService(
	chunkRepo *repository.ChunkRepository,
	client *embedding.BailianClient,
) *EmbeddingService {
	return &EmbeddingService{
		ChunkRepo: chunkRepo,
		Client:    client,
	}
}

func (s *EmbeddingService) ProcessPendingChunks(
	ctx context.Context,
) (int, error) {
	chunks, err := s.ChunkRepo.ListWithoutEmbedding()
	if err != nil {
		return 0, fmt.Errorf("list chunks without embedding: %w", err)
	}

	processed := 0

	for _, chunk := range chunks {
		vector, err := s.Client.Embed(ctx, chunk.Content)
		if err != nil {
			return processed, fmt.Errorf(
				"embed chunk %d: %w",
				chunk.ID,
				err,
			)
		}

		if err := s.ChunkRepo.UpdateEmbedding(
			chunk.ID,
			vector,
		); err != nil {
			return processed, fmt.Errorf(
				"update embedding for chunk %d: %w",
				chunk.ID,
				err,
			)
		}

		processed++
	}

	return processed, nil
}

func (s *EmbeddingService) Search(
	ctx context.Context,
	query string,
	limit int,
) ([]repository.ChunkSearchResult, error) {
	vector, err := s.Client.Embed(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("embed search query: %w", err)
	}

	results, err := s.ChunkRepo.SearchSimilar(vector, limit)
	if err != nil {
		return nil, fmt.Errorf("search similar chunks: %w", err)
	}

	return results, nil
}
