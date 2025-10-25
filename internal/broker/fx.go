package broker

import (
	"context"
	"time"

	brokerService "personalfinancedss/internal/broker/service"
	"personalfinancedss/internal/broker/worker"
	"personalfinancedss/internal/config"
	accountRepo "personalfinancedss/internal/module/cashflow/account/repository"
	assetRepo "personalfinancedss/internal/module/investment/investment_asset/repository"
	investmentTxnRepo "personalfinancedss/internal/module/investment/investment_transaction/repository"
	"personalfinancedss/internal/service"

	"go.uber.org/fx"
	"go.uber.org/zap"
)

// Module provides broker sync dependencies
var Module = fx.Module("broker",
	fx.Provide(
		provideSyncService,
		provideSyncWorker,
	),
	fx.Invoke(registerWorkerLifecycle),
)

// provideSyncService creates the sync service
func provideSyncService(
	accountRepo accountRepo.Repository,
	assetRepo assetRepo.Repository,
	investmentTxnRepo investmentTxnRepo.Repository,
	encryptionService *service.EncryptionService,
) *brokerService.SyncService {
	return brokerService.NewSyncService(accountRepo, assetRepo, investmentTxnRepo, encryptionService)
}

// provideSyncWorker creates the sync worker
func provideSyncWorker(
	cfg *config.Config,
	accountRepo accountRepo.Repository,
	syncService *brokerService.SyncService,
	logger *zap.Logger,
) *worker.SyncWorker {
	workerConfig := worker.SyncWorkerConfig{
		Enabled:       cfg.BrokerSync.Enabled,
		Interval:      time.Duration(cfg.BrokerSync.IntervalMin) * time.Minute,
		MaxConcurrent: cfg.BrokerSync.MaxConcurrent,
		SyncTimeout:   time.Duration(cfg.BrokerSync.TimeoutMin) * time.Minute,
	}

	return worker.NewSyncWorker(workerConfig, accountRepo, syncService, logger)
}

// registerWorkerLifecycle registers the worker with FX lifecycle
func registerWorkerLifecycle(
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
