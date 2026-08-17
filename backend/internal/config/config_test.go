package config

import (
	"strings"
	"testing"
	"time"
)

func TestChatHistoryLimit(t *testing.T) {
	t.Setenv("CHAT_HISTORY_LIMIT", "7")
	if got := Load().ChatHistoryLimit; got != 7 {
		t.Fatalf("got %d", got)
	}
}
func TestValidateMissingRequiredConfiguration(t *testing.T) {
	cfg := Config{GinMode: "release", RateLimitRequests: 1, RateLimitWindow: time.Second, JSONBodyLimit: 1}
	err := cfg.Validate()
	if err == nil || !strings.Contains(err.Error(), "DATABASE_URL") {
		t.Fatalf("unexpected error %v", err)
	}
}
func TestChatHistoryLimitDefault(t *testing.T) {
	t.Setenv("CHAT_HISTORY_LIMIT", "")
	if got := Load().ChatHistoryLimit; got != 20 {
		t.Fatalf("got %d", got)
	}
}
