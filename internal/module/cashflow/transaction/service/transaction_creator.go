package service

import (
	"context"
	"time"

	"personalfinancedss/internal/module/cashflow/transaction/domain"
	"personalfinancedss/internal/module/cashflow/transaction/dto"
	"personalfinancedss/internal/shared"

	"github.com/google/uuid"
)

// CreateTransaction creates a new transaction
func (s *transactionService) CreateTransaction(ctx context.Context, userID string, req dto.CreateTransactionRequest) (*domain.Transaction, error) {
	// Parse and validate user ID
	userUUID, err := parseUUID(userID, "user_id")
	if err != nil {
		return nil, err
	}

	// Parse and validate account ID
	accountUUID, err := parseUUID(req.AccountID, "accountId")
	if err != nil {
		return nil, err
	}

	// Validate enum fields
	direction, err := validateDirection(req.Direction)
	if err != nil {
		return nil, err
	}

	instrument, err := validateInstrument(req.Instrument)
	if err != nil {
		return nil, err
	}

	source, err := validateSource(req.Source)
	if err != nil {
		return nil, err
	}

	channel, err := validateChannel(req.Channel)
	if err != nil {
		return nil, err
	}

	// Validate business rules
	if err := validateCashTransaction(instrument, source); err != nil {
		return nil, err
	}

	if err := validateBankTransaction(instrument, source, req.BankCode); err != nil {
		return nil, err
	}

	// Build transaction entity
	transaction := &domain.Transaction{
		ID:          uuid.New(),
		UserID:      userUUID,
		AccountID:   accountUUID,
		Direction:   direction,
		Instrument:  instrument,
		Source:      source,
		Channel:     channel,
		BankCode:    req.BankCode,
		ExternalID:  req.ExternalID,
		Amount:      req.Amount,
		Currency:    getDefaultCurrency(req.Currency),
		BookingDate: req.BookingDate,
		ValueDate:   getDefaultValueDate(req.ValueDate, req.BookingDate),
		Description: req.Description,
		UserNote:    req.UserNote,
		Reference:   req.Reference,
		CreatedAt:   time.Now(),
	}

	// Set running balance if provided
	if req.RunningBalance != nil {
		transaction.RunningBalance = req.RunningBalance
	}

	// Set imported timestamp for imported transactions
	if source == domain.SourceBankAPI || source == domain.SourceCsvImport || source == domain.SourceJsonImport {
		now := time.Now()
		transaction.ImportedAt = &now
	}

	// Build counterparty
	transaction.Counterparty = buildCounterparty(
		req.CounterpartyName,
		req.CounterpartyAccountNumber,
		req.CounterpartyBankName,
		req.CounterpartyType,
	)

	// Set user category ID
	if req.UserCategoryID != "" {
		categoryUUID, err := parseUserCategoryID(req.UserCategoryID)
		if err != nil {
			return nil, err
		}
		transaction.UserCategoryID = categoryUUID
	}

	// Build links
	var links domain.TransactionLinks
	if len(req.Links) > 0 {
		links = make(domain.TransactionLinks, 0, len(req.Links))
		for _, linkDTO := range req.Links {
			links = append(links, domain.TransactionLink{
				Type: domain.LinkType(linkDTO.Type),
				ID:   linkDTO.ID,
			})
		}
		transaction.Links = &links

		// Validate links before creating transaction
		if s.linkProcessor != nil {
			if err := s.linkProcessor.ValidateLinks(ctx, userUUID, []domain.TransactionLink(links)); err != nil {
				return nil, err
			}
		}
	}

	// Build metadata
	transaction.Meta = buildMetadata(req.CheckImageAvailability, nil)

	// Begin database transaction for ACID guarantee
	tx := s.db.WithContext(ctx).Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}
	}()

	// 1. Create transaction within database transaction
	if err := s.repo.CreateWithTx(tx, transaction); err != nil {
		tx.Rollback()
		return nil, shared.ErrInternal.WithError(err)
	}

	// 2. Update account balance atomically (within same transaction)
	balanceDelta := float64(req.Amount)
	if direction == domain.DirectionDebit {
		balanceDelta = -balanceDelta
	}
	if err := s.accountRepo.UpdateBalanceWithTx(tx, accountUUID.String(), balanceDelta); err != nil {
		tx.Rollback()
		return nil, shared.ErrInternal.WithError(err)
	}

	// Commit transaction (ACID: transaction + account balance update)
	if err := tx.Commit().Error; err != nil {
		return nil, shared.ErrInternal.WithError(err)
	}

	// 3. Process links after transaction is committed (side effect, not part of ACID)
	// If this fails, transaction and account balance are already committed
	if s.linkProcessor != nil && len(links) > 0 {
		if err := s.linkProcessor.ProcessLinks(ctx, userUUID, req.Amount, direction, links); err != nil {
			// Log the error but don't fail - transaction and balance are already committed
			// TODO: Consider implementing compensation/rollback for link processing failures
		}
	}

	// Retrieve the created transaction
	created, err := s.repo.GetByID(ctx, transaction.ID)
	if err != nil {
		return nil, shared.ErrInternal.WithError(err)
	}

	return created, nil
}
