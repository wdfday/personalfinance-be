package cmd

import (
	"log"

	"personalfinancedss/internal/config"
	"personalfinancedss/internal/fx"

	"github.com/spf13/cobra"
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the API server",
	Long:  `Start the Personal Finance DSS API server with all services.`,
	Run: func(cmd *cobra.Command, args []string) {
		runServe()
	},
}

func init() {
	rootCmd.AddCommand(serveCmd)
}

func runServe() {
	log.Println("========================================")
	log.Println("  Personal Finance DSS API Server")
	log.Println("========================================")
	log.Println()

	// Load configuration
	log.Println("📋 Loading configuration...")
	cfg := config.Load()

	// Validate configuration
	log.Println("🔍 Validating configuration...")
	if err := config.ValidateConfig(); err != nil {
		log.Fatalf("❌ Configuration validation failed: %v", err)
	}

	// Print configuration
	log.Println("⚙️  Configuration Summary")
	config.PrintConfig()

	log.Println()
	log.Println("🚀 Starting application...")
	log.Printf("   Server: http://%s:%s", cfg.Server.Host, cfg.Server.Port)
	log.Printf("   Swagger: http://%s:%s/swagger/index.html", cfg.Server.Host, cfg.Server.Port)

	if config.IsDevelopment() {
		log.Println("   Mode: DEVELOPMENT 🛠")
	} else {
		log.Println("   Mode: PRODUCTION 🏭")
	}

	log.Println()
	log.Println("📦 Initializing dependency injection (Uber FX)...")

	// Run FX application
	fx.Application().Run()
}
