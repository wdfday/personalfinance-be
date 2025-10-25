package domain

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Category represents a transaction category for organizing and tracking expenses/income
type Category struct {
	ID uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4();primaryKey" json:"id"`

	UserID *uuid.UUID `gorm:"type:uuid;index;column:user_id" json:"user_id,omitempty"` // NULL for system categories

	// Category details
	Name        string       `gorm:"type:varchar(100);not null;column:name" json:"name"`
	Description *string      `gorm:"type:varchar(500);column:description" json:"description,omitempty"`
	Type        CategoryType `gorm:"type:varchar(20);not null;index;column:type" json:"type"` // income, expense, both

	// Hierarchical structure
	ParentID *uuid.UUID `gorm:"type:uuid;index;column:parent_id" json:"parent_id,omitempty"`
	Level    int        `gorm:"default:0;column:level" json:"level"` // 0 for root, 1 for sub-category, etc.

	// Visual representation
	Icon  *string `gorm:"type:varchar(50);column:icon" json:"icon,omitempty"`   // Icon identifier (e.g., "food", "transport")
	Color *string `gorm:"type:varchar(20);column:color" json:"color,omitempty"` // Hex color (e.g., "#FF5733")

	// Flags
	IsDefault bool `gorm:"default:false;column:is_default" json:"is_default"` // System-provided categories
	IsActive  bool `gorm:"default:true;column:is_active" json:"is_active"`

	// Budget tracking
	MonthlyBudget *float64 `gorm:"type:decimal(15,2);column:monthly_budget" json:"monthly_budget,omitempty"`

	// Statistics (can be calculated on-the-fly)
	TransactionCount int     `gorm:"-" json:"transaction_count,omitempty"` // Computed field
	TotalAmount      float64 `gorm:"-" json:"total_amount,omitempty"`      // Computed field

	// Display order
	DisplayOrder int `gorm:"default:0;column:display_order" json:"display_order"`

	// Relationships (not stored in DB, loaded on demand)
	Parent   *Category   `gorm:"foreignKey:ParentID" json:"parent,omitempty"`
	Children []*Category `gorm:"foreignKey:ParentID" json:"children,omitempty"`

	CreatedAt time.Time      `gorm:"autoCreateTime;column:created_at" json:"created_at"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime;column:updated_at" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index;column:deleted_at" json:"deleted_at,omitempty"`
}

// TableName specifies the database table name
func (Category) TableName() string {
	return "categories"
}

// IsSystemCategory returns true if this is a system-provided category
func (c *Category) IsSystemCategory() bool {
	return c.IsDefault || c.UserID == nil
}

// IsUserCategory returns true if this is a user-created category
func (c *Category) IsUserCategory() bool {
	return !c.IsSystemCategory()
}

// IsRootCategory returns true if this category has no parent
func (c *Category) IsRootCategory() bool {
	return c.ParentID == nil
}

// HasChildren returns true if this category has sub-categories
func (c *Category) HasChildren() bool {
	return len(c.Children) > 0
}
