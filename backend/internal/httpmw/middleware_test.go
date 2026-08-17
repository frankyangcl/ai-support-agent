package httpmw

import (
	"bytes"
	"github.com/auth0/go-jwt-middleware/v3/core"
	"github.com/auth0/go-jwt-middleware/v3/validator"
	"github.com/gin-gonic/gin"
	"io"
	"log/slog"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func claims(sub string) gin.HandlerFunc {
	return func(c *gin.Context) {
		v := &validator.ValidatedClaims{}
		v.RegisteredClaims.Subject = sub
		c.Request = c.Request.WithContext(core.SetClaims(c.Request.Context(), v))
		c.Next()
	}
}
func TestRequestIDGeneratedAndReturned(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(RequestID())
	r.GET("/", func(c *gin.Context) { c.Status(204) })
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest("GET", "/", nil))
	if w.Header().Get("X-Request-ID") == "" {
		t.Fatal("missing request id")
	}
}
func TestRequestIDAcceptsSafeIncomingValue(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(RequestID())
	r.GET("/", func(c *gin.Context) { c.Status(204) })
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("X-Request-ID", "safe-id_1")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if got := w.Header().Get("X-Request-ID"); got != "safe-id_1" {
		t.Fatalf("got %q", got)
	}
}
func TestSecurityHeaders(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(SecurityHeaders())
	r.GET("/api/x", func(c *gin.Context) { c.JSON(200, gin.H{}) })
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest("GET", "/api/x", nil))
	if w.Header().Get("X-Content-Type-Options") != "nosniff" || w.Header().Get("Cache-Control") != "no-store" {
		t.Fatal("security headers missing")
	}
}
func TestRecoveryReturnsSafeError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	var logs bytes.Buffer
	r := gin.New()
	r.Use(RequestID(), Recovery(slog.New(slog.NewJSONHandler(&logs, nil))))
	r.GET("/", func(*gin.Context) { panic("secret panic") })
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest("GET", "/", nil))
	if w.Code != 500 || !strings.Contains(w.Body.String(), "internal server error") {
		t.Fatalf("%d %q", w.Code, w.Body.String())
	}
}
func TestLimiterKeysByUserAndReturns429(t *testing.T) {
	gin.SetMode(gin.TestMode)
	l := NewLimiter(1, time.Minute)
	r := gin.New()
	r.Use(claims("user-a"), l.Middleware())
	r.GET("/", func(c *gin.Context) { c.Status(204) })
	for i, want := range []int{204, 429} {
		w := httptest.NewRecorder()
		r.ServeHTTP(w, httptest.NewRequest("GET", "/", nil))
		if w.Code != want {
			t.Fatalf("request %d got %d", i, w.Code)
		}
	}
	r2 := gin.New()
	r2.Use(claims("user-b"), l.Middleware())
	r2.GET("/", func(c *gin.Context) { c.Status(204) })
	w := httptest.NewRecorder()
	r2.ServeHTTP(w, httptest.NewRequest("GET", "/", nil))
	if w.Code != 204 {
		t.Fatalf("other user got %d", w.Code)
	}
}
func TestBodyLimit(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(BodyLimit(8))
	r.POST("/", func(c *gin.Context) {
		_, err := io.ReadAll(c.Request.Body)
		if err != nil {
			c.JSON(413, gin.H{"error": "too large"})
			return
		}
		c.Status(204)
	})
	req := httptest.NewRequest("POST", "/", strings.NewReader("0123456789"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != 413 {
		t.Fatalf("got %d", w.Code)
	}
}
func TestAccessLogPrivacy(t *testing.T) {
	gin.SetMode(gin.TestMode)
	var logs bytes.Buffer
	r := gin.New()
	r.Use(RequestID(), AccessLog(slog.New(slog.NewJSONHandler(&logs, nil))))
	r.POST("/api/chat", func(c *gin.Context) { c.Status(204) })
	req := httptest.NewRequest("POST", "/api/chat", strings.NewReader(`{"question":"private question"}`))
	req.Header.Set("Authorization", "Bearer secret-token")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	for _, secret := range []string{"secret-token", "private question", "Authorization"} {
		if strings.Contains(logs.String(), secret) {
			t.Fatalf("log leaked %q", secret)
		}
	}
}
