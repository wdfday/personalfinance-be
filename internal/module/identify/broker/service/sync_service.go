package service

import (
	"context"
	"fmt"
	accountDomain "personalfinancedss/internal/module/cashflow/account/domain"
	accountRepo "personalfinancedss/internal/module/cashflow/account/repository"
	transactionDomain "personalfinancedss/internal/module/cashflow/transaction/domain"
	transactionRepo "personalfinancedss/internal/module/cashflow/transaction/repository"
	"personalfinancedss/internal/module/identify/broker/client"
	"personalfinancedss/internal/module/identify/broker/client/okx"
	"personalfinancedss/internal/module/identify/broker/client/sepay"
	"personalfinancedss/internal/module/identify/broker/client/ssi"
	"personalfinancedss/internal/module/identify/broker/domain"
	"personalfinancedss/internal/module/identify/broker/repository"
	internalService "personalfinancedss/internal/service"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// SyncService handles syncing data from broker connections
type SyncService struct {
	brokerRepo        repository.BrokerConnectionRepository
	accountRepo       accountRepo.Repository
	transactionRepo   transactionRepo.Repository
	encryptionService *internalService.EncryptionService
	ssiClient         *ssi.SSIClient
	okxClient         *okx.OKXClient
	sepayClient       *sepay.Client
	logger            *zap.Logger
}

// NewSyncService creates a new sync service
func NewSyncService(
	brokerRepo repository.BrokerConnectionRepository,
	accountRepo accountRepo.Repository,
	transactionRepo transactionRepo.Repository,
	encryptionService *internalService.EncryptionService,
	ssiClient *ssi.SSIClient,
	okxClient *okx.OKXClient,
	sepayClient *sepay.Client,
	logger *zap.Logger,
) *SyncService {
	return &SyncService{
		brokerRepo:        brokerRepo,
		accountRepo:       accountRepo,
		transactionRepo:   transactionRepo,
		encryptionService: encryptionService,
		ssiClient:         ssiClient,
		okxClient:         okxClient,
		sepayClient:       sepayClient,
		logger:            logger.Named("broker.sync"),
	}
}

// SyncBrokerConnection syncs data from a broker connection based on broker type
func (s *SyncService) SyncBrokerConnection(ctx context.Context, connection *domain.BrokerConnection) (*SyncResult, error) {
	result := &SyncResult{
		SyncedAt: time.Now(),
		Success:  false,
		Details:  make(map[string]interface{}),
	}

	s.logger.Info("🔄 Starting sync for broker connection",
		zap.String("connection_id", connection.ID.String()),
		zap.String("broker_type", string(connection.BrokerType)),
		zap.String("broker_name", connection.BrokerName),
	)

	// Get broker client
	brokerClient, err := s.getBrokerClient(connection.BrokerType)
	if err != nil {
		errMsg := fmt.Sprintf("failed to get broker client: %v", err)
		result.Error = &errMsg
		return result, err
	}

	// Decrypt access token
	accessToken, err := s.getDecryptedAccessToken(connection)
	if err != nil {
		errMsg := fmt.Sprintf("failed to decrypt access token: %v", err)
		result.Error = &errMsg
		return result, err
	}

	// Sync based on broker type
	switch connection.BrokerType {
	case domain.BrokerTypeSSI, domain.BrokerTypeOKX:
		// Investment brokers: sync balance only with a single linked account
		account, err := s.findOrCreateLinkedAccount(ctx, connection)
		if err != nil {
			errMsg := fmt.Sprintf("failed to get linked account: %v", err)
			result.Error = &errMsg
			return result, err
		}

		if connection.SyncBalance {
			err = s.syncAccountBalance(ctx, brokerClient, accessToken, account)
			if err != nil {
				result.Details["balance_error"] = err.Error()
			} else {
				result.BalanceUpdated = true
			}
		}

	case domain.BrokerTypeSePay:
		// Banking: sync multiple accounts + transactions
		bankingClient, ok := brokerClient.(client.BankingBrokerClient)
		if !ok {
			errMsg := "broker client does not support banking operations"
			result.Error = &errMsg
			return result, fmt.Errorf(errMsg)
		}

		// Get all bank accounts from SePay (accessToken is actually the API key for SePay)
		bankAccounts, err := bankingClient.GetBankAccounts(ctx, accessToken)
		if err != nil {
			errMsg := fmt.Sprintf("failed to get bank accounts: %v", err)
			result.Error = &errMsg
			return result, err
		}

		s.logger.Info("Found bank accounts from SePay",
			zap.Int("count", len(bankAccounts)),
			zap.String("connection_id", connection.ID.String()),
		)

		// For initial validation (TotalSyncs == 0), only verify API connection
		// Don't create accounts or sync transactions until connection is saved
		if connection.TotalSyncs == 0 {
			s.logger.Info("Initial validation: API connection verified",
				zap.Int("bank_accounts_found", len(bankAccounts)),
			)
			result.Success = len(bankAccounts) > 0
			break
		}

		// Sync each bank account (only for established connections)
		for _, bankAcc := range bankAccounts {
			if !bankAcc.IsActive {
				s.logger.Debug("Skipping inactive bank account",
					zap.String("account_number", maskAccountNumber(bankAcc.AccountNumber)),
				)
				continue
			}

			// Find or create account for this bank account
			linkedAccount, err := s.findOrCreateAccountForBankAccount(ctx, connection, bankAcc)
			if err != nil {
				s.logger.Warn("Failed to get linked account for bank account",
					zap.String("account_number", maskAccountNumber(bankAcc.AccountNumber)),
					zap.Error(err),
				)
				continue
			}

			// Sync balance from bank account
			if connection.SyncBalance {
				linkedAccount.CurrentBalance = bankAcc.Balance
				now := time.Now()
				linkedAccount.LastSyncedAt = &now
				syncStatus := accountDomain.SyncStatusActive
				linkedAccount.SyncStatus = &syncStatus

				if err := s.accountRepo.Update(ctx, linkedAccount); err != nil {
					s.logger.Warn("Failed to update account balance",
						zap.String("account_id", linkedAccount.ID.String()),
						zap.Error(err),
					)
				} else {
					result.BalanceUpdated = true
					s.logger.Debug("Account balance updated",
						zap.String("account_id", linkedAccount.ID.String()),
						zap.Float64("balance", linkedAccount.CurrentBalance),
					)
				}
			}

			// Sync transactions for this specific account
			if connection.SyncTransactions {
				count, err := s.syncAccountTransactions(ctx, bankingClient, accessToken, connection, linkedAccount, bankAcc.AccountNumber)
				if err != nil {
					s.logger.Warn("Failed to sync transactions for account",
						zap.String("account_id", linkedAccount.ID.String()),
						zap.Error(err),
					)
				} else {
					result.TransactionsCount += count
				}
			}
		}
	}

	// Determine overall success
	result.Success = result.BalanceUpdated || result.TransactionsCount > 0

	// Update sync status on connection
	var syncError *string
	if !result.Success && len(result.Details) > 0 {
		errStr := fmt.Sprintf("%v", result.Details)
		syncError = &errStr
	}
	connection.UpdateSyncStatus(result.Success, syncError)

	// Save connection state only if connection already exists in DB
	// (skip update for initial validation sync before connection is saved)
	if connection.TotalSyncs > 0 {
		if err := s.brokerRepo.Update(ctx, connection); err != nil {
			s.logger.Error("Failed to update connection after sync", zap.Error(err))
		}
	}

	s.logger.Info("✅ Sync completed",
		zap.String("connection_id", connection.ID.String()),
		zap.Bool("success", result.Success),
		zap.Bool("balance_updated", result.BalanceUpdated),
		zap.Int("transactions_synced", result.TransactionsCount),
	)

	return result, nil
}

// syncAccountBalance syncs the account balance from broker
func (s *SyncService) syncAccountBalance(ctx context.Context, brokerClient client.BrokerClient, accessToken string, account *accountDomain.Account) error {
	portfolio, err := brokerClient.GetPortfolio(ctx, accessToken)
	if err != nil {
		return fmt.Errorf("failed to get portfolio: %w", err)
	}

	account.CurrentBalance = portfolio.TotalValue
	if portfolio.CashBalance > 0 {
		account.AvailableBalance = &portfolio.CashBalance
	}

	now := time.Now()
	account.LastSyncedAt = &now
	syncStatus := accountDomain.SyncStatusActive
	account.SyncStatus = &syncStatus

	if err := s.accountRepo.Update(ctx, account); err != nil {
		return fmt.Errorf("failed to update account balance: %w", err)
	}

	s.logger.Debug("Account balance updated",
		zap.String("account_id", account.ID.String()),
		zap.Float64("balance", account.CurrentBalance),
	)

	return nil
}

// syncBankTransactions syncs transactions from SePay banking API
func (s *SyncService) syncBankTransactions(ctx context.Context, brokerClient client.BrokerClient, accessToken string, connection *domain.BrokerConnection, account *accountDomain.Account) (int, error) {
	// Determine date range
	endDate := time.Now()
	startDate := endDate.AddDate(0, 0, -30) // Default: last 30 days

	if connection.LastSyncAt != nil {
		// Only sync since last sync
		startDate = *connection.LastSyncAt
	}

	s.logger.Debug("Fetching transactions",
		zap.Time("start_date", startDate),
		zap.Time("end_date", endDate),
	)

	// Get transactions from broker
	brokerTxns, err := brokerClient.GetTransactions(ctx, accessToken, startDate, endDate)
	if err != nil {
		return 0, fmt.Errorf("failed to get transactions: %w", err)
	}

	s.logger.Debug("Fetched transactions from broker", zap.Int("count", len(brokerTxns)))

	count := 0
	for _, txn := range brokerTxns {
		// Check for duplicate using external ID
		existing, _ := s.transactionRepo.GetByExternalID(ctx, account.UserID, txn.ExternalID)
		if existing != nil {
			s.logger.Debug("Skipping duplicate transaction", zap.String("external_id", txn.ExternalID))
			continue
		}

		// Map broker transaction to domain transaction
		direction := transactionDomain.DirectionCredit // deposit = money in
		if txn.TransactionType == "withdrawal" {
			direction = transactionDomain.DirectionDebit
		}

		// Convert amount to int64 (VND doesn't have decimals)
		amountInt := int64(txn.Amount)

		transaction := &transactionDomain.Transaction{
			ID:          uuid.New(),
			UserID:      account.UserID,
			AccountID:   account.ID,
			BankCode:    txn.Symbol, // Bank brand name
			Source:      transactionDomain.SourceBankAPI,
			ExternalID:  txn.ExternalID,
			Direction:   direction,
			Channel:     transactionDomain.ChannelUnknown,
			Instrument:  transactionDomain.InstrumentBankAccount,
			BookingDate: txn.TransactionDate,
			ValueDate:   txn.TransactionDate,
			Amount:      amountInt,
			Currency:    txn.Currency,
			Description: txn.Notes,
		}

		if err := s.transactionRepo.Create(ctx, transaction); err != nil {
			s.logger.Warn("Failed to create transaction",
				zap.String("external_id", txn.ExternalID),
				zap.Error(err),
			)
			continue
		}

		count++
	}

	s.logger.Info("Transactions synced",
		zap.Int("new_transactions", count),
		zap.Int("total_fetched", len(brokerTxns)),
	)

	return count, nil
}

// findOrCreateLinkedAccount finds or creates an account linked to the broker connection
func (s *SyncService) findOrCreateLinkedAccount(ctx context.Context, connection *domain.BrokerConnection) (*accountDomain.Account, error) {
	// Find existing account linked to this broker connection
	accounts, err := s.accountRepo.ListByUserID(ctx, connection.UserID.String(), accountDomain.ListAccountsFilter{})
	if err != nil {
		return nil, fmt.Errorf("failed to list user accounts: %w", err)
	}

	// Check if any account is linked to this connection
	for i := range accounts {
		if accounts[i].BrokerConnectionID != nil && *accounts[i].BrokerConnectionID == connection.ID {
			return &accounts[i], nil
		}
	}

	// No linked account found, create one
	account, err := s.createLinkedAccount(ctx, connection)
	if err != nil {
		return nil, err
	}

	return account, nil
}

// createLinkedAccount creates a new account linked to the broker connection
func (s *SyncService) createLinkedAccount(ctx context.Context, connection *domain.BrokerConnection) (*accountDomain.Account, error) {
	// Determine account type based on broker type
	var accountType accountDomain.AccountType
	switch connection.BrokerType {
	case domain.BrokerTypeSSI:
		accountType = accountDomain.AccountTypeInvestment
	case domain.BrokerTypeOKX:
		accountType = accountDomain.AccountTypeCryptoWallet
	case domain.BrokerTypeSePay:
		accountType = accountDomain.AccountTypeBank
	default:
		accountType = accountDomain.AccountTypeBank
	}

	// Build account name
	accountName := connection.BrokerName
	if connection.ExternalAccountNumber != nil {
		accountName = fmt.Sprintf("%s - %s", connection.BrokerName, *connection.ExternalAccountNumber)
	}

	account := &accountDomain.Account{
		ID:                 uuid.New(),
		UserID:             connection.UserID,
		AccountName:        accountName,
		AccountType:        accountType,
		Currency:           accountDomain.CurrencyVND,
		IsActive:           true,
		IsAutoSync:         true,
		IncludeInNetWorth:  true,
		BrokerConnectionID: &connection.ID,
	}

	if err := s.accountRepo.Create(ctx, account); err != nil {
		return nil, fmt.Errorf("failed to create linked account: %w", err)
	}

	s.logger.Info("Created linked account for broker connection",
		zap.String("account_id", account.ID.String()),
		zap.String("connection_id", connection.ID.String()),
		zap.String("account_name", account.AccountName),
	)

	return account, nil
}

// getDecryptedAccessToken decrypts the access token from connection
func (s *SyncService) getDecryptedAccessToken(connection *domain.BrokerConnection) (string, error) {
	if connection.AccessToken == nil {
		return "", fmt.Errorf("no access token available")
	}

	accessToken, err := s.encryptionService.Decrypt(*connection.AccessToken)
	if err != nil {
		return "", fmt.Errorf("failed to decrypt access token: %w", err)
	}

	return accessToken, nil
}

// getBrokerClient returns the appropriate broker client
func (s *SyncService) getBrokerClient(brokerType domain.BrokerType) (client.BrokerClient, error) {
	switch brokerType {
	case domain.BrokerTypeSSI:
		return s.ssiClient, nil
	case domain.BrokerTypeOKX:
		return s.okxClient, nil
	case domain.BrokerTypeSePay:
		return s.sepayClient, nil
	default:
		return nil, fmt.Errorf("unsupported broker type: %s", brokerType)
	}
}

// maskAccountNumber masks account number showing only last 4 digits
func maskAccountNumber(accountNumber string) string {
	if len(accountNumber) <= 4 {
		return "****"
	}
	return "****" + accountNumber[len(accountNumber)-4:]
}

// findOrCreateAccountForBankAccount finds or creates an account for a specific bank account
func (s *SyncService) findOrCreateAccountForBankAccount(
	ctx context.Context,
	connection *domain.BrokerConnection,
	bankAcc client.BankAccount,
) (*accountDomain.Account, error) {
	maskedNumber := maskAccountNumber(bankAcc.AccountNumber)

	// Find existing account linked to this connection with matching masked number
	accounts, err := s.accountRepo.ListByUserID(ctx, connection.UserID.String(), accountDomain.ListAccountsFilter{})
	if err != nil {
		return nil, fmt.Errorf("failed to list user accounts: %w", err)
	}

	for i := range accounts {
		if accounts[i].BrokerConnectionID != nil &&
			*accounts[i].BrokerConnectionID == connection.ID &&
			accounts[i].AccountNumberMasked != nil &&
			*accounts[i].AccountNumberMasked == maskedNumber {
			return &accounts[i], nil
		}
	}

	// No matching account found, create one
	accountName := fmt.Sprintf("%s - %s", bankAcc.BankName, maskedNumber)
	institutionName := bankAcc.BankName

	account := &accountDomain.Account{
		ID:                  uuid.New(),
		UserID:              connection.UserID,
		AccountName:         accountName,
		AccountType:         accountDomain.AccountTypeBank,
		InstitutionName:     &institutionName,
		Currency:            accountDomain.CurrencyVND,
		CurrentBalance:      bankAcc.Balance,
		IsActive:            true,
		IsAutoSync:          true,
		IncludeInNetWorth:   true,
		BrokerConnectionID:  &connection.ID,
		AccountNumberMasked: &maskedNumber,
	}

	if err := s.accountRepo.Create(ctx, account); err != nil {
		return nil, fmt.Errorf("failed to create linked account: %w", err)
	}

	s.logger.Info("Created linked account for bank account",
		zap.String("account_id", account.ID.String()),
		zap.String("connection_id", connection.ID.String()),
		zap.String("account_name", account.AccountName),
		zap.String("bank_name", bankAcc.BankName),
	)

	return account, nil
}

// syncAccountTransactions syncs transactions for a specific bank account
func (s *SyncService) syncAccountTransactions(
	ctx context.Context,
	bankingClient client.BankingBrokerClient,
	accessToken string,
	connection *domain.BrokerConnection,
	account *accountDomain.Account,
	accountNumber string,
) (int, error) {
	// Determine date range
	endDate := time.Now()
	startDate := endDate.AddDate(0, 0, -30) // Default: last 30 days

	if connection.LastSyncAt != nil {
		// Only sync since last sync
		startDate = *connection.LastSyncAt
	}

	s.logger.Debug("Fetching transactions for account",
		zap.String("account_id", account.ID.String()),
		zap.String("account_number", maskAccountNumber(accountNumber)),
		zap.Time("start_date", startDate),
		zap.Time("end_date", endDate),
	)

	// Get transactions from broker for this specific account
	brokerTxns, err := bankingClient.GetAccountTransactions(ctx, accessToken, accountNumber, startDate, endDate)
	if err != nil {
		return 0, fmt.Errorf("failed to get transactions: %w", err)
	}

	s.logger.Debug("Fetched transactions from broker", zap.Int("count", len(brokerTxns)))

	count := 0
	for _, txn := range brokerTxns {
		// Check for duplicate using external ID
		existing, _ := s.transactionRepo.GetByExternalID(ctx, account.UserID, txn.ExternalID)
		if existing != nil {
			continue
		}

		// Map broker transaction to domain transaction
		direction := transactionDomain.DirectionCredit // deposit = money in
		if txn.TransactionType == "withdrawal" {
			direction = transactionDomain.DirectionDebit
		}

		// Convert amount to int64 (VND doesn't have decimals)
		amountInt := int64(txn.Amount)
		runningBalance := int64(txn.RunningBalance)

		transaction := &transactionDomain.Transaction{
			ID:             uuid.New(),
			UserID:         account.UserID,
			AccountID:      account.ID,
			BankCode:       txn.Symbol,
			Source:         transactionDomain.SourceBankAPI,
			ExternalID:     txn.ExternalID,
			Direction:      direction,
			Channel:        transactionDomain.ChannelUnknown,
			Instrument:     transactionDomain.InstrumentBankAccount,
			BookingDate:    txn.TransactionDate,
			ValueDate:      txn.TransactionDate,
			Amount:         amountInt,
			Currency:       txn.Currency,
			Description:    txn.Notes,
			Reference:      txn.ReferenceCode,
			RunningBalance: &runningBalance,
		}

		if err := s.transactionRepo.Create(ctx, transaction); err != nil {
			s.logger.Warn("Failed to create transaction",
				zap.String("external_id", txn.ExternalID),
				zap.Error(err),
			)
			continue
		}

		count++
	}

	s.logger.Info("Transactions synced for account",
		zap.String("account_id", account.ID.String()),
		zap.Int("new_transactions", count),
		zap.Int("total_fetched", len(brokerTxns)),
	)

	return count, nil
}
