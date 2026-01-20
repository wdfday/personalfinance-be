package handler

import (
	"net/http"

	"personalfinancedss/internal/module/analytics/goal_prioritization/dto"
	"personalfinancedss/internal/module/analytics/goal_prioritization/service"
	"personalfinancedss/internal/shared"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// Handler handles AHP HTTP requests
type Handler struct {
	service service.Service
	logger  *zap.Logger
}

// NewHandler creates a new AHP handler
func NewHandler(service service.Service, logger *zap.Logger) *Handler {
	return &Handler{
		service: service,
		logger:  logger,
	}
}

// RegisterRoutes registers AHP routes
func (h *Handler) RegisterRoutes(router *gin.Engine) {
	analytics := router.Group("/api/v1/analytics")
	{
		// Full AHP with pairwise comparisons
		analytics.POST("/goal-prioritization", h.ExecuteAHP)

		// Simplified methods
		analytics.POST("/goal-prioritization/auto", h.ExecuteAutoScoring)
		analytics.POST("/goal-prioritization/direct-rating", h.ExecuteDirectRating)

		// Auto-score with user review flow (2-step)
		analytics.POST("/goal-prioritization/auto-scores", h.GetAutoScores)
		analytics.POST("/goal-prioritization/direct-rating-with-overrides", h.ExecuteDirectRatingWithOverrides)
	}
}

// ExecuteAHP godoc
// @Summary Execute AHP Goal Prioritization
// @Description Run Analytic Hierarchy Process to prioritize financial goals
// @Tags analytics
// @Accept json
// @Produce json
// @Param input body dto.AHPInput true "AHP Input"
// @Success 200 {object} dto.AHPOutput
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/analytics/goal-prioritization [post]
func (h *Handler) ExecuteAHP(c *gin.Context) {
	var input dto.AHPInput
	if err := c.ShouldBindJSON(&input); err != nil {
		h.logger.Error("Failed to bind AHP input", zap.Error(err))
		shared.RespondWithError(c, http.StatusBadRequest, err.Error())
		return
	}

	// Get user ID from context if available
	if userID, exists := c.Get("user_id"); exists {
		input.UserID = userID.(uuid.UUID).String()
	}

	// Execute AHP
	output, err := h.service.ExecuteAHP(c.Request.Context(), &input)
	if err != nil {
		h.logger.Error("Failed to execute AHP", zap.Error(err))
		shared.HandleError(c, err)
		return
	}

	shared.RespondWithSuccess(c, http.StatusOK, "Goal prioritization executed successfully", output)
}

// GetAutoScores godoc
// @Summary Get auto-calculated scores for user review
// @Description Step 1 of auto-score flow: Calculate scores automatically, return for user review
// @Tags analytics
// @Accept json
// @Produce json
// @Param input body dto.AutoScoresRequest true "Auto Scores Request"
// @Success 200 {object} dto.AutoScoresResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/analytics/goal-prioritization/auto-scores [post]
func (h *Handler) GetAutoScores(c *gin.Context) {
	var input dto.AutoScoresRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		h.logger.Error("Failed to bind auto-scores input", zap.Error(err))
		shared.RespondWithError(c, http.StatusBadRequest, err.Error())
		return
	}

	// Get user ID from context if available
	if userID, exists := c.Get("user_id"); exists {
		input.UserID = userID.(uuid.UUID).String()
	}

	output, err := h.service.GetAutoScores(c.Request.Context(), &input)
	if err != nil {
		h.logger.Error("Failed to get auto-scores", zap.Error(err))
		shared.HandleError(c, err)
		return
	}

	shared.RespondWithSuccess(c, http.StatusOK, "Auto-scores calculated successfully", output)
}

// ExecuteDirectRatingWithOverrides godoc
// @Summary Execute goal prioritization with user-modified scores
// @Description Step 2 of auto-score flow: Execute with user-approved/modified scores
// @Tags analytics
// @Accept json
// @Produce json
// @Param input body dto.DirectRatingWithOverridesInput true "Direct Rating with Overrides Input"
// @Success 200 {object} dto.AHPOutput
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/analytics/goal-prioritization/direct-rating-with-overrides [post]
func (h *Handler) ExecuteDirectRatingWithOverrides(c *gin.Context) {
	var input dto.DirectRatingWithOverridesInput
	if err := c.ShouldBindJSON(&input); err != nil {
		h.logger.Error("Failed to bind direct rating with overrides input", zap.Error(err))
		shared.RespondWithError(c, http.StatusBadRequest, err.Error())
		return
	}

	// Get user ID from context if available
	if userID, exists := c.Get("user_id"); exists {
		input.UserID = userID.(uuid.UUID).String()
	}

	output, err := h.service.ExecuteDirectRatingWithOverrides(c.Request.Context(), &input)
	if err != nil {
		h.logger.Error("Failed to execute direct rating with overrides", zap.Error(err))
		shared.HandleError(c, err)
		return
	}

	shared.RespondWithSuccess(c, http.StatusOK, "Goal prioritization with overrides executed successfully", output)
}
