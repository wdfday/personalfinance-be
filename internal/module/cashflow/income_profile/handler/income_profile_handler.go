package handler

import (
	"net/http"
	"strconv"

	"personalfinancedss/internal/middleware"
	"personalfinancedss/internal/module/cashflow/income_profile/dto"
	"personalfinancedss/internal/module/cashflow/income_profile/service"
	"personalfinancedss/internal/shared"

	"github.com/gin-gonic/gin"
)

// Handler handles income profile-related HTTP requests
type Handler struct {
	service service.Service
}

// NewHandler creates a new income profile handler
func NewHandler(service service.Service) *Handler {
	return &Handler{
		service: service,
	}
}

// RegisterRoutes registers all income profile routes
func (h *Handler) RegisterRoutes(r *gin.Engine, authMiddleware *middleware.Middleware) {
	incomeProfiles := r.Group("/api/v1/income-profiles")
	incomeProfiles.Use(authMiddleware.AuthMiddleware())
	{
		incomeProfiles.POST("", h.createIncomeProfile)
		incomeProfiles.GET("", h.listIncomeProfiles)
		incomeProfiles.GET("/:id", h.getIncomeProfile)
		incomeProfiles.GET("/:year/:month", h.getIncomeProfileByPeriod)
		incomeProfiles.PUT("/:id", h.updateIncomeProfile)
		incomeProfiles.DELETE("/:id", h.deleteIncomeProfile)
	}
}

// CreateIncomeProfile godoc
// @Summary Create a new income profile
// @Description Create a new income profile for a specific month
// @Tags income-profiles
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param income_profile body dto.CreateIncomeProfileRequest true \"Income profile data\"
// @Success 201 {object} dto.IncomeProfileResponse
// @Failure 400 {object} shared.ErrorResponse
// @Failure 401 {object} shared.ErrorResponse
// @Failure 409 {object} shared.ErrorResponse
// @Router /api/v1/income-profiles [post]
func (h *Handler) createIncomeProfile(c *gin.Context) {
	// Get user from context
	user, exists := middleware.GetCurrentUser(c)
	if !exists {
		shared.RespondWithError(c, http.StatusUnauthorized, "user not found in context")
		return
	}

	// Parse request
	var req dto.CreateIncomeProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		shared.RespondWithError(c, http.StatusBadRequest, "invalid request data: "+err.Error())
		return
	}

	// Create income profile
	ip, err := h.service.CreateIncomeProfile(c.Request.Context(), user.ID.String(), req)
	if err != nil {
		shared.HandleError(c, err)
		return
	}

	// Convert to response
	response := dto.ToIncomeProfileResponse(ip, true)
	shared.RespondWithSuccess(c, http.StatusCreated, "Income profile created successfully", response)
}

// ListIncomeProfiles godoc
// @Summary List income profiles
// @Description Get a list of income profiles with optional filters
// @Tags income-profiles
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param year query int false \"Filter by year\"
// @Param is_actual query bool false \"Filter by actual/projected status\"
// @Success 200 {object} dto.IncomeProfileListResponse
// @Failure 400 {object} shared.ErrorResponse
// @Failure 401 {object} shared.ErrorResponse
// @Router /api/v1/income-profiles [get]
func (h *Handler) listIncomeProfiles(c *gin.Context) {
	// Get user from context
	user, exists := middleware.GetCurrentUser(c)
	if !exists {
		shared.RespondWithError(c, http.StatusUnauthorized, "user not found in context")
		return
	}

	// Parse query parameters
	var query dto.ListIncomeProfilesQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		shared.RespondWithError(c, http.StatusBadRequest, "invalid query parameters: "+err.Error())
		return
	}

	// Get income profiles
	profiles, err := h.service.ListIncomeProfiles(c.Request.Context(), user.ID.String(), query)
	if err != nil {
		shared.HandleError(c, err)
		return
	}

	// Convert to response
	response := dto.ToIncomeProfileListResponse(profiles, true)
	shared.RespondWithSuccess(c, http.StatusOK, "Income profiles retrieved successfully", response)
}

// GetIncomeProfile godoc
// @Summary Get income profile by ID
// @Description Get detailed information about a specific income profile
// @Tags income-profiles
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true \"Income Profile ID\"
// @Success 200 {object} dto.IncomeProfileResponse
// @Failure 400 {object} shared.ErrorResponse
// @Failure 401 {object} shared.ErrorResponse
// @Failure 404 {object} shared.ErrorResponse
// @Router /api/v1/income-profiles/{id} [get]
func (h *Handler) getIncomeProfile(c *gin.Context) {
	// Get user from context
	user, exists := middleware.GetCurrentUser(c)
	if !exists {
		shared.RespondWithError(c, http.StatusUnauthorized, "user not found in context")
		return
	}

	// Get income profile ID from path
	profileID := c.Param("id")

	// Get income profile
	ip, err := h.service.GetIncomeProfile(c.Request.Context(), user.ID.String(), profileID)
	if err != nil {
		shared.HandleError(c, err)
		return
	}

	// Convert to response
	response := dto.ToIncomeProfileResponse(ip, true)
	shared.RespondWithSuccess(c, http.StatusOK, "Income profile retrieved successfully", response)
}

// GetIncomeProfileByPeriod godoc
// @Summary Get income profile by period
// @Description Get income profile for a specific year and month
// @Tags income-profiles
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param year path int true \"Year\"
// @Param month path int true \"Month (1-12)\"
// @Success 200 {object} dto.IncomeProfileResponse
// @Failure 400 {object} shared.ErrorResponse
// @Failure 401 {object} shared.ErrorResponse
// @Failure 404 {object} shared.ErrorResponse
// @Router /api/v1/income-profiles/{year}/{month} [get]
func (h *Handler) getIncomeProfileByPeriod(c *gin.Context) {
	// Get user from context
	user, exists := middleware.GetCurrentUser(c)
	if !exists {
		shared.RespondWithError(c, http.StatusUnauthorized, "user not found in context")
		return
	}

	// Parse year and month from path
	year, err := strconv.Atoi(c.Param("year"))
	if err != nil {
		shared.RespondWithError(c, http.StatusBadRequest, "invalid year parameter")
		return
	}

	month, err := strconv.Atoi(c.Param("month"))
	if err != nil {
		shared.RespondWithError(c, http.StatusBadRequest, "invalid month parameter")
		return
	}

	// Get income profile
	ip, err := h.service.GetIncomeProfileByPeriod(c.Request.Context(), user.ID.String(), year, month)
	if err != nil {
		shared.HandleError(c, err)
		return
	}

	// Convert to response
	response := dto.ToIncomeProfileResponse(ip, true)
	shared.RespondWithSuccess(c, http.StatusOK, "Income profile retrieved successfully", response)
}

// UpdateIncomeProfile godoc
// @Summary Update an income profile
// @Description Update an existing income profile
// @Tags income-profiles
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true \"Income Profile ID\"
// @Param income_profile body dto.UpdateIncomeProfileRequest true \"Updated income profile data\"
// @Success 200 {object} dto.IncomeProfileResponse
// @Failure 400 {object} shared.ErrorResponse
// @Failure 401 {object} shared.ErrorResponse
// @Failure 403 {object} shared.ErrorResponse
// @Failure 404 {object} shared.ErrorResponse
// @Router /api/v1/income-profiles/{id} [put]
func (h *Handler) updateIncomeProfile(c *gin.Context) {
	// Get user from context
	user, exists := middleware.GetCurrentUser(c)
	if !exists {
		shared.RespondWithError(c, http.StatusUnauthorized, "user not found in context")
		return
	}

	// Get income profile ID from path
	profileID := c.Param("id")

	// Parse request
	var req dto.UpdateIncomeProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		shared.RespondWithError(c, http.StatusBadRequest, "invalid request data: "+err.Error())
		return
	}

	// Update income profile
	ip, err := h.service.UpdateIncomeProfile(c.Request.Context(), user.ID.String(), profileID, req)
	if err != nil {
		shared.HandleError(c, err)
		return
	}

	// Convert to response
	response := dto.ToIncomeProfileResponse(ip, true)
	shared.RespondWithSuccess(c, http.StatusOK, "Income profile updated successfully", response)
}

// DeleteIncomeProfile godoc
// @Summary Delete an income profile
// @Description Delete an income profile
// @Tags income-profiles
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true \"Income Profile ID\"
// @Success 200 {object} shared.Success
// @Failure 400 {object} shared.ErrorResponse
// @Failure 401 {object} shared.ErrorResponse
// @Failure 403 {object} shared.ErrorResponse
// @Failure 404 {object} shared.ErrorResponse
// @Router /api/v1/income-profiles/{id} [delete]
func (h *Handler) deleteIncomeProfile(c *gin.Context) {
	// Get user from context
	user, exists := middleware.GetCurrentUser(c)
	if !exists {
		shared.RespondWithError(c, http.StatusUnauthorized, "user not found in context")
		return
	}

	// Get income profile ID from path
	profileID := c.Param("id")

	// Delete income profile
	if err := h.service.DeleteIncomeProfile(c.Request.Context(), user.ID.String(), profileID); err != nil {
		shared.HandleError(c, err)
		return
	}

	shared.RespondWithSuccessNoData(c, http.StatusOK, "Income profile deleted successfully")
}
