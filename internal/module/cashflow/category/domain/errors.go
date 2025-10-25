package domain

import "errors"

var (
	// ErrInvalidCategoryType is returned when category type is invalid
	ErrInvalidCategoryType = errors.New("invalid category type")

	// ErrInvalidParentID is returned when parent ID is invalid
	ErrInvalidParentID = errors.New("invalid parent ID")

	// ErrCategoryNotFound is returned when category is not found
	ErrCategoryNotFound = errors.New("category not found")

	// ErrCategoryNameExists is returned when category name already exists
	ErrCategoryNameExists = errors.New("category name already exists")

	// ErrCannotDeleteSystemCategory is returned when trying to delete system category
	ErrCannotDeleteSystemCategory = errors.New("cannot delete system category")

	// ErrCannotUpdateSystemCategory is returned when trying to update system category
	ErrCannotUpdateSystemCategory = errors.New("cannot update system category")

	// ErrCategoryHasChildren is returned when trying to delete category with children
	ErrCategoryHasChildren = errors.New("category has children")

	// ErrCircularReference is returned when parent would create circular reference
	ErrCircularReference = errors.New("circular reference detected")
)
