package handler

import (
	"net/http"
	"personalfinancedss/internal/middleware"
	"personalfinancedss/internal/module/notification/domain"
	notificationdto "personalfinancedss/internal/module/notification/dto"
	notificationservice "personalfinancedss/internal/module/notification/service"
	"personalfinancedss/internal/shared"

	"github.com/gin-gonic/gin"
)

// PreferenceHandler manages notification preference endpoints
type PreferenceHandler struct {
	preferenceService notificationservice.NotificationPreferenceService
}

// NewPreferenceHandler creates a new preference handler
func NewPreferenceHandler(preferenceService notificationservice.NotificationPreferenceService) *PreferenceHandler {
	return &PreferenceHandler{
		preferenceService: preferenceService,
	}
}

// RegisterRoutes registers preference routes
func (h *PreferenceHandler) RegisterRoutes(r *gin.Engine, authMiddleware *middleware.Middleware) {
	preferences := r.Group("/api/v1/notification-preferences")
	preferences.Use(authMiddleware.AuthMiddleware())
	{
		preferences.GET("", h.listPreferences)
		preferences.POST("", h.createOrUpdatePreference)
		preferences.GET("/:type", h.getPreferenceByType)
		preferences.DELETE("/:id", h.deletePreference)
	}
}

// listPreferences godoc
// @Summary List notification preferences
// @Description Get all notification preferences for authenticated user
// @Tags notification-preferences
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {array} notificationdto.NotificationPreferenceResponse
// @Failure 401 {object} shared.ErrorResponse
// @Failure 500 {object} shared.ErrorResponse
// @Router /api/v1/notification-preferences [get]
func (h *PreferenceHandler) listPreferences(c *gin.Context) {
	currentUser, exists := middleware.GetCurrentUser(c)
	if !exists {
		shared.RespondWithError(c, http.StatusUnauthorized, "user not found in context")
		return
	}

	prefs, err := h.preferenceService.ListUserPreferences(c.Request.Context(), currentUser.ID.String())
	if err != nil {
		shared.HandleError(c, err)
		return
	}

	responses := make([]notificationdto.NotificationPreferenceResponse, len(prefs))
	for i, pref := range prefs {
		responses[i] = notificationdto.ToNotificationPreferenceResponse(pref)
	}

	shared.RespondWithSuccess(c, http.StatusOK, "Preferences retrieved successfully", responses)
}

// createOrUpdatePreference godoc
// @Summary Create or update notification preference
// @Description Create or update notification preference for a specific type
// @Tags notification-preferences
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body notificationdto.CreateNotificationPreferenceRequest true "Preference data"
// @Success 200 {object} notificationdto.NotificationPreferenceResponse
// @Failure 400 {object} shared.ErrorResponse
// @Failure 401 {object} shared.ErrorResponse
// @Failure 500 {object} shared.ErrorResponse
// @Router /api/v1/notification-preferences [post]
func (h *PreferenceHandler) createOrUpdatePreference(c *gin.Context) {
	currentUser, exists := middleware.GetCurrentUser(c)
	if !exists {
		shared.RespondWithError(c, http.StatusUnauthorized, "user not found in context")
		return
	}

	var req notificationdto.CreateNotificationPreferenceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		shared.RespondWithError(c, http.StatusBadRequest, err.Error())
		return
	}

	// Convert channels
	channels := make([]domain.NotificationChannel, len(req.PreferredChannels))
	for i, ch := range req.PreferredChannels {
		channels[i] = domain.NotificationChannel(ch)
	}

	// Set defaults
	if req.Timezone == "" {
		req.Timezone = "UTC"
	}

	pref := &domain.NotificationPreference{
		UserID:            currentUser.ID,
		Type:              domain.NotificationType(req.Type),
		Enabled:           req.Enabled,
		PreferredChannels: channels,
		MinInterval:       req.MinInterval,
		QuietHoursFrom:    req.QuietHoursFrom,
		QuietHoursTo:      req.QuietHoursTo,
		Timezone:          req.Timezone,
	}

	if err := h.preferenceService.UpsertPreference(c.Request.Context(), pref); err != nil {
		shared.HandleError(c, err)
		return
	}

	shared.RespondWithSuccess(c, http.StatusOK, "Preference saved successfully", notificationdto.ToNotificationPreferenceResponse(*pref))
}

// getPreferenceByType godoc
// @Summary Get preference by notification type
// @Description Get notification preference for a specific type
// @Tags notification-preferences
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param type path string true "Notification type"
// @Success 200 {object} notificationdto.NotificationPreferenceResponse
// @Failure 401 {object} shared.ErrorResponse
// @Failure 404 {object} shared.ErrorResponse
// @Failure 500 {object} shared.ErrorResponse
// @Router /api/v1/notification-preferences/{type} [get]
func (h *PreferenceHandler) getPreferenceByType(c *gin.Context) {
	currentUser, exists := middleware.GetCurrentUser(c)
	if !exists {
		shared.RespondWithError(c, http.StatusUnauthorized, "user not found in context")
		return
	}

	notifType := c.Param("type")
	if notifType == "" {
		shared.RespondWithError(c, http.StatusBadRequest, "notification type is required")
		return
	}

	pref, err := h.preferenceService.GetUserPreferenceForType(
		c.Request.Context(),
		currentUser.ID.String(),
		domain.NotificationType(notifType),
	)
	if err != nil {
		shared.HandleError(c, err)
		return
	}

	shared.RespondWithSuccess(c, http.StatusOK, "Preference retrieved successfully", notificationdto.ToNotificationPreferenceResponse(*pref))
}

// deletePreference godoc
// @Summary Delete notification preference
// @Description Delete a notification preference
// @Tags notification-preferences
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Preference ID"
// @Success 200 {object} shared.Success
// @Failure 401 {object} shared.ErrorResponse
// @Failure 404 {object} shared.ErrorResponse
// @Failure 500 {object} shared.ErrorResponse
// @Router /api/v1/notification-preferences/{id} [delete]
func (h *PreferenceHandler) deletePreference(c *gin.Context) {
	currentUser, exists := middleware.GetCurrentUser(c)
	if !exists {
		shared.RespondWithError(c, http.StatusUnauthorized, "user not found in context")
		return
	}

	prefID := c.Param("id")
	if prefID == "" {
		shared.RespondWithError(c, http.StatusBadRequest, "preference id is required")
		return
	}

	// Verify ownership
	pref, err := h.preferenceService.GetPreference(c.Request.Context(), prefID)
	if err != nil {
		shared.HandleError(c, err)
		return
	}

	if pref.UserID != currentUser.ID {
		shared.RespondWithError(c, http.StatusForbidden, "access denied")
		return
	}

	if err := h.preferenceService.DeletePreference(c.Request.Context(), prefID); err != nil {
		shared.HandleError(c, err)
		return
	}

	shared.RespondWithSuccessNoData(c, http.StatusOK, "Preference deleted successfully")
}
