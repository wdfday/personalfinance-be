package fx

import (
	"fmt"
	"net/http"
	"time"

	"personalfinancedss/internal/config"
	"personalfinancedss/internal/logger"
	"personalfinancedss/internal/middleware"
	"personalfinancedss/internal/service"
	"personalfinancedss/internal/shared"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

// CoreModule provides core application dependencies
var CoreModule = fx.Module("core",
	fx.Provide(
		// Configuration
		config.Load,

		// Logger (must be early)
		NewLogger,

		// Database
		NewDatabase,

		// Gin router
		NewGinRouter,

		// Services
		NewEncryptionService,

		// Middlewares
		middleware.NewMiddleware,
		middleware.NewEmailVerificationMiddleware,
		middleware.NewCORS,
	),
)

// NewLogger creates a new zap logger based on config
func NewLogger(cfg *config.Config) (*zap.Logger, error) {
	log, err := logger.NewLogger(cfg.Logging.Level, cfg.Logging.Format)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize logger: %w", err)
	}

	log.Info("Logger initialized",
		zap.String("level", cfg.Logging.Level),
		zap.String("format", cfg.Logging.Format),
	)

	return log, nil
}

// NewDatabase creates a new database connection
func NewDatabase(cfg *config.Config, log *zap.Logger) (*gorm.DB, error) {
	log.Info("Connecting to database...",
		zap.String("host", cfg.Database.Host),
		zap.Int("port", cfg.Database.Port),
		zap.String("database", cfg.Database.Name),
		zap.String("user", cfg.Database.User),
	)
	var dsn string

	// Use DATABASE_URL if available, otherwise construct from components
	if cfg.Database.URL != "" {
		dsn = cfg.Database.URL
	} else {
		dsn = fmt.Sprintf(
			"host=%s port=%d user=%s password=%s dbname=%s sslmode=disable TimeZone=Asia/Ho_Chi_Minh",
			cfg.Database.Host,
			cfg.Database.Port,
			cfg.Database.User,
			cfg.Database.Pass,
			cfg.Database.Name,
		)
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		NowFunc: func() time.Time {
			return time.Now().UTC()
		},
	})

	if err != nil {
		log.Error("Failed to connect to database", zap.Error(err))
		return nil, fmt.Errorf("database connection failed: %w", err)
	}

	// Get underlying *sql.DB to configure connection pool
	sqlDB, err := db.DB()
	if err != nil {
		log.Error("Failed to get database instance", zap.Error(err))
		return nil, fmt.Errorf("failed to get database instance: %w", err)
	}

	// Set connection pool settings
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	log.Info("Successfully connected to database",
		zap.Int("max_idle_conns", 10),
		zap.Int("max_open_conns", 100),
		zap.Duration("conn_max_lifetime", time.Hour),
	)
	return db, nil
}

// NewGinRouter creates a new Gin router with basic configuration
func NewGinRouter(cfg *config.Config, log *zap.Logger) *gin.Engine {
	// Set Gin mode based on config
	if config.IsProduction() {
		gin.SetMode(gin.ReleaseMode)
	} else {
		gin.SetMode(gin.DebugMode)
	}

	r := gin.New()

	// Apply logger middleware first so it's available in all subsequent middleware
	r.Use(middleware.LoggerMiddleware(log))

	// Apply recovery middleware
	r.Use(middleware.RecoveryMiddleware())

	// Apply error handler middleware
	r.Use(middleware.ErrorHandlerMiddleware())

	// Apply CORS middleware
	corsMiddleware := middleware.NewCORS(cfg.CORS.Origins)
	r.Use(corsMiddleware)

	// Apply rate limiting middleware (global IP-based rate limiting)
	// Allow 100 requests per second with burst of 200
	rateLimiter := middleware.IPRateLimiter(100, 200)
	r.Use(rateLimiter)

	// Request logging middleware (only in debug mode)
	if config.IsDevelopment() {
		r.Use(gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
			return fmt.Sprintf("[%s] %s %s %d %s \"%s\" %s\n",
				param.TimeStamp.Format("2006/01/02 - 15:04:05"),
				param.ClientIP,
				param.Method,
				param.StatusCode,
				param.Latency,
				param.Path,
				param.ErrorMessage,
			)
		}))
	}

	// Health check endpoint
	r.GET("/health", func(c *gin.Context) {
		shared.RespondWithSuccess(c, http.StatusOK, "Service is healthy", gin.H{
			"status":    "ok",
			"timestamp": time.Now().UTC().Format(time.RFC3339),
		})
	})

	// Serve Swagger 2.0 spec files at separate path to avoid route conflict
	r.StaticFile("/openapi/swagger.yaml", "./docs/swagger.yaml")
	r.StaticFile("/openapi/swagger.json", "./docs/swagger.json")

	// Swagger UI pointing to Swagger 2.0 YAML file
	url := ginSwagger.URL("/openapi/swagger.yaml") // Point to Swagger 2.0 YAML file
	swaggerHandler := ginSwagger.WrapHandler(swaggerFiles.Handler, url,
		ginSwagger.PersistAuthorization(true), // Persist authorization across page refresh
		ginSwagger.DocExpansion("list"),
		ginSwagger.DefaultModelsExpandDepth(-1),
	)

	// Support both /swagger and /swagger-ui paths
	r.GET("/swagger/*any", swaggerHandler)
	r.GET("/swagger-ui/*any", swaggerHandler)

	return r
}

// NewEncryptionService creates a new encryption service
func NewEncryptionService(cfg *config.Config, log *zap.Logger) (*service.EncryptionService, error) {
	encryptionService, err := service.NewEncryptionService(cfg.Encryption.Key)
	if err != nil {
		log.Error("Failed to initialize encryption service", zap.Error(err))
		return nil, fmt.Errorf("encryption service initialization failed: %w", err)
	}

	log.Info("Encryption service initialized (AES-256-GCM)")
	return encryptionService, nil
}
