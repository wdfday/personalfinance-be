package fx

import (
	"context"
	"net/http"
	"personalfinancedss/internal/config"
	"personalfinancedss/internal/database"
	"personalfinancedss/internal/middleware"

	allocationHandler "personalfinancedss/internal/module/analytics/budget_allocation/handler"
	ahpHandler "personalfinancedss/internal/module/analytics/goal_prioritization/handler"

	eventHandler "personalfinancedss/internal/module/calendar/event/handler"
	monthHandler "personalfinancedss/internal/module/calendar/month/handler"
	accountHandler "personalfinancedss/internal/module/cashflow/account/handler"
	budgetHandler "personalfinancedss/internal/module/cashflow/budget/handler"
	budgetProfileHandler "personalfinancedss/internal/module/cashflow/budget_profile/handler"
	categoryHandler "personalfinancedss/internal/module/cashflow/category/handler"
	debtHandler "personalfinancedss/internal/module/cashflow/debt/handler"
	goalHandler "personalfinancedss/internal/module/cashflow/goal/handler"
	incomeProfileHandler "personalfinancedss/internal/module/cashflow/income_profile/handler"
	transactionHandler "personalfinancedss/internal/module/cashflow/transaction/handler"

	// chatbotHandler "personalfinancedss/internal/module/chatbot/handler" // Temporarily disabled
	authHandler "personalfinancedss/internal/module/identify/auth/handler"
	authService "personalfinancedss/internal/module/identify/auth/service"
	brokerHandler "personalfinancedss/internal/module/identify/broker/handler"
	profileHandler "personalfinancedss/internal/module/identify/profile/handler"
	userHandler "personalfinancedss/internal/module/identify/user/handler"
	userService "personalfinancedss/internal/module/identify/user/service"
	assetHandler "personalfinancedss/internal/module/investment/investment_asset/handler"
	investmentTransactionHandler "personalfinancedss/internal/module/investment/investment_transaction/handler"
	snapshotHandler "personalfinancedss/internal/module/investment/portfolio_snapshot/handler"
	notificationHandler "personalfinancedss/internal/module/notification/handler"
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

		// Register routes
		RegisterRoutes,

		// Start server
		StartServer,
	),
)

// RegisterRoutes registers all API routes
func RegisterRoutes(
	router *gin.Engine,
	authH *authHandler.Handler,
	usersH *userHandler.Handler,
	profileH *profileHandler.Handler,
	accountH *accountHandler.Handler,
	categoryH *categoryHandler.Handler,
	incomeProfileH *incomeProfileHandler.Handler,
	budgetH *budgetHandler.Handler,
	budgetProfileH *budgetProfileHandler.Handler,
	debtH *debtHandler.Handler,
	goalH *goalHandler.Handler,
	transactionH *transactionHandler.Handler,
	assetH *assetHandler.Handler,
	investmentTransH *investmentTransactionHandler.Handler,
	snapshotH *snapshotHandler.Handler,
	notificationH *notificationHandler.Handler,
	wsHandler *notificationHandler.WebSocketHandler,
	preferenceH *notificationHandler.PreferenceHandler,
	eventH *eventHandler.Handler,
	monthH *monthHandler.Handler,
	allocationH *allocationHandler.Handler,
	ahpH *ahpHandler.Handler,
	// chatbotH *chatbotHandler.Handler, // Temporarily disabled
	brokerH *brokerHandler.BrokerConnectionHandler,
	authMiddleware *middleware.Middleware,
	emailVerificationMiddleware *middleware.EmailVerificationMiddleware,
	logger *zap.Logger,
) {
	logger.Info("=== Route Registration Phase ===")

	// Register module routes
	logger.Info("Registering auth routes...")
	authH.RegisterRoutes(router, authMiddleware, emailVerificationMiddleware)

	logger.Info("Registering user routes...")
	usersH.RegisterRoutes(router, authMiddleware)

	logger.Info("Registering profile routes...")
	profileH.RegisterRoutes(router, authMiddleware)

	logger.Info("Registering account routes...")
	accountH.RegisterRoutes(router, authMiddleware)

	logger.Info("Registering category routes...")
	categoryH.RegisterRoutes(router, authMiddleware)

	logger.Info("Registering income profile routes...")
	incomeProfileH.RegisterRoutes(router, authMiddleware)

	logger.Info("Registering budget constraint routes...")
	budgetProfileH.RegisterRoutes(router, authMiddleware)

	logger.Info("Registering budget routes...")
	budgetH.RegisterRoutes(router, authMiddleware)

	logger.Info("Registering goal routes...")
	goalH.RegisterRoutes(router, authMiddleware)

	logger.Info("Registering debt routes...")
	debtH.RegisterRoutes(router, authMiddleware)

	logger.Info("Registering transaction routes...")
	transactionH.RegisterRoutes(router, authMiddleware)

	logger.Info("Registering investment asset routes...")
	assetH.RegisterRoutes(router, authMiddleware)

	logger.Info("Registering investment transaction routes...")
	investmentTransH.RegisterRoutes(router, authMiddleware)

	logger.Info("Registering portfolio snapshot routes...")
	snapshotH.RegisterRoutes(router, authMiddleware)

	logger.Info("Registering notification routes...")
	notificationH.RegisterRoutes(router, authMiddleware)

	logger.Info("Registering WebSocket routes...")
	wsHandler.RegisterRoutes(router, authMiddleware)

	logger.Info("Registering notification preference routes...")
	preferenceH.RegisterRoutes(router, authMiddleware)

	// logger.Info("Registering calendar event routes...")
	// eventH.RegisterRoutes(router, authMiddleware)

	logger.Info("Registering budget allocation routes...")
	allocationH.RegisterRoutes(router)

	logger.Info("Registering goal prioritization (AHP) routes...")
	ahpH.RegisterRoutes(router)

	// logger.Info("Registering chatbot routes...")
	// chatbotH.RegisterRoutes(router, authMiddleware) // Temporarily disabled

	logger.Info("Registering calendar event routes...")
	eventH.RegisterRoutes(router, authMiddleware)

	logger.Info("Registering month routes...")
	monthH.RegisterRoutes(router, authMiddleware)

	logger.Info("Registering broker routes...")
	brokerH.RegisterRoutes(router, authMiddleware)

	logger.Info("✅ All routes registered successfully")
}

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
	if err := database.AutoMigrate(db, logger); err != nil {
		logger.Fatal("Failed to run migrations", zap.Error(err))
	}

	// Run seeding (only in development or if AUTO_SEED is true)
	if config.IsDevelopment() {
		logger.Info("Running database seeding (development mode)...")
		seeder := database.NewSeeder(db, passwordService, userSvc, cfg.Seeding.AdminEmail, cfg.Seeding.AdminPassword, logger)
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
				logger.Info("🚀 Starting HTTP server",
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
