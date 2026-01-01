package dto

import (
	"time"

	"personalfinancedss/internal/module/calendar/event/domain"
)

// ToEventResponse converts a domain Event to EventResponse DTO
func ToEventResponse(event *domain.Event) EventResponse {
	resp := EventResponse{
		ID:          event.ID.String(),
		Name:        event.Name,
		Description: event.Description,
		Type:        string(event.Type),
		Source:      string(event.Source),
		StartDate:   event.StartDate.Format(time.RFC3339),
		AllDay:      event.AllDay,
		Color:       event.Color,
		IsRecurring: event.IsRecurring,
		IsMultiDay:  event.IsMultiDay(),
		Duration:    event.Duration(),
		CreatedAt:   event.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   event.UpdatedAt.Format(time.RFC3339),
	}

	if event.EndDate != nil {
		end := event.EndDate.Format(time.RFC3339)
		resp.EndDate = &end
	}

	if len(event.Tags) > 0 {
		tags := string(event.Tags)
		resp.Tags = &tags
	}

	return resp
}

// ToEventListResponse converts a slice of events to EventListResponse
func ToEventListResponse(events []*domain.Event, from, to string) EventListResponse {
	responses := make([]EventResponse, 0, len(events))
	for _, e := range events {
		responses = append(responses, ToEventResponse(e))
	}
	return EventListResponse{
		Events: responses,
		Count:  len(responses),
		From:   from,
		To:     to,
	}
}
