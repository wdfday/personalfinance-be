package handler

import (
	"net/http"

	"personalfinancedss/internal/middleware"
	"personalfinancedss/internal/module/cashflow/goal/dto"
	"personalfinancedss/internal/shared"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// ArchiveGoal godoc
// @Summary Archive goal
// @Description Archive a goal (soft delete)
// @Tags goals
// @Security BearerAuth
// @Param id path string true "Goal ID"
// @Success 200 {object} shared.SuccessResponse
// @Failure 400 {object} shared.ErrorResponse
// @Failure 401 {object} shared.ErrorResponse
// @Failure 404 {object} shared.ErrorResponse
// @Failure 500 {object} shared.ErrorResponse
// @Router /api/v1/goals/{id}/archive [put]
func (h *Handler) ArchiveGoal(c *gin.Context) {
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

	if err := h.service.ArchiveGoal(c.Request.Context(), id); err != nil {
		shared.HandleError(c, err)
		return
	}

	shared.RespondWithSuccess(c, http.StatusOK, "Goal archived successfully", struct{}{})
}

// UnarchiveGoal godoc
// @Summary Unarchive goal
// @Description Restore an archived goal
// @Tags goals
// @Security BearerAuth
// @Param id path string true "Goal ID"
// @Success 200 {object} shared.SuccessResponse
// @Failure 400 {object} shared.ErrorResponse
// @Failure 401 {object} shared.ErrorResponse
// @Failure 404 {object} shared.ErrorResponse
// @Failure 500 {object} shared.ErrorResponse
// @Router /api/v1/goals/{id}/unarchive [put]
func (h *Handler) UnarchiveGoal(c *gin.Context) {
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

	if err := h.service.UnarchiveGoal(c.Request.Context(), id); err != nil {
		shared.HandleError(c, err)
		return
	}

	shared.RespondWithSuccess(c, http.StatusOK, "Goal unarchived successfully", struct{}{})
}

// GetArchivedGoals godoc
// @Summary Get archived goals
// @Description Get all archived goals for the authenticated user
// @Tags goals
// @Produce json
// @Security BearerAuth
// @Success 200 {object} shared.SuccessResponse{data=[]dto.GoalResponse}
// @Failure 401 {object} shared.ErrorResponse
// @Failure 500 {object} shared.ErrorResponse
// @Router /api/v1/goals/archived [get]
func (h *Handler) GetArchivedGoals(c *gin.Context) {
	user, exists := middleware.GetCurrentUser(c)
	if !exists {
		shared.RespondWithError(c, http.StatusUnauthorized, "user not found in context")
		return
	}

	goals, err := h.service.GetArchivedGoals(c.Request.Context(), user.ID)
	if err != nil {
		shared.HandleError(c, err)
		return
	}

	shared.RespondWithSuccess(c, http.StatusOK, "Archived goals retrieved successfully", dto.ToGoalResponseList(goals))
}
