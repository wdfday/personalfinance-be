package cmd

import (
	"log"

	"personalfinancedss/internal/database"

	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var dbCmd = &cobra.Command{
	Use:   "db",
	Short: "Database management commands",
	Long:  `Manage database operations`,
}

var dbCleanCmd = &cobra.Command{
	Use:   "clean",
	Short: "Clean database (drop all tables + fresh migrations, NO seed)",
	Long:  `WARNING: Drops ALL tables and creates fresh empty database. No data will be seeded.`,
	Run: func(cmd *cobra.Command, args []string) {
		runDBClean()
	},
}

var dbResetCmd = &cobra.Command{
	Use:   "reset",
	Short: "Complete database reset (drop + migrate + seed)",
	Long:  `WARNING: Drops ALL tables, runs fresh migrations, and seeds data.`,
	Run: func(cmd *cobra.Command, args []string) {
		runDBReset()
	},
}

func init() {
	rootCmd.AddCommand(dbCmd)
	dbCmd.AddCommand(dbCleanCmd)
	dbCmd.AddCommand(dbResetCmd)
}

func runDBClean() {
	// Load env
	_ = loadEnvFile()

	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	log.Println("🧹 ========================================")
	log.Println("🧹 CLEANING DATABASE")
	log.Println("🧹 Dropping all tables + fresh migrations")
	log.Println("🧹 NO DATA WILL BE SEEDED")
	log.Println("🧹 ========================================")

	dsn := getDSN()
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("❌ Failed to connect to database: %v", err)
	}

	// Step 1: Drop all tables
	log.Println("\n📋 Step 1/2: Dropping all tables...")
	if err := database.DropAllTables(db, logger); err != nil {
		log.Fatalf("❌ Failed to drop tables: %v", err)
	}
	log.Println("✅ Tables dropped")

	// Step 2: Run migrations
	log.Println("\n📋 Step 2/2: Running fresh migrations...")
	if err := database.AutoMigrate(db, logger); err != nil {
		log.Fatalf("❌ Migration failed: %v", err)
	}
	log.Println("✅ Migrations completed")

	log.Println("\n✨ ========================================")
	log.Println("✨ DATABASE CLEANED!")
	log.Println("✨ All tables dropped and recreated")
	log.Println("✨ Database is now EMPTY (no data)")
	log.Println("✨ ========================================")
}

func runDBReset() {
	// Load env
	_ = loadEnvFile()

	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	log.Println("⚠️  ========================================")
	log.Println("⚠️  COMPLETE DATABASE RESET")
	log.Println("⚠️  This will DELETE ALL DATA!")
	log.Println("⚠️  ========================================")

	dsn := getDSN()
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("❌ Failed to connect to database: %v", err)
	}

	// Step 1: Drop all tables
	log.Println("\n📋 Step 1/3: Dropping all tables...")
	if err := database.DropAllTables(db, logger); err != nil {
		log.Fatalf("❌ Failed to drop tables: %v", err)
	}
	log.Println("✅ Tables dropped")

	// Step 2: Run migrations
	log.Println("\n📋 Step 2/3: Running fresh migrations...")
	if err := database.AutoMigrate(db, logger); err != nil {
		log.Fatalf("❌ Migration failed: %v", err)
	}
	log.Println("✅ Migrations completed")

	// Step 3: Seed data
	log.Println("\n📋 Step 3/3: Seeding database...")
	seeder := &CLISeeder{db: db, logger: logger}

	err = db.Transaction(func(tx *gorm.DB) error {
		// Seed categories
		logger.Info("Seeding categories...")
		if err := seeder.seedCategories(tx); err != nil {
			return err
		}

		// Seed users
		logger.Info("Seeding users...")
		if err := seeder.seedUsers(tx); err != nil {
			return err
		}

		// Seed financial data
		logger.Info("Seeding financial data...")
		mainSeeder := database.NewSeeder(
			tx,
			&bcryptHasher{},
			&noopUserService{},
			getEnvCLI("ADMIN_EMAIL", "admin@example.com"),
			getEnvCLI("ADMIN_PASSWORD", "Admin@123"),
			logger,
		)
		if err := mainSeeder.SeedMonthDSSData(tx); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		log.Fatalf("❌ Seeding failed: %v", err)
	}

	log.Println("✅ Seeding completed")
	log.Println("\n🎉 ========================================")
	log.Println("🎉 DATABASE FULLY RESET AND SEEDED!")
	log.Println("🎉 ========================================")
	log.Println("\n📊 Seeded data:")
	log.Println("   - Default categories")
	log.Println("   - 5 users (1 admin + 4 profiles)")
	log.Println("   - Financial data for 4 user profiles:")
	log.Println("     • Salaried (john.doe@example.com)")
	log.Println("     • Freelancer (jane.smith@example.com)")
	log.Println("     • Mixed (alice.johnson@example.com)")
	log.Println("     • Student (bob.wilson@example.com)")
	log.Println("\n🔐 Login credentials:")
	log.Println("   - Admin: admin@example.com / Admin@123")
	log.Println("   - Users: Password123!")
}
