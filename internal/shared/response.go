package shared

// Note: ErrorResponse is defined in errors.go to avoid duplication

// SuccessResponse represents a successful response with data
type SuccessResponse[T any] struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
	Detail  string `json:"detail,omitempty"`
	Data    T      `json:"data,omitempty"`
}

// Success represents a successful response without data
type Success struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
	Detail  string `json:"detail,omitempty"`
}

// Pagination represents paginated response structure
type Page[T any] struct {
	TotalItems   int64 `json:"totalItems"`
	TotalPages   int   `json:"totalPages"`
	CurrentPage  int   `json:"currentPage"`
	ItemsPerPage int   `json:"itemsPerPage"`
	Data         []T   `json:"data"`
}

// PaginationTimeCursor represents time-based cursor pagination
type PaginationTimeCursor[T any] struct {
	TimeCursor   string `json:"timeCursor"`
	HasMore      bool   `json:"hasMore"`
	ItemsPerPage int    `json:"itemsPerPage"`
	Data         *T     `json:"data"`
}
