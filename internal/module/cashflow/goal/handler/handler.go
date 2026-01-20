package handler

import (
	"net/http"

	"personalfinancedss/internal/middleware"
	"personalfinancedss/internal/module/cashflow/goal/domain"
	"personalfinancedss/internal/module/cashflow/goal/dto"
	"personalfinancedss/internal/module/cashflow/goal/service"
	"personalfinancedss/internal/shared"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type Handler struct {
	service service.Service
	logger  *zap.Logger
}

// NewHandler creates a new goal handler
func NewHandler(service service.Service, logger *zap.Logger) *Handler {
	return &Handler{
		service: service,
		logger:  logger,
	}
}

// RegisterRoutes registers goal routes
func (h *Handler) RegisterRoutes(router *gin.Engine, authMiddleware *middleware.Middleware) {
	goals := router.Group("/api/v1/goals")
	goals.Use(authMiddleware.AuthMiddleware())
	{
		goals.POST("", h.CreateGoal)
		goals.GET("", h.GetUserGoals)
		goals.GET("/active", h.GetActiveGoals)
		goals.GET("/completed", h.GetCompletedGoals)
		goals.GET("/archived", h.GetArchivedGoals)
		goals.GET("/summary", h.GetGoalSummary)
		goals.GET("/:id", h.GetGoalByID)
		goals.PUT("/:id", h.UpdateGoal)
		goals.PUT("/:id/archive", h.ArchiveGoal)
		goals.PUT("/:id/unarchive", h.UnarchiveGoal)
		goals.DELETE("/:id", h.DeleteGoal)
		goals.POST("/:id/contribute", h.AddContribution)
		goals.POST("/:id/complete", h.MarkAsCompleted)
	}
}

// CreateGoal godoc
// @Summary Create a new goal
// @Description Create a new financial goal for the authenticated user
// @Tags goals
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param goal body dto.CreateGoalRequest true "Goal details"
// @Success 201 {object} dto.GoalResponse
// @Failure 400 {object} shared.ErrorResponse
// @Failure 401 {object} shared.ErrorResponse
// @Failure 500 {object} shared.ErrorResponse
// @Router /api/v1/goals [post]
func (h *Handler) CreateGoal(c *gin.Context) {
	user, exists := middleware.GetCurrentUser(c)
	if !exists {
		shared.RespondWithError(c, http.StatusUnauthorized, "user not found in context")
		return
	}

	var req dto.CreateGoalRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		shared.RespondWithError(c, http.StatusBadRequest, "invalid request data: "+err.Error())
		return
	}

	goal := &domain.Goal{
		UserID:                  user.ID,
		Name:                    req.Name,
		Description:             req.Description,
		Behavior:                req.Behavior,
		Category:                req.Category,
		Priority:                req.Priority,
		TargetAmount:            req.TargetAmount,
		Currency:                req.Currency,
		StartDate:               req.StartDate,
		TargetDate:              req.TargetDate,
		ContributionFrequency:   req.ContributionFrequency,
		AutoContribute:          req.AutoContribute,
		AutoContributeAmount:    req.AutoContributeAmount,
		AutoContributeAccountID: req.AutoContributeAccountID,
		AccountID:               req.AccountID,
		EnableReminders:         req.EnableReminders,
		ReminderFrequency:       req.ReminderFrequency,
		Notes:                   req.Notes,
		Tags:                    req.Tags,
	}

	if err := h.service.CreateGoal(c.Request.Context(), goal); err != nil {
		shared.HandleError(c, err)
		return
	}

	shared.RespondWithSuccess(c, http.StatusCreated, "Goal created successfully", dto.ToGoalResponse(goal))
}

// GetUserGoals godoc
// @Summary Get all user goals
// @Description Get all goals for the authenticated user
// @Tags goals
// @Produce json
// @Security BearerAuth
// @Success 200 {array} dto.GoalResponse
// @Failure 401 {object} shared.ErrorResponse
// @Failure 500 {object} shared.ErrorResponse
// @Router /api/v1/goals [get]
func (h *Handler) GetUserGoals(c *gin.Context) {
	user, exists := middleware.GetCurrentUser(c)
	if !exists {
		shared.RespondWithError(c, http.StatusUnauthorized, "user not found in context")
		return
	}

	goals, err := h.service.GetUserGoals(c.Request.Context(), user.ID)
	if err != nil {
		shared.HandleError(c, err)
		return
	}

	shared.RespondWithSuccess(c, http.StatusOK, "Goals retrieved successfully", dto.ToGoalResponseList(goals))
}

// GetActiveGoals godoc
// @Summary Get active goals
// @Description Get all active goals for the authenticated user
// @Tags goals
// @Produce json
// @Security BearerAuth
// @Success 200 {array} dto.GoalResponse
// @Failure 401 {object} shared.ErrorResponse
// @Failure 500 {object} shared.ErrorResponse
// @Router /api/v1/goals/active [get]
func (h *Handler) GetActiveGoals(c *gin.Context) {
	user, exists := middleware.GetCurrentUser(c)
	if !exists {
		shared.RespondWithError(c, http.StatusUnauthorized, "user not found in context")
		return
	}

	goals, err := h.service.GetActiveGoals(c.Request.Context(), user.ID)
	if err != nil {
		shared.HandleError(c, err)
		return
	}

	shared.RespondWithSuccess(c, http.StatusOK, "Active goals retrieved successfully", dto.ToGoalResponseList(goals))
}

// GetCompletedGoals godoc
// @Summary Get completed goals
// @Description Get all completed goals for the authenticated user
// @Tags goals
// @Produce json
// @Security BearerAuth
// @Success 200 {array} dto.GoalResponse
// @Failure 401 {object} shared.ErrorResponse
// @Failure 500 {object} shared.ErrorResponse
// @Router /api/v1/goals/completed [get]
func (h *Handler) GetCompletedGoals(c *gin.Context) {
	user, exists := middleware.GetCurrentUser(c)
	if !exists {
		shared.RespondWithError(c, http.StatusUnauthorized, "user not found in context")
		return
	}

	goals, err := h.service.GetCompletedGoals(c.Request.Context(), user.ID)
	if err != nil {
		shared.HandleError(c, err)
		return
	}

	shared.RespondWithSuccess(c, http.StatusOK, "Completed goals retrieved successfully", dto.ToGoalResponseList(goals))
}

// GetGoalByID godoc
// @Summary Get goal by ID
// @Description Get a specific goal by its ID
// @Tags goals
// @Produce json
// @Security BearerAuth
// @Param id path string true "Goal ID"
// @Success 200 {object} dto.GoalResponse
// @Failure 400 {object} shared.ErrorResponse
// @Failure 401 {object} shared.ErrorResponse
// @Failure 404 {object} shared.ErrorResponse
// @Router /api/v1/goals/{id} [get]
func (h *Handler) GetGoalByID(c *gin.Context) {
	_, exists := middleware.GetCurrentUser(c)
	if !exists {
		shared.RespondWithError(c, http.StatusUnauthorized, "user not found in context")
		return
	}

	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		shared.RespondWithError(c, http.StatusBadRequest, "invalid goal ID")
		return
	}

	goal, err := h.service.GetGoalByID(c.Request.Context(), id)
	if err != nil {
		shared.HandleError(c, err)
		return
	}

	shared.RespondWithSuccess(c, http.StatusOK, "Goal retrieved successfully", dto.ToGoalResponse(goal))
}

// UpdateGoal godoc
// @Summary Update goal
// @Description Update an existing goal
// @Tags goals
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Goal ID"
// @Param goal body dto.UpdateGoalRequest true "Goal details"
// @Success 200 {object} dto.GoalResponse
// @Failure 400 {object} shared.ErrorResponse
// @Failure 401 {object} shared.ErrorResponse
// @Failure 404 {object} shared.ErrorResponse
// @Failure 500 {object} shared.ErrorResponse
// @Router /api/v1/goals/{id} [put]
func (h *Handler) UpdateGoal(c *gin.Context) {
	_, exists := middleware.GetCurrentUser(c)
	if !exists {
		shared.RespondWithError(c, http.StatusUnauthorized, "user not found in context")
		return
	}

	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		shared.RespondWithError(c, http.StatusBadRequest, "invalid goal ID")
		return
	}

	var req dto.UpdateGoalRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		shared.RespondWithError(c, http.StatusBadRequest, "invalid request data: "+err.Error())
		return
	}

	goal, err := h.service.GetGoalByID(c.Request.Context(), id)
	if err != nil {
		shared.HandleError(c, err)
		return
	}

	// Apply update request to goal
	req.ApplyTo(goal)

	if err := h.service.UpdateGoal(c.Request.Context(), goal); err != nil {
		shared.HandleError(c, err)
		return
	}

	shared.RespondWithSuccess(c, http.StatusOK, "Goal updated successfully", dto.ToGoalResponse(goal))
}

// DeleteGoal godoc
// @Summary Delete goal
// @Description Delete a goal
// @Tags goals
// @Security BearerAuth
// @Param id path string true "Goal ID"
// @Success 204
// @Failure 400 {object} shared.ErrorResponse
// @Failure 401 {object} shared.ErrorResponse
// @Failure 500 {object} shared.ErrorResponse
// @Router /api/v1/goals/{id} [delete]
func (h *Handler) DeleteGoal(c *gin.Context) {
	_, exists := middleware.GetCurrentUser(c)
	if !exists {
		shared.RespondWithError(c, http.StatusUnauthorized, "user not found in context")
		return
	}

	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		shared.RespondWithError(c, http.StatusBadRequest, "invalid goal ID")
		return
	}

	if err := h.service.DeleteGoal(c.Request.Context(), id); err != nil {
		shared.HandleError(c, err)
		return
	}

	shared.RespondWithNoContent(c)
}

// AddContribution godoc
// @Summary Add contribution to goal
// @Description Add a contribution amount to a goal
// @Tags goals
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Goal ID"
// @Param contribution body dto.AddContributionRequest true "Contribution details"
// @Success 200 {object} dto.GoalResponse
// @Failure 400 {object} shared.ErrorResponse
// @Failure 401 {object} shared.ErrorResponse
// @Failure 500 {object} shared.ErrorResponse
// @Router /api/v1/goals/{id}/contribute [post]
func (h *Handler) AddContribution(c *gin.Context) {
	_, exists := middleware.GetCurrentUser(c)
	if !exists {
		shared.RespondWithError(c, http.StatusUnauthorized, "user not found in context")
		return
	}

	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		shared.RespondWithError(c, http.StatusBadRequest, "invalid goal ID")
		return
	}

	var req dto.AddContributionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		shared.RespondWithError(c, http.StatusBadRequest, "invalid request data: "+err.Error())
		return
	}

	// Default source to "manual" if not provided
	source := "manual"
	if req.Source != nil {
		source = *req.Source
	}

	goal, err := h.service.AddContribution(c.Request.Context(), id, req.Amount, req.Note, source)
	if err != nil {
		shared.HandleError(c, err)
		return
	}

	shared.RespondWithSuccess(c, http.StatusOK, "Contribution added successfully", dto.ToGoalResponse(goal))
}

// MarkAsCompleted godoc
// @Summary Mark goal as completed
// @Description Mark a goal as completed
// @Tags goals
// @Security BearerAuth
// @Param id path string true "Goal ID"
// @Success 200 {object} dto.GoalResponse
// @Failure 400 {object} shared.ErrorResponse
// @Failure 401 {object} shared.ErrorResponse
// @Failure 404 {object} shared.ErrorResponse
// @Failure 500 {object} shared.ErrorResponse
// @Router /api/v1/goals/{id}/complete [post]
func (h *Handler) MarkAsCompleted(c *gin.Context) {
	_, exists := middleware.GetCurrentUser(c)
	if !exists {
		shared.RespondWithError(c, http.StatusUnauthorized, "user not found in context")
		return
	}

	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		shared.RespondWithError(c, http.StatusBadRequest, "invalid goal ID")
		return
	}

	if err := h.service.MarkAsCompleted(c.Request.Context(), id); err != nil {
		shared.HandleError(c, err)
		return
	}

	goal, err := h.service.GetGoalByID(c.Request.Context(), id)
	if err != nil {
		shared.HandleError(c, err)
		return
	}

	shared.RespondWithSuccess(c, http.StatusOK, "Goal marked as completed", dto.ToGoalResponse(goal))
}

// GetGoalSummary godoc
// @Summary Get goal summary
// @Description Get a summary of all goals for the authenticated user
// @Tags goals
// @Produce json
// @Security BearerAuth
// @Success 200 {object} dto.GoalSummaryResponse
// @Failure 401 {object} shared.ErrorResponse
// @Failure 500 {object} shared.ErrorResponse
// @Router /api/v1/goals/summary [get]
func (h *Handler) GetGoalSummary(c *gin.Context) {
	user, exists := middleware.GetCurrentUser(c)
	if !exists {
		shared.RespondWithError(c, http.StatusUnauthorized, "user not found in context")
		return
	}

	summary, err := h.service.GetGoalSummary(c.Request.Context(), user.ID)
	if err != nil {
		shared.HandleError(c, err)
		return
	}

	shared.RespondWithSuccess(c, http.StatusOK, "Goal summary retrieved successfully", dto.ToGoalSummaryResponse(summary))
}
