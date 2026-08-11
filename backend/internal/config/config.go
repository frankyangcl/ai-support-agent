package config

import "os"

type Config struct {
	DatabaseURL     string
	ServerAddr      string
	DeepSeekAPIKey  string
	DashScopeAPIKey string
	BailianBaseURL  string
}

func Load() Config {
	deepSeekAPIKey := os.Getenv("DEEPSEEK_API_KEY")
	dashScopeAPIKey := os.Getenv("DASHSCOPE_API_KEY")
	bailianBaseURL := os.Getenv("BAILIAN_BASE_URL")

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		databaseURL = "postgres://postgres:postgres@localhost:5433/ai_support_agent?sslmode=disable"
	}

	serverAddr := os.Getenv("SERVER_ADDR")
	if serverAddr == "" {
		serverAddr = ":8080"
	}

	return Config{
		DatabaseURL:     databaseURL,
		ServerAddr:      serverAddr,
		DeepSeekAPIKey:  deepSeekAPIKey,
		DashScopeAPIKey: dashScopeAPIKey,
		BailianBaseURL:  bailianBaseURL,
	}

}
