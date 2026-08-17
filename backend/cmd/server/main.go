package main

import (
	"log"

	"github.com/frankyangcl/ai-support-agent/backend/internal/config"
	"github.com/frankyangcl/ai-support-agent/backend/internal/database"
	"github.com/frankyangcl/ai-support-agent/backend/internal/router"
)

func main() {
	cfg := config.Load()

	db, err := database.Connect(cfg.DatabaseURL)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	r, err := router.Setup(db, cfg)
	if err != nil {
		log.Fatal(err)
	}
	if err := r.Run(cfg.ServerAddr); err != nil {
		log.Fatal(err)
	}
}
