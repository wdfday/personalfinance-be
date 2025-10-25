package service

import (
	"context"
	"strings"
	"time"

	"personalfinancedss/internal/module/cashflow/account/domain"
	accountdto "personalfinancedss/internal/module/cashflow/account/dto"
	"personalfinancedss/internal/shared"
)

// UpdateAccount updates an existing account
func (s *accountService) UpdateAccount(ctx context.Context, id, userID string, req accountdto.UpdateAccountRequest) (*domain.Account, error) {
	account, err := s.repo.GetByIDAndUserID(ctx, id, userID)
	if err != nil {
		if err == shared.ErrNotFound {
			return nil, err
		}
		return nil, shared.ErrInternal.WithError(err)
	}

	updates := make(map[string]any)

	if req.AccountName != nil {
		updates["account_name"] = strings.TrimSpace(*req.AccountName)
	}
	if req.AccountType != nil {
		accountType, err := parseAccountType(*req.AccountType)
		if err != nil {
			return nil, err
		}
		updates["account_type"] = string(accountType)
	}
	if req.InstitutionName != nil {
		updates["institution_name"] = normalizeNullableString(req.InstitutionName)
	}
	if req.CurrentBalance != nil {
		updates["current_balance"] = *req.CurrentBalance
	}
	if req.AvailableBalance != nil {
		updates["available_balance"] = req.AvailableBalance
	}
	if req.Currency != nil {
		updates["currency"] = strings.ToUpper(strings.TrimSpace(*req.Currency))
	}
	if req.AccountNumberMasked != nil {
		updates["account_number_masked"] = normalizeNullableString(req.AccountNumberMasked)
	}
	if req.IsActive != nil {
		updates["is_active"] = *req.IsActive
	}
	if req.IsPrimary != nil {
		if *req.IsPrimary && !account.IsPrimary {
			if err := s.unsetPrimaryAccount(ctx, userID); err != nil {
				return nil, err
			}
		}
		updates["is_primary"] = *req.IsPrimary
	}
	if req.IncludeInNetWorth != nil {
		updates["include_in_net_worth"] = *req.IncludeInNetWorth
	}
	if req.SyncStatus != nil {
		status, err := parseSyncStatus(*req.SyncStatus)
		if err != nil {
			return nil, err
		}
		updates["sync_status"] = string(status)
		if status == domain.SyncStatusActive {
			now := time.Now().UTC()
			updates["last_synced_at"] = &now
		}
	}
	if req.SyncErrorMessage != nil {
		updates["sync_error_message"] = normalizeNullableString(req.SyncErrorMessage)
	}

	if len(updates) == 0 {
		return account, nil
	}

	if err := s.repo.UpdateColumns(ctx, id, updates); err != nil {
		if err == shared.ErrNotFound {
			return nil, err
		}
		return nil, shared.ErrInternal.WithError(err)
	}

	return s.GetByID(ctx, id, userID)
}

// unsetPrimaryAccount removes primary flag from all user accounts
func (s *accountService) unsetPrimaryAccount(ctx context.Context, userID string) error {
	accounts, err := s.repo.ListByUserID(ctx, userID, domain.ListAccountsFilter{IsPrimary: boolPtr(true)})
	if err != nil {
		return shared.ErrInternal.WithError(err)
	}

	for _, acc := range accounts {
		if acc.IsPrimary {
			if err := s.repo.UpdateColumns(ctx, acc.ID.String(), map[string]any{"is_primary": false}); err != nil {
				return err
			}
		}
	}

	return nil
}
