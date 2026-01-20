package handler

import (
	"net/http"
	"personalfinancedss/internal/middleware"
	"personalfinancedss/internal/module/cashflow/debt/domain"
	"personalfinancedss/internal/module/cashflow/debt/dto"
	"personalfinancedss/internal/module/cashflow/debt/service"
	"personalfinancedss/internal/shared"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type Handler struct {
	service service.Service
	logger  *zap.Logger
}

// NewHandler creates a new debt handler
func NewHandler(service service.Service, logger *zap.Logger) *Handler {
	return &Handler{
		service: service,
		logger:  logger,
	}
}

// RegisterRoutes registers debt routes
func (h *Handler) RegisterRoutes(router *gin.Engine, authMiddleware *middleware.Middleware) {
	debts := router.Group("/api/v1/debts")
	debts.Use(authMiddleware.AuthMiddleware())
	{
		debts.POST("", h.CreateDebt)
		debts.GET("", h.GetUserDebts)
		debts.GET("/active", h.GetActiveDebts)
		debts.GET("/paid-off", h.GetPaidOffDebts)
		debts.GET("/summary", h.GetDebtSummary)
		debts.GET("/:id", h.GetDebtByID)
		debts.PUT("/:id", h.UpdateDebt)
		debts.DELETE("/:id", h.DeleteDebt)
		debts.POST("/:id/payment", h.AddPayment)
		debts.POST("/:id/paid-off", h.MarkAsPaidOff)
	}
}

// CreateDebt godoc
// @Summary Create a new debt
// @Description Create a new debt for the authenticated user
// @Tags debts
// @Accept json
// @Produce json
// @Param debt body dto.CreateDebtRequest true "Debt details"
// @Success 201 {object} dto.DebtResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/debts [post]
func (h *Handler) CreateDebt(c *gin.Context) {
	var req dto.CreateDebtRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		shared.RespondWithError(c, http.StatusUnauthorized, "user not authenticated")
		return
	}

	status := domain.DebtStatusActive
	if req.Status != nil {
		status = *req.Status
	}

	debt := &domain.Debt{
		UserID:            userID.(uuid.UUID),
		Name:              req.Name,
		Description:       req.Description,
		Type:              req.Type,
		Behavior:          req.Behavior,
		Status:            status,
		PrincipalAmount:   req.PrincipalAmount,
		CurrentBalance:    req.CurrentBalance,
		InterestRate:      req.InterestRate,
		MinimumPayment:    req.MinimumPayment,
		PaymentAmount:     req.PaymentAmount,
		Currency:          req.Currency,
		PaymentFrequency:  req.PaymentFrequency,
		NextPaymentDate:   req.NextPaymentDate,
		StartDate:         req.StartDate,
		DueDate:           req.DueDate,
		CreditorName:      req.CreditorName,
		AccountNumber:     req.AccountNumber,
		LinkedAccountID:   req.LinkedAccountID,
		EnableReminders:   req.EnableReminders,
		ReminderFrequency: req.ReminderFrequency,
		Notes:             req.Notes,
		Tags:              req.Tags,
	}

	if err := h.service.CreateDebt(c.Request.Context(), debt); err != nil {
		h.logger.Error("Failed to create debt", zap.Error(err))
		shared.HandleError(c, err)
		return
	}

	shared.RespondWithSuccess(c, http.StatusCreated, "Debt created successfully", dto.ToDebtResponse(debt))
}

// GetUserDebts godoc
// @Summary Get all user debts
// @Description Get all debts for the authenticated user
// @Tags debts
// @Produce json
// @Success 200 {array} dto.DebtResponse
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/debts [get]
func (h *Handler) GetUserDebts(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		return
	}

	debts, err := h.service.GetUserDebts(c.Request.Context(), userID.(uuid.UUID))
	if err != nil {
		h.logger.Error("Failed to get user debts", zap.Error(err))
		shared.HandleError(c, err)
		return
	}

	shared.RespondWithSuccess(c, http.StatusOK, "User debts retrieved successfully", dto.ToDebtResponseList(debts))
}

// GetActiveDebts godoc
// @Summary Get active debts
// @Description Get all active debts for the authenticated user
// @Tags debts
// @Produce json
// @Success 200 {array} dto.DebtResponse
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/debts/active [get]
func (h *Handler) GetActiveDebts(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		return
	}

	debts, err := h.service.GetActiveDebts(c.Request.Context(), userID.(uuid.UUID))
	if err != nil {
		h.logger.Error("Failed to get active debts", zap.Error(err))
		shared.HandleError(c, err)
		return
	}

	shared.RespondWithSuccess(c, http.StatusOK, "Active debts retrieved successfully", dto.ToDebtResponseList(debts))
}

// GetPaidOffDebts godoc
// @Summary Get paid off debts
// @Description Get all paid off debts for the authenticated user
// @Tags debts
// @Produce json
// @Success 200 {array} dto.DebtResponse
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/debts/paid-off [get]
func (h *Handler) GetPaidOffDebts(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		return
	}

	debts, err := h.service.GetPaidOffDebts(c.Request.Context(), userID.(uuid.UUID))
	if err != nil {
		h.logger.Error("Failed to get paid off debts", zap.Error(err))
		shared.HandleError(c, err)
		return
	}

	shared.RespondWithSuccess(c, http.StatusOK, "Paid off debts retrieved successfully", dto.ToDebtResponseList(debts))
}

// GetDebtByID godoc
// @Summary Get debt by ID
// @Description Get a specific debt by its ID
// @Tags debts
// @Produce json
// @Param id path string true "Debt ID"
// @Success 200 {object} dto.DebtResponse
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/debts/{id} [get]
func (h *Handler) GetDebtByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid debt ID"})
		return
	}

	debt, err := h.service.GetDebtByID(c.Request.Context(), id)
	if err != nil {
		shared.HandleError(c, err)
		return
	}

	shared.RespondWithSuccess(c, http.StatusOK, "Debt retrieved successfully", dto.ToDebtResponse(debt))
}

// UpdateDebt godoc
// @Summary Update debt
// @Description Update an existing debt
// @Tags debts
// @Accept json
// @Produce json
// @Param id path string true "Debt ID"
// @Param debt body dto.UpdateDebtRequest true "Debt details"
// @Success 200 {object} dto.DebtResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/debts/{id} [put]
func (h *Handler) UpdateDebt(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid debt ID"})
		return
	}

	var req dto.UpdateDebtRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	debt, err := h.service.GetDebtByID(c.Request.Context(), id)
	if err != nil {
		shared.HandleError(c, err)
		return
	}

	// Update fields if provided
	if req.Name != nil {
		debt.Name = *req.Name
	}
	if req.Description != nil {
		debt.Description = req.Description
	}
	if req.Type != nil {
		debt.Type = *req.Type
	}
	if req.Behavior != nil {
		debt.Behavior = *req.Behavior
	}
	if req.Status != nil {
		debt.Status = *req.Status
	}
	if req.PrincipalAmount != nil {
		debt.PrincipalAmount = *req.PrincipalAmount
	}
	if req.CurrentBalance != nil {
		debt.CurrentBalance = *req.CurrentBalance
	}
	if req.InterestRate != nil {
		debt.InterestRate = *req.InterestRate
	}
	if req.MinimumPayment != nil {
		debt.MinimumPayment = *req.MinimumPayment
	}
	if req.PaymentAmount != nil {
		debt.PaymentAmount = *req.PaymentAmount
	}
	if req.Currency != nil {
		debt.Currency = *req.Currency
	}
	if req.PaymentFrequency != nil {
		debt.PaymentFrequency = req.PaymentFrequency
	}
	if req.NextPaymentDate != nil {
		debt.NextPaymentDate = req.NextPaymentDate
	}
	if req.LastPaymentDate != nil {
		debt.LastPaymentDate = req.LastPaymentDate
	}
	if req.StartDate != nil {
		debt.StartDate = *req.StartDate
	}
	if req.DueDate != nil {
		debt.DueDate = req.DueDate
	}
	if req.CreditorName != nil {
		debt.CreditorName = req.CreditorName
	}
	if req.AccountNumber != nil {
		debt.AccountNumber = req.AccountNumber
	}
	if req.LinkedAccountID != nil {
		debt.LinkedAccountID = req.LinkedAccountID
	}
	if req.EnableReminders != nil {
		debt.EnableReminders = *req.EnableReminders
	}
	if req.ReminderFrequency != nil {
		debt.ReminderFrequency = req.ReminderFrequency
	}
	if req.Notes != nil {
		debt.Notes = req.Notes
	}
	if req.Tags != nil {
		debt.Tags = req.Tags
	}

	if err := h.service.UpdateDebt(c.Request.Context(), debt); err != nil {
		h.logger.Error("Failed to update debt", zap.Error(err))
		shared.HandleError(c, err)
		return
	}

	shared.RespondWithSuccess(c, http.StatusOK, "Debt updated successfully", dto.ToDebtResponse(debt))
}

// DeleteDebt godoc
// @Summary Delete debt
// @Description Delete a debt
// @Tags debts
// @Param id path string true "Debt ID"
// @Success 204
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/debts/{id} [delete]
func (h *Handler) DeleteDebt(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid debt ID"})
		return
	}

	if err := h.service.DeleteDebt(c.Request.Context(), id); err != nil {
		h.logger.Error("Failed to delete debt", zap.Error(err))
		shared.HandleError(c, err)
		return
	}

	shared.RespondWithNoContent(c)
}

// AddPayment godoc
// @Summary Add payment to debt
// @Description Add a payment amount to a debt
// @Tags debts
// @Accept json
// @Produce json
// @Param id path string true "Debt ID"
// @Param payment body dto.AddPaymentRequest true "Payment details"
// @Success 200 {object} dto.DebtResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/debts/{id}/payment [post]
func (h *Handler) AddPayment(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid debt ID"})
		return
	}

	var req dto.AddPaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	debt, err := h.service.AddPayment(c.Request.Context(), id, req.Amount)
	if err != nil {
		h.logger.Error("Failed to add payment", zap.Error(err))
		shared.HandleError(c, err)
		return
	}

	shared.RespondWithSuccess(c, http.StatusOK, "Payment added successfully", dto.ToDebtResponse(debt))
}

// MarkAsPaidOff godoc
// @Summary Mark debt as paid off
// @Description Mark a debt as paid off
// @Tags debts
// @Param id path string true "Debt ID"
// @Success 200 {object} dto.DebtResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/debts/{id}/paid-off [post]
func (h *Handler) MarkAsPaidOff(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid debt ID"})
		return
	}

	if err := h.service.MarkAsPaidOff(c.Request.Context(), id); err != nil {
		h.logger.Error("Failed to mark debt as paid off", zap.Error(err))
		shared.HandleError(c, err)
		return
	}

	debt, err := h.service.GetDebtByID(c.Request.Context(), id)
	if err != nil {
		shared.HandleError(c, err)
		return
	}

	shared.RespondWithSuccess(c, http.StatusOK, "Debt marked as paid off", dto.ToDebtResponse(debt))
}

// GetDebtSummary godoc
// @Summary Get debt summary
// @Description Get a summary of all debts for the authenticated user
// @Tags debts
// @Produce json
// @Success 200 {object} dto.DebtSummaryResponse
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/debts/summary [get]
func (h *Handler) GetDebtSummary(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		return
	}

	summary, err := h.service.GetDebtSummary(c.Request.Context(), userID.(uuid.UUID))
	if err != nil {
		h.logger.Error("Failed to get debt summary", zap.Error(err))
		shared.HandleError(c, err)
		return
	}

	shared.RespondWithSuccess(c, http.StatusOK, "Debt summary retrieved successfully", dto.ToDebtSummaryResponse(summary))
}
