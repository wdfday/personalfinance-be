package dto

import (
	"encoding/json"
	"fmt"
	"time"
)

// FlexibleTime is a time.Time that can be unmarshaled from multiple formats
type FlexibleTime struct {
	time.Time
}

// UnmarshalJSON implements json.Unmarshaler to parse multiple time formats
func (ft *FlexibleTime) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}

	if s == "" {
		return fmt.Errorf("empty time string")
	}

	// Try multiple formats
	formats := []string{
		time.RFC3339,           // "2006-01-02T15:04:05Z07:00"
		"2006-01-02T15:04:05Z", // "2006-01-02T15:04:05Z"
		"2006-01-02T15:04:05",  // "2006-01-02T15:04:05" (no timezone)
		"2006-01-02T15:04",     // "2006-01-02T15:04" (no seconds, no timezone)
		"2006-01-02 15:04:05",  // "2006-01-02 15:04:05"
		"2006-01-02 15:04",     // "2006-01-02 15:04"
		"2006-01-02",           // "2006-01-02"
	}

	for _, format := range formats {
		if t, err := time.Parse(format, s); err == nil {
			// If no timezone info, assume local timezone (or UTC)
			if format == "2006-01-02T15:04:05" || format == "2006-01-02T15:04" ||
				format == "2006-01-02 15:04:05" || format == "2006-01-02 15:04" {
				// Use UTC for times without timezone
				ft.Time = t.UTC()
			} else {
				ft.Time = t
			}
			return nil
		}
	}

	return fmt.Errorf("invalid time format: %s (supported formats: RFC3339, 2006-01-02T15:04:05, 2006-01-02T15:04, 2006-01-02)", s)
}

// MarshalJSON implements json.Marshaler
func (ft FlexibleTime) MarshalJSON() ([]byte, error) {
	return json.Marshal(ft.Time.Format(time.RFC3339))
}
