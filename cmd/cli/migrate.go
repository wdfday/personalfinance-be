package cmd

import (
	"log"
	"os"

	"personalfinancedss/internal/database"

	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var migrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "Run database migrations",
	Long:  `Run automatic database migrations for all entities.`,
	Run: func(cmd *cobra.Command, args []string) {
		runMigrate()
	},
}

var migrateResetCmd = &cobra.Command{
	Use:   "reset",
	Short: "Drop all tables and re-migrate",
	Long:  `WARNING: This will delete all data! Drops all tables and runs migrations fresh.`,
	Run: func(cmd *cobra.Command, args []string) {
		runMigrateReset()
	},
}

func init() {
	rootCmd.AddCommand(migrateCmd)
	migrateCmd.AddCommand(migrateResetCmd)
}

func connectDB() (*gorm.DB, error) {
	// Use same logic as main app - read from env
	dsn := getDSN()
	return gorm.Open(postgres.Open(dsn), &gorm.Config{})
}

func getDSN() string {
	// Read from environment variables with defaults matching docker-compose
	host := getEnv("DB_HOST", "localhost")
	user := getEnv("DB_USER", "finance_user")
	password := getEnv("DB_PASSWORD", "finance_password")
	dbname := getEnv("DB_NAME", "finance_dss")
	port := getEnv("DB_PORT", "5432")

	return "host=" + host + " user=" + user + " password=" + password + " dbname=" + dbname + " port=" + port + " sslmode=disable"
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func runMigrate() {
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	log.Println("🔧 Running database migrations...")

	db, err := connectDB()
	if err != nil {
		log.Fatalf("❌ Failed to connect to database: %v", err)
	}

	if err := database.AutoMigrate(db, logger); err != nil {
		log.Fatalf("❌ Migration failed: %v", err)
	}

	log.Println("✅ Migrations completed successfully!")
}

func runMigrateReset() {
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	log.Println("⚠️  WARNING: Dropping all tables...")

	db, err := connectDB()
	if err != nil {
		log.Fatalf("❌ Failed to connect to database: %v", err)
	}

	if err := database.DropAllTables(db, logger); err != nil {
		log.Fatalf("❌ Failed to drop tables: %v", err)
	}

	log.Println("🔧 Running fresh migrations...")

	if err := database.AutoMigrate(db, logger); err != nil {
		log.Fatalf("❌ Migration failed: %v", err)
	}

	log.Println("✅ Database reset completed!")
}
