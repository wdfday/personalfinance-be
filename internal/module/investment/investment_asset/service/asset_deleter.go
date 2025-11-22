package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"
)

// DeleteAsset soft deletes an asset
func (s *assetService) DeleteAsset(ctx context.Context, userID string, assetID string) error {
	// Parse IDs
	uid, err := uuid.Parse(userID)
	if err != nil {
		return fmt.Errorf("invalid user ID: %w", err)
	}

	aid, err := uuid.Parse(assetID)
	if err != nil {
		return fmt.Errorf("invalid asset ID: %w", err)
	}

	// Verify asset exists and belongs to user
	_, err = s.repo.GetByUserID(ctx, aid, uid)
	if err != nil {
		return fmt.Errorf("failed to get asset: %w", err)
	}

	// Delete asset
	if err := s.repo.Delete(ctx, aid); err != nil {
		return fmt.Errorf("failed to delete asset: %w", err)
	}

	return nil
}
