package broker

import (
	"context"
	"time"

	"personalfinancedss/internal/broker/client/okx"
	"personalfinancedss/internal/broker/client/ssi"
	"personalfinancedss/internal/config"
	"personalfinancedss/internal/module/identify/broker/client/sepay"
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
		ssi.NewClient,
		okx.NewClient,
		sepay.NewClient,

		// Repository
		provideBrokerConnectionRepository,

		// Service
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

// provideBrokerConnectionService creates the broker connection service
func provideBrokerConnectionService(
	repo repository2.BrokerConnectionRepository,
	encryptionService internalService.EncryptionService,
	ssiClient *ssi.Client,
	okxClient *okx.Client,
	sepayClient *sepay.Client,
) service2.BrokerConnectionService {
	return service2.NewBrokerConnectionService(
		repo,
		encryptionService,
		ssiClient,
		okxClient,
		sepayClient,
	)
}

// provideBrokerConnectionHandler creates the broker connection handler
func provideBrokerConnectionHandler(
	service service2.BrokerConnectionService,
) *handler.BrokerConnectionHandler {
	return handler.NewBrokerConnectionHandler(service)
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
) {
	handler.RegisterRoutes(router)
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
