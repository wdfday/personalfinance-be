package handler

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"personalfinancedss/internal/middleware"
	eventdomain "personalfinancedss/internal/module/calendar/event/domain"
	"personalfinancedss/internal/module/calendar/event/dto"
	eventservice "personalfinancedss/internal/module/calendar/event/service"
	"personalfinancedss/internal/shared"
)

// Handler manages user calendar event endpoints.
type Handler struct {
	service eventservice.Service
	logger  *zap.Logger
}

// NewHandler builds a new event handler.
func NewHandler(service eventservice.Service, logger *zap.Logger) *Handler {
	return &Handler{
		service: service,
		logger:  logger.Named("calendar.event.handler"),
	}
}

// RegisterRoutes wires calendar event routes under /api/v1/calendar/events.
func (h *Handler) RegisterRoutes(router *gin.Engine, auth *middleware.Middleware) {
	group := router.Group("/api/v1/calendar/events")
	group.Use(auth.AuthMiddleware(middleware.WithEmailVerified()))

	// Event CRUD
	group.POST("", h.createEvent)
	group.PUT("/:id", h.updateEvent)
	group.GET("/:id", h.getEvent)
	group.DELETE("/:id", h.deleteEvent)

	// Calendar views
	group.GET("", h.listEvents)                  // Main calendar view with date range
	group.GET("/upcoming", h.listUpcomingEvents) // Legacy endpoint

	// Holiday management
	group.POST("/holidays/generate", h.generateHolidays)
}

type createEventRequest struct {
	Name        string  `json:"name" binding:"required"`
	Description *string `json:"description"`
	Type        *string `json:"type"`                          // personal, birthday, etc.
	StartDate   string  `json:"start_date" binding:"required"` // ISO 8601 format
	EndDate     *string `json:"end_date"`
	AllDay      *bool   `json:"all_day"`
	Color       *string `json:"color"`
	Tags        *string `json:"tags"` // JSON array string
}

type updateEventRequest struct {
	Name        *string `json:"name"`
	Description *string `json:"description"`
	Type        *string `json:"type"`
	StartDate   *string `json:"start_date"`
	EndDate     *string `json:"end_date"`
	AllDay      *bool   `json:"all_day"`
	Color       *string `json:"color"`
	Tags        *string `json:"tags"`
}

type generateHolidaysRequest struct {
	Year *int `json:"year"` // Optional, defaults to current year
}

func (h *Handler) createEvent(c *gin.Context) {
	var req createEventRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		shared.RespondWithError(c, http.StatusBadRequest, err.Error())
		return
	}

	currentUser, exists := middleware.GetCurrentUser(c)
	if !exists {
		shared.RespondWithError(c, http.StatusUnauthorized, "user not found in context")
		return
	}

	// Parse start date (supports both date and datetime)
	start, err := parseDateTime(req.StartDate)
	if err != nil {
		shared.RespondWithError(c, http.StatusBadRequest, "invalid start_date format, expected ISO 8601 (e.g., 2025-03-15 or 2025-03-15T14:30:00Z)")
		return
	}

	var endPtr *time.Time
	if req.EndDate != nil && *req.EndDate != "" {
		end, err := parseDateTime(*req.EndDate)
		if err != nil {
			shared.RespondWithError(c, http.StatusBadRequest, "invalid end_date format")
			return
		}
		endPtr = &end
	}

	// Parse type
	eventType := eventdomain.EventTypePersonal
	if req.Type != nil {
		eventType = eventdomain.EventType(*req.Type)
	}

	// Parse all_day
	allDay := true
	if req.AllDay != nil {
		allDay = *req.AllDay
	}

	// Parse tags
	var tagsBytes []byte
	if req.Tags != nil && *req.Tags != "" {
		tagsBytes = []byte(*req.Tags)
	}

	event := &eventdomain.Event{
		ID:          uuid.New(),
		UserID:      currentUser.ID,
		Name:        req.Name,
		Description: req.Description,
		Type:        eventType,
		Source:      eventdomain.SourceUserCreated,
		StartDate:   start,
		EndDate:     endPtr,
		AllDay:      allDay,
		Color:       req.Color,
		Tags:        tagsBytes,
	}

	if err := h.service.CreateEvent(c.Request.Context(), event); err != nil {
		h.logger.Error("failed to create event", zap.Error(err))
		shared.RespondWithError(c, http.StatusBadRequest, err.Error())
		return
	}

	shared.RespondWithSuccess(c, http.StatusCreated, "Event created successfully", dto.ToEventResponse(event))
}

func (h *Handler) updateEvent(c *gin.Context) {
	currentUser, exists := middleware.GetCurrentUser(c)
	if !exists {
		shared.RespondWithError(c, http.StatusUnauthorized, "user not found in context")
		return
	}

	eventID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		shared.RespondWithError(c, http.StatusBadRequest, "invalid event id")
		return
	}

	// Get existing event
	event, err := h.service.GetEvent(c.Request.Context(), eventID, currentUser.ID)
	if err != nil {
		shared.HandleError(c, err)
		return
	}

	var req updateEventRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		shared.RespondWithError(c, http.StatusBadRequest, err.Error())
		return
	}

	// Apply updates
	if req.Name != nil {
		event.Name = *req.Name
	}
	if req.Description != nil {
		event.Description = req.Description
	}
	if req.Type != nil {
		event.Type = eventdomain.EventType(*req.Type)
	}
	if req.StartDate != nil {
		start, err := parseDateTime(*req.StartDate)
		if err != nil {
			shared.RespondWithError(c, http.StatusBadRequest, "invalid start_date format")
			return
		}
		event.StartDate = start
	}
	if req.EndDate != nil {
		if *req.EndDate == "" {
			event.EndDate = nil
		} else {
			end, err := parseDateTime(*req.EndDate)
			if err != nil {
				shared.RespondWithError(c, http.StatusBadRequest, "invalid end_date format")
				return
			}
			event.EndDate = &end
		}
	}
	if req.AllDay != nil {
		event.AllDay = *req.AllDay
	}
	if req.Color != nil {
		event.Color = req.Color
	}
	if req.Tags != nil {
		if *req.Tags == "" {
			event.Tags = nil
		} else {
			event.Tags = []byte(*req.Tags)
		}
	}

	if err := h.service.UpdateEvent(c.Request.Context(), event); err != nil {
		h.logger.Error("failed to update event", zap.Error(err))
		shared.RespondWithError(c, http.StatusBadRequest, err.Error())
		return
	}

	shared.RespondWithSuccess(c, http.StatusOK, "Event updated successfully", dto.ToEventResponse(event))
}

func (h *Handler) getEvent(c *gin.Context) {
	currentUser, exists := middleware.GetCurrentUser(c)
	if !exists {
		shared.RespondWithError(c, http.StatusUnauthorized, "user not found in context")
		return
	}

	eventID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		shared.RespondWithError(c, http.StatusBadRequest, "invalid event id")
		return
	}

	event, err := h.service.GetEvent(c.Request.Context(), eventID, currentUser.ID)
	if err != nil {
		shared.HandleError(c, err)
		return
	}

	shared.RespondWithSuccess(c, http.StatusOK, "Event retrieved successfully", dto.ToEventResponse(event))
}

// listEvents is the main calendar view endpoint
// GET /api/v1/calendar/events?from=2025-03-01&to=2025-03-31
func (h *Handler) listEvents(c *gin.Context) {
	currentUser, exists := middleware.GetCurrentUser(c)
	if !exists {
		shared.RespondWithError(c, http.StatusUnauthorized, "user not found in context")
		return
	}

	// Parse date range
	fromStr := c.Query("from")
	toStr := c.Query("to")

	if fromStr == "" || toStr == "" {
		shared.RespondWithError(c, http.StatusBadRequest, "from and to query parameters are required (format: YYYY-MM-DD)")
		return
	}

	from, err := time.Parse("2006-01-02", fromStr)
	if err != nil {
		shared.RespondWithError(c, http.StatusBadRequest, "invalid from date, expected YYYY-MM-DD")
		return
	}
	from = from.UTC()

	to, err := time.Parse("2006-01-02", toStr)
	if err != nil {
		shared.RespondWithError(c, http.StatusBadRequest, "invalid to date, expected YYYY-MM-DD")
		return
	}
	// Set to end of day
	to = time.Date(to.Year(), to.Month(), to.Day(), 23, 59, 59, 0, time.UTC)

	events, err := h.service.ListEventsByDateRange(c.Request.Context(), currentUser.ID, from, to)
	if err != nil {
		h.logger.Error("failed to list events", zap.Error(err))
		shared.RespondWithError(c, http.StatusInternalServerError, "failed to fetch events")
		return
	}

	responses := make([]dto.EventResponse, 0, len(events))
	for _, e := range events {
		responses = append(responses, dto.ToEventResponse(e))
	}

	shared.RespondWithSuccess(c, http.StatusOK, "Events retrieved successfully", gin.H{
		"events": responses,
		"count":  len(responses),
		"from":   fromStr,
		"to":     toStr,
	})
}

func (h *Handler) listUpcomingEvents(c *gin.Context) {
	currentUser, exists := middleware.GetCurrentUser(c)
	if !exists {
		shared.RespondWithError(c, http.StatusUnauthorized, "user not found in context")
		return
	}

	var from time.Time
	if fromStr := c.Query("from"); fromStr != "" {
		parsed, err := time.Parse("2006-01-02", fromStr)
		if err != nil {
			shared.RespondWithError(c, http.StatusBadRequest, "invalid from date, expected YYYY-MM-DD")
			return
		}
		from = parsed.UTC()
	}

	var to time.Time
	if toStr := c.Query("to"); toStr != "" {
		parsed, err := time.Parse("2006-01-02", toStr)
		if err != nil {
			shared.RespondWithError(c, http.StatusBadRequest, "invalid to date, expected YYYY-MM-DD")
			return
		}
		to = parsed.UTC()
	}

	events, err := h.service.ListUpcomingEvents(c.Request.Context(), currentUser.ID, from, to)
	if err != nil {
		h.logger.Error("failed to list events", zap.Error(err))
		shared.RespondWithError(c, http.StatusInternalServerError, "failed to fetch events")
		return
	}

	responses := make([]dto.EventResponse, 0, len(events))
	for _, e := range events {
		responses = append(responses, dto.ToEventResponse(&e))
	}

	shared.RespondWithSuccess(c, http.StatusOK, "Upcoming events fetched successfully", gin.H{
		"events": responses,
	})
}

func (h *Handler) deleteEvent(c *gin.Context) {
	currentUser, exists := middleware.GetCurrentUser(c)
	if !exists {
		shared.RespondWithError(c, http.StatusUnauthorized, "user not found in context")
		return
	}

	eventID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		shared.RespondWithError(c, http.StatusBadRequest, "invalid event id")
		return
	}

	if err := h.service.DeleteEvent(c.Request.Context(), eventID, currentUser.ID); err != nil {
		shared.HandleError(c, err)
		return
	}

	shared.RespondWithNoContent(c)
}

// generateHolidays generates Vietnam holidays for a user
// POST /api/v1/calendar/events/holidays/generate
// Body: { "year": 2025 } (optional, defaults to current year)
func (h *Handler) generateHolidays(c *gin.Context) {
	currentUser, exists := middleware.GetCurrentUser(c)
	if !exists {
		shared.RespondWithError(c, http.StatusUnauthorized, "user not found in context")
		return
	}

	var req generateHolidaysRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		// If no body, use current year
		req.Year = nil
	}

	year := time.Now().Year()
	if req.Year != nil {
		year = *req.Year
	}

	if year < 2020 || year > 2100 {
		shared.RespondWithError(c, http.StatusBadRequest, "year must be between 2020 and 2100")
		return
	}

	if err := h.service.GenerateHolidaysForYear(c.Request.Context(), currentUser.ID, year); err != nil {
		h.logger.Error("failed to generate holidays",
			zap.String("user_id", currentUser.ID.String()),
			zap.Int("year", year),
			zap.Error(err))
		shared.RespondWithError(c, http.StatusInternalServerError, "failed to generate holidays")
		return
	}

	h.logger.Info("generated holidays",
		zap.String("user_id", currentUser.ID.String()),
		zap.Int("year", year))

	shared.RespondWithSuccess(c, http.StatusCreated, "Holidays generated successfully", gin.H{
		"year": year,
	})
}

// Helper functions

// parseDateTime parses various datetime formats
// Supports: "2025-03-15", "2025-03-15T14:30:00Z", "2025-03-15T14:30:00+07:00"
func parseDateTime(s string) (time.Time, error) {
	formats := []string{
		time.RFC3339,
		"2006-01-02T15:04:05Z07:00",
		"2006-01-02T15:04:05",
		"2006-01-02",
	}

	for _, format := range formats {
		if t, err := time.Parse(format, s); err == nil {
			return t.UTC(), nil
		}
	}

	return time.Time{}, fmt.Errorf("invalid datetime format: %s", s)
}
