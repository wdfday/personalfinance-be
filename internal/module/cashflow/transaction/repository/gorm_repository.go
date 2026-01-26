package repository

import (
	"context"
	"fmt"
	"strings"
	"time"

	"personalfinancedss/internal/module/cashflow/transaction/domain"
	"personalfinancedss/internal/module/cashflow/transaction/dto"
	"personalfinancedss/internal/shared"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type gormRepository struct {
	db *gorm.DB
}

// NewGormRepository creates a new GORM-based transaction repository
func NewGormRepository(db *gorm.DB) Repository {
	return &gormRepository{db: db}
}

// Create creates a new transaction
func (r *gormRepository) Create(ctx context.Context, transaction *domain.Transaction) error {
	if err := r.db.WithContext(ctx).Create(transaction).Error; err != nil {
		return err
	}
	return nil
}

// GetByID retrieves a transaction by ID
func (r *gormRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Transaction, error) {
	var transaction domain.Transaction
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&transaction).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, shared.ErrNotFound
		}
		return nil, err
	}
	return &transaction, nil
}

// GetByUserID retrieves a transaction by ID and user ID
func (r *gormRepository) GetByUserID(ctx context.Context, id, userID uuid.UUID) (*domain.Transaction, error) {
	var transaction domain.Transaction
	if err := r.db.WithContext(ctx).
		Where("id = ? AND user_id = ?", id, userID).
		First(&transaction).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, shared.ErrNotFound
		}
		return nil, err
	}
	return &transaction, nil
}

// GetByExternalID retrieves a transaction by external ID for a user (for import deduplication)
func (r *gormRepository) GetByExternalID(ctx context.Context, userID uuid.UUID, externalID string) (*domain.Transaction, error) {
	if externalID == "" {
		return nil, shared.ErrNotFound
	}

	var transaction domain.Transaction
	if err := r.db.WithContext(ctx).
		Where("user_id = ? AND external_id = ?", userID, externalID).
		First(&transaction).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, shared.ErrNotFound
		}
		return nil, err
	}
	return &transaction, nil
}

// List retrieves transactions with filters and pagination
func (r *gormRepository) List(ctx context.Context, userID uuid.UUID, query dto.ListTransactionsQuery) ([]*domain.Transaction, int64, error) {
	var transactions []*domain.Transaction
	var total int64

	db := r.db.WithContext(ctx).Model(&domain.Transaction{}).Where("user_id = ?", userID)

	// Apply filters
	db = r.applyFilters(db, query)

	// Get total count
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Apply pagination
	page := query.Page
	if page < 1 {
		page = 1
	}
	pageSize := query.PageSize
	if pageSize < 1 {
		pageSize = 20
	} else if pageSize > 100 {
		pageSize = 100
	}
	offset := (page - 1) * pageSize

	// Apply sorting (default to booking_date DESC)
	sortBy := query.SortBy
	if sortBy == "" {
		sortBy = "booking_date"
	}
	sortOrder := strings.ToUpper(query.SortOrder)
	if sortOrder != "ASC" && sortOrder != "DESC" {
		sortOrder = "DESC"
	}
	orderClause := fmt.Sprintf("%s %s", sortBy, sortOrder)

	// Execute query
	if err := db.Order(orderClause).Limit(pageSize).Offset(offset).Find(&transactions).Error; err != nil {
		return nil, 0, err
	}

	return transactions, total, nil
}

// applyFilters applies query filters to the database query
func (r *gormRepository) applyFilters(db *gorm.DB, query dto.ListTransactionsQuery) *gorm.DB {
	// Account filter
	if query.AccountID != nil {
		accountUUID, err := uuid.Parse(*query.AccountID)
		if err == nil {
			db = db.Where("account_id = ?", accountUUID)
		}
	}

	// Transaction type filters (new model)
	if query.Direction != nil {
		db = db.Where("direction = ?", *query.Direction)
	}

	if query.Instrument != nil {
		db = db.Where("instrument = ?", *query.Instrument)
	}

	if query.Source != nil {
		db = db.Where("source = ?", *query.Source)
	}

	// Bank filters
	if query.BankCode != nil && *query.BankCode != "" {
		db = db.Where("bank_code = ?", *query.BankCode)
	}

	// Date range filters (booking date)
	if query.StartBookingDate != nil {
		db = db.Where("booking_date >= ?", *query.StartBookingDate)
	}

	if query.EndBookingDate != nil {
		db = db.Where("booking_date <= ?", *query.EndBookingDate)
	}

	// Date range filters (value date)
	if query.StartValueDate != nil {
		db = db.Where("value_date >= ?", *query.StartValueDate)
	}

	if query.EndValueDate != nil {
		db = db.Where("value_date <= ?", *query.EndValueDate)
	}

	// Amount range filters
	if query.MinAmount != nil {
		db = db.Where("amount >= ?", *query.MinAmount)
	}

	if query.MaxAmount != nil {
		db = db.Where("amount <= ?", *query.MaxAmount)
	}

	// Classification filters
	if query.UserCategoryID != nil {
		categoryUUID, err := uuid.Parse(*query.UserCategoryID)
		if err == nil {
			db = db.Where("user_category_id = ?", categoryUUID)
		}
	}

	// Text search (description, userNote, counterparty name)
	if query.Search != nil && *query.Search != "" {
		searchPattern := "%" + *query.Search + "%"
		db = db.Where("description ILIKE ? OR user_note ILIKE ? OR counterparty->>'name' ILIKE ?",
			searchPattern, searchPattern, searchPattern)
	}

	return db
}

// Update updates a transaction
func (r *gormRepository) Update(ctx context.Context, transaction *domain.Transaction) error {
	if err := r.db.WithContext(ctx).Save(transaction).Error; err != nil {
		return err
	}
	return nil
}

// UpdateColumns updates specific columns of a transaction
func (r *gormRepository) UpdateColumns(ctx context.Context, id uuid.UUID, columns map[string]interface{}) error {
	result := r.db.WithContext(ctx).Model(&domain.Transaction{}).Where("id = ?", id).Updates(columns)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return shared.ErrNotFound
	}
	return nil
}

// Delete soft deletes a transaction
func (r *gormRepository) Delete(ctx context.Context, id uuid.UUID) error {
	result := r.db.WithContext(ctx).Delete(&domain.Transaction{}, "id = ?", id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return shared.ErrNotFound
	}
	return nil
}

// GetAccountBalance calculates the current balance for an account based on transactions
func (r *gormRepository) GetAccountBalance(ctx context.Context, accountID uuid.UUID) (int64, error) {
	var result struct {
		Balance int64
	}

	// Calculate balance: sum of CREDIT transactions minus sum of DEBIT transactions
	query := `
		SELECT
			COALESCE(SUM(CASE
				WHEN direction = 'CREDIT' THEN amount
				WHEN direction = 'DEBIT' THEN -amount
				ELSE 0
			END), 0) as balance
		FROM transactions
		WHERE account_id = ? AND deleted_at IS NULL
	`

	if err := r.db.WithContext(ctx).Raw(query, accountID).Scan(&result).Error; err != nil {
		return 0, err
	}

	return result.Balance, nil
}

// GetTransactionsByDateRange gets transactions within a date range (using booking_date)
func (r *gormRepository) GetTransactionsByDateRange(ctx context.Context, userID uuid.UUID, accountID *uuid.UUID, startDate, endDate time.Time) ([]*domain.Transaction, error) {
	var transactions []*domain.Transaction

	db := r.db.WithContext(ctx).
		Where("user_id = ? AND booking_date >= ? AND booking_date <= ?", userID, startDate, endDate)

	if accountID != nil {
		db = db.Where("account_id = ?", *accountID)
	}

	if err := db.Order("booking_date DESC").Find(&transactions).Error; err != nil {
		return nil, err
	}

	return transactions, nil
}

// GetSummary calculates transaction summary for given filters
func (r *gormRepository) GetSummary(ctx context.Context, userID uuid.UUID, query dto.ListTransactionsQuery) (*dto.TransactionSummary, error) {
	summary := &dto.TransactionSummary{
		ByInstrument: make(map[string]dto.InstrumentSummary),
		BySource:     make(map[string]dto.SourceSummary),
	}

	db := r.db.WithContext(ctx).Model(&domain.Transaction{}).Where("user_id = ?", userID)

	// Apply same filters as List
	db = r.applyFilters(db, query)

	// Calculate overall summary by direction
	type directionResult struct {
		Direction string
		Total     int64
		Count     int64
	}

	var dirResults []directionResult
	if err := db.Select("direction, SUM(amount) as total, COUNT(*) as count").
		Group("direction").
		Scan(&dirResults).Error; err != nil {
		return nil, err
	}

	for _, r := range dirResults {
		summary.Count += r.Count
		switch r.Direction {
		case string(domain.DirectionCredit):
			summary.TotalCredit = r.Total
		case string(domain.DirectionDebit):
			summary.TotalDebit = r.Total
		}
	}

	summary.NetAmount = summary.TotalCredit - summary.TotalDebit

	// Calculate breakdown by instrument
	type instrumentResult struct {
		Instrument string
		Direction  string
		Total      int64
		Count      int64
	}

	var instResults []instrumentResult
	if err := db.Select("instrument, direction, SUM(amount) as total, COUNT(*) as count").
		Group("instrument, direction").
		Scan(&instResults).Error; err == nil {
		for _, r := range instResults {
			s, ok := summary.ByInstrument[r.Instrument]
			if !ok {
				s = dto.InstrumentSummary{}
			}
			s.Count += r.Count
			if r.Direction == string(domain.DirectionCredit) {
				s.Credit = r.Total
			} else {
				s.Debit = r.Total
			}
			summary.ByInstrument[r.Instrument] = s
		}
	}

	// Calculate breakdown by source
	type sourceResult struct {
		Source    string
		Direction string
		Total     int64
		Count     int64
	}

	var srcResults []sourceResult
	if err := db.Select("source, direction, SUM(amount) as total, COUNT(*) as count").
		Group("source, direction").
		Scan(&srcResults).Error; err == nil {
		for _, r := range srcResults {
			s, ok := summary.BySource[r.Source]
			if !ok {
				s = dto.SourceSummary{}
			}
			s.Count += r.Count
			if r.Direction == string(domain.DirectionCredit) {
				s.Credit = r.Total
			} else {
				s.Debit = r.Total
			}
			summary.BySource[r.Source] = s
		}
	}

	return summary, nil
}

// GetRecurringTransactions gets all manual recurring transaction templates (for future use)
func (r *gormRepository) GetRecurringTransactions(ctx context.Context, userID uuid.UUID) ([]*domain.Transaction, error) {
	var transactions []*domain.Transaction

	// For now, this could be used for manual recurring patterns
	// In the new model, bank transactions don't have is_recurring flag
	if err := r.db.WithContext(ctx).
		Where("user_id = ? AND source = ?", userID, domain.SourceManual).
		Order("booking_date DESC").
		Find(&transactions).Error; err != nil {
		return nil, err
	}

	return transactions, nil
}
