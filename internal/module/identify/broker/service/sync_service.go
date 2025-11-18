package service

import (
	"context"
	"fmt"
	accountDomain "personalfinancedss/internal/module/cashflow/account/domain"
	accountRepo "personalfinancedss/internal/module/cashflow/account/repository"
	"personalfinancedss/internal/module/identify/broker/client"
	"personalfinancedss/internal/module/identify/broker/client/okx"
	"personalfinancedss/internal/module/identify/broker/client/ssi"
	assetDomain "personalfinancedss/internal/module/investment/investment_asset/domain"
	assetRepo "personalfinancedss/internal/module/investment/investment_asset/repository"
	investmentTxnDomain "personalfinancedss/internal/module/investment/investment_transaction/domain"
	investmentTxnRepo "personalfinancedss/internal/module/investment/investment_transaction/repository"
	"personalfinancedss/internal/service"
	"time"
)

// SyncService handles syncing data from brokers
type SyncService struct {
	accountRepo       accountRepo.Repository
	assetRepo         assetRepo.Repository
	investmentTxnRepo investmentTxnRepo.Repository
	ssiClient         client.BrokerClient
	okxClient         client.BrokerClient
	encryptionService *service.EncryptionService
}

// NewSyncService creates a new sync service
func NewSyncService(
	accountRepo accountRepo.Repository,
	assetRepo assetRepo.Repository,
	investmentTxnRepo investmentTxnRepo.Repository,
	encryptionService *service.EncryptionService,
) *SyncService {
	return &SyncService{
		accountRepo:       accountRepo,
		assetRepo:         assetRepo,
		investmentTxnRepo: investmentTxnRepo,
		ssiClient:         ssi.NewSSIClient(),
		okxClient:         okx.NewOKXClient(),
		encryptionService: encryptionService,
	}
}

// GetSSIClient returns the SSI broker client
func (s *SyncService) GetSSIClient() client.BrokerClient {
	return s.ssiClient
}

// GetOKXClient returns the OKX broker client
func (s *SyncService) GetOKXClient() client.BrokerClient {
	return s.okxClient
}

// SyncAccount syncs data from broker for a specific account
func (s *SyncService) SyncAccount(ctx context.Context, account *accountDomain.Account) (*client.SyncResult, error) {
	result := &client.SyncResult{
		ConnectionID: account.ID,
		SyncedAt:     time.Now(),
		Success:      false,
		Details:      make(map[string]interface{}),
	}

	// Get broker integration
	integration, err := account.GetBrokerIntegration()
	if err != nil || integration == nil {
		errMsg := "no broker integration configured"
		result.Error = &errMsg
		return result, fmt.Errorf(errMsg)
	}

	if !integration.IsActive {
		errMsg := "broker integration is not active"
		result.Error = &errMsg
		return result, fmt.Errorf(errMsg)
	}

	// Select appropriate client
	var brokerClient client.BrokerClient
	switch integration.BrokerType {
	case accountDomain.BrokerTypeSSI:
		brokerClient = s.ssiClient
	case accountDomain.BrokerTypeOKX:
		brokerClient = s.okxClient
	default:
		errMsg := fmt.Sprintf("unsupported broker type: %s", integration.BrokerType)
		result.Error = &errMsg
		return result, fmt.Errorf(errMsg)
	}

	// Refresh token if needed
	if !account.IsTokenValid() && integration.RefreshToken != nil {
		authResp, err := brokerClient.RefreshToken(ctx, *integration.RefreshToken)
		if err != nil {
			errMsg := fmt.Sprintf("failed to refresh token: %v", err)
			result.Error = &errMsg
			return result, err
		}

		if err := account.RefreshAccessToken(authResp.AccessToken, authResp.ExpiresIn); err != nil {
			errMsg := fmt.Sprintf("failed to update token: %v", err)
			result.Error = &errMsg
			return result, err
		}

		// Save updated token
		if err := s.accountRepo.Update(ctx, account); err != nil {
			errMsg := fmt.Sprintf("failed to save updated token: %v", err)
			result.Error = &errMsg
			return result, err
		}

		// Update integration after token refresh
		integration, _ = account.GetBrokerIntegration()
	}

	// Get access token once
	accessToken := integration.AccessToken

	// Sync portfolio balance
	if integration.SyncBalance && accessToken != nil {
		if err := s.syncPortfolioBalance(ctx, brokerClient, account, *accessToken); err != nil {
			result.Details["balance_error"] = err.Error()
		} else {
			result.BalanceUpdated = true
		}
	}

	// Sync assets/positions
	if integration.SyncAssets && accessToken != nil {
		count, err := s.syncAssets(ctx, brokerClient, account, *accessToken)
		if err != nil {
			result.Details["assets_error"] = err.Error()
		} else {
			result.AssetsCount = count
		}
	}

	// Sync transactions
	if integration.SyncTransactions && accessToken != nil {
		count, err := s.syncTransactions(ctx, brokerClient, account, *accessToken)
		if err != nil {
			result.Details["transactions_error"] = err.Error()
		} else {
			result.TransactionsCount = count
		}
	}

	// Sync market prices
	if integration.SyncPrices {
		count, err := s.syncMarketPrices(ctx, brokerClient, account)
		if err != nil {
			result.Details["prices_error"] = err.Error()
		} else {
			result.UpdatedPricesCount = count
		}
	}

	// Determine overall success
	result.Success = result.BalanceUpdated || result.AssetsCount > 0 || result.TransactionsCount > 0 || result.UpdatedPricesCount > 0

	// Update sync status
	var syncError *string
	if !result.Success && result.Error != nil {
		syncError = result.Error
	}
	account.UpdateSyncStatus(result.Success, syncError)

	// Save account
	if err := s.accountRepo.Update(ctx, account); err != nil {
		return result, fmt.Errorf("failed to update account after sync: %w", err)
	}

	return result, nil
}

// syncPortfolioBalance syncs the portfolio balance to the account
func (s *SyncService) syncPortfolioBalance(ctx context.Context, brokerClient client.BrokerClient, account *accountDomain.Account, accessToken string) error {
	portfolio, err := brokerClient.GetPortfolio(ctx, accessToken)
	if err != nil {
		return fmt.Errorf("failed to get portfolio: %w", err)
	}

	// Update account balance
	account.CurrentBalance = portfolio.TotalValue
	availableBalance := portfolio.CashBalance
	account.AvailableBalance = &availableBalance

	return nil
}

// syncAssets syncs positions to investment_assets
func (s *SyncService) syncAssets(ctx context.Context, brokerClient client.BrokerClient, account *accountDomain.Account, accessToken string) (int, error) {
	positions, err := brokerClient.GetPositions(ctx, accessToken)
	if err != nil {
		return 0, fmt.Errorf("failed to get positions: %w", err)
	}

	count := 0
	for _, pos := range positions {
		// Check if asset already exists
		existing, err := s.assetRepo.GetBySymbol(ctx, account.UserID, pos.Symbol)
		if err != nil {
			continue
		}

		if existing != nil {
			// Update existing asset
			existing.Quantity = pos.Quantity
			existing.CurrentPrice = pos.CurrentPrice
			existing.AverageCostPerUnit = pos.AverageCostPerUnit
			existing.CalculateMetrics()

			if err := s.assetRepo.Update(ctx, existing); err != nil {
				continue
			}
		} else {
			// Create new asset
			assetType := s.mapToAssetType(pos.AssetType)
			exchange := pos.Exchange
			asset := &assetDomain.InvestmentAsset{
				UserID:             account.UserID,
				Symbol:             pos.Symbol,
				Name:               pos.Name,
				AssetType:          assetType,
				Currency:           pos.Currency,
				Quantity:           pos.Quantity,
				AverageCostPerUnit: pos.AverageCostPerUnit,
				CurrentPrice:       pos.CurrentPrice,
				TotalCost:          pos.Quantity * pos.AverageCostPerUnit,
				Status:             assetDomain.AssetStatusActive,
				AutoUpdatePrice:    true,
				ExternalID:         &pos.ExternalID,
				Exchange:           &exchange,
				Sector:             pos.Sector,
				Industry:           pos.Industry,
			}
			asset.CalculateMetrics()

			if err := s.assetRepo.Create(ctx, asset); err != nil {
				continue
			}
		}
		count++
	}

	return count, nil
}

// syncTransactions syncs recent transactions to investment_transactions
func (s *SyncService) syncTransactions(ctx context.Context, brokerClient client.BrokerClient, account *accountDomain.Account, accessToken string) (int, error) {
	// Sync last 30 days of transactions
	endDate := time.Now()
	startDate := endDate.AddDate(0, 0, -30)

	if account.LastSyncedAt != nil {
		// Only sync since last sync
		startDate = *account.LastSyncedAt
	}

	transactions, err := brokerClient.GetTransactions(ctx, accessToken, startDate, endDate)
	if err != nil {
		return 0, fmt.Errorf("failed to get transactions: %w", err)
	}

	count := 0
	for _, txn := range transactions {
		// Get asset ID
		asset, err := s.assetRepo.GetBySymbol(ctx, account.UserID, txn.Symbol)
		if err != nil || asset == nil {
			continue
		}

		// Map transaction type
		txnType := s.mapToTransactionType(txn.TransactionType)

		// Create investment transaction
		investmentTxn := &investmentTxnDomain.InvestmentTransaction{
			UserID:          account.UserID,
			AssetID:         asset.ID,
			TransactionType: txnType,
			Quantity:        txn.Quantity,
			PricePerUnit:    txn.Price,
			TotalAmount:     txn.Amount,
			Currency:        txn.Currency,
			Fees:            txn.Fee,
			Commission:      txn.Commission,
			Tax:             txn.Tax,
			TransactionDate: txn.TransactionDate,
			SettlementDate:  txn.SettlementDate,
			Status:          investmentTxnDomain.TransactionStatusCompleted,
			ExternalID:      &txn.ExternalID,
		}
		investmentTxn.CalculateTotalCost()

		if err := s.investmentTxnRepo.Create(ctx, investmentTxn); err != nil {
			// Skip if duplicate (based on external_id)
			continue
		}

		count++
	}

	return count, nil
}

// syncMarketPrices updates prices for all user's assets
func (s *SyncService) syncMarketPrices(ctx context.Context, brokerClient client.BrokerClient, account *accountDomain.Account) (int, error) {
	// Get all active assets for this user
	assets, err := s.assetRepo.GetActiveAssets(ctx, account.UserID)
	if err != nil {
		return 0, fmt.Errorf("failed to get assets: %w", err)
	}

	if len(assets) == 0 {
		return 0, nil
	}

	// Get symbols
	symbols := make([]string, 0, len(assets))
	for _, asset := range assets {
		symbols = append(symbols, asset.Symbol)
	}

	// Batch get prices
	prices, err := brokerClient.GetBatchMarketPrices(ctx, symbols)
	if err != nil {
		return 0, fmt.Errorf("failed to get prices: %w", err)
	}

	count := 0
	for _, asset := range assets {
		if price, ok := prices[asset.Symbol]; ok {
			asset.UpdatePrice(price.Price)
			if err := s.assetRepo.Update(ctx, asset); err != nil {
				continue
			}
			count++
		}
	}

	return count, nil
}

// Helper functions
func (s *SyncService) mapToAssetType(brokerType string) assetDomain.AssetType {
	switch brokerType {
	case "stock":
		return assetDomain.AssetTypeStock
	case "crypto":
		return assetDomain.AssetTypeCrypto
	case "etf":
		return assetDomain.AssetTypeETF
	default:
		return assetDomain.AssetTypeOther
	}
}

func (s *SyncService) mapToTransactionType(txnType string) investmentTxnDomain.TransactionType {
	switch txnType {
	case "buy":
		return investmentTxnDomain.TransactionTypeBuy
	case "sell":
		return investmentTxnDomain.TransactionTypeSell
	case "dividend":
		return investmentTxnDomain.TransactionTypeDividend
	default:
		return investmentTxnDomain.TransactionTypeOther
	}
}
