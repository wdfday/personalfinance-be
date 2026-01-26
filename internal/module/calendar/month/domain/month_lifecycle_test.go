package domain

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMonth_Close_Success(t *testing.T) {
	month := &Month{
		ID:     uuid.New(),
		UserID: uuid.New(),
		Month:  "2024-01",
		Status: StatusOpen,
	}

	err := month.Close()
	require.NoError(t, err)
	assert.Equal(t, StatusClosed, month.Status)
}

func TestMonth_Close_AlreadyClosed(t *testing.T) {
	month := &Month{
		Status: StatusClosed,
	}

	err := month.Close()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already closed")
}

func TestMonth_CanBeModified(t *testing.T) {
	tests := []struct {
		name     string
		status   MonthStatus
		expected bool
	}{
		{"Open month can be modified", StatusOpen, true},
		{"Closed month cannot be modified", StatusClosed, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			month := &Month{Status: tt.status}
			assert.Equal(t, tt.expected, month.CanBeModified())
		})
	}
}

func TestMonth_IncrementVersion(t *testing.T) {
	month := &Month{Version: 1}

	month.IncrementVersion()

	assert.Equal(t, int64(2), month.Version)
}

func TestMonth_TableName(t *testing.T) {
	month := &Month{}
	assert.Equal(t, "monthly_budgets", month.TableName())
}

func TestMonthStatus_String(t *testing.T) {
	assert.Equal(t, "OPEN", string(StatusOpen))
	assert.Equal(t, "CLOSED", string(StatusClosed))
}
