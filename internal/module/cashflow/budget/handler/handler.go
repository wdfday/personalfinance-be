package handler

import (
	"net/http"
	"personalfinancedss/internal/module/cashflow/budget/domain"
	"personalfinancedss/internal/module/cashflow/budget/dto"
	"personalfinancedss/internal/module/cashflow/budget/service"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type Handler struct {
	service service.Service
	logger  *zap.Logger
}

// NewHandler creates a new budget handler
func NewHandler(service service.Service, logger *zap.Logger) *Handler {
	return &Handler{
		service: service,
		logger:  logger,
	}
}

// RegisterRoutes registers budget routes
func (h *Handler) RegisterRoutes(router *gin.Engine) {
	budgets := router.Group("/api/v1/budgets")
	{
		budgets.POST("", h.CreateBudget)
		budgets.GET("", h.GetUserBudgets)
		budgets.GET("/active", h.GetActiveBudgets)
		budgets.GET("/summary", h.GetBudgetSummary)
		budgets.GET("/:id", h.GetBudgetByID)
		budgets.PUT("/:id", h.UpdateBudget)
		budgets.DELETE("/:id", h.DeleteBudget)
		budgets.POST("/:id/recalculate", h.RecalculateBudget)
		budgets.POST("/recalculate-all", h.RecalculateAllBudgets)
	}
}

// CreateBudget godoc
// @Summary Create a new budget
// @Description Create a new budget for the authenticated user
// @Tags budgets
// @Accept json
// @Produce json
// @Param budget body dto.CreateBudgetRequest true "Budget details"
// @Success 201 {object} dto.BudgetResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/budgets [post]
func (h *Handler) CreateBudget(c *gin.Context) {
	var req dto.CreateBudgetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		return
	}

	budget := &domain.Budget{
		UserID:               userID.(uuid.UUID),
		Name:                 req.Name,
		Description:          req.Description,
		Amount:               req.Amount,
		Currency:             req.Currency,
		Period:               req.Period,
		StartDate:            req.StartDate,
		EndDate:              req.EndDate,
		CategoryID:           req.CategoryID,
		AccountID:            req.AccountID,
		EnableAlerts:         req.EnableAlerts,
		AlertThresholds:      req.AlertThresholds,
		AllowRollover:        req.AllowRollover,
		CarryOverPercent:     req.CarryOverPercent,
		AutoAdjust:           req.AutoAdjust,
		AutoAdjustPercentage: req.AutoAdjustPercentage,
		AutoAdjustBasedOn:    req.AutoAdjustBasedOn,
	}

	if err := h.service.CreateBudget(c.Request.Context(), budget); err != nil {
		h.logger.Error("Failed to create budget", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, dto.ToBudgetResponse(budget))
}

// GetUserBudgets godoc
// @Summary Get all user budgets
// @Description Get all budgets for the authenticated user
// @Tags budgets
// @Produce json
// @Success 200 {array} dto.BudgetResponse
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/budgets [get]
func (h *Handler) GetUserBudgets(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		return
	}

	budgets, err := h.service.GetUserBudgets(c.Request.Context(), userID.(uuid.UUID))
	if err != nil {
		h.logger.Error("Failed to get user budgets", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, dto.ToBudgetResponseList(budgets))
}

// GetActiveBudgets godoc
// @Summary Get active budgets
// @Description Get all active budgets for the authenticated user
// @Tags budgets
// @Produce json
// @Success 200 {array} dto.BudgetResponse
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/budgets/active [get]
func (h *Handler) GetActiveBudgets(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		return
	}

	budgets, err := h.service.GetActiveBudgets(c.Request.Context(), userID.(uuid.UUID))
	if err != nil {
		h.logger.Error("Failed to get active budgets", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, dto.ToBudgetResponseList(budgets))
}

// GetBudgetByID godoc
// @Summary Get budget by ID
// @Description Get a specific budget by its ID
// @Tags budgets
// @Produce json
// @Param id path string true "Budget ID"
// @Success 200 {object} dto.BudgetResponse
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/budgets/{id} [get]
func (h *Handler) GetBudgetByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid budget ID"})
		return
	}

	budget, err := h.service.GetBudgetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, dto.ToBudgetResponse(budget))
}

// UpdateBudget godoc
// @Summary Update budget
// @Description Update an existing budget
// @Tags budgets
// @Accept json
// @Produce json
// @Param id path string true "Budget ID"
// @Param budget body dto.UpdateBudgetRequest true "Budget details"
// @Success 200 {object} dto.BudgetResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/budgets/{id} [put]
func (h *Handler) UpdateBudget(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid budget ID"})
		return
	}

	var req dto.UpdateBudgetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	budget, err := h.service.GetBudgetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	// Update fields if provided
	if req.Name != nil {
		budget.Name = *req.Name
	}
	if req.Description != nil {
		budget.Description = req.Description
	}
	if req.Amount != nil {
		budget.Amount = *req.Amount
	}
	if req.Currency != nil {
		budget.Currency = *req.Currency
	}
	if req.Period != nil {
		budget.Period = *req.Period
	}
	if req.StartDate != nil {
		budget.StartDate = *req.StartDate
	}
	if req.EndDate != nil {
		budget.EndDate = req.EndDate
	}
	if req.CategoryID != nil {
		budget.CategoryID = req.CategoryID
	}
	if req.AccountID != nil {
		budget.AccountID = req.AccountID
	}
	if req.EnableAlerts != nil {
		budget.EnableAlerts = *req.EnableAlerts
	}
	if len(req.AlertThresholds) > 0 {
		budget.AlertThresholds = req.AlertThresholds
	}
	if req.AllowRollover != nil {
		budget.AllowRollover = *req.AllowRollover
	}
	if req.CarryOverPercent != nil {
		budget.CarryOverPercent = req.CarryOverPercent
	}
	if req.AutoAdjust != nil {
		budget.AutoAdjust = *req.AutoAdjust
	}
	if req.AutoAdjustPercentage != nil {
		budget.AutoAdjustPercentage = req.AutoAdjustPercentage
	}
	if req.AutoAdjustBasedOn != nil {
		budget.AutoAdjustBasedOn = req.AutoAdjustBasedOn
	}
	if req.Status != nil {
		budget.Status = *req.Status
	}

	if err := h.service.UpdateBudget(c.Request.Context(), budget); err != nil {
		h.logger.Error("Failed to update budget", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, dto.ToBudgetResponse(budget))
}

// DeleteBudget godoc
// @Summary Delete budget
// @Description Delete a budget
// @Tags budgets
// @Param id path string true "Budget ID"
// @Success 204
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/budgets/{id} [delete]
func (h *Handler) DeleteBudget(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid budget ID"})
		return
	}

	if err := h.service.DeleteBudget(c.Request.Context(), id); err != nil {
		h.logger.Error("Failed to delete budget", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

// RecalculateBudget godoc
// @Summary Recalculate budget spending
// @Description Recalculate spent amount for a budget
// @Tags budgets
// @Param id path string true "Budget ID"
// @Success 200 {object} dto.BudgetResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/budgets/{id}/recalculate [post]
func (h *Handler) RecalculateBudget(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid budget ID"})
		return
	}

	if err := h.service.RecalculateBudgetSpending(c.Request.Context(), id); err != nil {
		h.logger.Error("Failed to recalculate budget", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	budget, err := h.service.GetBudgetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, dto.ToBudgetResponse(budget))
}

// RecalculateAllBudgets godoc
// @Summary Recalculate all budgets
// @Description Recalculate spent amounts for all user budgets
// @Tags budgets
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/budgets/recalculate-all [post]
func (h *Handler) RecalculateAllBudgets(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		return
	}

	if err := h.service.RecalculateAllBudgets(c.Request.Context(), userID.(uuid.UUID)); err != nil {
		h.logger.Error("Failed to recalculate all budgets", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "all budgets recalculated successfully"})
}

// GetBudgetSummary godoc
// @Summary Get budget summary
// @Description Get a summary of budget performance for the authenticated user
// @Tags budgets
// @Produce json
// @Success 200 {object} dto.BudgetSummaryResponse
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/budgets/summary [get]
func (h *Handler) GetBudgetSummary(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		return
	}

	summary, err := h.service.GetBudgetSummary(c.Request.Context(), userID.(uuid.UUID), time.Now())
	if err != nil {
		h.logger.Error("Failed to get budget summary", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, dto.ToBudgetSummaryResponse(summary))
}
