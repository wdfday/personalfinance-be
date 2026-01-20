package domain

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMonth_BeforeCreate_GeneratesUUIDv7(t *testing.T) {
	month := &Month{
		UserID:    uuid.New(),
		Month:     "2024-01",
		StartDate: time.Now(),
		EndDate:   time.Now().AddDate(0, 1, 0),
		Status:    StatusOpen,
	}

	err := month.BeforeCreate(nil)
	require.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, month.ID)

	// Verify it's a valid UUID
	assert.NotEmpty(t, month.ID.String())
}

func TestMonth_BeforeCreate_PreservesExistingID(t *testing.T) {
	existingID := uuid.New()
	month := &Month{
		ID:        existingID,
		UserID:    uuid.New(),
		Month:     "2024-01",
		StartDate: time.Now(),
		EndDate:   time.Now().AddDate(0, 1, 0),
	}

	err := month.BeforeCreate(nil)
	require.NoError(t, err)
	assert.Equal(t, existingID, month.ID)
}

func TestMonth_EnsureState_CreatesDefaultState(t *testing.T) {
	month := &Month{
		ID:        uuid.New(),
		UserID:    uuid.New(),
		Month:     "2024-01",
		StartDate: time.Now(),
		EndDate:   time.Now().AddDate(0, 1, 0),
		Status:    StatusOpen,
		States:    []MonthState{},
	}

	month.EnsureState()

	assert.Len(t, month.States, 1)
	assert.NotNil(t, month.CurrentState())
}

func TestMonth_CurrentState_ReturnsLatestState(t *testing.T) {
	month := &Month{
		States: []MonthState{
			{Version: 1, CreatedAt: time.Now().Add(-2 * time.Hour)},
			{Version: 2, CreatedAt: time.Now().Add(-1 * time.Hour)},
			{Version: 3, CreatedAt: time.Now()},
		},
	}

	current := month.CurrentState()
	assert.Equal(t, 3, current.Version)
}

func TestMonth_AppendState_IncrementsVersion(t *testing.T) {
	month := &Month{
		States: []MonthState{
			{Version: 1},
		},
	}

	newState := MonthState{
		ToBeBudgeted:   10000,
		CategoryStates: make(map[uuid.UUID]*CategoryState),
	}

	month.AppendState(newState)

	assert.Len(t, month.States, 2)
	assert.Equal(t, 2, month.States[1].Version)
}

func TestMonth_Validate_RequiresUserID(t *testing.T) {
	month := &Month{
		Month:     "2024-01",
		StartDate: time.Now(),
		EndDate:   time.Now().AddDate(0, 1, 0),
	}

	err := month.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "user_id")
}

func TestMonth_Validate_RequiresMonth(t *testing.T) {
	month := &Month{
		UserID:    uuid.New(),
		StartDate: time.Now(),
		EndDate:   time.Now().AddDate(0, 1, 0),
	}

	err := month.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "month")
}

func TestMonth_Validate_ValidStatus(t *testing.T) {
	month := &Month{
		UserID:    uuid.New(),
		Month:     "2024-01",
		StartDate: time.Now(),
		EndDate:   time.Now().AddDate(0, 1, 0),
		Status:    MonthStatus("INVALID"),
	}

	err := month.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "status")
}

func TestMonth_Validate_Success(t *testing.T) {
	month := &Month{
		UserID:    uuid.New(),
		Month:     "2024-01",
		StartDate: time.Now(),
		EndDate:   time.Now().AddDate(0, 1, 0),
		Status:    StatusOpen,
	}

	err := month.Validate()
	assert.NoError(t, err)
}

func TestMonthStatus_IsValid(t *testing.T) {
	tests := []struct {
		status MonthStatus
		valid  bool
	}{
		{StatusOpen, true},
		{StatusClosed, true},
		{MonthStatus("INVALID"), false},
		{MonthStatus(""), false},
	}

	for _, tt := range tests {
		t.Run(string(tt.status), func(t *testing.T) {
			assert.Equal(t, tt.valid, tt.status.IsValid())
		})
	}
}

func TestMonth_CalculateClosedSnapshot(t *testing.T) {
	userID := uuid.New()
	month := &Month{
		ID:     uuid.New(),
		UserID: userID,
		Month:  "2024-01",
		States: []MonthState{
			{
				Version:      1,
				ToBeBudgeted: 5000,
				ActualIncome: 20000,
				CategoryStates: map[uuid.UUID]*CategoryState{
					uuid.New(): {
						CategoryID: uuid.New(),
						Type:       ItemTypeConstraint,
						Rollover:   1000,
						Assigned:   5000,
						Activity:   -4000,
					},
				},
				Input: &InputSnapshot{
					Totals: InputTotals{
						ProjectedIncome: 18000,
					},
				},
			},
		},
	}

	closedBy := uuid.New()
	snapshot := month.CalculateClosedSnapshot(&closedBy, false)

	assert.NotNil(t, snapshot)
	assert.Equal(t, closedBy, snapshot.ClosedBy)
	assert.False(t, snapshot.IsAutoClose)
	assert.Equal(t, float64(20000), snapshot.TotalIncomeReceived)
	assert.Equal(t, float64(18000), snapshot.ProjectedIncome)
	assert.Len(t, snapshot.CategoryRollovers, 1)
}
