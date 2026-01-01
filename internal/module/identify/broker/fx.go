package broker

import (
	"context"
	"time"

	"personalfinancedss/internal/config"
	"personalfinancedss/internal/middleware"
	accountRepo "personalfinancedss/internal/module/cashflow/account/repository"
	transactionRepo "personalfinancedss/internal/module/cashflow/transaction/repository"
	"personalfinancedss/internal/module/identify/broker/client/okx"
	"personalfinancedss/internal/module/identify/broker/client/sepay"
	"personalfinancedss/internal/module/identify/broker/client/ssi"
	"personalfinancedss/internal/module/identify/broker/handler"
	repository2 "personalfinancedss/internal/module/identify/broker/repository"
	service2 "personalfinancedss/internal/module/identify/broker/service"
	"personalfinancedss/internal/module/identify/broker/worker"
	internalService "personalfinancedss/internal/service"

	"github.com/gin-gonic/gin"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// Module provides broker connection management dependencies
var Module = fx.Module("broker",
	fx.Provide(
		// Broker clients
		ssi.NewSSIClient,
		okx.NewOKXClient,
		sepay.NewClient,

		// Repository
		provideBrokerConnectionRepository,

		// Services
		provideSyncService,
		provideBrokerConnectionService,

		// Handler
		provideBrokerConnectionHandler,

		// Worker
		provideSyncWorker,
	),
	fx.Invoke(
		registerBrokerRoutes,
		registerSyncWorkerLifecycle,
	),
)

// provideBrokerConnectionRepository creates the broker connection repository
func provideBrokerConnectionRepository(db *gorm.DB) repository2.BrokerConnectionRepository {
	return repository2.NewGormBrokerConnectionRepository(db)
}

// provideSyncService creates the sync service
func provideSyncService(
	brokerRepo repository2.BrokerConnectionRepository,
	accRepo accountRepo.Repository,
	txnRepo transactionRepo.Repository,
	encryptionService *internalService.EncryptionService,
	ssiClient *ssi.SSIClient,
	okxClient *okx.OKXClient,
	sepayClient *sepay.Client,
	logger *zap.Logger,
) *service2.SyncService {
	return service2.NewSyncService(
		brokerRepo,
		accRepo,
		txnRepo,
		encryptionService,
		ssiClient,
		okxClient,
		sepayClient,
		logger,
	)
}

// provideBrokerConnectionService creates the broker connection service
func provideBrokerConnectionService(
	repo repository2.BrokerConnectionRepository,
	encryptionService *internalService.EncryptionService,
	ssiClient *ssi.SSIClient,
	okxClient *okx.OKXClient,
	sepayClient *sepay.Client,
	syncService *service2.SyncService,
) service2.BrokerConnectionService {
	return service2.NewBrokerConnectionService(
		repo,
		encryptionService,
		ssiClient,
		okxClient,
		sepayClient,
		syncService,
	)
}

// provideBrokerConnectionHandler creates the broker connection handler
func provideBrokerConnectionHandler(
	service service2.BrokerConnectionService,
	logger *zap.Logger,
) *handler.BrokerConnectionHandler {
	return handler.NewBrokerConnectionHandler(service, logger)
}

// provideSyncWorker creates the broker sync worker
func provideSyncWorker(
	cfg *config.Config,
	repo repository2.BrokerConnectionRepository,
	svc service2.BrokerConnectionService,
	logger *zap.Logger,
) *worker.SyncWorker {
	workerConfig := worker.SyncWorkerConfig{
		Enabled:       cfg.BrokerSync.Enabled,
		Interval:      time.Duration(cfg.BrokerSync.IntervalMin) * time.Minute,
		MaxConcurrent: cfg.BrokerSync.MaxConcurrent,
		SyncTimeout:   time.Duration(cfg.BrokerSync.TimeoutMin) * time.Minute,
	}

	return worker.NewSyncWorker(workerConfig, repo, svc, logger)
}

// registerBrokerRoutes registers broker connection routes
func registerBrokerRoutes(
	router *gin.Engine,
	handler *handler.BrokerConnectionHandler,
	authMiddleware *middleware.Middleware,
) {
	handler.RegisterRoutes(router, authMiddleware)
}

// registerSyncWorkerLifecycle registers the sync worker lifecycle hooks
func registerSyncWorkerLifecycle(
	lc fx.Lifecycle,
	w *worker.SyncWorker,
	logger *zap.Logger,
) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			logger.Info("🚀 Starting broker sync worker...")
			return w.Start(ctx)
		},
		OnStop: func(ctx context.Context) error {
			logger.Info("🛑 Stopping broker sync worker...")
			return w.Stop(ctx)
		},
	})
}
