package handler

import (
	"net/http"

	"personalfinancedss/internal/middleware"
	"personalfinancedss/internal/module/cashflow/income_profile/dto"
	"personalfinancedss/internal/module/cashflow/income_profile/service"
	"personalfinancedss/internal/shared"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
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
		// CRUD operations
		incomeProfiles.POST("", h.createIncomeProfile)
		incomeProfiles.GET("", h.listIncomeProfiles)
		incomeProfiles.GET("/:id", h.getIncomeProfile)
		incomeProfiles.PUT("/:id", h.updateIncomeProfile) // Creates new version
		incomeProfiles.DELETE("/:id", h.deleteIncomeProfile)

		// Versioning endpoints
		incomeProfiles.GET("/:id/history", h.getIncomeProfileHistory)

		// Status endpoints
		incomeProfiles.GET("/active", h.getActiveIncomes)
		incomeProfiles.GET("/archived", h.getArchivedIncomes)
		incomeProfiles.GET("/recurring", h.getRecurringIncomes)

		// Actions
		incomeProfiles.POST("/:id/verify", h.verifyIncomeProfile)
		incomeProfiles.POST("/:id/archive", h.archiveIncomeProfile)
		incomeProfiles.POST("/:id/dss-metadata", h.updateDSSMetadata)
		incomeProfiles.POST("/check-ended", h.checkAndArchiveEnded)
	}
}

// CreateIncomeProfile godoc
// @Summary Create a new income profile
// @Description Create a new income profile with source, amount, frequency, and period
// @Tags income-profiles
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param income_profile body dto.CreateIncomeProfileRequest true "Income profile data"
// @Success 201 {object} dto.IncomeProfileResponse
// @Failure 400 {object} shared.ErrorResponse
// @Failure 401 {object} shared.ErrorResponse
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
// @Param status query string false "Filter by status (active, pending, ended, archived, paused)"
// @Param is_recurring query bool false "Filter by recurring status"
// @Param is_verified query bool false "Filter by verified status"
// @Param source query string false "Filter by source"
// @Param include_archived query bool false "Include archived profiles"
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

	// Convert to response with summary
	response := dto.ToIncomeProfileListResponse(profiles, true, true)
	shared.RespondWithSuccess(c, http.StatusOK, "Income profiles retrieved successfully", response)
}

// GetIncomeProfile godoc
// @Summary Get income profile by ID
// @Description Get detailed information about a specific income profile
// @Tags income-profiles
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Income Profile ID"
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
	if _, err := uuid.Parse(profileID); err != nil {
		shared.RespondWithError(c, http.StatusBadRequest, "invalid profile id")
		return
	}

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

// GetIncomeProfileHistory godoc
// @Summary Get income profile version history
// @Description Get income profile with all its version history
// @Tags income-profiles
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Income Profile ID"
// @Success 200 {object} dto.IncomeProfileWithHistoryResponse
// @Failure 400 {object} shared.ErrorResponse
// @Failure 401 {object} shared.ErrorResponse
// @Failure 404 {object} shared.ErrorResponse
// @Router /api/v1/income-profiles/{id}/history [get]
func (h *Handler) getIncomeProfileHistory(c *gin.Context) {
	// Get user from context
	user, exists := middleware.GetCurrentUser(c)
	if !exists {
		shared.RespondWithError(c, http.StatusUnauthorized, "user not found in context")
		return
	}

	// Get income profile ID from path
	profileID := c.Param("id")

	// Get income profile with history
	current, history, err := h.service.GetIncomeProfileWithHistory(c.Request.Context(), user.ID.String(), profileID)
	if err != nil {
		shared.HandleError(c, err)
		return
	}

	// Convert to response
	response := dto.ToIncomeProfileWithHistoryResponse(current, history, true)
	shared.RespondWithSuccess(c, http.StatusOK, "Income profile history retrieved successfully", response)
}

// GetActiveIncomes godoc
// @Summary Get active income profiles
// @Description Get all currently active income profiles for the user
// @Tags income-profiles
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} dto.IncomeProfileListResponse
// @Failure 401 {object} shared.ErrorResponse
// @Router /api/v1/income-profiles/active [get]
func (h *Handler) getActiveIncomes(c *gin.Context) {
	// Get user from context
	user, exists := middleware.GetCurrentUser(c)
	if !exists {
		shared.RespondWithError(c, http.StatusUnauthorized, "user not found in context")
		return
	}

	// Get active incomes
	profiles, err := h.service.GetActiveIncomes(c.Request.Context(), user.ID.String())
	if err != nil {
		shared.HandleError(c, err)
		return
	}

	// Convert to response
	response := dto.ToIncomeProfileListResponse(profiles, true, true)
	shared.RespondWithSuccess(c, http.StatusOK, "Active income profiles retrieved successfully", response)
}

// GetArchivedIncomes godoc
// @Summary Get archived income profiles
// @Description Get all archived income profiles for the user
// @Tags income-profiles
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} dto.IncomeProfileListResponse
// @Failure 401 {object} shared.ErrorResponse
// @Router /api/v1/income-profiles/archived [get]
func (h *Handler) getArchivedIncomes(c *gin.Context) {
	// Get user from context
	user, exists := middleware.GetCurrentUser(c)
	if !exists {
		shared.RespondWithError(c, http.StatusUnauthorized, "user not found in context")
		return
	}

	// Get archived incomes
	profiles, err := h.service.GetArchivedIncomes(c.Request.Context(), user.ID.String())
	if err != nil {
		shared.HandleError(c, err)
		return
	}

	// Convert to response
	response := dto.ToIncomeProfileListResponse(profiles, true, false)
	shared.RespondWithSuccess(c, http.StatusOK, "Archived income profiles retrieved successfully", response)
}

// GetRecurringIncomes godoc
// @Summary Get recurring income profiles
// @Description Get all recurring income profiles for the user
// @Tags income-profiles
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} dto.IncomeProfileListResponse
// @Failure 401 {object} shared.ErrorResponse
// @Router /api/v1/income-profiles/recurring [get]
func (h *Handler) getRecurringIncomes(c *gin.Context) {
	// Get user from context
	user, exists := middleware.GetCurrentUser(c)
	if !exists {
		shared.RespondWithError(c, http.StatusUnauthorized, "user not found in context")
		return
	}

	// Get recurring incomes
	profiles, err := h.service.GetRecurringIncomes(c.Request.Context(), user.ID.String())
	if err != nil {
		shared.HandleError(c, err)
		return
	}

	// Convert to response
	response := dto.ToIncomeProfileListResponse(profiles, true, true)
	shared.RespondWithSuccess(c, http.StatusOK, "Recurring income profiles retrieved successfully", response)
}

// UpdateIncomeProfile godoc
// @Summary Update an income profile (creates new version)
// @Description Update an income profile by creating a new version and archiving the old one
// @Tags income-profiles
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Income Profile ID"
// @Param income_profile body dto.UpdateIncomeProfileRequest true "Updated income profile data"
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

	// Update income profile (creates new version)
	newVersion, err := h.service.UpdateIncomeProfile(c.Request.Context(), user.ID.String(), profileID, req)
	if err != nil {
		shared.HandleError(c, err)
		return
	}

	// Convert to response
	response := dto.ToIncomeProfileResponse(newVersion, true)
	shared.RespondWithSuccess(c, http.StatusOK, "Income profile updated (new version created)", response)
}

// VerifyIncomeProfile godoc
// @Summary Verify an income profile
// @Description Mark an income profile as verified (user confirmed actual receipt)
// @Tags income-profiles
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Income Profile ID"
// @Param request body dto.VerifyIncomeRequest true "Verification data"
// @Success 200 {object} dto.IncomeProfileResponse
// @Failure 400 {object} shared.ErrorResponse
// @Failure 401 {object} shared.ErrorResponse
// @Failure 404 {object} shared.ErrorResponse
// @Router /api/v1/income-profiles/{id}/verify [post]
func (h *Handler) verifyIncomeProfile(c *gin.Context) {
	// Get user from context
	user, exists := middleware.GetCurrentUser(c)
	if !exists {
		shared.RespondWithError(c, http.StatusUnauthorized, "user not found in context")
		return
	}

	// Get income profile ID from path
	profileID := c.Param("id")

	// Parse request
	var req dto.VerifyIncomeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		shared.RespondWithError(c, http.StatusBadRequest, "invalid request data: "+err.Error())
		return
	}

	// Verify income profile
	ip, err := h.service.VerifyIncomeProfile(c.Request.Context(), user.ID.String(), profileID, req.Verified)
	if err != nil {
		shared.HandleError(c, err)
		return
	}

	// Convert to response
	response := dto.ToIncomeProfileResponse(ip, true)
	shared.RespondWithSuccess(c, http.StatusOK, "Income profile verification updated", response)
}

// ArchiveIncomeProfile godoc
// @Summary Archive an income profile
// @Description Manually archive an income profile
// @Tags income-profiles
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Income Profile ID"
// @Success 200 {object} dto.MessageResponse
// @Failure 400 {object} shared.ErrorResponse
// @Failure 401 {object} shared.ErrorResponse
// @Failure 404 {object} shared.ErrorResponse
// @Router /api/v1/income-profiles/{id}/archive [post]
func (h *Handler) archiveIncomeProfile(c *gin.Context) {
	// Get user from context
	user, exists := middleware.GetCurrentUser(c)
	if !exists {
		shared.RespondWithError(c, http.StatusUnauthorized, "user not found in context")
		return
	}

	// Get income profile ID from path
	profileID := c.Param("id")

	// Archive income profile
	if err := h.service.ArchiveIncomeProfile(c.Request.Context(), user.ID.String(), profileID); err != nil {
		shared.HandleError(c, err)
		return
	}

	shared.RespondWithSuccessNoData(c, http.StatusOK, "Income profile archived successfully")
}

// UpdateDSSMetadata godoc
// @Summary Update DSS analysis metadata
// @Description Update DSS (Decision Support System) analysis metadata for an income profile
// @Tags income-profiles
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Income Profile ID"
// @Param metadata body dto.UpdateDSSMetadataRequest true "DSS metadata"
// @Success 200 {object} dto.IncomeProfileResponse
// @Failure 400 {object} shared.ErrorResponse
// @Failure 401 {object} shared.ErrorResponse
// @Failure 404 {object} shared.ErrorResponse
// @Router /api/v1/income-profiles/{id}/dss-metadata [post]
func (h *Handler) updateDSSMetadata(c *gin.Context) {
	// Get user from context
	user, exists := middleware.GetCurrentUser(c)
	if !exists {
		shared.RespondWithError(c, http.StatusUnauthorized, "user not found in context")
		return
	}

	// Get income profile ID from path
	profileID := c.Param("id")

	// Parse request
	var req dto.UpdateDSSMetadataRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		shared.RespondWithError(c, http.StatusBadRequest, "invalid request data: "+err.Error())
		return
	}

	// Update DSS metadata
	ip, err := h.service.UpdateDSSMetadata(c.Request.Context(), user.ID.String(), profileID, req)
	if err != nil {
		shared.HandleError(c, err)
		return
	}

	// Convert to response
	response := dto.ToIncomeProfileResponse(ip, true)
	shared.RespondWithSuccess(c, http.StatusOK, "DSS metadata updated successfully", response)
}

// CheckAndArchiveEnded godoc
// @Summary Check and archive ended incomes
// @Description Automatically check and archive income profiles that have reached their end date
// @Tags income-profiles
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]int
// @Failure 401 {object} shared.ErrorResponse
// @Router /api/v1/income-profiles/check-ended [post]
func (h *Handler) checkAndArchiveEnded(c *gin.Context) {
	// Get user from context
	user, exists := middleware.GetCurrentUser(c)
	if !exists {
		shared.RespondWithError(c, http.StatusUnauthorized, "user not found in context")
		return
	}

	// Check and archive ended incomes
	count, err := h.service.CheckAndArchiveEnded(c.Request.Context(), user.ID.String())
	if err != nil {
		shared.HandleError(c, err)
		return
	}

	shared.RespondWithSuccess(c, http.StatusOK, "Ended income profiles checked and archived", gin.H{
		"archived_count": count,
	})
}

// DeleteIncomeProfile godoc
// @Summary Delete an income profile
// @Description Soft delete an income profile
// @Tags income-profiles
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Income Profile ID"
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
