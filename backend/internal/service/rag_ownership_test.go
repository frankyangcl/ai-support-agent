package service

import (
	"context"
	"github.com/frankyangcl/ai-support-agent/backend/internal/repository"
	"testing"
)

type ownerSearcher struct{ seen []string }

func (s *ownerSearcher) Search(_ context.Context, owner, _ string, _ int) ([]repository.ChunkSearchResult, error) {
	s.seen = append(s.seen, owner)
	return nil, nil
}

type noCallChat struct{}

func (noCallChat) Chat(context.Context, string, string) (string, error)                 { return "", nil }
func (noCallChat) StreamChat(context.Context, string, string, func(string) error) error { return nil }
func TestRAGAskFiltersRetrievalByOwner(t *testing.T) {
	search := &ownerSearcher{}
	svc := NewRAGService(search, noCallChat{})
	if _, err := svc.Ask(context.Background(), "auth0|a", "question"); err != nil {
		t.Fatal(err)
	}
	if len(search.seen) != 1 || search.seen[0] != "auth0|a" {
		t.Fatalf("owners %#v", search.seen)
	}
}
func TestRAGStreamFiltersRetrievalByOwner(t *testing.T) {
	search := &ownerSearcher{}
	svc := NewRAGService(search, noCallChat{})
	_, err := svc.Stream(context.Background(), "auth0|b", "question", func([]repository.ChunkSearchResult) error { return nil }, func(string) error { return nil })
	if err != nil {
		t.Fatal(err)
	}
	if len(search.seen) != 1 || search.seen[0] != "auth0|b" {
		t.Fatalf("owners %#v", search.seen)
	}
}
