package handler

import (
	"net/http"
	"personalfinancedss/internal/middleware"
	notificationdto "personalfinancedss/internal/module/notification/dto"
	notificationservice "personalfinancedss/internal/module/notification/service"
	"personalfinancedss/internal/shared"
	"strconv"

	"github.com/gin-gonic/gin"
)

// Handler manages notification endpoints
type Handler struct {
	userNotificationService notificationservice.UserNotificationService
}

// NewHandler creates a new notification handler
func NewHandler(userNotificationService notificationservice.UserNotificationService) *Handler {
	return &Handler{
		userNotificationService: userNotificationService,
	}
}

// RegisterRoutes registers notification routes
func (h *Handler) RegisterRoutes(r *gin.Engine, authMiddleware *middleware.Middleware) {
	notifications := r.Group("/api/v1/notifications")
	notifications.Use(authMiddleware.AuthMiddleware())
	{
		notifications.GET("", h.listNotifications)
		notifications.GET("/unread-count", h.getUnreadCount)
		notifications.GET("/:id", h.getNotification)
		notifications.PUT("/:id/read", h.markAsRead)
		notifications.PUT("/read-all", h.markAllAsRead)
	}
}

// listNotifications godoc
// @Summary List user notifications
// @Description Get paginated list of notifications for authenticated user
// @Tags notifications
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param limit query int false "Number of items per page (max 100)" default(20)
// @Param offset query int false "Offset for pagination" default(0)
// @Success 200 {object} notificationdto.NotificationListResponse
// @Failure 401 {object} shared.ErrorResponse
// @Failure 500 {object} shared.ErrorResponse
// @Router /api/v1/notifications [get]
func (h *Handler) listNotifications(c *gin.Context) {
	currentUser, exists := middleware.GetCurrentUser(c)
	if !exists {
		shared.RespondWithError(c, http.StatusUnauthorized, "user not found in context")
		return
	}

	// Parse query parameters
	limit := 20
	offset := 0

	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil {
			limit = l
		}
	}

	if offsetStr := c.Query("offset"); offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil {
			offset = o
		}
	}

	// Get notifications
	notifications, err := h.userNotificationService.GetNotifications(
		c.Request.Context(),
		currentUser.ID.String(),
		limit,
		offset,
	)
	if err != nil {
		shared.HandleError(c, err)
		return
	}

	// Get unread count
	unreadCount, err := h.userNotificationService.GetUnreadCount(
		c.Request.Context(),
		currentUser.ID.String(),
	)
	if err != nil {
		shared.HandleError(c, err)
		return
	}

	// Calculate page
	page := (offset / limit) + 1
	total := int64(len(notifications))

	response := notificationdto.ToNotificationListResponse(
		notifications,
		total,
		unreadCount,
		page,
		limit,
	)

	shared.RespondWithSuccess(c, http.StatusOK, "Notifications retrieved successfully", response)
}

// getUnreadCount godoc
// @Summary Get unread notification count
// @Description Get count of unread notifications for authenticated user
// @Tags notifications
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]int64
// @Failure 401 {object} shared.ErrorResponse
// @Failure 500 {object} shared.ErrorResponse
// @Router /api/v1/notifications/unread-count [get]
func (h *Handler) getUnreadCount(c *gin.Context) {
	currentUser, exists := middleware.GetCurrentUser(c)
	if !exists {
		shared.RespondWithError(c, http.StatusUnauthorized, "user not found in context")
		return
	}

	count, err := h.userNotificationService.GetUnreadCount(
		c.Request.Context(),
		currentUser.ID.String(),
	)
	if err != nil {
		shared.HandleError(c, err)
		return
	}

	shared.RespondWithSuccess(c, http.StatusOK, "Unread count retrieved successfully", gin.H{
		"unread_count": count,
	})
}

// getNotification godoc
// @Summary Get notification by ID
// @Description Get a single notification by ID for authenticated user
// @Tags notifications
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Notification ID"
// @Success 200 {object} notificationdto.NotificationResponse
// @Failure 401 {object} shared.ErrorResponse
// @Failure 404 {object} shared.ErrorResponse
// @Failure 500 {object} shared.ErrorResponse
// @Router /api/v1/notifications/{id} [get]
func (h *Handler) getNotification(c *gin.Context) {
	currentUser, exists := middleware.GetCurrentUser(c)
	if !exists {
		shared.RespondWithError(c, http.StatusUnauthorized, "user not found in context")
		return
	}

	notificationID := c.Param("id")
	if notificationID == "" {
		shared.RespondWithError(c, http.StatusBadRequest, "notification id is required")
		return
	}

	notification, err := h.userNotificationService.GetNotificationByID(
		c.Request.Context(),
		notificationID,
	)
	if err != nil {
		shared.HandleError(c, err)
		return
	}

	// Verify ownership
	if notification.UserID != currentUser.ID {
		shared.RespondWithError(c, http.StatusForbidden, "access denied")
		return
	}

	shared.RespondWithSuccess(c, http.StatusOK, "Notification retrieved successfully", notificationdto.ToNotificationResponse(*notification))
}

// markAsRead godoc
// @Summary Mark notification as read
// @Description Mark a single notification as read for authenticated user
// @Tags notifications
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Notification ID"
// @Success 200 {object} shared.Success
// @Failure 401 {object} shared.ErrorResponse
// @Failure 404 {object} shared.ErrorResponse
// @Failure 500 {object} shared.ErrorResponse
// @Router /api/v1/notifications/{id}/read [put]
func (h *Handler) markAsRead(c *gin.Context) {
	currentUser, exists := middleware.GetCurrentUser(c)
	if !exists {
		shared.RespondWithError(c, http.StatusUnauthorized, "user not found in context")
		return
	}

	notificationID := c.Param("id")
	if notificationID == "" {
		shared.RespondWithError(c, http.StatusBadRequest, "notification id is required")
		return
	}

	// Verify ownership first
	notification, err := h.userNotificationService.GetNotificationByID(
		c.Request.Context(),
		notificationID,
	)
	if err != nil {
		shared.HandleError(c, err)
		return
	}

	if notification.UserID != currentUser.ID {
		shared.RespondWithError(c, http.StatusForbidden, "access denied")
		return
	}

	// Mark as read
	if err := h.userNotificationService.MarkAsRead(c.Request.Context(), notificationID); err != nil {
		shared.HandleError(c, err)
		return
	}

	shared.RespondWithSuccessNoData(c, http.StatusOK, "Notification marked as read")
}

// markAllAsRead godoc
// @Summary Mark all notifications as read
// @Description Mark all unread notifications as read for authenticated user
// @Tags notifications
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} shared.Success
// @Failure 401 {object} shared.ErrorResponse
// @Failure 500 {object} shared.ErrorResponse
// @Router /api/v1/notifications/read-all [put]
func (h *Handler) markAllAsRead(c *gin.Context) {
	currentUser, exists := middleware.GetCurrentUser(c)
	if !exists {
		shared.RespondWithError(c, http.StatusUnauthorized, "user not found in context")
		return
	}

	if err := h.userNotificationService.MarkAllAsRead(
		c.Request.Context(),
		currentUser.ID.String(),
	); err != nil {
		shared.HandleError(c, err)
		return
	}

	shared.RespondWithSuccessNoData(c, http.StatusOK, "All notifications marked as read")
}
