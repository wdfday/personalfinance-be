package app

import (
	"context"
	"net/http"
	"personalfinancedss/internal/config"
	database2 "personalfinancedss/internal/infra/database"
	authService "personalfinancedss/internal/module/identify/auth/service"
	userService "personalfinancedss/internal/module/identify/user/service"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// AppModule provides the main application dependencies
var AppModule = fx.Module("app",
	fx.Invoke(
		// Run migrations and seeding (must run before server starts)
		RunMigrationsAndSeeding,

		// Start server
		StartServer,
	),
)

// RunMigrationsAndSeeding runs database migrations and seeding
func RunMigrationsAndSeeding(
	db *gorm.DB,
	cfg *config.Config,
	passwordService authService.IPasswordService,
	userSvc userService.IUserService,
	logger *zap.Logger,
) {
	logger.Info("=== Database Migration & Seeding Phase ===")

	// Run auto migrations
	logger.Info("Starting database migrations...")
	if err := database2.AutoMigrate(db, logger); err != nil {
		logger.Fatal("Failed to run migrations", zap.Error(err))
	}

	// Run seeding (only in development or if AUTO_SEED is true)
	if config.IsDevelopment() {
		logger.Info("Running database seeding (development mode)...")
		seeder := database2.NewSeeder(db, passwordService, userSvc, cfg.Seeding.AdminEmail, cfg.Seeding.AdminPassword, logger)
		if err := seeder.SeedAll(); err != nil {
			logger.Warn("⚠️  Seeding failed", zap.Error(err))
			// Don't fatal on seeding errors, just warn
		}
	} else {
		logger.Info("Skipping database seeding (production mode)")
	}

	logger.Info("=== Migration & Seeding Complete ===")
}

// StartServer starts the HTTP server with graceful shutdown
func StartServer(lc fx.Lifecycle, router *gin.Engine, cfg *config.Config, logger *zap.Logger) {
	server := &http.Server{
		Addr:         ":" + cfg.Server.Port,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			go func() {
				logger.Info("Starting HTTP server",
					zap.String("addr", server.Addr),
					zap.Duration("read_timeout", 15*time.Second),
					zap.Duration("write_timeout", 15*time.Second),
					zap.Duration("idle_timeout", 60*time.Second),
				)
				logger.Info("Server URLs",
					zap.String("base", "http://"+cfg.Server.Host+":"+cfg.Server.Port),
					zap.String("swagger", "http://"+cfg.Server.Host+":"+cfg.Server.Port+"/swagger/index.html"),
					zap.String("health", "http://"+cfg.Server.Host+":"+cfg.Server.Port+"/health"),
				)

				if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
					logger.Fatal("Failed to start server", zap.Error(err))
				}
			}()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			logger.Info("Shutting down HTTP server...")
			shutdownCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
			defer cancel()

			if err := server.Shutdown(shutdownCtx); err != nil {
				logger.Error("Server forced to shutdown", zap.Error(err))
				return err
			}

			logger.Info("✅ Server gracefully stopped")
			return nil
		},
	})
}
