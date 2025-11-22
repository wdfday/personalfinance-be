package service

import (
	"context"
	"fmt"
	"time"

	"personalfinancedss/internal/module/investment/portfolio_snapshot/domain"
	"personalfinancedss/internal/module/investment/portfolio_snapshot/dto"
	"personalfinancedss/internal/module/investment/portfolio_snapshot/repository"

	"github.com/google/uuid"
)

// Service defines all portfolio snapshot operations
type Service interface {
	CreateSnapshot(ctx context.Context, userID string, req dto.CreateSnapshotRequest) (*domain.PortfolioSnapshot, error)
	GetSnapshot(ctx context.Context, userID string, snapshotID string) (*domain.PortfolioSnapshot, error)
	GetLatestSnapshot(ctx context.Context, userID string) (*domain.PortfolioSnapshot, error)
	ListSnapshots(ctx context.Context, userID string, query dto.ListSnapshotsQuery) (*dto.SnapshotListResponse, error)
	GetSnapshotsByPeriod(ctx context.Context, userID string, period string, limit int) ([]*domain.PortfolioSnapshot, error)
	DeleteSnapshot(ctx context.Context, userID string, snapshotID string) error
	GetPerformanceMetrics(ctx context.Context, userID string, startDate, endDate string) (*dto.PerformanceMetrics, error)
}

type snapshotService struct {
	repo repository.Repository
}

func NewService(repo repository.Repository) Service {
	return &snapshotService{
		repo: repo,
	}
}

func (s *snapshotService) CreateSnapshot(ctx context.Context, userID string, req dto.CreateSnapshotRequest) (*domain.PortfolioSnapshot, error) {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID: %w", err)
	}

	var snapshotDate time.Time
	if req.SnapshotDate != "" {
		snapshotDate, err = time.Parse("2006-01-02", req.SnapshotDate)
		if err != nil {
			return nil, fmt.Errorf("invalid snapshot date format: %w", err)
		}
	} else {
		snapshotDate = time.Now()
	}

	snapshotType := domain.SnapshotTypeManual
	if req.SnapshotType != nil {
		snapshotType = *req.SnapshotType
	}

	snapshot := &domain.PortfolioSnapshot{
		UserID:       uid,
		SnapshotDate: snapshotDate,
		SnapshotType: snapshotType,
		Period:       req.Period,
		Notes:        req.Notes,
	}

	// TODO: Calculate portfolio values from current asset data
	// This would typically fetch all active assets and calculate totals
	// For now, we'll create the snapshot with zeros

	snapshot.CalculateMetrics()

	if err := s.repo.Create(ctx, snapshot); err != nil {
		return nil, fmt.Errorf("failed to create snapshot: %w", err)
	}

	return snapshot, nil
}

func (s *snapshotService) GetSnapshot(ctx context.Context, userID string, snapshotID string) (*domain.PortfolioSnapshot, error) {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID: %w", err)
	}

	sid, err := uuid.Parse(snapshotID)
	if err != nil {
		return nil, fmt.Errorf("invalid snapshot ID: %w", err)
	}

	snapshot, err := s.repo.GetByUserID(ctx, sid, uid)
	if err != nil {
		return nil, fmt.Errorf("failed to get snapshot: %w", err)
	}

	return snapshot, nil
}

func (s *snapshotService) GetLatestSnapshot(ctx context.Context, userID string) (*domain.PortfolioSnapshot, error) {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID: %w", err)
	}

	snapshot, err := s.repo.GetLatest(ctx, uid)
	if err != nil {
		return nil, fmt.Errorf("failed to get latest snapshot: %w", err)
	}

	return snapshot, nil
}

func (s *snapshotService) ListSnapshots(ctx context.Context, userID string, query dto.ListSnapshotsQuery) (*dto.SnapshotListResponse, error) {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID: %w", err)
	}

	if query.Page < 1 {
		query.Page = 1
	}
	if query.PageSize < 1 {
		query.PageSize = 30
	}

	snapshots, total, err := s.repo.List(ctx, uid, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list snapshots: %w", err)
	}

	response := dto.ToSnapshotListResponse(snapshots, total, query.Page, query.PageSize)
	return &response, nil
}

func (s *snapshotService) GetSnapshotsByPeriod(ctx context.Context, userID string, period string, limit int) ([]*domain.PortfolioSnapshot, error) {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID: %w", err)
	}

	p := domain.SnapshotPeriod(period)
	if !p.IsValid() {
		return nil, fmt.Errorf("invalid period: %s", period)
	}

	if limit < 1 {
		limit = 30
	}
	if limit > 365 {
		limit = 365
	}

	snapshots, err := s.repo.GetByPeriod(ctx, uid, p, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get snapshots by period: %w", err)
	}

	return snapshots, nil
}

func (s *snapshotService) DeleteSnapshot(ctx context.Context, userID string, snapshotID string) error {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return fmt.Errorf("invalid user ID: %w", err)
	}

	sid, err := uuid.Parse(snapshotID)
	if err != nil {
		return fmt.Errorf("invalid snapshot ID: %w", err)
	}

	_, err = s.repo.GetByUserID(ctx, sid, uid)
	if err != nil {
		return fmt.Errorf("failed to get snapshot: %w", err)
	}

	if err := s.repo.Delete(ctx, sid); err != nil {
		return fmt.Errorf("failed to delete snapshot: %w", err)
	}

	return nil
}

func (s *snapshotService) GetPerformanceMetrics(ctx context.Context, userID string, startDate, endDate string) (*dto.PerformanceMetrics, error) {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID: %w", err)
	}

	start, err := time.Parse("2006-01-02", startDate)
	if err != nil {
		return nil, fmt.Errorf("invalid start date format: %w", err)
	}

	end, err := time.Parse("2006-01-02", endDate)
	if err != nil {
		return nil, fmt.Errorf("invalid end date format: %w", err)
	}

	metrics, err := s.repo.GetPerformanceMetrics(ctx, uid, start, end)
	if err != nil {
		return nil, fmt.Errorf("failed to get performance metrics: %w", err)
	}

	return metrics, nil
}
