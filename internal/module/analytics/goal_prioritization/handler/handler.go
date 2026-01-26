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

// ExecuteAutoScoring godoc
// @Summary Execute Auto-Scoring AHP
// @Description Automatically score and prioritize goals using AI-based criteria calculation
// @Tags analytics
// @Accept json
// @Produce json
// @Param input body dto.AutoScoringInput true \"Auto-Scoring Input\"
// @Success 200 {object} dto.AHPOutput
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/analytics/goal-prioritization/auto [post]
func (h *Handler) ExecuteAutoScoring(c *gin.Context) {
	var input dto.AutoScoringInput
	if err := c.ShouldBindJSON(&input); err != nil {
		h.logger.Error("Failed to bind auto-scoring input", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid input: " + err.Error(),
		})
		return
	}

	// Get user ID from context if available
	if userID, exists := c.Get("user_id"); exists {
		input.UserID = userID.(uuid.UUID).String()
	}

	h.logger.Info("Executing auto-scoring AHP",
		zap.String("user_id", input.UserID),
		zap.Int("goal_count", len(input.Goals)))

	// Execute auto-scoring
	output, err := h.service.ExecuteAutoScoring(c.Request.Context(), &input)
	if err != nil {
		h.logger.Error("Failed to execute auto-scoring AHP", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to prioritize goals: " + err.Error(),
		})
		return
	}

	h.logger.Info("Auto-scoring AHP completed successfully",
		zap.String("user_id", input.UserID),
		zap.Int("goals_ranked", len(output.Ranking)))

	c.JSON(http.StatusOK, output)
}

// ExecuteDirectRating godoc
// @Summary Execute Direct Rating AHP
// @Description Prioritize goals using simplified direct rating method (18 inputs vs 33)
// @Tags analytics
// @Accept json
// @Produce json
// @Param input body dto.DirectRatingInput true \"Direct Rating Input\"
// @Success 200 {object} dto.AHPOutput
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/analytics/goal-prioritization/direct-rating [post]
func (h *Handler) ExecuteDirectRating(c *gin.Context) {
	var input dto.DirectRatingInput
	if err := c.ShouldBindJSON(&input); err != nil {
		h.logger.Error("Failed to bind direct rating input", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid input: " + err.Error(),
		})
		return
	}

	// Get user ID from context if available
	if userID, exists := c.Get("user_id"); exists {
		input.UserID = userID.(uuid.UUID).String()
	}

	h.logger.Info("Executing direct rating AHP",
		zap.String("user_id", input.UserID),
		zap.Int("goal_count", len(input.Goals)))

	// Validate criteria ratings
	if err := input.Validate(); err != nil {
		h.logger.Error("Direct rating validation failed", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Validation failed: " + err.Error(),
		})
		return
	}

	// Execute direct rating
	output, err := h.service.ExecuteDirectRating(c.Request.Context(), &input)
	if err != nil {
		h.logger.Error("Failed to execute direct rating AHP", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to prioritize goals: " + err.Error(),
		})
		return
	}

	h.logger.Info("Direct rating AHP completed successfully",
		zap.String("user_id", input.UserID),
		zap.Int("goals_ranked", len(output.Ranking)))

	c.JSON(http.StatusOK, output)
}
