package domain

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
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

	// DSS Metadata for Analytics & Budget Recommendations
	DSSMetadata datatypes.JSON `gorm:"type:jsonb;column:dss_metadata" json:"dss_metadata,omitempty"`
	// Structure:
	// {
	//   "avg_monthly_spending": 5000000,      // Average monthly spending
	//   "spending_trend": "increasing",       // increasing, stable, decreasing
	//   "volatility": 0.2,                    // Spending volatility (std dev / mean)
	//   "transaction_count": 45,              // Avg transactions per month
	//   "avg_transaction_amount": 111111,     // Average transaction amount
	//   "last_transaction_date": "2024-01-15",
	//   "typical_amount_range": [100000, 500000], // [min, max] typical amounts
	//   "spending_pattern": "regular",        // regular, irregular, seasonal
	//   "seasonality": {
	//     "has_seasonality": true,
	//     "peak_months": [12, 1],             // December, January
	//     "low_months": [6, 7]                // June, July
	//   },
	//   "default_budget_percent": 0.15,       // Recommended % of income
	//   "recommended_budget": 6000000,        // Recommended monthly budget
	//   "priority": 5,                        // Priority (1-10)
	//   "is_necessity": true,                 // Essential vs discretionary
	//   "is_discretionary": false,
	//   "optimization_potential": 0.3,        // Potential to reduce (0-1)
	//   "merchant_diversity": 0.7,            // How many different merchants
	//   "top_merchants": [
	//     {"name": "Merchant A", "amount": 2000000, "count": 10},
	//     {"name": "Merchant B", "amount": 1500000, "count": 8}
	//   ],
	//   "last_analyzed": "2024-01-15T10:00:00Z"
	// }

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
