package httpmw

import (
	"crypto/rand"
	"encoding/hex"
	"github.com/frankyangcl/ai-support-agent/backend/internal/auth"
	"github.com/gin-gonic/gin"
	"log/slog"
	"net/http"
	"runtime/debug"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode"
)

const requestIDKey = "request_id"

func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.GetHeader("X-Request-ID")
		if !validID(id) {
			b := make([]byte, 16)
			if _, err := rand.Read(b); err == nil {
				id = hex.EncodeToString(b)
			} else {
				id = strconv.FormatInt(time.Now().UnixNano(), 10)
			}
		}
		c.Set(requestIDKey, id)
		c.Header("X-Request-ID", id)
		c.Next()
	}
}
func RequestIDValue(c *gin.Context) string { v, _ := c.Get(requestIDKey); s, _ := v.(string); return s }
func validID(s string) bool {
	if len(s) < 1 || len(s) > 128 {
		return false
	}
	for _, r := range s {
		if !(unicode.IsLetter(r) || unicode.IsDigit(r) || strings.ContainsRune("-_.:", r)) {
			return false
		}
	}
	return true
}
func SecurityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("X-Content-Type-Options", "nosniff")
		if strings.HasPrefix(c.Request.URL.Path, "/api/") {
			c.Header("Cache-Control", "no-store")
		}
		c.Next()
	}
}
func BodyLimit(limit int64) gin.HandlerFunc {
	return func(c *gin.Context) {
		if strings.HasPrefix(c.GetHeader("Content-Type"), "application/json") {
			c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, limit)
		}
		c.Next()
	}
}
func AccessLog(logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		logger.Info("http_request", "request_id", RequestIDValue(c), "method", c.Request.Method, "path", c.Request.URL.Path, "status", c.Writer.Status(), "latency_ms", time.Since(start).Milliseconds())
	}
}
func Recovery(logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if recovered := recover(); recovered != nil {
				logger.Error("http_panic", "request_id", RequestIDValue(c), "panic", recovered, "stack", string(debug.Stack()))
				if !c.Writer.Written() {
					c.AbortWithStatusJSON(500, gin.H{"error": "internal server error"})
				} else {
					c.Abort()
				}
			}
		}()
		c.Next()
	}
}

type window struct {
	start time.Time
	count int
}
type Limiter struct {
	mu       sync.Mutex
	items    map[string]window
	requests int
	duration time.Duration
	now      func() time.Time
}

func NewLimiter(requests int, duration time.Duration) *Limiter {
	return &Limiter{items: make(map[string]window), requests: requests, duration: duration, now: time.Now}
}
func (l *Limiter) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		sub, err := auth.Subject(c.Request)
		if err != nil {
			c.AbortWithStatusJSON(401, gin.H{"error": "unauthorized"})
			return
		}
		now := l.now()
		l.mu.Lock()
		w := l.items[sub]
		if w.start.IsZero() || now.Sub(w.start) >= l.duration {
			w = window{start: now}
		}
		allowed := w.count < l.requests
		if allowed {
			w.count++
			l.items[sub] = w
		}
		retry := int((l.duration - now.Sub(w.start)).Seconds())
		l.mu.Unlock()
		if !allowed {
			if retry < 1 {
				retry = 1
			}
			c.Header("Retry-After", strconv.Itoa(retry))
			c.AbortWithStatusJSON(429, gin.H{"error": "rate limit exceeded"})
			return
		}
		c.Next()
	}
}
