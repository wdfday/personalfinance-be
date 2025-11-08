package handler

import (
	"net/http"

	"personalfinancedss/internal/middleware"
	"personalfinancedss/internal/module/cashflow/budget_profile/dto"
	"personalfinancedss/internal/module/cashflow/budget_profile/service"
	"personalfinancedss/internal/shared"

	"github.com/gin-gonic/gin"
)

// Handler handles budget constraint-related HTTP requests
type Handler struct {
	service service.Service
}

// NewHandler creates a new budget constraint handler
func NewHandler(service service.Service) *Handler {
	return &Handler{
		service: service,
	}
}

// RegisterRoutes registers all budget constraint routes
func (h *Handler) RegisterRoutes(r *gin.Engine, authMiddleware *middleware.Middleware) {
	budgetConstraints := r.Group("/api/v1/budget-constraints")
	budgetConstraints.Use(authMiddleware.AuthMiddleware())
	{
		budgetConstraints.POST("", h.createBudgetConstraint)
		budgetConstraints.GET("", h.listBudgetConstraints)
		budgetConstraints.GET("/summary", h.getBudgetConstraintSummary)
		budgetConstraints.GET("/:id", h.getBudgetConstraint)
		budgetConstraints.GET("/category/:category_id", h.getBudgetConstraintByCategory)
		budgetConstraints.PUT("/:id", h.updateBudgetConstraint)
		budgetConstraints.DELETE("/:id", h.deleteBudgetConstraint)
	}
}

// CreateBudgetConstraint godoc
// @Summary Create a new budget constraint
// @Description Create a new budget constraint for a category
// @Tags budget-constraints
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param budget_constraint body dto.CreateBudgetConstraintRequest true \"Budget constraint data\"
// @Success 201 {object} dto.BudgetConstraintResponse
// @Failure 400 {object} shared.ErrorResponse
// @Failure 401 {object} shared.ErrorResponse
// @Failure 409 {object} shared.ErrorResponse
// @Router /api/v1/budget-constraints [post]
func (h *Handler) createBudgetConstraint(c *gin.Context) {
	// Get user from context
	user, exists := middleware.GetCurrentUser(c)
	if !exists {
		shared.RespondWithError(c, http.StatusUnauthorized, "user not found in context")
		return
	}

	// Parse request
	var req dto.CreateBudgetConstraintRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		shared.RespondWithError(c, http.StatusBadRequest, "invalid request data: "+err.Error())
		return
	}

	// Create budget constraint
	bc, err := h.service.CreateBudgetConstraint(c.Request.Context(), user.ID.String(), req)
	if err != nil {
		shared.HandleError(c, err)
		return
	}

	// Convert to response
	response := dto.ToBudgetConstraintResponse(bc, true)
	shared.RespondWithSuccess(c, http.StatusCreated, "Budget constraint created successfully", response)
}

// ListBudgetConstraints godoc
// @Summary List budget constraints
// @Description Get a list of budget constraints with optional filters
// @Tags budget-constraints
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param category_id query string false \"Filter by category ID\"
// @Param is_flexible query bool false \"Filter by flexible status\"
// @Success 200 {object} dto.BudgetConstraintListResponse
// @Failure 400 {object} shared.ErrorResponse
// @Failure 401 {object} shared.ErrorResponse
// @Router /api/v1/budget-constraints [get]
func (h *Handler) listBudgetConstraints(c *gin.Context) {
	// Get user from context
	user, exists := middleware.GetCurrentUser(c)
	if !exists {
		shared.RespondWithError(c, http.StatusUnauthorized, "user not found in context")
		return
	}

	// Parse query parameters
	var query dto.ListBudgetConstraintsQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		shared.RespondWithError(c, http.StatusBadRequest, "invalid query parameters: "+err.Error())
		return
	}

	// Get budget constraints
	constraints, err := h.service.ListBudgetConstraints(c.Request.Context(), user.ID.String(), query)
	if err != nil {
		shared.HandleError(c, err)
		return
	}

	// Convert to response
	response := dto.ToBudgetConstraintListResponse(constraints, true)
	shared.RespondWithSuccess(c, http.StatusOK, "Budget constraints retrieved successfully", response)
}

// GetBudgetConstraintSummary godoc
// @Summary Get budget constraint summary
// @Description Get summary of all budget constraints including total mandatory expenses
// @Tags budget-constraints
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} dto.BudgetConstraintSummaryResponse
// @Failure 400 {object} shared.ErrorResponse
// @Failure 401 {object} shared.ErrorResponse
// @Router /api/v1/budget-constraints/summary [get]
func (h *Handler) getBudgetConstraintSummary(c *gin.Context) {
	// Get user from context
	user, exists := middleware.GetCurrentUser(c)
	if !exists {
		shared.RespondWithError(c, http.StatusUnauthorized, "user not found in context")
		return
	}

	// Get summary
	summary, err := h.service.GetBudgetConstraintSummary(c.Request.Context(), user.ID.String())
	if err != nil {
		shared.HandleError(c, err)
		return
	}

	shared.RespondWithSuccess(c, http.StatusOK, "Budget constraint summary retrieved successfully", summary)
}

// GetBudgetConstraint godoc
// @Summary Get budget constraint by ID
// @Description Get detailed information about a specific budget constraint
// @Tags budget-constraints
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true \"Budget Constraint ID\"
// @Success 200 {object} dto.BudgetConstraintResponse
// @Failure 400 {object} shared.ErrorResponse
// @Failure 401 {object} shared.ErrorResponse
// @Failure 404 {object} shared.ErrorResponse
// @Router /api/v1/budget-constraints/{id} [get]
func (h *Handler) getBudgetConstraint(c *gin.Context) {
	// Get user from context
	user, exists := middleware.GetCurrentUser(c)
	if !exists {
		shared.RespondWithError(c, http.StatusUnauthorized, "user not found in context")
		return
	}

	// Get budget constraint ID from path
	constraintID := c.Param("id")

	// Get budget constraint
	bc, err := h.service.GetBudgetConstraint(c.Request.Context(), user.ID.String(), constraintID)
	if err != nil {
		shared.HandleError(c, err)
		return
	}

	// Convert to response
	response := dto.ToBudgetConstraintResponse(bc, true)
	shared.RespondWithSuccess(c, http.StatusOK, "Budget constraint retrieved successfully", response)
}

// GetBudgetConstraintByCategory godoc
// @Summary Get budget constraint by category
// @Description Get budget constraint for a specific category
// @Tags budget-constraints
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param category_id path string true \"Category ID\"
// @Success 200 {object} dto.BudgetConstraintResponse
// @Failure 400 {object} shared.ErrorResponse
// @Failure 401 {object} shared.ErrorResponse
// @Failure 404 {object} shared.ErrorResponse
// @Router /api/v1/budget-constraints/category/{category_id} [get]
func (h *Handler) getBudgetConstraintByCategory(c *gin.Context) {
	// Get user from context
	user, exists := middleware.GetCurrentUser(c)
	if !exists {
		shared.RespondWithError(c, http.StatusUnauthorized, "user not found in context")
		return
	}

	// Get category ID from path
	categoryID := c.Param("category_id")

	// Get budget constraint
	bc, err := h.service.GetBudgetConstraintByCategory(c.Request.Context(), user.ID.String(), categoryID)
	if err != nil {
		shared.HandleError(c, err)
		return
	}

	// Convert to response
	response := dto.ToBudgetConstraintResponse(bc, true)
	shared.RespondWithSuccess(c, http.StatusOK, "Budget constraint retrieved successfully", response)
}

// UpdateBudgetConstraint godoc
// @Summary Update a budget constraint
// @Description Update an existing budget constraint
// @Tags budget-constraints
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true \"Budget Constraint ID\"
// @Param budget_constraint body dto.UpdateBudgetConstraintRequest true \"Updated budget constraint data\"
// @Success 200 {object} dto.BudgetConstraintResponse
// @Failure 400 {object} shared.ErrorResponse
// @Failure 401 {object} shared.ErrorResponse
// @Failure 403 {object} shared.ErrorResponse
// @Failure 404 {object} shared.ErrorResponse
// @Router /api/v1/budget-constraints/{id} [put]
func (h *Handler) updateBudgetConstraint(c *gin.Context) {
	// Get user from context
	user, exists := middleware.GetCurrentUser(c)
	if !exists {
		shared.RespondWithError(c, http.StatusUnauthorized, "user not found in context")
		return
	}

	// Get budget constraint ID from path
	constraintID := c.Param("id")

	// Parse request
	var req dto.UpdateBudgetConstraintRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		shared.RespondWithError(c, http.StatusBadRequest, "invalid request data: "+err.Error())
		return
	}

	// Update budget constraint
	bc, err := h.service.UpdateBudgetConstraint(c.Request.Context(), user.ID.String(), constraintID, req)
	if err != nil {
		shared.HandleError(c, err)
		return
	}

	// Convert to response
	response := dto.ToBudgetConstraintResponse(bc, true)
	shared.RespondWithSuccess(c, http.StatusOK, "Budget constraint updated successfully", response)
}

// DeleteBudgetConstraint godoc
// @Summary Delete a budget constraint
// @Description Delete a budget constraint
// @Tags budget-constraints
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true \"Budget Constraint ID\"
// @Success 200 {object} shared.Success
// @Failure 400 {object} shared.ErrorResponse
// @Failure 401 {object} shared.ErrorResponse
// @Failure 403 {object} shared.ErrorResponse
// @Failure 404 {object} shared.ErrorResponse
// @Router /api/v1/budget-constraints/{id} [delete]
func (h *Handler) deleteBudgetConstraint(c *gin.Context) {
	// Get user from context
	user, exists := middleware.GetCurrentUser(c)
	if !exists {
		shared.RespondWithError(c, http.StatusUnauthorized, "user not found in context")
		return
	}

	// Get budget constraint ID from path
	constraintID := c.Param("id")

	// Delete budget constraint
	if err := h.service.DeleteBudgetConstraint(c.Request.Context(), user.ID.String(), constraintID); err != nil {
		shared.HandleError(c, err)
		return
	}

	shared.RespondWithSuccessNoData(c, http.StatusOK, "Budget constraint deleted successfully")
}
