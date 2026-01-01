package handler

import (
	"net/http"
	"time"

	"personalfinancedss/internal/middleware"
	"personalfinancedss/internal/module/cashflow/budget/domain"
	"personalfinancedss/internal/module/cashflow/budget/dto"
	"personalfinancedss/internal/module/cashflow/budget/service"
	"personalfinancedss/internal/shared"

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
func (h *Handler) RegisterRoutes(router *gin.Engine, authMiddleware *middleware.Middleware) {
	budgets := router.Group("/api/v1/budgets")
	budgets.Use(authMiddleware.AuthMiddleware())
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
// @Security BearerAuth
// @Param budget body dto.CreateBudgetRequest true "Budget details"
// @Success 201 {object} dto.BudgetResponse
// @Failure 400 {object} shared.ErrorResponse
// @Failure 401 {object} shared.ErrorResponse
// @Failure 500 {object} shared.ErrorResponse
// @Router /api/v1/budgets [post]
func (h *Handler) CreateBudget(c *gin.Context) {
	user, exists := middleware.GetCurrentUser(c)
	if !exists {
		shared.RespondWithError(c, http.StatusUnauthorized, "user not found in context")
		return
	}

	var req dto.CreateBudgetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		shared.RespondWithError(c, http.StatusBadRequest, "invalid request data: "+err.Error())
		return
	}

	budget := &domain.Budget{
		UserID:               user.ID,
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
		shared.HandleError(c, err)
		return
	}

	shared.RespondWithSuccess(c, http.StatusCreated, "Budget created successfully", dto.ToBudgetResponse(budget))
}

// GetUserBudgets godoc
// @Summary Get all user budgets
// @Description Get all budgets for the authenticated user
// @Tags budgets
// @Produce json
// @Security BearerAuth
// @Success 200 {array} dto.BudgetResponse
// @Failure 401 {object} shared.ErrorResponse
// @Failure 500 {object} shared.ErrorResponse
// @Router /api/v1/budgets [get]
func (h *Handler) GetUserBudgets(c *gin.Context) {
	user, exists := middleware.GetCurrentUser(c)
	if !exists {
		shared.RespondWithError(c, http.StatusUnauthorized, "user not found in context")
		return
	}

	budgets, err := h.service.GetUserBudgets(c.Request.Context(), user.ID)
	if err != nil {
		shared.HandleError(c, err)
		return
	}

	shared.RespondWithSuccess(c, http.StatusOK, "Budgets retrieved successfully", dto.ToBudgetResponseList(budgets))
}

// GetActiveBudgets godoc
// @Summary Get active budgets
// @Description Get all active budgets for the authenticated user
// @Tags budgets
// @Produce json
// @Security BearerAuth
// @Success 200 {array} dto.BudgetResponse
// @Failure 401 {object} shared.ErrorResponse
// @Failure 500 {object} shared.ErrorResponse
// @Router /api/v1/budgets/active [get]
func (h *Handler) GetActiveBudgets(c *gin.Context) {
	user, exists := middleware.GetCurrentUser(c)
	if !exists {
		shared.RespondWithError(c, http.StatusUnauthorized, "user not found in context")
		return
	}

	budgets, err := h.service.GetActiveBudgets(c.Request.Context(), user.ID)
	if err != nil {
		shared.HandleError(c, err)
		return
	}

	shared.RespondWithSuccess(c, http.StatusOK, "Active budgets retrieved successfully", dto.ToBudgetResponseList(budgets))
}

// GetBudgetByID godoc
// @Summary Get budget by ID
// @Description Get a specific budget by its ID
// @Tags budgets
// @Produce json
// @Security BearerAuth
// @Param id path string true "Budget ID"
// @Success 200 {object} dto.BudgetResponse
// @Failure 400 {object} shared.ErrorResponse
// @Failure 401 {object} shared.ErrorResponse
// @Failure 404 {object} shared.ErrorResponse
// @Router /api/v1/budgets/{id} [get]
func (h *Handler) GetBudgetByID(c *gin.Context) {
	_, exists := middleware.GetCurrentUser(c)
	if !exists {
		shared.RespondWithError(c, http.StatusUnauthorized, "user not found in context")
		return
	}

	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		shared.RespondWithError(c, http.StatusBadRequest, "invalid budget ID")
		return
	}

	budget, err := h.service.GetBudgetByID(c.Request.Context(), id)
	if err != nil {
		shared.HandleError(c, err)
		return
	}

	shared.RespondWithSuccess(c, http.StatusOK, "Budget retrieved successfully", dto.ToBudgetResponse(budget))
}

// UpdateBudget godoc
// @Summary Update budget
// @Description Update an existing budget
// @Tags budgets
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Budget ID"
// @Param budget body dto.UpdateBudgetRequest true "Budget details"
// @Success 200 {object} dto.BudgetResponse
// @Failure 400 {object} shared.ErrorResponse
// @Failure 401 {object} shared.ErrorResponse
// @Failure 404 {object} shared.ErrorResponse
// @Failure 500 {object} shared.ErrorResponse
// @Router /api/v1/budgets/{id} [put]
func (h *Handler) UpdateBudget(c *gin.Context) {
	_, exists := middleware.GetCurrentUser(c)
	if !exists {
		shared.RespondWithError(c, http.StatusUnauthorized, "user not found in context")
		return
	}

	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		shared.RespondWithError(c, http.StatusBadRequest, "invalid budget ID")
		return
	}

	var req dto.UpdateBudgetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		shared.RespondWithError(c, http.StatusBadRequest, "invalid request data: "+err.Error())
		return
	}

	budget, err := h.service.GetBudgetByID(c.Request.Context(), id)
	if err != nil {
		shared.HandleError(c, err)
		return
	}

	// Apply update request to budget
	req.ApplyTo(budget)

	if err := h.service.UpdateBudget(c.Request.Context(), budget); err != nil {
		shared.HandleError(c, err)
		return
	}

	shared.RespondWithSuccess(c, http.StatusOK, "Budget updated successfully", dto.ToBudgetResponse(budget))
}

// DeleteBudget godoc
// @Summary Delete budget
// @Description Delete a budget
// @Tags budgets
// @Security BearerAuth
// @Param id path string true "Budget ID"
// @Success 204
// @Failure 400 {object} shared.ErrorResponse
// @Failure 401 {object} shared.ErrorResponse
// @Failure 500 {object} shared.ErrorResponse
// @Router /api/v1/budgets/{id} [delete]
func (h *Handler) DeleteBudget(c *gin.Context) {
	_, exists := middleware.GetCurrentUser(c)
	if !exists {
		shared.RespondWithError(c, http.StatusUnauthorized, "user not found in context")
		return
	}

	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		shared.RespondWithError(c, http.StatusBadRequest, "invalid budget ID")
		return
	}

	if err := h.service.DeleteBudget(c.Request.Context(), id); err != nil {
		shared.HandleError(c, err)
		return
	}

	shared.RespondWithNoContent(c)
}

// RecalculateBudget godoc
// @Summary Recalculate budget spending
// @Description Recalculate spent amount for a budget
// @Tags budgets
// @Security BearerAuth
// @Param id path string true "Budget ID"
// @Success 200 {object} dto.BudgetResponse
// @Failure 400 {object} shared.ErrorResponse
// @Failure 401 {object} shared.ErrorResponse
// @Failure 404 {object} shared.ErrorResponse
// @Failure 500 {object} shared.ErrorResponse
// @Router /api/v1/budgets/{id}/recalculate [post]
func (h *Handler) RecalculateBudget(c *gin.Context) {
	_, exists := middleware.GetCurrentUser(c)
	if !exists {
		shared.RespondWithError(c, http.StatusUnauthorized, "user not found in context")
		return
	}

	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		shared.RespondWithError(c, http.StatusBadRequest, "invalid budget ID")
		return
	}

	if err := h.service.RecalculateBudgetSpending(c.Request.Context(), id); err != nil {
		shared.HandleError(c, err)
		return
	}

	budget, err := h.service.GetBudgetByID(c.Request.Context(), id)
	if err != nil {
		shared.HandleError(c, err)
		return
	}

	shared.RespondWithSuccess(c, http.StatusOK, "Budget recalculated successfully", dto.ToBudgetResponse(budget))
}

// RecalculateAllBudgets godoc
// @Summary Recalculate all budgets
// @Description Recalculate spent amounts for all user budgets
// @Tags budgets
// @Security BearerAuth
// @Success 200 {object} shared.Success
// @Failure 401 {object} shared.ErrorResponse
// @Failure 500 {object} shared.ErrorResponse
// @Router /api/v1/budgets/recalculate-all [post]
func (h *Handler) RecalculateAllBudgets(c *gin.Context) {
	user, exists := middleware.GetCurrentUser(c)
	if !exists {
		shared.RespondWithError(c, http.StatusUnauthorized, "user not found in context")
		return
	}

	if err := h.service.RecalculateAllBudgets(c.Request.Context(), user.ID); err != nil {
		shared.HandleError(c, err)
		return
	}

	shared.RespondWithSuccessNoData(c, http.StatusOK, "All budgets recalculated successfully")
}

// GetBudgetSummary godoc
// @Summary Get budget summary
// @Description Get a summary of budget performance for the authenticated user
// @Tags budgets
// @Produce json
// @Security BearerAuth
// @Success 200 {object} dto.BudgetSummaryResponse
// @Failure 401 {object} shared.ErrorResponse
// @Failure 500 {object} shared.ErrorResponse
// @Router /api/v1/budgets/summary [get]
func (h *Handler) GetBudgetSummary(c *gin.Context) {
	user, exists := middleware.GetCurrentUser(c)
	if !exists {
		shared.RespondWithError(c, http.StatusUnauthorized, "user not found in context")
		return
	}

	summary, err := h.service.GetBudgetSummary(c.Request.Context(), user.ID, time.Now())
	if err != nil {
		shared.HandleError(c, err)
		return
	}

	shared.RespondWithSuccess(c, http.StatusOK, "Budget summary retrieved successfully", dto.ToBudgetSummaryResponse(summary))
}
