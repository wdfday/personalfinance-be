package dto

// EventResponse represents a calendar event in API responses
type EventResponse struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Description *string `json:"description,omitempty"`
	Type        string  `json:"type"`
	Source      string  `json:"source"`
	StartDate   string  `json:"start_date"` // RFC3339 format
	EndDate     *string `json:"end_date,omitempty"`
	AllDay      bool    `json:"all_day"`
	Color       *string `json:"color,omitempty"`
	Tags        *string `json:"tags,omitempty"` // JSON array string
	IsRecurring bool    `json:"is_recurring"`
	IsMultiDay  bool    `json:"is_multi_day"`
	Duration    int     `json:"duration_days"`
	CreatedAt   string  `json:"created_at"`
	UpdatedAt   string  `json:"updated_at"`
}

// EventListResponse wraps a list of events with metadata
type EventListResponse struct {
	Events []EventResponse `json:"events"`
	Count  int             `json:"count"`
	From   string          `json:"from,omitempty"`
	To     string          `json:"to,omitempty"`
}
