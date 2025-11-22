package handler

import (
	"fmt"
	"net/http"

	"personalfinancedss/internal/middleware"
	"personalfinancedss/internal/module/investment/portfolio_snapshot/dto"
	"personalfinancedss/internal/module/investment/portfolio_snapshot/service"
	"personalfinancedss/internal/shared"

	"github.com/gin-gonic/gin"
)

// Handler handles portfolio snapshot-related HTTP requests
type Handler struct {
	service service.Service
}

// NewHandler creates a new portfolio snapshot handler
func NewHandler(service service.Service) *Handler {
	return &Handler{
		service: service,
	}
}

// RegisterRoutes registers all portfolio snapshot routes
func (h *Handler) RegisterRoutes(r *gin.Engine, authMiddleware *middleware.Middleware) {
	snapshots := r.Group("/api/v1/investment/snapshots")
	snapshots.Use(authMiddleware.AuthMiddleware())
	{
		snapshots.POST("", h.createSnapshot)
		snapshots.GET("", h.listSnapshots)
		snapshots.GET("/latest", h.getLatestSnapshot)
		snapshots.GET("/period/:period", h.getSnapshotsByPeriod)
		snapshots.GET("/performance", h.getPerformanceMetrics)
		snapshots.GET("/:id", h.getSnapshot)
		snapshots.DELETE("/:id", h.deleteSnapshot)
	}
}

// CreateSnapshot godoc
// @Summary Create a portfolio snapshot
// @Description Create a new portfolio snapshot
// @Tags portfolio-snapshots
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param snapshot body dto.CreateSnapshotRequest true "Snapshot data"
// @Success 201 {object} dto.SnapshotResponse
// @Failure 400 {object} shared.ErrorResponse
// @Failure 401 {object} shared.ErrorResponse
// @Router /api/v1/investment/snapshots [post]
func (h *Handler) createSnapshot(c *gin.Context) {
	user, exists := middleware.GetCurrentUser(c)
	if !exists {
		shared.RespondWithError(c, http.StatusUnauthorized, "user not found in context")
		return
	}

	var req dto.CreateSnapshotRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		shared.RespondWithError(c, http.StatusBadRequest, "invalid request data: "+err.Error())
		return
	}

	snapshot, err := h.service.CreateSnapshot(c.Request.Context(), user.ID.String(), req)
	if err != nil {
		shared.HandleError(c, err)
		return
	}

	response := dto.ToSnapshotResponse(snapshot)
	shared.RespondWithSuccess(c, http.StatusCreated, "Snapshot created successfully", response)
}

// ListSnapshots godoc
// @Summary List portfolio snapshots
// @Description Get a paginated list of portfolio snapshots
// @Tags portfolio-snapshots
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param snapshot_type query string false "Filter by snapshot type"
// @Param period query string false "Filter by period"
// @Param start_date query string false "Start date (YYYY-MM-DD)"
// @Param end_date query string false "End date (YYYY-MM-DD)"
// @Param page query int false "Page number (default: 1)"
// @Param page_size query int false "Page size (default: 30, max: 100)"
// @Success 200 {object} dto.SnapshotListResponse
// @Failure 400 {object} shared.ErrorResponse
// @Failure 401 {object} shared.ErrorResponse
// @Router /api/v1/investment/snapshots [get]
func (h *Handler) listSnapshots(c *gin.Context) {
	user, exists := middleware.GetCurrentUser(c)
	if !exists {
		shared.RespondWithError(c, http.StatusUnauthorized, "user not found in context")
		return
	}

	var query dto.ListSnapshotsQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		shared.RespondWithError(c, http.StatusBadRequest, "invalid query parameters: "+err.Error())
		return
	}

	response, err := h.service.ListSnapshots(c.Request.Context(), user.ID.String(), query)
	if err != nil {
		shared.HandleError(c, err)
		return
	}

	shared.RespondWithSuccess(c, http.StatusOK, "Snapshots retrieved successfully", response)
}

// GetSnapshot godoc
// @Summary Get snapshot by ID
// @Description Get detailed information about a specific portfolio snapshot
// @Tags portfolio-snapshots
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Snapshot ID"
// @Success 200 {object} dto.SnapshotResponse
// @Failure 400 {object} shared.ErrorResponse
// @Failure 401 {object} shared.ErrorResponse
// @Failure 404 {object} shared.ErrorResponse
// @Router /api/v1/investment/snapshots/{id} [get]
func (h *Handler) getSnapshot(c *gin.Context) {
	user, exists := middleware.GetCurrentUser(c)
	if !exists {
		shared.RespondWithError(c, http.StatusUnauthorized, "user not found in context")
		return
	}

	snapshotID := c.Param("id")

	snapshot, err := h.service.GetSnapshot(c.Request.Context(), user.ID.String(), snapshotID)
	if err != nil {
		shared.HandleError(c, err)
		return
	}

	response := dto.ToSnapshotResponse(snapshot)
	shared.RespondWithSuccess(c, http.StatusOK, "Snapshot retrieved successfully", response)
}

// GetLatestSnapshot godoc
// @Summary Get latest snapshot
// @Description Get the most recent portfolio snapshot
// @Tags portfolio-snapshots
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} dto.SnapshotResponse
// @Failure 401 {object} shared.ErrorResponse
// @Failure 404 {object} shared.ErrorResponse
// @Router /api/v1/investment/snapshots/latest [get]
func (h *Handler) getLatestSnapshot(c *gin.Context) {
	user, exists := middleware.GetCurrentUser(c)
	if !exists {
		shared.RespondWithError(c, http.StatusUnauthorized, "user not found in context")
		return
	}

	snapshot, err := h.service.GetLatestSnapshot(c.Request.Context(), user.ID.String())
	if err != nil {
		shared.HandleError(c, err)
		return
	}

	if snapshot == nil {
		shared.RespondWithError(c, http.StatusNotFound, "no snapshots found")
		return
	}

	response := dto.ToSnapshotResponse(snapshot)
	shared.RespondWithSuccess(c, http.StatusOK, "Latest snapshot retrieved successfully", response)
}

// GetSnapshotsByPeriod godoc
// @Summary Get snapshots by period
// @Description Get snapshots for a specific period (daily, weekly, monthly)
// @Tags portfolio-snapshots
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param period path string true "Period (daily, weekly, monthly, yearly)"
// @Param limit query int false "Limit (default: 30, max: 365)"
// @Success 200 {array} dto.SnapshotResponse
// @Failure 400 {object} shared.ErrorResponse
// @Failure 401 {object} shared.ErrorResponse
// @Router /api/v1/investment/snapshots/period/{period} [get]
func (h *Handler) getSnapshotsByPeriod(c *gin.Context) {
	user, exists := middleware.GetCurrentUser(c)
	if !exists {
		shared.RespondWithError(c, http.StatusUnauthorized, "user not found in context")
		return
	}

	period := c.Param("period")
	limit := 30
	if l := c.Query("limit"); l != "" {
		fmt.Sscanf(l, "%d", &limit)
	}

	snapshots, err := h.service.GetSnapshotsByPeriod(c.Request.Context(), user.ID.String(), period, limit)
	if err != nil {
		shared.HandleError(c, err)
		return
	}

	responses := make([]dto.SnapshotResponse, 0, len(snapshots))
	for _, snapshot := range snapshots {
		responses = append(responses, dto.ToSnapshotResponse(snapshot))
	}

	shared.RespondWithSuccess(c, http.StatusOK, "Snapshots retrieved successfully", responses)
}

// GetPerformanceMetrics godoc
// @Summary Get performance metrics
// @Description Get portfolio performance metrics over a time period
// @Tags portfolio-snapshots
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param start_date query string true "Start date (YYYY-MM-DD)"
// @Param end_date query string true "End date (YYYY-MM-DD)"
// @Success 200 {object} dto.PerformanceMetrics
// @Failure 400 {object} shared.ErrorResponse
// @Failure 401 {object} shared.ErrorResponse
// @Router /api/v1/investment/snapshots/performance [get]
func (h *Handler) getPerformanceMetrics(c *gin.Context) {
	user, exists := middleware.GetCurrentUser(c)
	if !exists {
		shared.RespondWithError(c, http.StatusUnauthorized, "user not found in context")
		return
	}

	startDate := c.Query("start_date")
	endDate := c.Query("end_date")

	if startDate == "" || endDate == "" {
		shared.RespondWithError(c, http.StatusBadRequest, "start_date and end_date are required")
		return
	}

	metrics, err := h.service.GetPerformanceMetrics(c.Request.Context(), user.ID.String(), startDate, endDate)
	if err != nil {
		shared.HandleError(c, err)
		return
	}

	shared.RespondWithSuccess(c, http.StatusOK, "Performance metrics retrieved successfully", metrics)
}

// DeleteSnapshot godoc
// @Summary Delete a portfolio snapshot
// @Description Soft delete a portfolio snapshot
// @Tags portfolio-snapshots
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Snapshot ID"
// @Success 200 {object} shared.Success
// @Failure 400 {object} shared.ErrorResponse
// @Failure 401 {object} shared.ErrorResponse
// @Failure 404 {object} shared.ErrorResponse
// @Router /api/v1/investment/snapshots/{id} [delete]
func (h *Handler) deleteSnapshot(c *gin.Context) {
	user, exists := middleware.GetCurrentUser(c)
	if !exists {
		shared.RespondWithError(c, http.StatusUnauthorized, "user not found in context")
		return
	}

	snapshotID := c.Param("id")

	if err := h.service.DeleteSnapshot(c.Request.Context(), user.ID.String(), snapshotID); err != nil {
		shared.HandleError(c, err)
		return
	}

	shared.RespondWithSuccessNoData(c, http.StatusOK, "Snapshot deleted successfully")
}
