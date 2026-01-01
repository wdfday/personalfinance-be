package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"personalfinancedss/internal/module/calendar/event/domain"
	"personalfinancedss/internal/module/calendar/event/repository"
)

const defaultUpcomingWindowDays = 60

// Service encapsulates calendar event operations.
type Service interface {
	// Event CRUD
	CreateEvent(ctx context.Context, event *domain.Event) error
	UpdateEvent(ctx context.Context, event *domain.Event) error
	GetEvent(ctx context.Context, id uuid.UUID, userID uuid.UUID) (*domain.Event, error)
	DeleteEvent(ctx context.Context, id uuid.UUID, userID uuid.UUID) error

	// Calendar views
	ListEventsByDateRange(ctx context.Context, userID uuid.UUID, from, to time.Time) ([]*domain.Event, error)
	ListUpcomingEvents(ctx context.Context, userID uuid.UUID, from, to time.Time) ([]domain.Event, error)

	// Holiday management
	GenerateHolidaysForNewUser(ctx context.Context, userID uuid.UUID) error
	GenerateHolidaysForYear(ctx context.Context, userID uuid.UUID, year int) error
}

type eventService struct {
	repo             repository.Repository
	holidayGenerator *HolidayGeneratorService
	logger           *zap.Logger
}

// NewService builds a new event service.
func NewService(repo repository.Repository, logger *zap.Logger) Service {
	return &eventService{
		repo:             repo,
		holidayGenerator: NewHolidayGeneratorService(repo, logger),
		logger:           logger.Named("calendar.event.service"),
	}
}

// CreateEvent creates a new event
func (s *eventService) CreateEvent(ctx context.Context, event *domain.Event) error {
	// Validate
	if err := event.Validate(); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	// Set defaults
	if event.ID == uuid.Nil {
		event.ID = uuid.New()
	}
	if event.Type == "" {
		event.Type = domain.EventTypePersonal
	}
	if event.Source == "" {
		event.Source = domain.SourceUserCreated
	}

	// For all-day events, normalize times
	if event.AllDay {
		event.StartDate = time.Date(
			event.StartDate.Year(),
			event.StartDate.Month(),
			event.StartDate.Day(),
			0, 0, 0, 0, time.UTC,
		)
		if event.EndDate != nil {
			endDate := time.Date(
				event.EndDate.Year(),
				event.EndDate.Month(),
				event.EndDate.Day(),
				23, 59, 59, 0, time.UTC,
			)
			event.EndDate = &endDate
		}
	}

	return s.repo.Create(ctx, event)
}

// UpdateEvent updates an existing event
func (s *eventService) UpdateEvent(ctx context.Context, event *domain.Event) error {
	if err := event.Validate(); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}
	return s.repo.Update(ctx, event)
}

// GetEvent retrieves a single event by ID and user
func (s *eventService) GetEvent(ctx context.Context, id uuid.UUID, userID uuid.UUID) (*domain.Event, error) {
	return s.repo.GetByIDAndUser(ctx, id, userID)
}

// ListEventsByDateRange returns all events within a date range (for calendar view)
func (s *eventService) ListEventsByDateRange(ctx context.Context, userID uuid.UUID, from, to time.Time) ([]*domain.Event, error) {
	return s.repo.ListByUserAndDateRange(ctx, userID, from, to)
}

// ListUpcomingEvents returns upcoming events (legacy method, kept for compatibility)
func (s *eventService) ListUpcomingEvents(ctx context.Context, userID uuid.UUID, from, to time.Time) ([]domain.Event, error) {
	now := time.Now().UTC().Truncate(24 * time.Hour)
	if from.IsZero() {
		from = now
	}
	if to.IsZero() {
		to = from.AddDate(0, 0, defaultUpcomingWindowDays)
	}
	if to.Before(from) {
		to = from
	}
	return s.repo.ListUpcomingByUser(ctx, userID, from, to)
}

// DeleteEvent deletes an event
func (s *eventService) DeleteEvent(ctx context.Context, id uuid.UUID, userID uuid.UUID) error {
	return s.repo.Delete(ctx, id, userID)
}

// GenerateHolidaysForNewUser generates holidays for a newly registered user
func (s *eventService) GenerateHolidaysForNewUser(ctx context.Context, userID uuid.UUID) error {
	return s.holidayGenerator.GenerateForNewUser(ctx, userID)
}

// GenerateHolidaysForYear generates holidays for a specific year
func (s *eventService) GenerateHolidaysForYear(ctx context.Context, userID uuid.UUID, year int) error {
	return s.holidayGenerator.GenerateForYear(ctx, userID, year)
}
