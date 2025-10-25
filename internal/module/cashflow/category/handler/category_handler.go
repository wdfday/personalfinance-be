package handler

import (
	"net/http"

	"personalfinancedss/internal/middleware"
	"personalfinancedss/internal/module/cashflow/category/dto"
	"personalfinancedss/internal/module/cashflow/category/service"
	"personalfinancedss/internal/shared"

	"github.com/gin-gonic/gin"
)

// Handler handles category-related HTTP requests
type Handler struct {
	service service.Service
}

// NewHandler creates a new category handler
func NewHandler(service service.Service) *Handler {
	return &Handler{
		service: service,
	}
}

// RegisterRoutes registers all category routes
func (h *Handler) RegisterRoutes(r *gin.Engine, authMiddleware *middleware.Middleware) {
	categories := r.Group("/api/v1/categories")
	categories.Use(authMiddleware.AuthMiddleware())
	{
		categories.POST("", h.createCategory)
		categories.GET("", h.listCategories)
		categories.GET("/tree", h.listCategoriesWithChildren)
		categories.GET("/:id", h.getCategory)
		categories.GET("/:id/stats", h.getCategoryStats)
		categories.PUT("/:id", h.updateCategory)
		categories.DELETE("/:id", h.deleteCategory)
	}
}

// CreateCategory godoc
// @Summary Create a new category
// @Description Create a new transaction category
// @Tags categories
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param category body dto.CreateCategoryRequest true "Category data"
// @Success 201 {object} dto.CategoryResponse
// @Failure 400 {object} shared.ErrorResponse
// @Failure 401 {object} shared.ErrorResponse
// @Failure 409 {object} shared.ErrorResponse
// @Router /api/v1/categories [post]
func (h *Handler) createCategory(c *gin.Context) {
	// Get user from context
	user, exists := middleware.GetCurrentUser(c)
	if !exists {
		shared.RespondWithError(c, http.StatusUnauthorized, "user not found in context")
		return
	}

	// Parse request
	var req dto.CreateCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		shared.RespondWithError(c, http.StatusBadRequest, "invalid request data: "+err.Error())
		return
	}

	// Create category
	category, err := h.service.CreateCategory(c.Request.Context(), user.ID.String(), req)
	if err != nil {
		shared.HandleError(c, err)
		return
	}

	// Convert to response
	response := dto.ToCategoryResponse(category, false)
	shared.RespondWithSuccess(c, http.StatusCreated, "Category created successfully", response)
}

// ListCategories godoc
// @Summary List categories
// @Description Get a list of categories with optional filters
// @Tags categories
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param type query string false "Filter by type (income, expense, both)"
// @Param parent_id query string false "Filter by parent ID"
// @Param is_root_only query bool false "Only root categories"
// @Param include_stats query bool false "Include transaction statistics"
// @Param is_active query bool false "Filter by active status"
// @Success 200 {object} dto.CategoryListResponse
// @Failure 400 {object} shared.ErrorResponse
// @Failure 401 {object} shared.ErrorResponse
// @Router /api/v1/categories [get]
func (h *Handler) listCategories(c *gin.Context) {
	// Get user from context
	user, exists := middleware.GetCurrentUser(c)
	if !exists {
		shared.RespondWithError(c, http.StatusUnauthorized, "user not found in context")
		return
	}

	// Parse query parameters
	var query dto.ListCategoriesQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		shared.RespondWithError(c, http.StatusBadRequest, "invalid query parameters: "+err.Error())
		return
	}

	// Get categories
	categories, err := h.service.ListCategories(c.Request.Context(), user.ID.String(), query)
	if err != nil {
		shared.HandleError(c, err)
		return
	}

	// Convert to response
	response := dto.ToCategoryListResponse(categories, false)
	shared.RespondWithSuccess(c, http.StatusOK, "Categories retrieved successfully", response)
}

// ListCategoriesWithChildren godoc
// @Summary List categories with children (tree)
// @Description Get a hierarchical tree of categories with their children
// @Tags categories
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param type query string false "Filter by type (income, expense, both)"
// @Param is_root_only query bool false "Only root categories"
// @Param include_stats query bool false "Include transaction statistics"
// @Param is_active query bool false "Filter by active status"
// @Success 200 {object} dto.CategoryListResponse
// @Failure 400 {object} shared.ErrorResponse
// @Failure 401 {object} shared.ErrorResponse
// @Router /api/v1/categories/tree [get]
func (h *Handler) listCategoriesWithChildren(c *gin.Context) {
	// Get user from context
	user, exists := middleware.GetCurrentUser(c)
	if !exists {
		shared.RespondWithError(c, http.StatusUnauthorized, "user not found in context")
		return
	}

	// Parse query parameters
	var query dto.ListCategoriesQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		shared.RespondWithError(c, http.StatusBadRequest, "invalid query parameters: "+err.Error())
		return
	}

	// Get categories with children
	categories, err := h.service.ListCategoriesWithChildren(c.Request.Context(), user.ID.String(), query)
	if err != nil {
		shared.HandleError(c, err)
		return
	}

	// Convert to response
	response := dto.ToCategoryListResponse(categories, true)
	shared.RespondWithSuccess(c, http.StatusOK, "Category tree retrieved successfully", response)
}

// GetCategory godoc
// @Summary Get category by ID
// @Description Get detailed information about a specific category
// @Tags categories
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Category ID"
// @Success 200 {object} dto.CategoryResponse
// @Failure 400 {object} shared.ErrorResponse
// @Failure 401 {object} shared.ErrorResponse
// @Failure 404 {object} shared.ErrorResponse
// @Router /api/v1/categories/{id} [get]
func (h *Handler) getCategory(c *gin.Context) {
	// Get user from context
	user, exists := middleware.GetCurrentUser(c)
	if !exists {
		shared.RespondWithError(c, http.StatusUnauthorized, "user not found in context")
		return
	}

	// Get category ID from path
	categoryID := c.Param("id")

	// Get category
	category, err := h.service.GetCategory(c.Request.Context(), user.ID.String(), categoryID)
	if err != nil {
		shared.HandleError(c, err)
		return
	}

	// Convert to response
	response := dto.ToCategoryResponse(category, false)
	shared.RespondWithSuccess(c, http.StatusOK, "Category retrieved successfully", response)
}

// GetCategoryStats godoc
// @Summary Get category statistics
// @Description Get transaction statistics for a specific category
// @Tags categories
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Category ID"
// @Success 200 {object} dto.CategoryStatsResponse
// @Failure 400 {object} shared.ErrorResponse
// @Failure 401 {object} shared.ErrorResponse
// @Failure 404 {object} shared.ErrorResponse
// @Router /api/v1/categories/{id}/stats [get]
func (h *Handler) getCategoryStats(c *gin.Context) {
	// Get user from context
	user, exists := middleware.GetCurrentUser(c)
	if !exists {
		shared.RespondWithError(c, http.StatusUnauthorized, "user not found in context")
		return
	}

	// Get category ID from path
	categoryID := c.Param("id")

	// Get category stats
	category, err := h.service.GetCategoryStats(c.Request.Context(), user.ID.String(), categoryID)
	if err != nil {
		shared.HandleError(c, err)
		return
	}

	// Convert to stats response
	response := dto.ToCategoryStatsResponse(category)
	shared.RespondWithSuccess(c, http.StatusOK, "Category stats retrieved successfully", response)
}

// UpdateCategory godoc
// @Summary Update a category
// @Description Update an existing category
// @Tags categories
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Category ID"
// @Param category body dto.UpdateCategoryRequest true "Updated category data"
// @Success 200 {object} dto.CategoryResponse
// @Failure 400 {object} shared.ErrorResponse
// @Failure 401 {object} shared.ErrorResponse
// @Failure 403 {object} shared.ErrorResponse
// @Failure 404 {object} shared.ErrorResponse
// @Failure 409 {object} shared.ErrorResponse
// @Router /api/v1/categories/{id} [put]
func (h *Handler) updateCategory(c *gin.Context) {
	// Get user from context
	user, exists := middleware.GetCurrentUser(c)
	if !exists {
		shared.RespondWithError(c, http.StatusUnauthorized, "user not found in context")
		return
	}

	// Get category ID from path
	categoryID := c.Param("id")

	// Parse request
	var req dto.UpdateCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		shared.RespondWithError(c, http.StatusBadRequest, "invalid request data: "+err.Error())
		return
	}

	// Update category
	category, err := h.service.UpdateCategory(c.Request.Context(), user.ID.String(), categoryID, req)
	if err != nil {
		shared.HandleError(c, err)
		return
	}

	// Convert to response
	response := dto.ToCategoryResponse(category, false)
	shared.RespondWithSuccess(c, http.StatusOK, "Category updated successfully", response)
}

// DeleteCategory godoc
// @Summary Delete a category
// @Description Soft delete a category
// @Tags categories
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Category ID"
// @Success 200 {object} shared.Success
// @Failure 400 {object} shared.ErrorResponse
// @Failure 401 {object} shared.ErrorResponse
// @Failure 403 {object} shared.ErrorResponse
// @Failure 404 {object} shared.ErrorResponse
// @Router /api/v1/categories/{id} [delete]
func (h *Handler) deleteCategory(c *gin.Context) {
	// Get user from context
	user, exists := middleware.GetCurrentUser(c)
	if !exists {
		shared.RespondWithError(c, http.StatusUnauthorized, "user not found in context")
		return
	}

	// Get category ID from path
	categoryID := c.Param("id")

	// Delete category
	if err := h.service.DeleteCategory(c.Request.Context(), user.ID.String(), categoryID); err != nil {
		shared.HandleError(c, err)
		return
	}

	shared.RespondWithSuccessNoData(c, http.StatusOK, "Category deleted successfully")
}
