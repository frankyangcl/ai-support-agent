package config

import "os"

type Config struct {
	DatabaseURL string
	ServerAddr  string
}

func Load() Config {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		databaseURL = "postgres://postgres:postgres@localhost:5432/ai_support_agent?sslmode=disable"
	}

	serverAddr := os.Getenv("SERVER_ADDR")
	if serverAddr == "" {
		serverAddr = ":8080"
	}

	return Config{
		DatabaseURL: databaseURL,
		ServerAddr:  serverAddr,
	}
}