package dto

type Emergency_fundInput struct {
	UserID string `json:"user_id"`
}

type Emergency_fundOutput struct {
	Result string `json:"result"`
}
