package mbms

import (
	"context"
	"time"
)

// Model là interface cơ bản mà tất cả decision models phải implement
// Interface này định nghĩa contract chung cho mọi model trong hệ thống
type Model interface {
	// Name trả về tên unique của model để registry có thể quản lý
	Name() string

	// Description trả về mô tả ngắn gọn về chức năng của model
	Description() string

	// Dependencies trả về list các models mà model này phụ thuộc
	// Ví dụ: Budget Allocation phụ thuộc Goal Prioritization, Debt Strategy
	Dependencies() []string

	// Validate kiểm tra input có hợp lệ không trước khi execute
	// Trả về error nếu input không đáp ứng requirements
	Validate(ctx context.Context, input interface{}) error

	// Execute chạy model với input đã validate và trả về output
	// Context cho phép cancel operation nếu cần (timeout, user cancel)
	Execute(ctx context.Context, input interface{}) (interface{}, error)
}

// ModelMetadata chứa thông tin meta về model execution
// Dùng để tracking, auditing, và debugging
type ModelMetadata struct {
	ModelName      string        `json:"model_name"`      // Tên model đã chạy
	ExecutionID    string        `json:"execution_id"`    // UUID unique cho lần execution này
	StartTime      time.Time     `json:"start_time"`      // Thời điểm bắt đầu
	EndTime        time.Time     `json:"end_time"`        // Thời điểm kết thúc
	Duration       time.Duration `json:"duration"`        // Thời gian thực thi
	Status         string        `json:"status"`          // "success", "failed", "cancelled"
	ErrorMessage   string        `json:"error_message"`   // Chi tiết lỗi nếu failed
	InputSnapshot  []byte        `json:"input_snapshot"`  // JSON snapshot của input (để audit)
	OutputSnapshot []byte        `json:"output_snapshot"` // JSON snapshot của output
}

// ModelResult là wrapper chứa cả output và metadata
// Cho phép tracking đầy đủ quá trình thực thi
type ModelResult struct {
	Output   interface{}   `json:"output"`
	Metadata ModelMetadata `json:"metadata"`
}

// ModelRegistry quản lý tất cả models trong hệ thống
// Implement pattern Registry để có thể plug-and-play models
type ModelRegistry interface {
	// Register thêm một model vào registry
	Register(model Model) error

	// Get lấy model theo tên
	Get(name string) (Model, error)

	// List trả về danh sách tất cả models đã đăng ký
	List() []string

	// Unregister xóa model khỏi registry
	Unregister(name string) error

	// GetMetadata lấy metadata của một model
	GetMetadata(name string) (*RegistryMetadata, error)

	// GetAllMetadata lấy metadata của tất cả models
	GetAllMetadata() map[string]*RegistryMetadata

	// EnableModel enables a model
	EnableModel(name string) error

	// DisableModel disables a model
	DisableModel(name string) error

	// ValidateRegistry kiểm tra registry health
	ValidateRegistry() []error
}

// RegistryMetadata chứa metadata về một model trong registry
type RegistryMetadata struct {
	Name            string   `json:"name"`
	Description     string   `json:"description"`
	Version         string   `json:"version"`
	Dependencies    []string `json:"dependencies"`
	Category        string   `json:"category"`          // "core", "supporting", "advanced"
	IsEnabled       bool     `json:"is_enabled"`        // Có thể enable/disable runtime
	AverageExecTime float64  `json:"average_exec_time"` // Milliseconds
	TotalExecutions int64    `json:"total_executions"`
}

// ModelOrchestrator điều phối việc thực thi nhiều models theo đúng thứ tự
// Xử lý dependencies và pipeline execution
type ModelOrchestrator interface {
	// ExecutePipeline chạy một chuỗi models theo thứ tự phụ thuộc
	// Input map chứa initial inputs, outputs sẽ được pass giữa các models
	ExecutePipeline(ctx context.Context, modelNames []string, inputs map[string]interface{}) (map[string]*ModelResult, error)

	// ExecuteSingle chạy một model đơn lẻ (Expert Mode)
	ExecuteSingle(ctx context.Context, modelName string, input interface{}) (*ModelResult, error)

	// ResolveDependencies tính toán thứ tự thực thi dựa trên dependency graph
	// Trả về ordered list hoặc error nếu có circular dependency
	ResolveDependencies(modelNames []string) ([]string, error)
}

// ResultCache interface cho việc cache kết quả models
// Tránh phải re-run models khi input không đổi
type ResultCache interface {
	// Set lưu result với TTL (time-to-live)
	Set(ctx context.Context, key string, result *ModelResult, ttl time.Duration) error

	// Get lấy cached result nếu còn valid
	Get(ctx context.Context, key string) (*ModelResult, error)

	// Invalidate xóa cached result (khi input thay đổi)
	Invalidate(ctx context.Context, key string) error

	// Clear xóa tất cả cache
	Clear(ctx context.Context) error
}

// DependencyResolver resolves model dependencies
type DependencyResolver interface {
	// Resolve returns ordered list of models to execute
	// Input: target models (models user muốn chạy)
	// Output: ordered list bao gồm cả target models và dependencies
	Resolve(targetModels []string) ([]string, error)

	// GetDependencyTree returns dependency tree cho visualization
	GetDependencyTree(modelName string) (*DependencyNode, error)

	// ValidateDependencies validates tất cả dependencies are resolvable
	ValidateDependencies() error
}

// DependencyNode represents một node trong dependency tree
type DependencyNode struct {
	Name         string            `json:"name"`
	Dependencies []*DependencyNode `json:"dependencies"`
}
