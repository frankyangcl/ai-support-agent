package auth_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/auth0/go-jwt-middleware/v3/core"
	"github.com/auth0/go-jwt-middleware/v3/validator"
	"github.com/frankyangcl/ai-support-agent/backend/internal/auth"
	"github.com/frankyangcl/ai-support-agent/backend/internal/handler"
	"github.com/gin-gonic/gin"
)

type fakeChecker struct{}

func (fakeChecker) CheckJWT(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Header.Get("Authorization") {
		case "Bearer valid-token":
			claims := &validator.ValidatedClaims{}
			claims.RegisteredClaims.Subject = "auth0|test-user"
			ctx := core.SetClaims(r.Context(), claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		default:
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			_, _ = w.Write([]byte(`{"error":"unauthorized"}`))
		}
	})
}

func newTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/health", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"status": "ok"}) })
	api := r.Group("/api", auth.GinMiddleware(fakeChecker{}))
	api.GET("/protected", func(c *gin.Context) { c.Status(http.StatusNoContent) })
	api.GET("/me", handler.Me)
	return r
}

func request(t *testing.T, r http.Handler, path, authorization string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, path, nil)
	if authorization != "" {
		req.Header.Set("Authorization", authorization)
	}
	resp := httptest.NewRecorder()
	r.ServeHTTP(resp, req)
	return resp
}

func TestProtectedEndpointRequiresToken(t *testing.T) {
	resp := request(t, newTestRouter(), "/api/protected", "")
	if resp.Code != http.StatusUnauthorized {
		t.Fatalf("got status %d, want 401", resp.Code)
	}
}

func TestProtectedEndpointRejectsMalformedBearerToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	middleware, err := auth.New("tenant.auth0.com", "test-audience")
	if err != nil {
		t.Fatalf("create middleware: %v", err)
	}
	r := gin.New()
	r.GET("/protected", middleware, func(c *gin.Context) { c.Status(http.StatusNoContent) })
	resp := request(t, r, "/protected", "Bearer")
	if resp.Code != http.StatusUnauthorized {
		t.Fatalf("got status %d, want 401", resp.Code)
	}
	if resp.Body.String() != "{\"error\":\"unauthorized\"}\n" {
		t.Fatalf("unexpected unauthorized response %q", resp.Body.String())
	}
}

func TestProtectedEndpointRejectsInvalidToken(t *testing.T) {
	resp := request(t, newTestRouter(), "/api/protected", "Bearer invalid-token")
	if resp.Code != http.StatusUnauthorized {
		t.Fatalf("got status %d, want 401", resp.Code)
	}
}

func TestValidTokenExecutesProtectedHandler(t *testing.T) {
	resp := request(t, newTestRouter(), "/api/protected", "Bearer valid-token")
	if resp.Code != http.StatusNoContent {
		t.Fatalf("got status %d, want 204", resp.Code)
	}
}

func TestMeReturnsValidatedSubject(t *testing.T) {
	resp := request(t, newTestRouter(), "/api/me", "Bearer valid-token")
	if resp.Code != http.StatusOK || resp.Body.String() != `{"sub":"auth0|test-user"}` {
		t.Fatalf("got status %d body %q", resp.Code, resp.Body.String())
	}
}

func TestHealthRemainsPublic(t *testing.T) {
	resp := request(t, newTestRouter(), "/health", "")
	if resp.Code != http.StatusOK {
		t.Fatalf("got status %d, want 200", resp.Code)
	}
}

func TestMeRejectsMissingValidatedClaims(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/me", handler.Me)
	req := httptest.NewRequest(http.MethodGet, "/me", nil).WithContext(context.Background())
	resp := httptest.NewRecorder()
	r.ServeHTTP(resp, req)
	if resp.Code != http.StatusUnauthorized {
		t.Fatalf("got status %d, want 401", resp.Code)
	}
}

func TestNewRequiresConfiguration(t *testing.T) {
	if _, err := auth.New("", "audience"); err == nil {
		t.Fatal("expected missing domain error")
	}
	if _, err := auth.New("tenant.auth0.com", ""); err == nil {
		t.Fatal("expected missing audience error")
	}
}
