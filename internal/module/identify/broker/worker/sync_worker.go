package worker

import (
	"context"
	"personalfinancedss/internal/module/identify/broker/domain"
	"personalfinancedss/internal/module/identify/broker/repository"
	"personalfinancedss/internal/module/identify/broker/service"
	"sync"
	"time"

	"go.uber.org/zap"
)

// SyncWorkerConfig holds configuration for the sync worker
type SyncWorkerConfig struct {
	Enabled       bool          // Enable/disable the worker
	Interval      time.Duration // How often to check for connections needing sync
	MaxConcurrent int           // Max number of concurrent syncs
	SyncTimeout   time.Duration // Timeout for each sync operation
}

// DefaultSyncWorkerConfig returns default configuration
func DefaultSyncWorkerConfig() SyncWorkerConfig {
	return SyncWorkerConfig{
		Enabled:       true,
		Interval:      1 * time.Minute, // Check every minute
		MaxConcurrent: 5,               // Max 5 concurrent syncs
		SyncTimeout:   2 * time.Minute, // 2 minute timeout per sync
	}
}

// SyncWorker handles periodic syncing of broker connections
type SyncWorker struct {
	config     SyncWorkerConfig
	brokerRepo repository.BrokerConnectionRepository
	brokerSvc  service.BrokerConnectionService
	logger     *zap.Logger
	stopChan   chan struct{}
	wg         sync.WaitGroup
	semaphore  chan struct{} // Limit concurrent syncs
}

// NewSyncWorker creates a new sync worker
func NewSyncWorker(
	config SyncWorkerConfig,
	brokerRepo repository.BrokerConnectionRepository,
	brokerSvc service.BrokerConnectionService,
	logger *zap.Logger,
) *SyncWorker {
	return &SyncWorker{
		config:     config,
		brokerRepo: brokerRepo,
		brokerSvc:  brokerSvc,
		logger:     logger.Named("broker.sync.worker"),
		stopChan:   make(chan struct{}),
		semaphore:  make(chan struct{}, config.MaxConcurrent),
	}
}

// Start starts the sync worker
func (w *SyncWorker) Start(ctx context.Context) error {
	if !w.config.Enabled {
		w.logger.Info("🔕 Broker sync worker is disabled")
		return nil
	}

	w.logger.Info("🚀 Starting broker sync worker",
		zap.Duration("interval", w.config.Interval),
		zap.Int("max_concurrent", w.config.MaxConcurrent),
		zap.Duration("sync_timeout", w.config.SyncTimeout),
	)

	// Start worker goroutine
	w.wg.Add(1)
	go w.run(ctx)

	return nil
}

// Stop stops the sync worker gracefully
func (w *SyncWorker) Stop(ctx context.Context) error {
	w.logger.Info("🛑 Stopping broker sync worker...")

	// Signal stop
	close(w.stopChan)

	// Wait for worker to finish with timeout
	done := make(chan struct{})
	go func() {
		w.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		w.logger.Info("✅ Broker sync worker stopped gracefully")
		return nil
	case <-ctx.Done():
		w.logger.Warn("⚠️  Broker sync worker shutdown timeout")
		return ctx.Err()
	}
}

// run is the main worker loop
func (w *SyncWorker) run(ctx context.Context) {
	defer w.wg.Done()

	ticker := time.NewTicker(w.config.Interval)
	defer ticker.Stop()

	w.logger.Info("✅ Broker sync worker started")

	// Run initial sync immediately
	w.syncConnections(ctx)

	for {
		select {
		case <-ticker.C:
			w.syncConnections(ctx)

		case <-w.stopChan:
			w.logger.Info("Broker sync worker received stop signal")
			return

		case <-ctx.Done():
			w.logger.Info("Broker sync worker context cancelled")
			return
		}
	}
}

// syncConnections finds and syncs all broker connections that need syncing
func (w *SyncWorker) syncConnections(ctx context.Context) {
	startTime := time.Now()

	w.logger.Debug("🔍 Checking for broker connections needing sync...")

	// Get connections that need syncing
	connections, err := w.brokerRepo.GetNeedingSync(ctx, 100) // Max 100 per cycle
	if err != nil {
		w.logger.Error("Failed to get connections needing sync", zap.Error(err))
		return
	}

	if len(connections) == 0 {
		w.logger.Debug("No broker connections need syncing at this time")
		return
	}

	w.logger.Info("📊 Found broker connections needing sync",
		zap.Int("count", len(connections)),
	)

	// Sync connections concurrently with rate limiting
	var syncWg sync.WaitGroup
	successCount := 0
	failureCount := 0
	var mu sync.Mutex

	for _, conn := range connections {
		// Acquire semaphore slot
		w.semaphore <- struct{}{}

		syncWg.Add(1)
		go func(connection *domain.BrokerConnection) {
			defer syncWg.Done()
			defer func() { <-w.semaphore }() // Release semaphore slot

			// Create timeout context for this sync
			syncCtx, cancel := context.WithTimeout(ctx, w.config.SyncTimeout)
			defer cancel()

			// Perform sync
			w.logger.Info("🔄 Syncing broker connection",
				zap.String("connection_id", connection.ID.String()),
				zap.String("broker_type", string(connection.BrokerType)),
				zap.String("broker_name", connection.BrokerName),
			)

			// Use SyncNow method from broker service
			result, err := w.brokerSvc.SyncNow(syncCtx, connection.ID, connection.UserID)

			mu.Lock()
			defer mu.Unlock()

			if err != nil || !result.Success {
				failureCount++
				w.logger.Error("❌ Broker connection sync failed",
					zap.String("connection_id", connection.ID.String()),
					zap.String("broker_type", string(connection.BrokerType)),
					zap.Error(err),
					zap.Any("result", result),
				)
			} else {
				successCount++
				w.logger.Info("✅ Broker connection synced successfully",
					zap.String("connection_id", connection.ID.String()),
					zap.String("broker_type", string(connection.BrokerType)),
					zap.Int("assets_synced", result.AssetsCount),
					zap.Int("transactions_synced", result.TransactionsCount),
					zap.Int("prices_updated", result.UpdatedPricesCount),
					zap.Bool("balance_updated", result.BalanceUpdated),
				)
			}
		}(conn)
	}

	// Wait for all syncs to complete
	syncWg.Wait()

	duration := time.Since(startTime)
	w.logger.Info("📈 Sync cycle completed",
		zap.Int("total_connections", len(connections)),
		zap.Int("successful", successCount),
		zap.Int("failed", failureCount),
		zap.Duration("duration", duration),
	)
}

// ForceSync triggers an immediate sync check (useful for testing or manual triggers)
func (w *SyncWorker) ForceSync(ctx context.Context) {
	w.logger.Info("🔧 Manual sync triggered")
	w.syncConnections(ctx)
}
