package domain

// MonthStatus represents the lifecycle status of a budgeting month
type MonthStatus string

const (
	// StatusOpen indicates the month is active and can be modified
	StatusOpen MonthStatus = "OPEN"
	// StatusClosed indicates the month is frozen for reporting
	StatusClosed MonthStatus = "CLOSED"
	// StatusArchived indicates the month is archived (read-only, hidden from UI)
	StatusArchived MonthStatus = "ARCHIVED"
)

// IsValid checks if the status is a valid MonthStatus
func (s MonthStatus) IsValid() bool {
	switch s {
	case StatusOpen, StatusClosed, StatusArchived:
		return true
	}
	return false
}

// EventType represents the type of domain event in the event log
type EventType string

const (
	// EventMonthCreated is emitted when a new month is created
	EventMonthCreated EventType = "month.created"
	// EventCategoryAssigned is emitted when money is assigned to a category
	EventCategoryAssigned EventType = "category.assigned"
	// EventMoneyMoved is emitted when money is moved between categories
	EventMoneyMoved EventType = "money.moved"
	// EventIncomeReceived is emitted when income is added to TBB
	EventIncomeReceived EventType = "income.received"
	// EventTransactionPosted is emitted when a transaction affects a category (external event)
	EventTransactionPosted EventType = "transaction.posted"
	// EventMonthClosed is emitted when a month is closed/frozen
	EventMonthClosed EventType = "month.closed"
)

// IsValid checks if the event type is valid
func (e EventType) IsValid() bool {
	switch e {
	case EventMonthCreated, EventCategoryAssigned, EventMoneyMoved,
		EventIncomeReceived, EventTransactionPosted, EventMonthClosed:
		return true
	}
	return false
}

// String returns the string representation
func (e EventType) String() string {
	return string(e)
}
