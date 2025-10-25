package service

import (
	"encoding/json"
	"strings"
	"time"

	"personalfinancedss/internal/module/cashflow/transaction/domain"
	"personalfinancedss/internal/shared"

	"github.com/google/uuid"
)

// Validation helpers for enum types

// validateDirection validates and returns a Direction enum value
func validateDirection(str string) (domain.Direction, error) {
	direction := domain.Direction(str)
	switch direction {
	case domain.DirectionDebit, domain.DirectionCredit:
		return direction, nil
	default:
		return "", shared.ErrBadRequest.WithDetails("field", "direction").WithDetails("reason", "invalid value: must be DEBIT or CREDIT")
	}
}

// validateInstrument validates and returns an Instrument enum value
func validateInstrument(str string) (domain.Instrument, error) {
	instrument := domain.Instrument(str)
	switch instrument {
	case domain.InstrumentCash, domain.InstrumentBankAccount,
		domain.InstrumentDebitCard, domain.InstrumentCreditCard,
		domain.InstrumentEWallet, domain.InstrumentCrypto,
		domain.InstrumentUnknown:
		return instrument, nil
	default:
		return "", shared.ErrBadRequest.WithDetails("field", "instrument").WithDetails("reason", "invalid instrument type")
	}
}

// validateSource validates and returns a TransactionSource enum value
func validateSource(str string) (domain.TransactionSource, error) {
	source := domain.TransactionSource(str)
	switch source {
	case domain.SourceBankAPI, domain.SourceCsvImport,
		domain.SourceJsonImport, domain.SourceManual:
		return source, nil
	default:
		return "", shared.ErrBadRequest.WithDetails("field", "source").WithDetails("reason", "invalid source type")
	}
}

// validateChannel validates and returns a Channel enum value
func validateChannel(str string) (domain.Channel, error) {
	if str == "" {
		return domain.ChannelUnknown, nil
	}

	channel := domain.Channel(str)
	switch channel {
	case domain.ChannelMobile, domain.ChannelWeb,
		domain.ChannelATM, domain.ChannelPOS,
		domain.ChannelUnknown:
		return channel, nil
	default:
		return "", shared.ErrBadRequest.WithDetails("field", "channel").WithDetails("reason", "invalid channel type")
	}
}

// Default value helpers

// getDefaultCurrency returns default currency if empty
func getDefaultCurrency(currency string) string {
	if currency == "" {
		return "VND"
	}
	return strings.ToUpper(currency)
}

// getDefaultValueDate returns value date or booking date if not provided
func getDefaultValueDate(valueDate *time.Time, bookingDate time.Time) time.Time {
	if valueDate != nil {
		return *valueDate
	}
	return bookingDate
}

// Business rule validation helpers

// validateCashTransaction validates business rules for cash transactions
func validateCashTransaction(instrument domain.Instrument, source domain.TransactionSource) error {
	if instrument == domain.InstrumentCash && source != domain.SourceManual {
		return shared.ErrBadRequest.WithDetails("reason", "cash transactions must have source=MANUAL")
	}
	return nil
}

// validateBankTransaction validates business rules for bank transactions
func validateBankTransaction(instrument domain.Instrument, source domain.TransactionSource, bankCode string) error {
	if instrument == domain.InstrumentBankAccount {
		if source != domain.SourceBankAPI && source != domain.SourceCsvImport && source != domain.SourceJsonImport {
			return shared.ErrBadRequest.WithDetails("reason", "bank transactions should be from BANK_API or imports")
		}
		if bankCode == "" && source == domain.SourceBankAPI {
			return shared.ErrBadRequest.WithDetails("reason", "bank transactions from BANK_API should have bankCode")
		}
	}
	return nil
}

// UUID parsing helpers

// parseUUID parses a UUID string and returns error with field context
func parseUUID(uuidStr string, fieldName string) (uuid.UUID, error) {
	parsed, err := uuid.Parse(uuidStr)
	if err != nil {
		return uuid.UUID{}, shared.ErrBadRequest.WithDetails("field", fieldName).WithDetails("reason", "invalid UUID format")
	}
	return parsed, nil
}

// parseOptionalUUID parses an optional UUID string
func parseOptionalUUID(uuidStr *string, fieldName string) (*uuid.UUID, error) {
	if uuidStr == nil || *uuidStr == "" {
		return nil, nil
	}

	parsed, err := uuid.Parse(*uuidStr)
	if err != nil {
		return nil, shared.ErrBadRequest.WithDetails("field", fieldName).WithDetails("reason", "invalid UUID format")
	}
	return &parsed, nil
}

// Nested structure builders

// buildCounterparty builds a Counterparty object from individual fields
func buildCounterparty(name, accountNumber, bankName, cpType string) *domain.Counterparty {
	// Only create if at least one field is provided
	if name == "" && accountNumber == "" && bankName == "" {
		return nil
	}

	return &domain.Counterparty{
		Name:          name,
		AccountNumber: accountNumber,
		BankName:      bankName,
		Type:          cpType,
	}
}

// buildClassification builds a Classification object from individual fields
func buildClassification(systemCategory, userCategoryID string, isTransfer, isRefund bool, tags []string) *domain.Classification {
	// Only create if at least one meaningful field is provided
	if systemCategory == "" && userCategoryID == "" && !isTransfer && !isRefund && len(tags) == 0 {
		return nil
	}

	return &domain.Classification{
		SystemCategory: systemCategory,
		UserCategoryID: userCategoryID,
		IsTransfer:     isTransfer,
		IsRefund:       isRefund,
		Tags:           tags,
	}
}

// buildMetadata builds a TransactionMeta object
func buildMetadata(checkImageAvailability string, rawData interface{}) *domain.TransactionMeta {
	if checkImageAvailability == "" && rawData == nil {
		return nil
	}

	meta := &domain.TransactionMeta{
		CheckImageAvailability: checkImageAvailability,
	}

	// Convert raw data to JSON
	if rawData != nil {
		if jsonData, err := json.Marshal(rawData); err == nil {
			meta.Raw = json.RawMessage(jsonData)
		}
	}

	return meta
}

// String utilities

// normalizeNullableString returns nil for empty strings, otherwise returns the pointer
func normalizeNullableString(s *string) *string {
	if s == nil || *s == "" {
		return nil
	}
	return s
}

// Timestamp utilities

// ensureTimestamps sets CreatedAt if not already set
func ensureTimestamps(t *domain.Transaction) {
	now := time.Now()
	if t.CreatedAt.IsZero() {
		t.CreatedAt = now
	}
}
