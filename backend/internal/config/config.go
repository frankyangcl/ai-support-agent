package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	DatabaseURL, ServerAddr, DeepSeekAPIKey, DashScopeAPIKey, BailianBaseURL string
	Auth0Domain, Auth0Audience, GinMode                                      string
	MaxUploadSizeMB                                                          int64
	ChatHistoryLimit                                                         int
	CORSAllowedOrigins, TrustedProxies                                       []string
	RateLimitRequests                                                        int
	RateLimitWindow                                                          time.Duration
	JSONBodyLimit                                                            int64
	DBMaxOpen, DBMaxIdle                                                     int
	DBConnMaxLifetime                                                        time.Duration
	ShutdownTimeout, ReadHeaderTimeout, IdleTimeout                          time.Duration
}

func Load() Config {
	port := env("PORT", "8080")
	addr := os.Getenv("SERVER_ADDR")
	if addr == "" {
		addr = ":" + port
	}
	return Config{DatabaseURL: os.Getenv("DATABASE_URL"), ServerAddr: addr, DeepSeekAPIKey: os.Getenv("DEEPSEEK_API_KEY"), DashScopeAPIKey: os.Getenv("DASHSCOPE_API_KEY"), BailianBaseURL: os.Getenv("BAILIAN_BASE_URL"), Auth0Domain: os.Getenv("AUTH0_DOMAIN"), Auth0Audience: os.Getenv("AUTH0_AUDIENCE"), GinMode: env("GIN_MODE", "debug"), MaxUploadSizeMB: int64(envInt("MAX_UPLOAD_SIZE_MB", 10)), ChatHistoryLimit: envInt("CHAT_HISTORY_LIMIT", 20), CORSAllowedOrigins: csv("CORS_ALLOWED_ORIGINS"), TrustedProxies: csv("TRUSTED_PROXIES"), RateLimitRequests: envInt("RATE_LIMIT_REQUESTS", 30), RateLimitWindow: envDuration("RATE_LIMIT_WINDOW", time.Minute), JSONBodyLimit: int64(envInt("JSON_BODY_LIMIT_KB", 256)) << 10, DBMaxOpen: envInt("DB_MAX_OPEN_CONNS", 20), DBMaxIdle: envInt("DB_MAX_IDLE_CONNS", 5), DBConnMaxLifetime: envDuration("DB_CONN_MAX_LIFETIME", 30*time.Minute), ShutdownTimeout: envDuration("SHUTDOWN_TIMEOUT", 20*time.Second), ReadHeaderTimeout: envDuration("READ_HEADER_TIMEOUT", 10*time.Second), IdleTimeout: envDuration("IDLE_TIMEOUT", 120*time.Second)}
}
func (c Config) Validate() error {
	var missing []string
	for name, value := range map[string]string{"DATABASE_URL": c.DatabaseURL, "DEEPSEEK_API_KEY": c.DeepSeekAPIKey, "DASHSCOPE_API_KEY": c.DashScopeAPIKey, "BAILIAN_BASE_URL": c.BailianBaseURL, "AUTH0_DOMAIN": c.Auth0Domain, "AUTH0_AUDIENCE": c.Auth0Audience} {
		if strings.TrimSpace(value) == "" {
			missing = append(missing, name)
		}
	}
	if len(missing) > 0 {
		return fmt.Errorf("missing required configuration: %s", strings.Join(missing, ", "))
	}
	if c.GinMode != "debug" && c.GinMode != "release" && c.GinMode != "test" {
		return errors.New("GIN_MODE must be debug, release, or test")
	}
	if c.RateLimitRequests < 1 || c.RateLimitWindow <= 0 || c.JSONBodyLimit < 1 {
		return errors.New("rate limit and body limit values must be positive")
	}
	return nil
}
func env(k, d string) string {
	if v := strings.TrimSpace(os.Getenv(k)); v != "" {
		return v
	}
	return d
}
func envInt(k string, d int) int {
	v, err := strconv.Atoi(os.Getenv(k))
	if err == nil && v > 0 {
		return v
	}
	return d
}
func envDuration(k string, d time.Duration) time.Duration {
	v, err := time.ParseDuration(os.Getenv(k))
	if err == nil && v > 0 {
		return v
	}
	return d
}
func csv(k string) []string {
	raw := strings.TrimSpace(os.Getenv(k))
	if raw == "" {
		return nil
	}
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if p = strings.TrimSpace(p); p != "" {
			out = append(out, p)
		}
	}
	return out
}
