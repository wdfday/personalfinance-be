package handler

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
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

// Handler handles budget-related HTTP requests
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
		budgets.GET("/constraint/:constraint_id", h.GetBudgetsByConstraint)
		budgets.GET("/:id", h.GetBudgetByID)
		budgets.GET("/:id/progress", h.GetBudgetProgress)
		budgets.GET("/:id/analytics", h.GetBudgetAnalytics)
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
		// Log validation error for debugging
		h.logger.Error("Budget request validation failed", zap.Error(err))
		shared.RespondWithError(c, http.StatusBadRequest, "invalid request data: "+err.Error())
		return
	}

	budget, err := h.service.CreateBudget(c.Request.Context(), user.ID, req)
	if err != nil {
		fmt.Printf("ERROR CreateBudget: %v\n", err)
		h.logger.Error("CreateBudget failed",
			zap.Error(err),
			zap.String("create_budget_error", err.Error()),
		)
		appErr := budgetCreateErrorToAppError(err)
		shared.RespondWithAppError(c, appErr)
		return
	}

	shared.RespondWithSuccess(c, http.StatusCreated, "Budget created successfully", dto.ToBudgetResponse(budget))
}

// budgetCreateErrorToAppError maps CreateBudget service errors to AppErrors for proper HTTP responses.
func budgetCreateErrorToAppError(err error) *shared.AppError {
	msg := err.Error()
	msgLower := strings.ToLower(msg)
	switch {
	case strings.Contains(msg, "user ID is required"),
		strings.Contains(msg, "budget name is required"),
		strings.Contains(msg, "amount must be greater"),
		strings.Contains(msg, "invalid budget period"),
		strings.Contains(msg, "unknown budget period"),
		strings.Contains(msg, "start date is required"),
		strings.Contains(msg, "end date must be after"),
		strings.Contains(msg, "cannot calculate end date"),
		strings.Contains(msg, "failed to auto-calculate"),
		strings.Contains(msg, "alert thresholds"):
		return shared.ErrValidation.WithError(err).WithDetails("reason", msg)
	case strings.Contains(msg, "conflicts with existing budget"),
		strings.Contains(msg, "another recurring budget"):
		return shared.ErrConflict.WithError(err).WithDetails("reason", msg)
	// DB constraint / schema errors -> 400 validation
	case strings.Contains(msgLower, "null value"),
		strings.Contains(msgLower, "violates not-null"),
		strings.Contains(msgLower, "not null"),
		strings.Contains(msgLower, "foreign key"),
		strings.Contains(msgLower, "violates foreign key"):
		return shared.ErrValidation.WithError(err).WithDetails("reason", msg)
	// Duplicate / unique -> 409 conflict
	case strings.Contains(msgLower, "duplicate key"),
		strings.Contains(msgLower, "unique constraint"),
		strings.Contains(msgLower, "already exists"):
		return shared.ErrConflict.WithError(err).WithDetails("reason", msg)
	default:
		return shared.ErrInternal.WithError(err)
	}
}

// GetUserBudgets godoc
// @Summary Get all user budgets
// @Description Get all budgets for the authenticated user with optional pagination
// @Tags budgets
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number (default: 1)"
// @Param page_size query int false "Page size (default: 10, max: 100)"
// @Success 200 {object} dto.PaginatedBudgetResponse
// @Failure 401 {object} shared.ErrorResponse
// @Failure 500 {object} shared.ErrorResponse
// @Router /api/v1/budgets [get]
func (h *Handler) GetUserBudgets(c *gin.Context) {
	user, exists := middleware.GetCurrentUser(c)
	if !exists {
		shared.RespondWithError(c, http.StatusUnauthorized, "user not found in context")
		return
	}

	// Parse pagination parameters
	page := 1
	pageSize := 10

	if pageStr := c.Query("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	if pageSizeStr := c.Query("page_size"); pageSizeStr != "" {
		if ps, err := strconv.Atoi(pageSizeStr); err == nil && ps > 0 {
			pageSize = ps
		}
	}

	result, err := h.service.GetUserBudgetsPaginated(c.Request.Context(), user.ID, page, pageSize)
	if err != nil {
		shared.HandleError(c, err)
		return
	}

	shared.RespondWithSuccess(c, http.StatusOK, "Budgets retrieved successfully", dto.ToPaginatedBudgetResponse(result))
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

// GetBudgetsByConstraint godoc
// @Summary Get budgets by constraint ID
// @Description Get all budgets for a specific constraint
// @Tags budgets
// @Produce json
// @Security BearerAuth
// @Param constraint_id path string true "Constraint ID"
// @Success 200 {array} dto.BudgetResponse
// @Failure 400 {object} shared.ErrorResponse
// @Failure 401 {object} shared.ErrorResponse
// @Failure 500 {object} shared.ErrorResponse
// @Router /api/v1/budgets/constraint/{constraint_id} [get]
func (h *Handler) GetBudgetsByConstraint(c *gin.Context) {
	user, exists := middleware.GetCurrentUser(c)
	if !exists {
		shared.RespondWithError(c, http.StatusUnauthorized, "user not found in context")
		return
	}

	constraintIDStr := c.Param("constraint_id")
	constraintID, err := uuid.Parse(constraintIDStr)
	if err != nil {
		shared.RespondWithError(c, http.StatusBadRequest, "invalid constraint ID")
		return
	}

	budgets, err := h.service.GetBudgetsByConstraint(c.Request.Context(), user.ID, constraintID)
	if err != nil {
		shared.HandleError(c, err)
		return
	}

	shared.RespondWithSuccess(c, http.StatusOK, "Budgets retrieved successfully", dto.ToBudgetResponseList(budgets))
}

// GetBudgetByID godoc
// @Summary Get budget by ID
// @Description Get a specific budget by its ID (only accessible by owner)
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
	user, exists := middleware.GetCurrentUser(c)
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

	// Use ownership-verified method
	budget, err := h.service.GetBudgetByIDForUser(c.Request.Context(), id, user.ID)
	if err != nil {
		if domain.IsBudgetNotFound(err) {
			shared.RespondWithError(c, http.StatusNotFound, "budget not found")
			return
		}
		shared.HandleError(c, err)
		return
	}

	shared.RespondWithSuccess(c, http.StatusOK, "Budget retrieved successfully", dto.ToBudgetResponse(budget))
}

// UpdateBudget godoc
// @Summary Update budget
// @Description Update an existing budget (only accessible by owner)
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
	user, exists := middleware.GetCurrentUser(c)
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

	// Fetch budget with ownership verification
	budget, err := h.service.GetBudgetByIDForUser(c.Request.Context(), id, user.ID)
	if err != nil {
		if domain.IsBudgetNotFound(err) {
			shared.RespondWithError(c, http.StatusNotFound, "budget not found")
			return
		}
		shared.HandleError(c, err)
		return
	}

	// Apply update request to budget
	req.ApplyTo(budget)

	// Update with ownership verification
	if err := h.service.UpdateBudgetForUser(c.Request.Context(), budget, user.ID); err != nil {
		shared.HandleError(c, err)
		return
	}

	shared.RespondWithSuccess(c, http.StatusOK, "Budget updated successfully", dto.ToBudgetResponse(budget))
}

// DeleteBudget godoc
// @Summary Delete budget
// @Description Delete a budget (only accessible by owner)
// @Tags budgets
// @Security BearerAuth
// @Param id path string true "Budget ID"
// @Success 204
// @Failure 400 {object} shared.ErrorResponse
// @Failure 401 {object} shared.ErrorResponse
// @Failure 404 {object} shared.ErrorResponse
// @Failure 500 {object} shared.ErrorResponse
// @Router /api/v1/budgets/{id} [delete]
func (h *Handler) DeleteBudget(c *gin.Context) {
	user, exists := middleware.GetCurrentUser(c)
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

	// Delete with ownership verification
	if err := h.service.DeleteBudgetForUser(c.Request.Context(), id, user.ID); err != nil {
		if domain.IsBudgetNotFound(err) {
			shared.RespondWithError(c, http.StatusNotFound, "budget not found")
			return
		}
		shared.HandleError(c, err)
		return
	}

	shared.RespondWithNoContent(c)
}

// RecalculateBudget godoc
// @Summary Recalculate budget spending
// @Description Recalculate spent amount for a budget (only accessible by owner)
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
	user, exists := middleware.GetCurrentUser(c)
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

	// Recalculate with ownership verification
	if err := h.service.RecalculateBudgetSpendingForUser(c.Request.Context(), id, user.ID); err != nil {
		if domain.IsBudgetNotFound(err) {
			shared.RespondWithError(c, http.StatusNotFound, "budget not found")
			return
		}
		shared.HandleError(c, err)
		return
	}

	// Fetch updated budget with ownership verification
	budget, err := h.service.GetBudgetByIDForUser(c.Request.Context(), id, user.ID)
	if err != nil {
		shared.HandleError(c, err)
		return
	}

	shared.RespondWithSuccess(c, http.StatusOK, "Budget recalculated successfully", dto.ToBudgetResponse(budget))
}

// GetBudgetProgress godoc
// @Summary Get budget progress
// @Description Get detailed progress for a budget (only accessible by owner)
// @Tags budgets
// @Produce json
// @Security BearerAuth
// @Param id path string true "Budget ID"
// @Success 200 {object} dto.BudgetProgressResponse
// @Failure 400 {object} shared.ErrorResponse
// @Failure 401 {object} shared.ErrorResponse
// @Failure 404 {object} shared.ErrorResponse
// @Router /api/v1/budgets/{id}/progress [get]
func (h *Handler) GetBudgetProgress(c *gin.Context) {
	user, exists := middleware.GetCurrentUser(c)
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

	progress, err := h.service.GetBudgetProgress(c.Request.Context(), id, user.ID)
	if err != nil {
		if domain.IsBudgetNotFound(err) {
			shared.RespondWithError(c, http.StatusNotFound, "budget not found")
			return
		}
		shared.HandleError(c, err)
		return
	}

	shared.RespondWithSuccess(c, http.StatusOK, "Budget progress retrieved successfully", dto.ToBudgetProgressResponse(progress))
}

// GetBudgetAnalytics godoc
// @Summary Get budget analytics
// @Description Get analytics for a budget (only accessible by owner)
// @Tags budgets
// @Produce json
// @Security BearerAuth
// @Param id path string true "Budget ID"
// @Success 200 {object} dto.BudgetAnalyticsResponse
// @Failure 400 {object} shared.ErrorResponse
// @Failure 401 {object} shared.ErrorResponse
// @Failure 404 {object} shared.ErrorResponse
// @Router /api/v1/budgets/{id}/analytics [get]
func (h *Handler) GetBudgetAnalytics(c *gin.Context) {
	user, exists := middleware.GetCurrentUser(c)
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

	analytics, err := h.service.GetBudgetAnalytics(c.Request.Context(), id, user.ID)
	if err != nil {
		if domain.IsBudgetNotFound(err) {
			shared.RespondWithError(c, http.StatusNotFound, "budget not found")
			return
		}
		shared.HandleError(c, err)
		return
	}

	shared.RespondWithSuccess(c, http.StatusOK, "Budget analytics retrieved successfully", dto.ToBudgetAnalyticsResponse(analytics))
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
