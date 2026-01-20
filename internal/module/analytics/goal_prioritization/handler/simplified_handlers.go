package handler

import (
	"net/http"

	"personalfinancedss/internal/module/analytics/goal_prioritization/dto"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

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
