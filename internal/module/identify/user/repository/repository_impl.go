package repository

import (
	"context"
	"errors"
	_ "fmt"
	"personalfinancedss/internal/module/identify/user/domain"
	"strings"
	"time"

	"personalfinancedss/internal/shared"

	"gorm.io/gorm"
	_ "gorm.io/gorm/clause"
)

type gormRepo struct {
	db *gorm.DB
}

func New(db *gorm.DB) Repository {
	return &gormRepo{db: db}
}

// helper: always filter out soft-deleted by default
func base(db *gorm.DB) *gorm.DB {
	return db.Where("deleted_at IS NULL")
}

func (r *gormRepo) Create(ctx context.Context, u *domain.User) error {
	return r.db.WithContext(ctx).Create(u).Error
}

func (r *gormRepo) GetByID(ctx context.Context, id string) (*domain.User, error) {
	var u domain.User
	if err := base(r.db).WithContext(ctx).First(&u, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, shared.ErrUserNotFound
		}
		return nil, err
	}
	return &u, nil
}

func (r *gormRepo) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	var u domain.User
	if err := base(r.db).WithContext(ctx).
		First(&u, "email = ?", strings.ToLower(email)).
		Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, shared.ErrUserNotFound
		}
		return nil, err
	}
	return &u, nil
}

func (r *gormRepo) Count(ctx context.Context, f domain.ListUsersFilter) (int64, error) {
	var cnt int64
	q := r.applyFilter(base(r.db), f)
	if err := q.WithContext(ctx).Model(&domain.User{}).Count(&cnt).Error; err != nil {
		return 0, err
	}
	return cnt, nil
}

func (r *gormRepo) List(ctx context.Context, f domain.ListUsersFilter, p shared.Pagination) (shared.Page[domain.User], error) {
	if p.Page <= 0 {
		p.Page = 1
	}
	if p.PerPage <= 0 || p.PerPage > 200 {
		p.PerPage = 20
	}
	q := r.applyFilter(base(r.db), f)

	// Sorting mặc định
	sort := p.Sort
	if strings.TrimSpace(sort) == "" {
		sort = "last_active_at desc"
	}

	var items []domain.User
	if err := q.WithContext(ctx).
		Order(sort).
		Limit(p.PerPage).
		Offset((p.Page - 1) * p.PerPage).
		Find(&items).Error; err != nil {
		return shared.Page[domain.User]{}, err
	}

	total, err := r.Count(ctx, f)
	if err != nil {
		return shared.Page[domain.User]{}, err
	}
	totalPages := int((total + int64(p.PerPage) - 1) / int64(p.PerPage))

	return shared.Page[domain.User]{
		Data:         items,
		TotalItems:   total,
		CurrentPage:  p.Page,
		ItemsPerPage: p.PerPage,
		TotalPages:   totalPages,
	}, nil
}

func (r *gormRepo) applyFilter(db *gorm.DB, f domain.ListUsersFilter) *gorm.DB {
	q := db
	if f.Query != "" {
		like := "%" + strings.ToLower(f.Query) + "%"
		q = q.Where(`
			lower(email) LIKE ? OR
			lower(full_name) LIKE ? OR
			lower(display_name) LIKE ?
		`, like, like, like)
	}
	if f.Role != nil {
		q = q.Where("role = ?", *f.Role)
	}
	if f.Status != nil {
		q = q.Where("status = ?", *f.Status)
	}
	if f.EmailVerified != nil {
		q = q.Where("email_verified = ?", *f.EmailVerified)
	}
	if !f.ActiveOnly {
		// cho phép lọc cả bản ghi đã xoá
		q = q.Session(&gorm.Session{}).Where("1=1") // giữ nguyên
	}
	return q
}

func (r *gormRepo) Update(ctx context.Context, u *domain.User) error {
	u.UpdatedAt = time.Now()
	return r.db.WithContext(ctx).Save(u).Error
}

func (r *gormRepo) UpdateColumns(ctx context.Context, id string, cols map[string]any) error {
	cols["updated_at"] = time.Now()
	tx := r.db.WithContext(ctx).Model(&domain.User{}).
		Where("id = ? AND deleted_at IS NULL", id).
		Updates(cols)
	if tx.Error != nil {
		return tx.Error
	}
	if tx.RowsAffected == 0 {
		return shared.ErrUserNotFound
	}
	return nil
}

func (r *gormRepo) SoftDelete(ctx context.Context, id string) error {
	now := time.Now()
	tx := r.db.WithContext(ctx).Model(&domain.User{}).
		Where("id = ? AND deleted_at IS NULL", id).
		Update("deleted_at", now)
	if tx.Error != nil {
		return tx.Error
	}
	if tx.RowsAffected == 0 {
		return shared.ErrUserNotFound
	}
	return nil
}

func (r *gormRepo) Restore(ctx context.Context, id string) error {
	tx := r.db.WithContext(ctx).Unscoped().Model(&domain.User{}).
		Where("id = ? AND deleted_at IS NOT NULL", id).
		Update("deleted_at", nil)
	if tx.Error != nil {
		return tx.Error
	}
	if tx.RowsAffected == 0 {
		return shared.ErrUserNotFound
	}
	return nil
}

func (r *gormRepo) HardDelete(ctx context.Context, id string) error {
	tx := r.db.WithContext(ctx).Unscoped().
		Where("id = ?", id).
		Delete(&domain.User{})
	if tx.Error != nil {
		return tx.Error
	}
	if tx.RowsAffected == 0 {
		return shared.ErrUserNotFound
	}
	return nil
}

func (r *gormRepo) MarkEmailVerified(ctx context.Context, id string, at time.Time) error {
	return r.UpdateColumns(ctx, id, map[string]any{
		"email_verified":    true,
		"email_verified_at": at,
	})
}

func (r *gormRepo) IncLoginAttempts(ctx context.Context, id string) error {
	tx := r.db.WithContext(ctx).Model(&domain.User{}).
		Where("id = ? AND deleted_at IS NULL", id).
		UpdateColumn("login_attempts", gorm.Expr("login_attempts + 1"))
	if tx.Error != nil {
		return tx.Error
	}
	if tx.RowsAffected == 0 {
		return shared.ErrUserNotFound
	}
	return nil
}

func (r *gormRepo) ResetLoginAttempts(ctx context.Context, id string) error {
	return r.UpdateColumns(ctx, id, map[string]any{"login_attempts": 0})
}

func (r *gormRepo) SetLockedUntil(ctx context.Context, id string, until *time.Time) error {
	return r.UpdateColumns(ctx, id, map[string]any{"locked_until": until})
}

func (r *gormRepo) UpdateLastLogin(ctx context.Context, id string, at time.Time, ip *string) error {
	return r.UpdateColumns(ctx, id, map[string]any{
		"last_login_at": at,
		"last_login_ip": ip,
	})
}
