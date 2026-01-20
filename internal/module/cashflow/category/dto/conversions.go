package dto

import (
	"personalfinancedss/internal/module/cashflow/category/domain"

	"github.com/google/uuid"
)

// ToCategoryResponse converts domain.Category to CategoryResponse
func ToCategoryResponse(c *domain.Category, includeRelations bool) *CategoryResponse {
	if c == nil {
		return nil
	}

	resp := &CategoryResponse{
		ID:               c.ID.String(),
		Name:             c.Name,
		Description:      c.Description,
		Type:             string(c.Type),
		Level:            c.Level,
		Icon:             c.Icon,
		Color:            c.Color,
		IsDefault:        c.IsDefault,
		IsActive:         c.IsActive,
		MonthlyBudget:    c.MonthlyBudget,
		TransactionCount: c.TransactionCount,
		TotalAmount:      c.TotalAmount,
		DisplayOrder:     c.DisplayOrder,
		CreatedAt:        c.CreatedAt,
		UpdatedAt:        c.UpdatedAt,
	}

	// Convert optional UUID fields
	if c.UserID != nil {
		userID := c.UserID.String()
		resp.UserID = &userID
	}

	if c.ParentID != nil {
		parentID := c.ParentID.String()
		resp.ParentID = &parentID
	}

	// Convert DeletedAt
	if c.DeletedAt.Valid {
		resp.DeletedAt = &c.DeletedAt.Time
	}

	// Include relationships if requested
	if includeRelations {
		// Parent
		if c.Parent != nil {
			resp.Parent = ToCategoryResponse(c.Parent, false) // Don't recurse parent's parent
		}

		// Children
		if len(c.Children) > 0 {
			resp.Children = make([]CategoryResponse, 0, len(c.Children))
			for _, child := range c.Children {
				if childResp := ToCategoryResponse(child, false); childResp != nil {
					resp.Children = append(resp.Children, *childResp)
				}
			}
		}
	}

	return resp
}

// ToCategoryListResponse converts a slice of categories to list response
func ToCategoryListResponse(categories []*domain.Category, includeRelations bool) *CategoryListResponse {
	resp := &CategoryListResponse{
		Categories: make([]CategoryResponse, 0, len(categories)),
		Count:      len(categories),
	}

	for _, c := range categories {
		if cr := ToCategoryResponse(c, includeRelations); cr != nil {
			resp.Categories = append(resp.Categories, *cr)
		}
	}

	return resp
}

// ToCategoryStatsResponse converts category with stats to stats response
func ToCategoryStatsResponse(c *domain.Category) *CategoryStatsResponse {
	if c == nil {
		return nil
	}

	resp := &CategoryStatsResponse{
		CategoryID:       c.ID.String(),
		CategoryName:     c.Name,
		TransactionCount: c.TransactionCount,
		TotalAmount:      c.TotalAmount,
		MonthlyBudget:    c.MonthlyBudget,
	}

	// Calculate average
	if c.TransactionCount > 0 {
		resp.AverageAmount = c.TotalAmount / float64(c.TransactionCount)
	}

	// Calculate budget usage if budget exists
	if c.MonthlyBudget != nil && *c.MonthlyBudget > 0 {
		used := (c.TotalAmount / *c.MonthlyBudget) * 100
		resp.BudgetUsed = &used

		remaining := *c.MonthlyBudget - c.TotalAmount
		resp.BudgetRemaining = &remaining
	}

	return resp
}

// ========== Request to Entity Conversions ==========

// FromCreateCategoryRequest converts CreateCategoryRequest to Category entity
func FromCreateCategoryRequest(req CreateCategoryRequest, userID uuid.UUID) (*domain.Category, error) {
	// Parse category type
	categoryType := domain.CategoryType(req.Type)
	if !domain.ValidateCategoryType(categoryType) {
		return nil, domain.ErrInvalidCategoryType
	}

	category := &domain.Category{
		UserID:        &userID,
		Name:          req.Name,
		Description:   req.Description,
		Type:          categoryType,
		Icon:          req.Icon,
		Color:         req.Color,
		IsDefault:     false,
		IsActive:      true,
		MonthlyBudget: req.MonthlyBudget,
		DisplayOrder:  getDefaultDisplayOrder(req.DisplayOrder),
		Level:         0, // Will be calculated if has parent
	}
	category.ID = uuid.New()

	// Parse parent ID if provided
	if req.ParentID != nil {
		parentUUID, err := uuid.Parse(*req.ParentID)
		if err != nil {
			return nil, domain.ErrInvalidParentID
		}
		category.ParentID = &parentUUID
	}

	return category, nil
}

// ApplyUpdateCategoryRequest applies update request to existing category
func ApplyUpdateCategoryRequest(req UpdateCategoryRequest) (map[string]any, error) {
	updates := make(map[string]any)

	if req.Name != nil {
		updates["name"] = *req.Name
	}

	if req.Description != nil {
		if *req.Description == "" {
			updates["description"] = nil
		} else {
			updates["description"] = *req.Description
		}
	}

	if req.Type != nil {
		categoryType := domain.CategoryType(*req.Type)
		if !domain.ValidateCategoryType(categoryType) {
			return nil, domain.ErrInvalidCategoryType
		}
		updates["type"] = string(categoryType)
	}

	if req.ParentID != nil {
		parentUUID, err := uuid.Parse(*req.ParentID)
		if err != nil {
			return nil, domain.ErrInvalidParentID
		}
		updates["parent_id"] = parentUUID
	}

	if req.Icon != nil {
		if *req.Icon == "" {
			updates["icon"] = nil
		} else {
			updates["icon"] = *req.Icon
		}
	}

	if req.Color != nil {
		if *req.Color == "" {
			updates["color"] = nil
		} else {
			updates["color"] = *req.Color
		}
	}

	if req.MonthlyBudget != nil {
		updates["monthly_budget"] = *req.MonthlyBudget
	}

	if req.DisplayOrder != nil {
		updates["display_order"] = *req.DisplayOrder
	}

	if req.IsActive != nil {
		updates["is_active"] = *req.IsActive
	}

	return updates, nil
}

// FromDefaultCategories converts default category definitions to entities
func FromDefaultCategories(userID uuid.UUID, includeIncome, includeExpense bool) []*domain.Category {
	categories := make([]*domain.Category, 0)

	var displayOrderCounter int

	// Helper to convert recursive definition
	var convert func(def domain.DefaultCategoryDef, level int, parentID *uuid.UUID) *domain.Category
	convert = func(def domain.DefaultCategoryDef, level int, parentID *uuid.UUID) *domain.Category {
		desc := def.Description
		icon := def.Icon
		color := def.Color

		cat := &domain.Category{
			ID:           uuid.New(),
			UserID:       &userID,
			Name:         def.Name,
			Description:  &desc,
			Type:         def.Type,
			Icon:         &icon,
			Color:        &color,
			IsDefault:    true,
			IsActive:     true,
			Level:        level,
			DisplayOrder: displayOrderCounter,
			ParentID:     parentID,
		}

		// Increment global display order counter for flat sorting fallback
		// (though hierarchical sorting usually relies on parent -> children order)
		displayOrderCounter++

		// Process children
		if len(def.SubCategories) > 0 {
			cat.Children = make([]*domain.Category, 0, len(def.SubCategories))
			for _, sub := range def.SubCategories {
				child := convert(sub, level+1, &cat.ID) // Pass parent ID for reference, though GORM can handle object ref
				child.Parent = cat                      // manual link
				cat.Children = append(cat.Children, child)
			}
		}

		return cat
	}

	// Add expense categories
	if includeExpense {
		expenseDefaults := domain.DefaultExpenseCategories()
		for _, def := range expenseDefaults {
			cat := convert(def, 0, nil)
			categories = append(categories, cat)
		}
	}

	// Add income categories
	if includeIncome {
		incomeDefaults := domain.DefaultIncomeCategories()
		for _, def := range incomeDefaults {
			cat := convert(def, 0, nil)
			categories = append(categories, cat)
		}
	}

	return categories
}

// ========== Helper Functions ==========

func getDefaultDisplayOrder(order *int) int {
	if order != nil {
		return *order
	}
	return 0
}
