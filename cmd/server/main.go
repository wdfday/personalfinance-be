package main

import (
	"log"
	"personalfinancedss/internal/app"
	"personalfinancedss/internal/config"
)

func main() {
	log.Println("========================================")
	log.Println("  Personal Finance DSS API Server")
	log.Println("========================================")

	cfg := config.Load()

	if err := config.ValidateConfig(); err != nil {
		log.Fatalf("Configuration validation failed: %v", err)
	}

	config.PrintConfig()

	log.Printf("Server: http://%s:%s", cfg.Server.Host, cfg.Server.Port)
	log.Printf("Swagger: http://%s:%s/swagger/index.html", cfg.Server.Host, cfg.Server.Port)

	if config.IsDevelopment() {
		log.Println("Mode: DEVELOPMENT")
	} else {
		log.Println("Mode: PRODUCTION")
	}

	app.Application().Run()
}
