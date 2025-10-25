package dto

import (
	"encoding/json"

	"personalfinancedss/internal/module/cashflow/transaction/domain"
)

// ToTransactionResponse converts domain.Transaction to TransactionResponse
func ToTransactionResponse(t *domain.Transaction) *TransactionResponse {
	if t == nil {
		return nil
	}

	resp := &TransactionResponse{
		ID:          t.ID.String(),
		UserID:      t.UserID.String(),
		AccountID:   t.AccountID.String(),
		Direction:   string(t.Direction),
		Instrument:  string(t.Instrument),
		Source:      string(t.Source),
		BankCode:    t.BankCode,
		ExternalID:  t.ExternalID,
		Channel:     string(t.Channel),
		Amount:      t.Amount,
		Currency:    t.Currency,
		BookingDate: t.BookingDate,
		ValueDate:   t.ValueDate,
		CreatedAt:   t.CreatedAt,
		ImportedAt:  t.ImportedAt,
		Description: t.Description,
		UserNote:    t.UserNote,
		Reference:   t.Reference,
	}

	// Convert running balance
	if t.RunningBalance != nil {
		resp.RunningBalance = t.RunningBalance
	}

	// Convert counterparty
	if t.Counterparty != nil {
		resp.Counterparty = &CounterpartyResponse{
			Name:          t.Counterparty.Name,
			AccountNumber: t.Counterparty.AccountNumber,
			BankName:      t.Counterparty.BankName,
			Type:          t.Counterparty.Type,
		}
	}

	// Convert classification
	if t.Classification != nil {
		resp.Classification = &ClassificationResponse{
			SystemCategory: t.Classification.SystemCategory,
			UserCategoryID: t.Classification.UserCategoryID,
			IsTransfer:     t.Classification.IsTransfer,
			IsRefund:       t.Classification.IsRefund,
			Tags:           t.Classification.Tags,
		}
	}

	// Convert links
	if t.Links != nil && len(*t.Links) > 0 {
		resp.Links = make([]TransactionLinkResponse, 0, len(*t.Links))
		for _, link := range *t.Links {
			resp.Links = append(resp.Links, TransactionLinkResponse{
				Type: string(link.Type),
				ID:   link.ID,
			})
		}
	}

	// Convert metadata
	if t.Meta != nil {
		meta := &TransactionMetaResponse{
			CheckImageAvailability: t.Meta.CheckImageAvailability,
		}

		// Convert raw JSON to map
		if len(t.Meta.Raw) > 0 {
			var rawMap map[string]interface{}
			if err := json.Unmarshal(t.Meta.Raw, &rawMap); err == nil {
				meta.Raw = rawMap
			}
		}

		resp.Meta = meta
	}

	return resp
}

// ToTransactionListResponse converts a slice of transactions to list response
func ToTransactionListResponse(transactions []*domain.Transaction, pagination PaginationInfo, summary *TransactionSummary) *TransactionListResponse {
	resp := &TransactionListResponse{
		Transactions: make([]TransactionResponse, 0, len(transactions)),
		Pagination:   pagination,
		Summary:      summary,
	}

	for _, t := range transactions {
		if tr := ToTransactionResponse(t); tr != nil {
			resp.Transactions = append(resp.Transactions, *tr)
		}
	}

	return resp
}

// FromCreateRequest converts CreateTransactionRequest to domain.Transaction
func FromCreateRequest(req CreateTransactionRequest) (*domain.Transaction, error) {
	t := &domain.Transaction{
		// Core fields parsed separately in service layer for UUID conversion
		Direction:   domain.Direction(req.Direction),
		Instrument:  domain.Instrument(req.Instrument),
		Source:      domain.TransactionSource(req.Source),
		Amount:      req.Amount,
		Currency:    req.Currency,
		BankCode:    req.BankCode,
		ExternalID:  req.ExternalID,
		BookingDate: req.BookingDate,
		Description: req.Description,
		UserNote:    req.UserNote,
		Reference:   req.Reference,
	}

	// Set value date (defaults to booking date if not provided)
	if req.ValueDate != nil {
		t.ValueDate = *req.ValueDate
	} else {
		t.ValueDate = req.BookingDate
	}

	// Set channel if provided
	if req.Channel != "" {
		t.Channel = domain.Channel(req.Channel)
	}

	// Set running balance if provided
	if req.RunningBalance != nil {
		t.RunningBalance = req.RunningBalance
	}

	// Build counterparty if any field is provided
	if req.CounterpartyName != "" || req.CounterpartyAccountNumber != "" || req.CounterpartyBankName != "" {
		t.Counterparty = &domain.Counterparty{
			Name:          req.CounterpartyName,
			AccountNumber: req.CounterpartyAccountNumber,
			BankName:      req.CounterpartyBankName,
			Type:          req.CounterpartyType,
		}
	}

	// Build classification if any field is provided
	if req.SystemCategory != "" || req.UserCategoryID != "" || req.IsTransfer || req.IsRefund || len(req.Tags) > 0 {
		t.Classification = &domain.Classification{
			SystemCategory: req.SystemCategory,
			UserCategoryID: req.UserCategoryID,
			IsTransfer:     req.IsTransfer,
			IsRefund:       req.IsRefund,
			Tags:           req.Tags,
		}
	}

	// Build links if provided
	if len(req.Links) > 0 {
		links := make([]domain.TransactionLink, 0, len(req.Links))
		for _, linkDTO := range req.Links {
			links = append(links, domain.TransactionLink{
				Type: domain.LinkType(linkDTO.Type),
				ID:   linkDTO.ID,
			})
		}
		t.Links = &links
	}

	// Build metadata if provided
	if req.CheckImageAvailability != "" {
		t.Meta = &domain.TransactionMeta{
			CheckImageAvailability: req.CheckImageAvailability,
		}
	}

	return t, nil
}

// ApplyUpdateRequest applies UpdateTransactionRequest to a transaction
// Returns a map of fields to update for repository
func ApplyUpdateRequest(req UpdateTransactionRequest) map[string]interface{} {
	updates := make(map[string]interface{})

	// Core fields
	if req.Direction != nil {
		updates["direction"] = domain.Direction(*req.Direction)
	}
	if req.Instrument != nil {
		updates["instrument"] = domain.Instrument(*req.Instrument)
	}
	if req.Source != nil {
		updates["source"] = domain.TransactionSource(*req.Source)
	}
	if req.Amount != nil {
		updates["amount"] = *req.Amount
	}
	if req.Currency != nil {
		updates["currency"] = *req.Currency
	}

	// Timestamps
	if req.BookingDate != nil {
		updates["booking_date"] = *req.BookingDate
	}
	if req.ValueDate != nil {
		updates["value_date"] = *req.ValueDate
	}

	// Descriptions
	if req.Description != nil {
		updates["description"] = *req.Description
	}
	if req.UserNote != nil {
		updates["user_note"] = *req.UserNote
	}
	if req.Reference != nil {
		updates["reference"] = *req.Reference
	}

	// Bank fields
	if req.BankCode != nil {
		updates["bank_code"] = *req.BankCode
	}
	if req.ExternalID != nil {
		updates["external_id"] = *req.ExternalID
	}
	if req.Channel != nil {
		updates["channel"] = domain.Channel(*req.Channel)
	}
	if req.RunningBalance != nil {
		updates["running_balance"] = *req.RunningBalance
	}

	// Counterparty (build if any field is updated)
	if req.CounterpartyName != nil || req.CounterpartyAccountNumber != nil ||
		req.CounterpartyBankName != nil || req.CounterpartyType != nil {
		counterparty := &domain.Counterparty{}
		if req.CounterpartyName != nil {
			counterparty.Name = *req.CounterpartyName
		}
		if req.CounterpartyAccountNumber != nil {
			counterparty.AccountNumber = *req.CounterpartyAccountNumber
		}
		if req.CounterpartyBankName != nil {
			counterparty.BankName = *req.CounterpartyBankName
		}
		if req.CounterpartyType != nil {
			counterparty.Type = *req.CounterpartyType
		}
		updates["counterparty"] = counterparty
	}

	// Classification (build if any field is updated)
	if req.SystemCategory != nil || req.UserCategoryID != nil ||
		req.IsTransfer != nil || req.IsRefund != nil || req.Tags != nil {
		classification := &domain.Classification{}
		if req.SystemCategory != nil {
			classification.SystemCategory = *req.SystemCategory
		}
		if req.UserCategoryID != nil {
			classification.UserCategoryID = *req.UserCategoryID
		}
		if req.IsTransfer != nil {
			classification.IsTransfer = *req.IsTransfer
		}
		if req.IsRefund != nil {
			classification.IsRefund = *req.IsRefund
		}
		if req.Tags != nil {
			classification.Tags = *req.Tags
		}
		updates["classification"] = classification
	}

	// Links
	if req.Links != nil {
		links := make([]domain.TransactionLink, 0, len(*req.Links))
		for _, linkDTO := range *req.Links {
			links = append(links, domain.TransactionLink{
				Type: domain.LinkType(linkDTO.Type),
				ID:   linkDTO.ID,
			})
		}
		updates["links"] = &links
	}

	// Metadata
	if req.CheckImageAvailability != nil {
		meta := &domain.TransactionMeta{
			CheckImageAvailability: *req.CheckImageAvailability,
		}
		updates["meta"] = meta
	}

	return updates
}
