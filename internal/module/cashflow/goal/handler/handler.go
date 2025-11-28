package handler

import (
	"net/http"
	"personalfinancedss/internal/module/cashflow/goal/domain"
	"personalfinancedss/internal/module/cashflow/goal/dto"
	"personalfinancedss/internal/module/cashflow/goal/service"

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
func (h *Handler) RegisterRoutes(router *gin.Engine) {
	goals := router.Group("/api/v1/goals")
	{
		goals.POST("", h.CreateGoal)
		goals.GET("", h.GetUserGoals)
		goals.GET("/active", h.GetActiveGoals)
		goals.GET("/completed", h.GetCompletedGoals)
		goals.GET("/summary", h.GetGoalSummary)
		goals.GET("/:id", h.GetGoalByID)
		goals.PUT("/:id", h.UpdateGoal)
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
// @Param goal body dto.CreateGoalRequest true "Goal details"
// @Success 201 {object} dto.GoalResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/goals [post]
func (h *Handler) CreateGoal(c *gin.Context) {
	var req dto.CreateGoalRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		return
	}

	goal := &domain.Goal{
		UserID:                  userID.(uuid.UUID),
		Name:                    req.Name,
		Description:             req.Description,
		Type:                    req.Type,
		Priority:                req.Priority,
		TargetAmount:            req.TargetAmount,
		Currency:                req.Currency,
		StartDate:               req.StartDate,
		TargetDate:              req.TargetDate,
		ContributionFrequency:   req.ContributionFrequency,
		AutoContribute:          req.AutoContribute,
		AutoContributeAmount:    req.AutoContributeAmount,
		AutoContributeAccountID: req.AutoContributeAccountID,
		LinkedAccountID:         req.LinkedAccountID,
		EnableReminders:         req.EnableReminders,
		ReminderFrequency:       req.ReminderFrequency,
		Notes:                   req.Notes,
		Tags:                    req.Tags,
	}

	if err := h.service.CreateGoal(c.Request.Context(), goal); err != nil {
		h.logger.Error("Failed to create goal", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, dto.ToGoalResponse(goal))
}

// GetUserGoals godoc
// @Summary Get all user goals
// @Description Get all goals for the authenticated user
// @Tags goals
// @Produce json
// @Success 200 {array} dto.GoalResponse
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/goals [get]
func (h *Handler) GetUserGoals(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		return
	}

	goals, err := h.service.GetUserGoals(c.Request.Context(), userID.(uuid.UUID))
	if err != nil {
		h.logger.Error("Failed to get user goals", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, dto.ToGoalResponseList(goals))
}

// GetActiveGoals godoc
// @Summary Get active goals
// @Description Get all active goals for the authenticated user
// @Tags goals
// @Produce json
// @Success 200 {array} dto.GoalResponse
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/goals/active [get]
func (h *Handler) GetActiveGoals(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		return
	}

	goals, err := h.service.GetActiveGoals(c.Request.Context(), userID.(uuid.UUID))
	if err != nil {
		h.logger.Error("Failed to get active goals", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, dto.ToGoalResponseList(goals))
}

// GetCompletedGoals godoc
// @Summary Get completed goals
// @Description Get all completed goals for the authenticated user
// @Tags goals
// @Produce json
// @Success 200 {array} dto.GoalResponse
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/goals/completed [get]
func (h *Handler) GetCompletedGoals(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		return
	}

	goals, err := h.service.GetCompletedGoals(c.Request.Context(), userID.(uuid.UUID))
	if err != nil {
		h.logger.Error("Failed to get completed goals", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, dto.ToGoalResponseList(goals))
}

// GetGoalByID godoc
// @Summary Get goal by ID
// @Description Get a specific goal by its ID
// @Tags goals
// @Produce json
// @Param id path string true "Goal ID"
// @Success 200 {object} dto.GoalResponse
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/goals/{id} [get]
func (h *Handler) GetGoalByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid goal ID"})
		return
	}

	goal, err := h.service.GetGoalByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, dto.ToGoalResponse(goal))
}

// UpdateGoal godoc
// @Summary Update goal
// @Description Update an existing goal
// @Tags goals
// @Accept json
// @Produce json
// @Param id path string true "Goal ID"
// @Param goal body dto.UpdateGoalRequest true "Goal details"
// @Success 200 {object} dto.GoalResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/goals/{id} [put]
func (h *Handler) UpdateGoal(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid goal ID"})
		return
	}

	var req dto.UpdateGoalRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	goal, err := h.service.GetGoalByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	// Update fields if provided
	if req.Name != nil {
		goal.Name = *req.Name
	}
	if req.Description != nil {
		goal.Description = req.Description
	}
	if req.Type != nil {
		goal.Type = *req.Type
	}
	if req.Priority != nil {
		goal.Priority = *req.Priority
	}
	if req.TargetAmount != nil {
		goal.TargetAmount = *req.TargetAmount
	}
	if req.Currency != nil {
		goal.Currency = *req.Currency
	}
	if req.StartDate != nil {
		goal.StartDate = *req.StartDate
	}
	if req.TargetDate != nil {
		goal.TargetDate = req.TargetDate
	}
	if req.ContributionFrequency != nil {
		goal.ContributionFrequency = req.ContributionFrequency
	}
	if req.AutoContribute != nil {
		goal.AutoContribute = *req.AutoContribute
	}
	if req.AutoContributeAmount != nil {
		goal.AutoContributeAmount = req.AutoContributeAmount
	}
	if req.AutoContributeAccountID != nil {
		goal.AutoContributeAccountID = req.AutoContributeAccountID
	}
	if req.LinkedAccountID != nil {
		goal.LinkedAccountID = req.LinkedAccountID
	}
	if req.EnableReminders != nil {
		goal.EnableReminders = *req.EnableReminders
	}
	if req.ReminderFrequency != nil {
		goal.ReminderFrequency = req.ReminderFrequency
	}
	if req.Status != nil {
		goal.Status = *req.Status
	}
	if req.Notes != nil {
		goal.Notes = req.Notes
	}
	if req.Tags != nil {
		goal.Tags = req.Tags
	}

	if err := h.service.UpdateGoal(c.Request.Context(), goal); err != nil {
		h.logger.Error("Failed to update goal", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, dto.ToGoalResponse(goal))
}

// DeleteGoal godoc
// @Summary Delete goal
// @Description Delete a goal
// @Tags goals
// @Param id path string true "Goal ID"
// @Success 204
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/goals/{id} [delete]
func (h *Handler) DeleteGoal(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid goal ID"})
		return
	}

	if err := h.service.DeleteGoal(c.Request.Context(), id); err != nil {
		h.logger.Error("Failed to delete goal", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

// AddContribution godoc
// @Summary Add contribution to goal
// @Description Add a contribution amount to a goal
// @Tags goals
// @Accept json
// @Produce json
// @Param id path string true "Goal ID"
// @Param contribution body dto.AddContributionRequest true "Contribution details"
// @Success 200 {object} dto.GoalResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/goals/{id}/contribute [post]
func (h *Handler) AddContribution(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid goal ID"})
		return
	}

	var req dto.AddContributionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.service.AddContribution(c.Request.Context(), id, req.Amount); err != nil {
		h.logger.Error("Failed to add contribution", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Fetch updated goal
	goal, err := h.service.GetGoalByID(c.Request.Context(), id)
	if err != nil {
		h.logger.Error("Failed to fetch updated goal", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, dto.ToGoalResponse(goal))
}

// MarkAsCompleted godoc
// @Summary Mark goal as completed
// @Description Mark a goal as completed
// @Tags goals
// @Param id path string true "Goal ID"
// @Success 200 {object} dto.GoalResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/goals/{id}/complete [post]
func (h *Handler) MarkAsCompleted(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid goal ID"})
		return
	}

	if err := h.service.MarkAsCompleted(c.Request.Context(), id); err != nil {
		h.logger.Error("Failed to mark goal as completed", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	goal, err := h.service.GetGoalByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, dto.ToGoalResponse(goal))
}

// GetGoalSummary godoc
// @Summary Get goal summary
// @Description Get a summary of all goals for the authenticated user
// @Tags goals
// @Produce json
// @Success 200 {object} dto.GoalSummaryResponse
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/goals/summary [get]
func (h *Handler) GetGoalSummary(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		return
	}

	summary, err := h.service.GetGoalSummary(c.Request.Context(), userID.(uuid.UUID))
	if err != nil {
		h.logger.Error("Failed to get goal summary", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, dto.ToGoalSummaryResponse(summary))
}
