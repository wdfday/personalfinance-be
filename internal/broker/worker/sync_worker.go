package worker

import (
	"context"
	"sync"
	"time"

	"personalfinancedss/internal/broker/service"
	accountDomain "personalfinancedss/internal/module/cashflow/account/domain"
	accountRepo "personalfinancedss/internal/module/cashflow/account/repository"

	"go.uber.org/zap"
)

// SyncWorkerConfig holds configuration for the sync worker
type SyncWorkerConfig struct {
	Enabled       bool          // Enable/disable the worker
	Interval      time.Duration // How often to check for accounts needing sync
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

// SyncWorker handles periodic syncing of broker accounts
type SyncWorker struct {
	config      SyncWorkerConfig
	accountRepo accountRepo.Repository
	syncService *service.SyncService
	logger      *zap.Logger
	stopChan    chan struct{}
	wg          sync.WaitGroup
	semaphore   chan struct{} // Limit concurrent syncs
}

// NewSyncWorker creates a new sync worker
func NewSyncWorker(
	config SyncWorkerConfig,
	accountRepo accountRepo.Repository,
	syncService *service.SyncService,
	logger *zap.Logger,
) *SyncWorker {
	return &SyncWorker{
		config:      config,
		accountRepo: accountRepo,
		syncService: syncService,
		logger:      logger,
		stopChan:    make(chan struct{}),
		semaphore:   make(chan struct{}, config.MaxConcurrent),
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
	w.syncAccounts(ctx)

	for {
		select {
		case <-ticker.C:
			w.syncAccounts(ctx)

		case <-w.stopChan:
			w.logger.Info("Broker sync worker received stop signal")
			return

		case <-ctx.Done():
			w.logger.Info("Broker sync worker context cancelled")
			return
		}
	}
}

// syncAccounts finds and syncs all accounts that need syncing
func (w *SyncWorker) syncAccounts(ctx context.Context) {
	startTime := time.Now()

	w.logger.Debug("🔍 Checking for accounts needing sync...")

	// Get accounts that need syncing
	accounts, err := w.accountRepo.GetAccountsNeedingSync(ctx)
	if err != nil {
		w.logger.Error("Failed to get accounts needing sync", zap.Error(err))
		return
	}

	if len(accounts) == 0 {
		w.logger.Debug("No accounts need syncing at this time")
		return
	}

	w.logger.Info("📊 Found accounts needing sync",
		zap.Int("count", len(accounts)),
	)

	// Sync accounts concurrently with rate limiting
	var syncWg sync.WaitGroup
	successCount := 0
	failureCount := 0
	var mu sync.Mutex

	for _, account := range accounts {
		// Acquire semaphore slot
		w.semaphore <- struct{}{}

		syncWg.Add(1)
		go func(acc *accountDomain.Account) {
			defer syncWg.Done()
			defer func() { <-w.semaphore }() // Release semaphore slot

			// Create timeout context for this sync
			syncCtx, cancel := context.WithTimeout(ctx, w.config.SyncTimeout)
			defer cancel()

			// Perform sync
			w.logger.Info("🔄 Syncing account",
				zap.String("account_id", acc.ID.String()),
				zap.String("account_name", acc.AccountName),
			)

			result, err := w.syncService.SyncAccount(syncCtx, acc)

			mu.Lock()
			defer mu.Unlock()

			if err != nil || !result.Success {
				failureCount++
				w.logger.Error("❌ Account sync failed",
					zap.String("account_id", acc.ID.String()),
					zap.String("account_name", acc.AccountName),
					zap.Error(err),
					zap.Any("result", result),
				)
			} else {
				successCount++
				w.logger.Info("✅ Account synced successfully",
					zap.String("account_id", acc.ID.String()),
					zap.String("account_name", acc.AccountName),
					zap.Int("assets_synced", result.AssetsCount),
					zap.Int("transactions_synced", result.TransactionsCount),
					zap.Int("prices_updated", result.UpdatedPricesCount),
					zap.Bool("balance_updated", result.BalanceUpdated),
				)
			}
		}(account)
	}

	// Wait for all syncs to complete
	syncWg.Wait()

	duration := time.Since(startTime)
	w.logger.Info("📈 Sync cycle completed",
		zap.Int("total_accounts", len(accounts)),
		zap.Int("successful", successCount),
		zap.Int("failed", failureCount),
		zap.Duration("duration", duration),
	)
}

// ForceSync triggers an immediate sync check (useful for testing or manual triggers)
func (w *SyncWorker) ForceSync(ctx context.Context) {
	w.logger.Info("🔧 Manual sync triggered")
	w.syncAccounts(ctx)
}
