package mbms

import (
	"errors"
	"fmt"
	"sync"
)

// modelRegistry implements ModelRegistry interface
type modelRegistry struct {
	// models map lưu trữ model instances
	// Key: model name, Value: model instance
	models map[string]Model

	// metadata map lưu trữ metadata của mỗi model
	metadata map[string]*RegistryMetadata

	// mutex để đảm bảo thread-safety
	// Vì registry có thể được access từ nhiều goroutines
	mu sync.RWMutex
}

// NewModelRegistry tạo một registry mới
func NewModelRegistry() ModelRegistry {
	return &modelRegistry{
		models:   make(map[string]Model),
		metadata: make(map[string]*RegistryMetadata),
	}
}

// Register thêm một model vào registry
// Method này sẽ được gọi khi application khởi động
func (r *modelRegistry) Register(model Model) error {
	// Validate model có implement đầy đủ interface không
	if model == nil {
		return errors.New("model cannot be nil")
	}

	name := model.Name()
	if name == "" {
		return errors.New("model name cannot be empty")
	}

	// Lock để write-safe
	r.mu.Lock()
	defer r.mu.Unlock()

	// Check nếu model đã tồn tại
	if _, exists := r.models[name]; exists {
		return fmt.Errorf("model '%s' is already registered", name)
	}

	// Store model instance
	r.models[name] = model

	// Create và store metadata
	r.metadata[name] = &RegistryMetadata{
		Name:         name,
		Description:  model.Description(),
		Version:      "1.0.0", // Có thể extend để models tự report version
		Dependencies: model.Dependencies(),
		Category:     r.inferCategory(model),
		IsEnabled:    true, // Default enabled
	}

	return nil
}

// Get lấy model theo tên
func (r *modelRegistry) Get(name string) (Model, error) {
	// Read lock vì chỉ đọc, không modify
	r.mu.RLock()
	defer r.mu.RUnlock()

	model, exists := r.models[name]
	if !exists {
		return nil, fmt.Errorf("model '%s' not found in registry", name)
	}

	// Check if model is enabled
	if meta, ok := r.metadata[name]; ok && !meta.IsEnabled {
		return nil, fmt.Errorf("model '%s' is currently disabled", name)
	}

	return model, nil
}

// List trả về danh sách tất cả model names
func (r *modelRegistry) List() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	names := make([]string, 0, len(r.models))
	for name := range r.models {
		names = append(names, name)
	}

	return names
}

// GetMetadata lấy metadata của một model
func (r *modelRegistry) GetMetadata(name string) (*RegistryMetadata, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	meta, exists := r.metadata[name]
	if !exists {
		return nil, fmt.Errorf("metadata for model '%s' not found", name)
	}

	// Return a copy để avoid external modifications
	metaCopy := *meta
	return &metaCopy, nil
}

// GetAllMetadata lấy metadata của tất cả models
func (r *modelRegistry) GetAllMetadata() map[string]*RegistryMetadata {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// Return copies
	result := make(map[string]*RegistryMetadata)
	for name, meta := range r.metadata {
		metaCopy := *meta
		result[name] = &metaCopy
	}

	return result
}

// UpdateMetadata cập nhật metadata (ví dụ: execution stats)
func (r *modelRegistry) updateMetadata(name string, updateFn func(*RegistryMetadata)) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	meta, exists := r.metadata[name]
	if !exists {
		return fmt.Errorf("metadata for model '%s' not found", name)
	}

	updateFn(meta)
	return nil
}

// EnableModel enables a model
func (r *modelRegistry) EnableModel(name string) error {
	return r.updateMetadata(name, func(m *RegistryMetadata) {
		m.IsEnabled = true
	})
}

// DisableModel disables a model
func (r *modelRegistry) DisableModel(name string) error {
	return r.updateMetadata(name, func(m *RegistryMetadata) {
		m.IsEnabled = false
	})
}

// Unregister removes model from registry
// Useful for testing hoặc hot-reload scenarios
func (r *modelRegistry) Unregister(name string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.models[name]; !exists {
		return fmt.Errorf("model '%s' not found", name)
	}

	delete(r.models, name)
	delete(r.metadata, name)

	return nil
}

// inferCategory tự động infer category dựa trên dependencies
// Helper function để categorize models
func (r *modelRegistry) inferCategory(model Model) string {
	deps := model.Dependencies()

	if len(deps) == 0 {
		return "core" // No dependencies = core model
	} else if len(deps) >= 3 {
		return "advanced" // Nhiều dependencies = advanced/integrated model
	} else {
		return "supporting" // Mid-level model
	}
}

// GetByCategory trả về models theo category
func (r *modelRegistry) GetByCategory(category string) []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make([]string, 0)
	for name, meta := range r.metadata {
		if meta.Category == category {
			result = append(result, name)
		}
	}

	return result
}

// ValidateRegistry kiểm tra registry health
// Useful để run lúc startup để ensure tất cả dependencies valid
func (r *modelRegistry) ValidateRegistry() []error {
	r.mu.RLock()
	defer r.mu.RUnlock()

	errs := make([]error, 0)

	// Check 1: Validate dependencies exist
	for name, model := range r.models {
		for _, dep := range model.Dependencies() {
			if _, exists := r.models[dep]; !exists {
				errs = append(errs,
					fmt.Errorf("model '%s' depends on '%s' which is not registered",
						name, dep))
			}
		}
	}

	return errs
}
