package handler

import (
	"net/http"

	"personalfinancedss/internal/middleware"
	"personalfinancedss/internal/module/investment/investment_transaction/dto"
	"personalfinancedss/internal/module/investment/investment_transaction/service"
	"personalfinancedss/internal/shared"

	"github.com/gin-gonic/gin"
)

// Handler handles investment transaction-related HTTP requests
type Handler struct {
	service service.Service
}

// NewHandler creates a new investment transaction handler
func NewHandler(service service.Service) *Handler {
	return &Handler{
		service: service,
	}
}

// RegisterRoutes registers all investment transaction routes
func (h *Handler) RegisterRoutes(r *gin.Engine, authMiddleware *middleware.Middleware) {
	transactions := r.Group("/api/v1/investment/transactions")
	transactions.Use(authMiddleware.AuthMiddleware())
	{
		transactions.GET("", h.listTransactions)
		transactions.GET("/summary", h.getTransactionSummary)
		transactions.GET("/asset/:assetId", h.getTransactionsByAsset)
		transactions.GET("/:id", h.getTransaction)
	}
}

// ListTransactions godoc
// @Summary List investment transactions
// @Description Get a paginated list of investment transactions with optional filters
// @Tags investment-transactions
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param asset_id query string false "Filter by asset ID"
// @Param transaction_type query string false "Filter by type (buy, sell, dividend)"
// @Param status query string false "Filter by status"
// @Param start_date query string false "Start date (YYYY-MM-DD)"
// @Param end_date query string false "End date (YYYY-MM-DD)"
// @Param min_amount query number false "Minimum amount"
// @Param max_amount query number false "Maximum amount"
// @Param broker query string false "Filter by broker"
// @Param page query int false "Page number (default: 1)"
// @Param page_size query int false "Page size (default: 20, max: 100)"
// @Param sort_by query string false "Sort by field (transaction_date, total_amount)"
// @Param sort_order query string false "Sort order (asc, desc)"
// @Success 200 {object} dto.TransactionListResponse
// @Failure 400 {object} shared.ErrorResponse
// @Failure 401 {object} shared.ErrorResponse
// @Router /api/v1/investment/transactions [get]
func (h *Handler) listTransactions(c *gin.Context) {
	user, exists := middleware.GetCurrentUser(c)
	if !exists {
		shared.RespondWithError(c, http.StatusUnauthorized, "user not found in context")
		return
	}

	var query dto.ListTransactionsQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		shared.RespondWithError(c, http.StatusBadRequest, "invalid query parameters: "+err.Error())
		return
	}

	response, err := h.service.ListTransactions(c.Request.Context(), user.ID.String(), query)
	if err != nil {
		shared.HandleError(c, err)
		return
	}

	shared.RespondWithSuccess(c, http.StatusOK, "Transactions retrieved successfully", response)
}

// GetTransaction godoc
// @Summary Get transaction by ID
// @Description Get detailed information about a specific investment transaction
// @Tags investment-transactions
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Transaction ID"
// @Success 200 {object} dto.TransactionResponse
// @Failure 400 {object} shared.ErrorResponse
// @Failure 401 {object} shared.ErrorResponse
// @Failure 404 {object} shared.ErrorResponse
// @Router /api/v1/investment/transactions/{id} [get]
func (h *Handler) getTransaction(c *gin.Context) {
	user, exists := middleware.GetCurrentUser(c)
	if !exists {
		shared.RespondWithError(c, http.StatusUnauthorized, "user not found in context")
		return
	}

	transactionID := c.Param("id")

	transaction, err := h.service.GetTransaction(c.Request.Context(), user.ID.String(), transactionID)
	if err != nil {
		shared.HandleError(c, err)
		return
	}

	response := dto.ToTransactionResponse(transaction)
	shared.RespondWithSuccess(c, http.StatusOK, "Transaction retrieved successfully", response)
}

// GetTransactionsByAsset godoc
// @Summary Get transactions by asset ID
// @Description Get all transactions for a specific asset
// @Tags investment-transactions
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param assetId path string true "Asset ID"
// @Success 200 {array} dto.TransactionResponse
// @Failure 400 {object} shared.ErrorResponse
// @Failure 401 {object} shared.ErrorResponse
// @Router /api/v1/investment/transactions/asset/{assetId} [get]
func (h *Handler) getTransactionsByAsset(c *gin.Context) {
	user, exists := middleware.GetCurrentUser(c)
	if !exists {
		shared.RespondWithError(c, http.StatusUnauthorized, "user not found in context")
		return
	}

	assetID := c.Param("assetId")

	transactions, err := h.service.GetByAsset(c.Request.Context(), user.ID.String(), assetID)
	if err != nil {
		shared.HandleError(c, err)
		return
	}

	responses := make([]dto.TransactionResponse, 0, len(transactions))
	for _, transaction := range transactions {
		responses = append(responses, dto.ToTransactionResponse(transaction))
	}

	shared.RespondWithSuccess(c, http.StatusOK, "Transactions retrieved successfully", responses)
}

// GetTransactionSummary godoc
// @Summary Get transaction summary
// @Description Get aggregate information about investment transactions
// @Tags investment-transactions
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param asset_id query string false "Filter by asset ID"
// @Param start_date query string false "Start date (YYYY-MM-DD)"
// @Param end_date query string false "End date (YYYY-MM-DD)"
// @Success 200 {object} dto.TransactionSummary
// @Failure 400 {object} shared.ErrorResponse
// @Failure 401 {object} shared.ErrorResponse
// @Router /api/v1/investment/transactions/summary [get]
func (h *Handler) getTransactionSummary(c *gin.Context) {
	user, exists := middleware.GetCurrentUser(c)
	if !exists {
		shared.RespondWithError(c, http.StatusUnauthorized, "user not found in context")
		return
	}

	var query dto.ListTransactionsQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		shared.RespondWithError(c, http.StatusBadRequest, "invalid query parameters: "+err.Error())
		return
	}

	summary, err := h.service.GetSummary(c.Request.Context(), user.ID.String(), query)
	if err != nil {
		shared.HandleError(c, err)
		return
	}

	shared.RespondWithSuccess(c, http.StatusOK, "Transaction summary retrieved successfully", summary)
}
