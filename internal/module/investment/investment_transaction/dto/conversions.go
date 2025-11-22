package dto

import (
	"personalfinancedss/internal/module/investment/investment_transaction/domain"
)

// ToTransactionResponse converts a domain transaction to a response DTO
func ToTransactionResponse(transaction *domain.InvestmentTransaction) TransactionResponse {
	if transaction == nil {
		return TransactionResponse{}
	}

	response := TransactionResponse{
		ID:              transaction.ID,
		UserID:          transaction.UserID,
		AssetID:         transaction.AssetID,
		CreatedAt:       transaction.CreatedAt,
		UpdatedAt:       transaction.UpdatedAt,
		TransactionType: string(transaction.TransactionType),
		Quantity:        transaction.Quantity,
		PricePerUnit:    transaction.PricePerUnit,
		TotalAmount:     transaction.TotalAmount,
		Currency:        transaction.Currency,
		Fees:            transaction.Fees,
		Commission:      transaction.Commission,
		Tax:             transaction.Tax,
		TotalCost:       transaction.TotalCost,
		RealizedGain:    transaction.RealizedGain,
		RealizedGainPct: transaction.RealizedGainPct,
		TransactionDate: transaction.TransactionDate,
		SettlementDate:  transaction.SettlementDate,
		Status:          string(transaction.Status),
		Description:     transaction.Description,
	}

	// Handle optional pointer fields
	if transaction.Notes != nil {
		response.Notes = *transaction.Notes
	}
	if transaction.Broker != nil {
		response.Broker = *transaction.Broker
	}
	if transaction.Exchange != nil {
		response.Exchange = *transaction.Exchange
	}
	if transaction.OrderID != nil {
		response.OrderID = *transaction.OrderID
	}
	if transaction.Tags != nil {
		response.Tags = *transaction.Tags
	}

	return response
}

// ToTransactionListResponse converts a list of domain transactions to a list response DTO
func ToTransactionListResponse(transactions []*domain.InvestmentTransaction, total int64, page, pageSize int) TransactionListResponse {
	transactionResponses := make([]TransactionResponse, 0, len(transactions))
	for _, transaction := range transactions {
		transactionResponses = append(transactionResponses, ToTransactionResponse(transaction))
	}

	return TransactionListResponse{
		Transactions: transactionResponses,
		Total:        total,
		Page:         page,
		PerPage:      pageSize,
	}
}
