package shared

type PageRequest struct {
	Page     int `json:"page" form:"page"`
	PageSize int `json:"pageSize" form:"pageSize"`
}

type Pagination struct {
	Page    int
	PerPage int
	Sort    string // ví dụ: "created_at desc", "last_active_at desc"
}
