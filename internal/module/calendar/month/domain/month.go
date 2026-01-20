package domain

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Month represents the Aggregate Root for a budgeting month (read model)
// This is the materialized view stored in the monthly_budgets table
type Month struct {
	ID     uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"` // UUIDv7 generated in BeforeCreate
	UserID uuid.UUID `gorm:"type:uuid;not null;index" json:"user_id"`
	Month  string    `gorm:"type:varchar(10);not null;index" json:"month"` // Display name: "2024-02"

	// Budget period dates - supports custom periods (e.g., pay period 15/01-14/02)
	StartDate time.Time `gorm:"type:date;not null;index" json:"start_date"`
	EndDate   time.Time `gorm:"type:date;not null;index" json:"end_date"`

	Status MonthStatus `gorm:"type:varchar(20);default:'OPEN'" json:"status"`

	// ===== THE BIG CHANGE: Array of States (Snapshots/Iterations) =====
	// Each MonthState is a complete planning snapshot
	// The LAST element is the current/active state
	// Previous elements are history for "what-if" comparisons
	States []MonthState `gorm:"type:jsonb;serializer:json;not null;default:'[]'" json:"states"`

	// ===== CLOSED MONTH SUMMARY (at Month level) =====
	ClosedAt       *time.Time           `gorm:"type:timestamptz" json:"closed_at,omitempty"`
	ClosedSnapshot *MonthClosedSnapshot `gorm:"type:jsonb;serializer:json" json:"closed_snapshot,omitempty"`

	Version int64 `gorm:"default:1" json:"version"` // For optimistic locking

	CreatedAt time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName specifies the database table name
func (Month) TableName() string {
	return "monthly_budgets"
}

// MonthState represents a complete snapshot/iteration state for planning
// Multiple MonthStates can exist in Month.States array for version history
type MonthState struct {
	// ===== METADATA =====
	Version   int       `json:"version"`    // 1, 2, 3... (iteration number)
	CreatedAt time.Time `json:"created_at"` // When this snapshot was created
	IsApplied bool      `json:"is_applied"` // Was this iteration applied to budget?

	// ===== BUDGET EXECUTION STATE (YNAB Core) =====
	ToBeBudgeted   float64         `json:"tbb"`             // To Be Budgeted
	ActualIncome   float64         `json:"actual_income"`   // Confirmed income from transactions
	CategoryStates []CategoryState `json:"category_states"` // Array of all budget line items (constraints/goals/debts)

	// ===== INPUT SNAPSHOT (Frozen from source tables) =====
	Input *InputSnapshot `json:"input,omitempty"`

	// ===== DSS WORKFLOW RESULTS =====
	DSSWorkflow *DSSWorkflowResults `json:"dss_workflow,omitempty"` // Sequential 4-step DSS workflow
}

// NewMonthState creates a new initialized MonthState
func NewMonthState() *MonthState {
	return &MonthState{
		ToBeBudgeted:   0,
		ActualIncome:   0,
		CategoryStates: []CategoryState{},
		Input: &InputSnapshot{
			CapturedAt:     time.Now(),
			IncomeProfiles: []IncomeSnapshot{},
			Constraints:    []ConstraintSnapshot{},
			Goals:          []GoalSnapshot{},
			Debts:          []DebtSnapshot{},
			AdHocExpenses:  []AdHocExpense{},
			AdHocIncome:    []AdHocIncome{},
			Totals:         InputTotals{},
		},
	}
}

// EnsureInitialized ensures all maps/slices are initialized
func (s *MonthState) EnsureInitialized() {
	if s.CategoryStates == nil {
		s.CategoryStates = []CategoryState{}
	}
	if s.Input == nil {
		s.Input = &InputSnapshot{
			CapturedAt:     time.Now(),
			IncomeProfiles: []IncomeSnapshot{},
			Constraints:    []ConstraintSnapshot{},
			Goals:          []GoalSnapshot{},
			Debts:          []DebtSnapshot{},
			AdHocExpenses:  []AdHocExpense{},
			AdHocIncome:    []AdHocIncome{},
			Totals:         InputTotals{},
		}
	}
}

// GoalStateSnapshot captures goal state at month start/current
type GoalStateSnapshot struct {
	GoalID        uuid.UUID `json:"goal_id"`
	Name          string    `json:"name"`
	TargetAmount  float64   `json:"target_amount"`
	CurrentAmount float64   `json:"current_amount"`
	Progress      float64   `json:"progress"` // 0-100
	Priority      int       `json:"priority"`
	Status        string    `json:"status"` // "active", "completed", etc.
	TargetDate    *string   `json:"target_date,omitempty"`
}

// DebtStateSnapshot captures debt state at month start/current
type DebtStateSnapshot struct {
	DebtID         uuid.UUID `json:"debt_id"`
	Name           string    `json:"name"`
	OriginalAmount float64   `json:"original_amount"`
	CurrentBalance float64   `json:"current_balance"`
	InterestRate   float64   `json:"interest_rate"`
	MinPayment     float64   `json:"min_payment"`
	Progress       float64   `json:"progress"` // % paid off
	Status         string    `json:"status"`   // "active", "paid_off"
}

// MonthClosedSnapshot contains aggregated data when month is closed
// This is the final summary that gets carried to next month
type MonthClosedSnapshot struct {
	ClosedAt    time.Time `json:"closed_at"`
	ClosedBy    uuid.UUID `json:"closed_by,omitempty"`
	IsAutoClose bool      `json:"is_auto_close"` // Auto-closed at month end vs manual

	// ===== INCOME SUMMARY =====
	TotalIncomeReceived float64 `json:"total_income_received"` // Actual income from transactions
	ProjectedIncome     float64 `json:"projected_income"`      // What was planned
	IncomeVariance      float64 `json:"income_variance"`       // Actual - Projected

	// ===== EXPENSE SUMMARY =====
	TotalSpent       float64 `json:"total_spent"`       // Sum of all negative activity
	TotalBudgeted    float64 `json:"total_budgeted"`    // Sum of all assigned
	SpendingVariance float64 `json:"spending_variance"` // Budgeted - Spent

	// ===== SAVINGS SUMMARY =====
	NetSavings  float64 `json:"net_savings"`  // Income - Spent
	SavingsRate float64 `json:"savings_rate"` // NetSavings / Income * 100

	// ===== TBB SUMMARY =====
	FinalTBB        float64 `json:"final_tbb"`          // TBB at close (carries to next)
	TBBStartOfMonth float64 `json:"tbb_start_of_month"` // TBB at beginning

	// ===== CATEGORY ROLLOVERS =====
	// Map of CategoryID -> Available (to become Rollover in next month)
	CategoryRollovers   map[uuid.UUID]float64 `json:"category_rollovers"`
	OverspentCategories []uuid.UUID           `json:"overspent_categories,omitempty"` // Categories with negative Available

	// ===== GOAL PROGRESS =====
	GoalContributions map[uuid.UUID]float64 `json:"goal_contributions,omitempty"` // Amount contributed to each goal
	GoalsCompleted    []uuid.UUID           `json:"goals_completed,omitempty"`    // Goals that reached 100%

	// ===== DEBT PROGRESS =====
	DebtPayments map[uuid.UUID]float64 `json:"debt_payments,omitempty"`  // Amount paid to each debt
	DebtsPaidOff []uuid.UUID           `json:"debts_paid_off,omitempty"` // Debts that reached 0

	// ===== TRANSACTION SUMMARY (for CSV export) =====
	TransactionCount    int                          `json:"transaction_count"`            // Total transactions
	IncomeTransactions  int                          `json:"income_transactions"`          // Inflow count
	ExpenseTransactions int                          `json:"expense_transactions"`         // Outflow count
	CategoryBreakdown   []CategoryTransactionSummary `json:"category_breakdown,omitempty"` // Per-category breakdown

	// ===== DSS USAGE =====
	DSSIterationCount int  `json:"dss_iteration_count"`         // How many times DSS was run
	AppliedIteration  *int `json:"applied_iteration,omitempty"` // Which iteration was applied
}

// CategoryTransactionSummary summarizes transactions for a single category
type CategoryTransactionSummary struct {
	CategoryID       uuid.UUID `json:"category_id"`
	CategoryName     string    `json:"category_name"`
	TransactionCount int       `json:"transaction_count"`
	TotalAmount      float64   `json:"total_amount"` // Sum of all transactions (negative = expense)
	Budgeted         float64   `json:"budgeted"`     // Assigned amount
	Rollover         float64   `json:"rollover"`     // From previous month
	Available        float64   `json:"available"`    // Final available
	Variance         float64   `json:"variance"`     // Budgeted - abs(TotalAmount)
}

// CanBeModified checks if the month is open and can be modified
func (m *Month) CanBeModified() bool {
	return m.Status == StatusOpen
}

// Close marks the month as closed (immutable after this)
func (m *Month) Close() error {
	if m.Status == StatusClosed {
		return errors.New("month already closed")
	}
	m.Status = StatusClosed
	return nil
}

// Validate validates the month entity
func (m *Month) Validate() error {
	if m.UserID == uuid.Nil {
		return errors.New("user_id is required")
	}
	if m.Month == "" {
		return errors.New("month is required")
	}
	if !m.Status.IsValid() {
		return errors.New("invalid month status")
	}
	return nil
}

// BeforeCreate GORM hook to generate UUIDv7 before inserting
func (m *Month) BeforeCreate(tx *gorm.DB) error {
	if m.ID == uuid.Nil {
		m.ID = uuid.Must(uuid.NewV7())
	}
	return nil
}

// EnsureState ensures States array has at least one element
func (m *Month) EnsureState() {
	if len(m.States) == 0 {
		m.States = append(m.States, *NewMonthState())
	}
	// Ensure the current state is initialized
	m.CurrentState().EnsureInitialized()
}

// CurrentState returns the most recent (active) state
func (m *Month) CurrentState() *MonthState {
	if len(m.States) == 0 {
		m.States = append(m.States, *NewMonthState())
	}
	return &m.States[len(m.States)-1]
}

// AppendState adds a new state snapshot to the history
func (m *Month) AppendState(state MonthState) {
	state.Version = len(m.States) + 1
	state.CreatedAt = time.Now()
	m.States = append(m.States, state)
}

// GetCategoryState retrieves a specific category state from the current state
// Note: Now requires both ID and Type to uniquely identify (since we use slice)
func (m *Month) GetCategoryState(categoryID uuid.UUID, itemType ItemType) (*CategoryState, error) {
	m.EnsureState()
	for i := range m.CurrentState().CategoryStates {
		cs := &m.CurrentState().CategoryStates[i]
		if cs.CategoryID == categoryID && cs.Type == itemType {
			return cs, nil
		}
	}
	return nil, errors.New("category not found in month state")
}

// UpdateCategoryState updates a specific category state in the current state
func (m *Month) UpdateCategoryState(categoryID uuid.UUID, itemType ItemType, catState CategoryState) {
	m.EnsureState()
	state := m.CurrentState()

	// Find and update existing
	for i := range state.CategoryStates {
		if state.CategoryStates[i].CategoryID == categoryID &&
			state.CategoryStates[i].Type == itemType {
			state.CategoryStates[i] = catState
			return
		}
	}

	// Not found, append new
	state.CategoryStates = append(state.CategoryStates, catState)
}

// IncrementVersion increments the version for optimistic locking
func (m *Month) IncrementVersion() {
	m.Version++
}

// CalculateClosedSnapshot calculates and returns the closed snapshot for this month
// This should be called before closing the month
func (m *Month) CalculateClosedSnapshot(closedBy *uuid.UUID, isAutoClose bool) *MonthClosedSnapshot {
	m.EnsureState()
	state := m.CurrentState()

	snapshot := &MonthClosedSnapshot{
		ClosedAt:          time.Now(),
		IsAutoClose:       isAutoClose,
		CategoryRollovers: make(map[uuid.UUID]float64),
	}

	if closedBy != nil {
		snapshot.ClosedBy = *closedBy
	}

	// Calculate totals from category states
	var totalAssigned, totalActivity float64
	for _, catState := range state.CategoryStates {
		totalAssigned += catState.Assigned
		totalActivity += catState.Activity

		// Available = Rollover + Assigned + Activity
		available := catState.Rollover + catState.Assigned + catState.Activity
		snapshot.CategoryRollovers[catState.CategoryID] = available

		// Track overspent categories
		if available < 0 {
			snapshot.OverspentCategories = append(snapshot.OverspentCategories, catState.CategoryID)
		}
	}

	// Income summary
	snapshot.TotalIncomeReceived = state.ActualIncome
	if state.Input != nil {
		snapshot.ProjectedIncome = state.Input.Totals.ProjectedIncome
	}
	snapshot.IncomeVariance = snapshot.TotalIncomeReceived - snapshot.ProjectedIncome

	// Expense summary (Activity is typically negative for expenses)
	snapshot.TotalSpent = -totalActivity // Flip sign: negative activity = positive spending
	if snapshot.TotalSpent < 0 {
		snapshot.TotalSpent = 0 // Guard against positive activity (refunds)
	}
	snapshot.TotalBudgeted = totalAssigned
	snapshot.SpendingVariance = snapshot.TotalBudgeted - snapshot.TotalSpent

	// Savings summary
	snapshot.NetSavings = snapshot.TotalIncomeReceived - snapshot.TotalSpent
	if snapshot.TotalIncomeReceived > 0 {
		snapshot.SavingsRate = (snapshot.NetSavings / snapshot.TotalIncomeReceived) * 100
	}

	// TBB summary
	snapshot.FinalTBB = state.ToBeBudgeted
	// TBBStartOfMonth would need to be tracked separately if needed

	// DSS usage
	snapshot.DSSIterationCount = len(m.States)
	for i := len(m.States) - 1; i >= 0; i-- {
		if m.States[i].IsApplied {
			appliedIdx := i + 1 // 1-indexed version
			snapshot.AppliedIteration = &appliedIdx
			break
		}
	}

	return snapshot
}

// CloseWithSnapshot closes the month and sets the closed snapshot
func (m *Month) CloseWithSnapshot(closedBy *uuid.UUID, isAutoClose bool) (*MonthClosedSnapshot, error) {
	if m.Status == StatusClosed {
		return nil, errors.New("month is already closed")
	}

	snapshot := m.CalculateClosedSnapshot(closedBy, isAutoClose)
	now := time.Now()
	m.ClosedAt = &now
	m.ClosedSnapshot = snapshot
	m.Status = StatusClosed

	return snapshot, nil
}
