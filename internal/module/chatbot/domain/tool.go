package domain

// Tool represents a function that can be called by the LLM
type Tool struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Parameters  ToolParameters `json:"parameters"`
}

// ToolParameters defines the JSON Schema for tool parameters
type ToolParameters struct {
	Type       string                  `json:"type"` // always "object"
	Properties map[string]ToolProperty `json:"properties"`
	Required   []string                `json:"required"`
}

// ToolProperty defines a single property in tool parameters
type ToolProperty struct {
	Type        string        `json:"type"`
	Description string        `json:"description"`
	Enum        []string      `json:"enum,omitempty"`
	Items       *ToolProperty `json:"items,omitempty"` // for arrays
	Default     any           `json:"default,omitempty"`
}

// ToolDefinitions contains all available tools for the chatbot
var ToolDefinitions = []Tool{
	{
		Name:        "analyze_budget_allocation",
		Description: "Phân tích và đề xuất phân bổ ngân sách dựa trên thu nhập và chi tiêu của người dùng",
		Parameters: ToolParameters{
			Type: "object",
			Properties: map[string]ToolProperty{
				"monthly_income": {
					Type:        "number",
					Description: "Thu nhập hàng tháng (VND)",
				},
				"fixed_expenses": {
					Type:        "number",
					Description: "Chi phí cố định hàng tháng (VND)",
				},
				"goals": {
					Type:        "array",
					Description: "Danh sách mục tiêu tài chính",
					Items: &ToolProperty{
						Type: "object",
					},
				},
			},
			Required: []string{"monthly_income"},
		},
	},
	{
		Name:        "analyze_debt_strategy",
		Description: "Phân tích và đề xuất chiến lược trả nợ (Avalanche hoặc Snowball)",
		Parameters: ToolParameters{
			Type: "object",
			Properties: map[string]ToolProperty{
				"debts": {
					Type:        "array",
					Description: "Danh sách các khoản nợ",
					Items: &ToolProperty{
						Type: "object",
					},
				},
				"monthly_payment": {
					Type:        "number",
					Description: "Số tiền có thể trả nợ hàng tháng (VND)",
				},
				"strategy": {
					Type:        "string",
					Description: "Chiến lược trả nợ",
					Enum:        []string{"avalanche", "snowball", "hybrid"},
					Default:     "avalanche",
				},
			},
			Required: []string{"debts", "monthly_payment"},
		},
	},
	{
		Name:        "calculate_emergency_fund",
		Description: "Tính toán quỹ khẩn cấp cần thiết",
		Parameters: ToolParameters{
			Type: "object",
			Properties: map[string]ToolProperty{
				"monthly_expenses": {
					Type:        "number",
					Description: "Chi tiêu hàng tháng (VND)",
				},
				"job_stability": {
					Type:        "string",
					Description: "Mức độ ổn định công việc",
					Enum:        []string{"stable", "moderate", "unstable"},
				},
				"dependents": {
					Type:        "number",
					Description: "Số người phụ thuộc",
				},
			},
			Required: []string{"monthly_expenses"},
		},
	},
	{
		Name:        "analyze_savings_vs_debt",
		Description: "Phân tích cân bằng giữa tiết kiệm và trả nợ",
		Parameters: ToolParameters{
			Type: "object",
			Properties: map[string]ToolProperty{
				"surplus_amount": {
					Type:        "number",
					Description: "Số tiền dư hàng tháng (VND)",
				},
				"debt_interest_rate": {
					Type:        "number",
					Description: "Lãi suất nợ (%/năm)",
				},
				"savings_interest_rate": {
					Type:        "number",
					Description: "Lãi suất tiết kiệm (%/năm)",
				},
			},
			Required: []string{"surplus_amount", "debt_interest_rate"},
		},
	},
	{
		Name:        "get_user_financial_summary",
		Description: "Lấy tóm tắt tình hình tài chính của người dùng",
		Parameters: ToolParameters{
			Type:       "object",
			Properties: map[string]ToolProperty{},
			Required:   []string{},
		},
	},
	{
		Name:        "prioritize_goals",
		Description: "Xếp hạng ưu tiên các mục tiêu tài chính sử dụng phương pháp AHP",
		Parameters: ToolParameters{
			Type: "object",
			Properties: map[string]ToolProperty{
				"goals": {
					Type:        "array",
					Description: "Danh sách mục tiêu cần xếp hạng",
					Items: &ToolProperty{
						Type: "object",
					},
				},
				"criteria": {
					Type:        "array",
					Description: "Tiêu chí đánh giá (urgency, importance, feasibility)",
					Items: &ToolProperty{
						Type: "string",
					},
				},
			},
			Required: []string{"goals"},
		},
	},
}
