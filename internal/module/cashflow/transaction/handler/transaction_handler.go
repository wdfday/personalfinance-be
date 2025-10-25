package handler

import (
	"net/http"

	"personalfinancedss/internal/middleware"
	"personalfinancedss/internal/module/cashflow/transaction/dto"
	"personalfinancedss/internal/module/cashflow/transaction/service"
	"personalfinancedss/internal/shared"

	"github.com/gin-gonic/gin"
)

// Handler handles transaction-related HTTP requests
type Handler struct {
	service service.Service
}

// NewHandler creates a new transaction handler
func NewHandler(service service.Service) *Handler {
	return &Handler{
		service: service,
	}
}

// RegisterRoutes registers all transaction routes
func (h *Handler) RegisterRoutes(r *gin.Engine, authMiddleware *middleware.Middleware) {
	transactions := r.Group("/api/v1/transactions")
	transactions.Use(authMiddleware.AuthMiddleware())
	{
		transactions.POST("", h.createTransaction)
		transactions.GET("", h.listTransactions)
		transactions.GET("/:id", h.getTransaction)
		transactions.PUT("/:id", h.updateTransaction)
		transactions.DELETE("/:id", h.deleteTransaction)
		transactions.GET("/summary", h.getTransactionSummary)

		// Import endpoints
		transactions.POST("/import/json", h.importJSONTransactions)
	}
}

// CreateTransaction godoc
// @Summary Create a new transaction
// @Description Create a new financial transaction with direction (DEBIT/CREDIT), instrument type, and source
// @Tags transactions
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param transaction body dto.CreateTransactionRequest true "Transaction data"
// @Success 201 {object} dto.TransactionResponse
// @Failure 400 {object} shared.ErrorResponse
// @Failure 401 {object} shared.ErrorResponse
// @Router /api/v1/transactions [post]
func (h *Handler) createTransaction(c *gin.Context) {
	// Get user from context
	user, exists := middleware.GetCurrentUser(c)
	if !exists {
		shared.RespondWithError(c, http.StatusUnauthorized, "user not found in context")
		return
	}

	// Parse request
	var req dto.CreateTransactionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		shared.RespondWithError(c, http.StatusBadRequest, "invalid request data: "+err.Error())
		return
	}

	// Create transaction
	transaction, err := h.service.CreateTransaction(c.Request.Context(), user.ID.String(), req)
	if err != nil {
		shared.HandleError(c, err)
		return
	}

	// Convert to response
	response := dto.ToTransactionResponse(transaction)
	shared.RespondWithSuccess(c, http.StatusCreated, "Transaction created successfully", response)
}

// ListTransactions godoc
// @Summary List transactions
// @Description Get a paginated list of transactions with optional filters
// @Tags transactions
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param accountId query string false "Filter by account ID"
// @Param direction query string false "Filter by direction (DEBIT, CREDIT)"
// @Param instrument query string false "Filter by instrument (CASH, BANK_ACCOUNT, E_WALLET, etc.)"
// @Param source query string false "Filter by source (BANK_API, CSV_IMPORT, MANUAL)"
// @Param bankCode query string false "Filter by bank code"
// @Param startBookingDate query string false "Start booking date (YYYY-MM-DD)"
// @Param endBookingDate query string false "End booking date (YYYY-MM-DD)"
// @Param startValueDate query string false "Start value date (YYYY-MM-DD)"
// @Param endValueDate query string false "End value date (YYYY-MM-DD)"
// @Param minAmount query number false "Minimum amount (in smallest currency unit)"
// @Param maxAmount query number false "Maximum amount (in smallest currency unit)"
// @Param categoryId query string false "Filter by user category ID"
// @Param isTransfer query boolean false "Filter transfers between own accounts"
// @Param isRefund query boolean false "Filter refund transactions"
// @Param tag query string false "Filter by specific tag"
// @Param search query string false "Search in description, userNote, counterparty name"
// @Param page query int false "Page number (default: 1)"
// @Param pageSize query int false "Page size (default: 20, max: 100)"
// @Param sortBy query string false "Sort by field (booking_date, value_date, amount, created_at)"
// @Param sortOrder query string false "Sort order (asc, desc)"
// @Success 200 {object} dto.TransactionListResponse
// @Failure 400 {object} shared.ErrorResponse
// @Failure 401 {object} shared.ErrorResponse
// @Router /api/v1/transactions [get]
func (h *Handler) listTransactions(c *gin.Context) {
	// Get user from context
	user, exists := middleware.GetCurrentUser(c)
	if !exists {
		shared.RespondWithError(c, http.StatusUnauthorized, "user not found in context")
		return
	}

	// Parse query parameters
	var query dto.ListTransactionsQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		shared.RespondWithError(c, http.StatusBadRequest, "invalid query parameters: "+err.Error())
		return
	}

	// Get transactions
	response, err := h.service.ListTransactions(c.Request.Context(), user.ID.String(), query)
	if err != nil {
		shared.HandleError(c, err)
		return
	}

	shared.RespondWithSuccess(c, http.StatusOK, "Transactions retrieved successfully", response)
}

// GetTransaction godoc
// @Summary Get transaction by ID
// @Description Get detailed information about a specific transaction
// @Tags transactions
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Transaction ID"
// @Success 200 {object} dto.TransactionResponse
// @Failure 400 {object} shared.ErrorResponse
// @Failure 401 {object} shared.ErrorResponse
// @Failure 404 {object} shared.ErrorResponse
// @Router /api/v1/transactions/{id} [get]
func (h *Handler) getTransaction(c *gin.Context) {
	// Get user from context
	user, exists := middleware.GetCurrentUser(c)
	if !exists {
		shared.RespondWithError(c, http.StatusUnauthorized, "user not found in context")
		return
	}

	// Get transaction ID from path
	transactionID := c.Param("id")

	// Get transaction
	transaction, err := h.service.GetTransaction(c.Request.Context(), user.ID.String(), transactionID)
	if err != nil {
		shared.HandleError(c, err)
		return
	}

	// Convert to response
	response := dto.ToTransactionResponse(transaction)
	shared.RespondWithSuccess(c, http.StatusOK, "Transaction retrieved successfully", response)
}

// UpdateTransaction godoc
// @Summary Update a transaction
// @Description Update an existing transaction
// @Tags transactions
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Transaction ID"
// @Param transaction body dto.UpdateTransactionRequest true "Updated transaction data"
// @Success 200 {object} dto.TransactionResponse
// @Failure 400 {object} shared.ErrorResponse
// @Failure 401 {object} shared.ErrorResponse
// @Failure 404 {object} shared.ErrorResponse
// @Router /api/v1/transactions/{id} [put]
func (h *Handler) updateTransaction(c *gin.Context) {
	// Get user from context
	user, exists := middleware.GetCurrentUser(c)
	if !exists {
		shared.RespondWithError(c, http.StatusUnauthorized, "user not found in context")
		return
	}

	// Get transaction ID from path
	transactionID := c.Param("id")

	// Parse request
	var req dto.UpdateTransactionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		shared.RespondWithError(c, http.StatusBadRequest, "invalid request data: "+err.Error())
		return
	}

	// Update transaction
	transaction, err := h.service.UpdateTransaction(c.Request.Context(), user.ID.String(), transactionID, req)
	if err != nil {
		shared.HandleError(c, err)
		return
	}

	// Convert to response
	response := dto.ToTransactionResponse(transaction)
	shared.RespondWithSuccess(c, http.StatusOK, "Transaction updated successfully", response)
}

// DeleteTransaction godoc
// @Summary Delete a transaction
// @Description Soft delete a transaction
// @Tags transactions
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Transaction ID"
// @Success 200 {object} shared.Success
// @Failure 400 {object} shared.ErrorResponse
// @Failure 401 {object} shared.ErrorResponse
// @Failure 404 {object} shared.ErrorResponse
// @Router /api/v1/transactions/{id} [delete]
func (h *Handler) deleteTransaction(c *gin.Context) {
	// Get user from context
	user, exists := middleware.GetCurrentUser(c)
	if !exists {
		shared.RespondWithError(c, http.StatusUnauthorized, "user not found in context")
		return
	}

	// Get transaction ID from path
	transactionID := c.Param("id")

	// Delete transaction
	if err := h.service.DeleteTransaction(c.Request.Context(), user.ID.String(), transactionID); err != nil {
		shared.HandleError(c, err)
		return
	}

	shared.RespondWithSuccessNoData(c, http.StatusOK, "Transaction deleted successfully")
}

// GetTransactionSummary godoc
// @Summary Get transaction summary
// @Description Get aggregate information about transactions with breakdowns by direction, instrument, and source
// @Tags transactions
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param accountId query string false "Filter by account ID"
// @Param direction query string false "Filter by direction"
// @Param instrument query string false "Filter by instrument"
// @Param source query string false "Filter by source"
// @Param startBookingDate query string false "Start booking date (YYYY-MM-DD)"
// @Param endBookingDate query string false "End booking date (YYYY-MM-DD)"
// @Success 200 {object} dto.TransactionSummary
// @Failure 400 {object} shared.ErrorResponse
// @Failure 401 {object} shared.ErrorResponse
// @Router /api/v1/transactions/summary [get]
func (h *Handler) getTransactionSummary(c *gin.Context) {
	// Get user from context
	user, exists := middleware.GetCurrentUser(c)
	if !exists {
		shared.RespondWithError(c, http.StatusUnauthorized, "user not found in context")
		return
	}

	// Parse query parameters
	var query dto.ListTransactionsQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		shared.RespondWithError(c, http.StatusBadRequest, "invalid query parameters: "+err.Error())
		return
	}

	// Get summary
	summary, err := h.service.GetTransactionSummary(c.Request.Context(), user.ID.String(), query)
	if err != nil {
		shared.HandleError(c, err)
		return
	}

	shared.RespondWithSuccess(c, http.StatusOK, "Transaction summary retrieved successfully", summary)
}

// ImportJSONTransactions godoc
// @Summary Import bank transactions from JSON
// @Description Import transactions from bank JSON export and automatically sync account balance
// @Tags transactions
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param import body dto.ImportJSONRequest true "Import data"
// @Success 200 {object} dto.ImportJSONResponse
// @Failure 400 {object} shared.ErrorResponse
// @Failure 401 {object} shared.ErrorResponse
// @Router /api/v1/transactions/import/json [post]
func (h *Handler) importJSONTransactions(c *gin.Context) {
	// Get user from context
	user, exists := middleware.GetCurrentUser(c)
	if !exists {
		shared.RespondWithError(c, http.StatusUnauthorized, "user not found in context")
		return
	}

	// Parse request
	var req dto.ImportJSONRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		shared.RespondWithError(c, http.StatusBadRequest, "invalid request data: "+err.Error())
		return
	}

	// Import transactions
	response, err := h.service.ImportJSONTransactions(c.Request.Context(), user.ID.String(), req)
	if err != nil {
		shared.HandleError(c, err)
		return
	}

	shared.RespondWithSuccess(c, http.StatusOK, "Transactions imported successfully", response)
}
