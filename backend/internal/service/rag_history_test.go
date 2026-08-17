package service

import (
	"context"
	"github.com/frankyangcl/ai-support-agent/backend/internal/repository"
	"strings"
	"testing"
)

type historySearcher struct{}

func (historySearcher) Search(context.Context, string, string, int) ([]repository.ChunkSearchResult, error) {
	return []repository.ChunkSearchResult{{DocumentID: 1, ChunkIndex: 0, Content: "policy", Filename: "policy.pdf"}}, nil
}

type promptCapture struct{ user string }

func (p *promptCapture) Chat(_ context.Context, _, user string) (string, error) {
	p.user = user
	return "answer", nil
}
func (p *promptCapture) StreamChat(_ context.Context, _, user string, delta func(string) error) error {
	p.user = user
	return delta("answer")
}
func TestRAGIncludesConversationHistory(t *testing.T) {
	llm := &promptCapture{}
	svc := NewRAGService(historySearcher{}, llm)
	_, err := svc.AskWithHistory(context.Background(), "auth0|a", "What about damage?", []HistoryMessage{{Role: "user", Content: "What is the refund period?"}, {Role: "assistant", Content: "30 days."}})
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"User: What is the refund period?", "Assistant: 30 days.", "Current user question:\nWhat about damage?"} {
		if !strings.Contains(llm.user, want) {
			t.Fatalf("prompt missing %q: %q", want, llm.user)
		}
	}
}
func TestRAGStreamIncludesConversationHistory(t *testing.T) {
	llm := &promptCapture{}
	svc := NewRAGService(historySearcher{}, llm)
	_, err := svc.StreamWithHistory(context.Background(), "auth0|a", "followup", []HistoryMessage{{Role: "assistant", Content: "prior"}}, func([]repository.ChunkSearchResult) error { return nil }, func(string) error { return nil })
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(llm.user, "Assistant: prior") {
		t.Fatalf("history missing: %q", llm.user)
	}
}
