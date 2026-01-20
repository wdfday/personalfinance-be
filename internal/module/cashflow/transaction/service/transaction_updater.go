package service

import (
	"context"

	"personalfinancedss/internal/module/cashflow/transaction/domain"
	"personalfinancedss/internal/module/cashflow/transaction/dto"
	"personalfinancedss/internal/shared"
)

// UpdateTransaction updates an existing transaction
func (s *transactionService) UpdateTransaction(ctx context.Context, userID string, transactionID string, req dto.UpdateTransactionRequest) (*domain.Transaction, error) {
	// Parse user ID
	userUUID, err := parseUUID(userID, "user_id")
	if err != nil {
		return nil, err
	}

	// Parse transaction ID
	transactionUUID, err := parseUUID(transactionID, "transaction_id")
	if err != nil {
		return nil, err
	}

	// Verify transaction exists and belongs to user
	existing, err := s.repo.GetByUserID(ctx, transactionUUID, userUUID)
	if err != nil {
		if err == shared.ErrNotFound {
			return nil, err
		}
		return nil, shared.ErrInternal.WithError(err)
	}

	// Collect updates with validation
	updates, err := collectTransactionUpdates(existing, req)
	if err != nil {
		return nil, err
	}

	// Apply updates if any
	if len(updates) > 0 {
		if err := s.repo.UpdateColumns(ctx, transactionUUID, updates); err != nil {
			if err == shared.ErrNotFound {
				return nil, err
			}
			return nil, shared.ErrInternal.WithError(err)
		}
	}

	// Retrieve updated transaction
	updated, err := s.repo.GetByUserID(ctx, transactionUUID, userUUID)
	if err != nil {
		return nil, shared.ErrInternal.WithError(err)
	}

	return updated, nil
}

// collectTransactionUpdates collects and validates update fields from request
func collectTransactionUpdates(existing *domain.Transaction, req dto.UpdateTransactionRequest) (map[string]interface{}, error) {
	updates := make(map[string]interface{})

	// Validate and update enum fields
	if req.Direction != nil {
		direction, err := validateDirection(*req.Direction)
		if err != nil {
			return nil, err
		}
		updates["direction"] = direction
	}

	if req.Instrument != nil {
		instrument, err := validateInstrument(*req.Instrument)
		if err != nil {
			return nil, err
		}
		updates["instrument"] = instrument
	}

	if req.Source != nil {
		// Validate: Don't allow changing source for bank-imported transactions
		if existing.Source == domain.SourceBankAPI && *req.Source != string(domain.SourceBankAPI) {
			return nil, shared.ErrBadRequest.WithDetails("reason", "cannot change source for bank-imported transactions")
		}

		source, err := validateSource(*req.Source)
		if err != nil {
			return nil, err
		}
		updates["source"] = source
	}

	if req.Channel != nil {
		channel, err := validateChannel(*req.Channel)
		if err != nil {
			return nil, err
		}
		updates["channel"] = channel
	}

	// Update core fields
	if req.AccountID != nil {
		accountUUID, err := parseUUID(*req.AccountID, "accountId")
		if err != nil {
			return nil, err
		}
		updates["account_id"] = accountUUID
	}

	if req.Amount != nil {
		updates["amount"] = *req.Amount
	}

	if req.Currency != nil {
		updates["currency"] = getDefaultCurrency(*req.Currency)
	}

	if req.BookingDate != nil {
		updates["booking_date"] = *req.BookingDate
	}

	if req.ValueDate != nil {
		updates["value_date"] = *req.ValueDate
	}

	// Update description fields
	if req.Description != nil {
		updates["description"] = *req.Description
	}

	if req.UserNote != nil {
		updates["user_note"] = *req.UserNote
	}

	if req.Reference != nil {
		updates["reference"] = *req.Reference
	}

	// Update bank-specific fields
	if req.BankCode != nil {
		updates["bank_code"] = *req.BankCode
	}

	if req.ExternalID != nil {
		// Validate: Don't allow changing externalId for already-imported transactions
		if existing.ExternalID != "" && existing.Source == domain.SourceBankAPI {
			return nil, shared.ErrBadRequest.WithDetails("reason", "cannot change externalId for bank-imported transactions")
		}
		updates["external_id"] = *req.ExternalID
	}

	if req.RunningBalance != nil {
		updates["running_balance"] = *req.RunningBalance
	}

	// Update counterparty
	if req.CounterpartyName != nil || req.CounterpartyAccountNumber != nil ||
		req.CounterpartyBankName != nil || req.CounterpartyType != nil {
		counterparty := buildCounterparty(
			getStringValue(req.CounterpartyName),
			getStringValue(req.CounterpartyAccountNumber),
			getStringValue(req.CounterpartyBankName),
			getStringValue(req.CounterpartyType),
		)
		updates["counterparty"] = counterparty
	}

	// Update classification
	if req.SystemCategory != nil || req.UserCategoryID != nil ||
		req.IsTransfer != nil || req.IsRefund != nil || req.Tags != nil {

		systemCategory := ""
		userCategoryID := ""
		isTransfer := false
		isRefund := false
		var tags []string

		if req.SystemCategory != nil {
			systemCategory = *req.SystemCategory
		}
		if req.UserCategoryID != nil {
			userCategoryID = *req.UserCategoryID
		}
		if req.IsTransfer != nil {
			isTransfer = *req.IsTransfer
		}
		if req.IsRefund != nil {
			isRefund = *req.IsRefund
		}
		if req.Tags != nil {
			tags = *req.Tags
		}

		classification := buildClassification(
			systemCategory,
			userCategoryID,
			isTransfer,
			isRefund,
			tags,
		)
		updates["classification"] = classification
	}

	// Update links - WRITE-ONCE CONSTRAINT
	// If existing transaction has links, reject any changes
	// If existing has no links, allow adding new links
	if req.Links != nil {
		if existing.Links != nil && len(*existing.Links) > 0 {
			// Existing transaction has links - reject any changes
			return nil, shared.ErrBadRequest.WithDetails("reason", "links cannot be modified after creation (write-once)")
		}

		// No existing links - allow adding new links
		links := make([]domain.TransactionLink, 0, len(*req.Links))
		for _, linkDTO := range *req.Links {
			links = append(links, domain.TransactionLink{
				Type: domain.LinkType(linkDTO.Type),
				ID:   linkDTO.ID,
			})
		}
		updates["links"] = &links
	}

	// Update metadata
	if req.CheckImageAvailability != nil {
		meta := buildMetadata(*req.CheckImageAvailability, nil)
		updates["meta"] = meta
	}

	return updates, nil
}

// Helper to get string value from pointer
func getStringValue(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
