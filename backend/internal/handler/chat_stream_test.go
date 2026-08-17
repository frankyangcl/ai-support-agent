package handler_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/auth0/go-jwt-middleware/v3/core"
	"github.com/auth0/go-jwt-middleware/v3/validator"
	"github.com/frankyangcl/ai-support-agent/backend/internal/auth"
	"github.com/frankyangcl/ai-support-agent/backend/internal/handler"
	"github.com/frankyangcl/ai-support-agent/backend/internal/repository"
	"github.com/frankyangcl/ai-support-agent/backend/internal/service"
	"github.com/gin-gonic/gin"
)

type authChecker struct{}

func (authChecker) CheckJWT(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer test-token" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		claims := &validator.ValidatedClaims{}
		claims.RegisteredClaims.Subject = "auth0|test-user"
		next.ServeHTTP(w, r.WithContext(core.SetClaims(r.Context(), claims)))
	})
}

type fakeChatService struct {
	ask    func(context.Context, string) (*service.RAGResult, error)
	stream func(context.Context, string, func([]repository.ChunkSearchResult) error, func(string) error) ([]repository.ChunkSearchResult, error)
}

func (f *fakeChatService) Ask(ctx context.Context, _, question string) (*service.RAGResult, error) {
	return f.ask(ctx, question)
}

func (f *fakeChatService) Stream(ctx context.Context, _, question string, start func([]repository.ChunkSearchResult) error, delta func(string) error) ([]repository.ChunkSearchResult, error) {
	return f.stream(ctx, question, start, delta)
}
func (f *fakeChatService) AskWithHistory(ctx context.Context, owner, question string, _ []service.HistoryMessage) (*service.RAGResult, error) {
	return f.Ask(ctx, owner, question)
}
func (f *fakeChatService) StreamWithHistory(ctx context.Context, owner, question string, _ []service.HistoryMessage, start func([]repository.ChunkSearchResult) error, delta func(string) error) ([]repository.ChunkSearchResult, error) {
	return f.Stream(ctx, owner, question, start, delta)
}

func streamRouter(chatService handler.ChatService) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := handler.NewChatHandler(chatService)
	api := r.Group("/api", auth.GinMiddleware(authChecker{}))
	api.POST("/chat", h.Chat)
	api.POST("/chat/stream", h.Stream)
	return r
}

func post(r http.Handler, path, body, token string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodPost, path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", token)
	}
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func successfulService() *fakeChatService {
	sources := []repository.ChunkSearchResult{{DocumentID: 7, Filename: "policy.pdf", ChunkIndex: 2, Content: "Refund policy", Distance: 0.1}}
	return &fakeChatService{
		ask: func(context.Context, string) (*service.RAGResult, error) {
			return &service.RAGResult{Answer: "answer", Sources: sources}, nil
		},
		stream: func(_ context.Context, _ string, start func([]repository.ChunkSearchResult) error, delta func(string) error) ([]repository.ChunkSearchResult, error) {
			if err := start(sources); err != nil {
				return nil, err
			}
			if err := delta("first "); err != nil {
				return nil, err
			}
			if err := delta("second"); err != nil {
				return nil, err
			}
			return sources, nil
		},
	}
}

func TestChatStreamRequiresAuthentication(t *testing.T) {
	w := post(streamRouter(successfulService()), "/api/chat/stream", `{"question":"hello"}`, "")
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("got %d, want 401", w.Code)
	}
}

func TestChatStreamRejectsInvalidRequest(t *testing.T) {
	w := post(streamRouter(successfulService()), "/api/chat/stream", `{}`, "Bearer test-token")
	if w.Code != http.StatusBadRequest {
		t.Fatalf("got %d, want 400", w.Code)
	}
}

func TestChatStreamEventsAreOrdered(t *testing.T) {
	w := post(streamRouter(successfulService()), "/api/chat/stream", `{"question":"hello"}`, "Bearer test-token")
	body := w.Body.String()
	if w.Header().Get("Content-Type") != "text/event-stream" {
		t.Fatalf("unexpected content type %q", w.Header().Get("Content-Type"))
	}
	events := []string{"event: start", `data: {"text":"first "}`, `data: {"text":"second"}`, "event: citations", `"filename":"policy.pdf"`, "event: done"}
	position := -1
	for _, event := range events {
		next := strings.Index(body[position+1:], event)
		if next < 0 {
			t.Fatalf("missing or out-of-order %q in %q", event, body)
		}
		position += next + 1
	}
}

func TestChatStreamProviderErrorEmitsErrorEvent(t *testing.T) {
	svc := successfulService()
	svc.stream = func(_ context.Context, _ string, start func([]repository.ChunkSearchResult) error, _ func(string) error) ([]repository.ChunkSearchResult, error) {
		if err := start(nil); err != nil {
			return nil, err
		}
		return nil, errors.New("provider secret detail")
	}
	w := post(streamRouter(svc), "/api/chat/stream", `{"question":"hello"}`, "Bearer test-token")
	if !strings.Contains(w.Body.String(), "event: error") || strings.Contains(w.Body.String(), "secret detail") {
		t.Fatalf("unsafe error response %q", w.Body.String())
	}
}

func TestChatStreamCancellationReachesService(t *testing.T) {
	seen := make(chan struct{})
	svc := successfulService()
	svc.stream = func(ctx context.Context, _ string, _ func([]repository.ChunkSearchResult) error, _ func(string) error) ([]repository.ChunkSearchResult, error) {
		<-ctx.Done()
		close(seen)
		return nil, ctx.Err()
	}
	r := streamRouter(svc)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	req := httptest.NewRequest(http.MethodPost, "/api/chat/stream", strings.NewReader(`{"question":"hello"}`)).WithContext(ctx)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer test-token")
	r.ServeHTTP(httptest.NewRecorder(), req)
	select {
	case <-seen:
	default:
		t.Fatal("cancellation did not reach service")
	}
}

func TestExistingChatRemainsCompatible(t *testing.T) {
	w := post(streamRouter(successfulService()), "/api/chat", `{"question":"hello"}`, "Bearer test-token")
	if w.Code != http.StatusOK || !strings.Contains(w.Body.String(), `"answer":"answer"`) || !strings.Contains(w.Body.String(), `"sources"`) {
		t.Fatalf("unexpected response %d %q", w.Code, w.Body.String())
	}
}
