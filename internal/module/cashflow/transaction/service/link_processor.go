package service

import (
	"context"
	"fmt"

	budgetService "personalfinancedss/internal/module/cashflow/budget/service"
	debtService "personalfinancedss/internal/module/cashflow/debt/service"
	incomeProfileService "personalfinancedss/internal/module/cashflow/income_profile/service"
	"personalfinancedss/internal/module/cashflow/transaction/domain"
	"personalfinancedss/internal/shared"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// LinkProcessor handles processing of transaction links to update related entities
type LinkProcessor struct {
	budgetService        budgetService.Service
	debtService          debtService.Service
	incomeProfileService incomeProfileService.Service
	logger               *zap.Logger
}

// NewLinkProcessor creates a new link processor
func NewLinkProcessor(
	budgetSvc budgetService.Service,
	debtSvc debtService.Service,
	incomeProfileSvc incomeProfileService.Service,
	logger *zap.Logger,
) *LinkProcessor {
	return &LinkProcessor{
		budgetService:        budgetSvc,
		debtService:          debtSvc,
		incomeProfileService: incomeProfileSvc,
		logger:               logger,
	}
}

// ProcessLinks processes transaction links and updates related entities
// For BUDGET: recalculates budget spending
// For DEBT: adds payment to the debt
// For INCOME_PROFILE: validates the link (no update needed, just for analytics)
func (p *LinkProcessor) ProcessLinks(ctx context.Context, userID uuid.UUID, amount int64, direction domain.Direction, links []domain.TransactionLink) error {
	if len(links) == 0 {
		p.logger.Debug("ProcessLinks: No links to process")
		return nil
	}

	p.logger.Info("ProcessLinks: Starting to process transaction links",
		zap.String("user_id", userID.String()),
		zap.Int64("amount", amount),
		zap.String("direction", string(direction)),
		zap.Int("links_count", len(links)),
		zap.Any("links", links),
	)

	for i, link := range links {
		p.logger.Info("ProcessLinks: Processing link",
			zap.Int("link_index", i+1),
			zap.Int("total_links", len(links)),
			zap.String("link_type", string(link.Type)),
			zap.String("link_id", link.ID),
		)
		if err := p.processLink(ctx, userID, amount, direction, link); err != nil {
			p.logger.Error("ProcessLinks: Failed to process link",
				zap.Int("link_index", i+1),
				zap.String("link_type", string(link.Type)),
				zap.String("link_id", link.ID),
				zap.Error(err),
			)
			return err
		}
		p.logger.Info("ProcessLinks: Successfully processed link",
			zap.Int("link_index", i+1),
			zap.String("link_type", string(link.Type)),
			zap.String("link_id", link.ID),
		)
	}

	p.logger.Info("ProcessLinks: Completed processing all links",
		zap.Int("total_links", len(links)),
	)

	return nil
}

// processLink processes a single link
func (p *LinkProcessor) processLink(ctx context.Context, userID uuid.UUID, amount int64, direction domain.Direction, link domain.TransactionLink) error {
	p.logger.Debug("processLink: Starting to process single link",
		zap.String("link_type", string(link.Type)),
		zap.String("link_id", link.ID),
		zap.String("user_id", userID.String()),
		zap.Int64("amount", amount),
		zap.String("direction", string(direction)),
	)

	linkID, err := uuid.Parse(link.ID)
	if err != nil {
		p.logger.Error("processLink: Invalid link ID",
			zap.String("link_id", link.ID),
			zap.Error(err),
		)
		return shared.ErrBadRequest.WithDetails("reason", fmt.Sprintf("invalid link ID: %s", link.ID))
	}

	switch link.Type {
	case domain.LinkBudget:
		p.logger.Debug("processLink: Routing to processBudgetLink")
		return p.processBudgetLink(ctx, userID, linkID)
	case domain.LinkDebt:
		p.logger.Debug("processLink: Routing to processDebtLink")
		return p.processDebtLink(ctx, linkID, amount, direction)
	case domain.LinkIncomeProfile:
		p.logger.Debug("processLink: Routing to processIncomeProfileLink")
		return p.processIncomeProfileLink(ctx, userID, linkID)
	default:
		p.logger.Error("processLink: Unknown link type",
			zap.String("link_type", string(link.Type)),
		)
		return shared.ErrBadRequest.WithDetails("reason", fmt.Sprintf("unknown link type: %s", link.Type))
	}
}

// processBudgetLink recalculates budget spending
func (p *LinkProcessor) processBudgetLink(ctx context.Context, userID uuid.UUID, budgetID uuid.UUID) error {
	p.logger.Info("processBudgetLink: Starting to process budget link",
		zap.String("budget_id", budgetID.String()),
		zap.String("user_id", userID.String()),
	)

	// Verify budget exists and belongs to user
	p.logger.Debug("processBudgetLink: Fetching budget",
		zap.String("budget_id", budgetID.String()),
		zap.String("user_id", userID.String()),
	)
	budget, err := p.budgetService.GetBudgetByIDForUser(ctx, budgetID, userID)
	if err != nil {
		p.logger.Error("processBudgetLink: Failed to get budget for link",
			zap.String("budget_id", budgetID.String()),
			zap.String("user_id", userID.String()),
			zap.Error(err),
		)
		return shared.ErrNotFound.WithDetails("reason", "budget not found")
	}

	p.logger.Debug("processBudgetLink: Budget found",
		zap.String("budget_id", budgetID.String()),
		zap.String("budget_name", budget.Name),
		zap.String("budget_user_id", budget.UserID.String()),
	)

	if budget.UserID != userID {
		p.logger.Error("processBudgetLink: Budget does not belong to user",
			zap.String("budget_id", budgetID.String()),
			zap.String("budget_user_id", budget.UserID.String()),
			zap.String("requested_user_id", userID.String()),
		)
		return shared.ErrForbidden.WithDetails("reason", "budget does not belong to user")
	}

	// Recalculate budget spending
	p.logger.Info("processBudgetLink: Calling RecalculateBudgetSpendingForUser",
		zap.String("budget_id", budgetID.String()),
		zap.String("user_id", userID.String()),
	)
	if err := p.budgetService.RecalculateBudgetSpendingForUser(ctx, budgetID, userID); err != nil {
		p.logger.Error("processBudgetLink: Failed to recalculate budget spending",
			zap.String("budget_id", budgetID.String()),
			zap.String("user_id", userID.String()),
			zap.Error(err),
		)
		return shared.ErrInternal.WithError(err)
	}

	p.logger.Info("processBudgetLink: Successfully recalculated budget spending",
		zap.String("budget_id", budgetID.String()),
		zap.String("user_id", userID.String()),
	)

	return nil
}

// processDebtLink adds payment to the debt
// Only DEBIT transactions pay off debt (money going out to pay debt)
func (p *LinkProcessor) processDebtLink(ctx context.Context, debtID uuid.UUID, amount int64, direction domain.Direction) error {
	p.logger.Info("processDebtLink: Starting to process debt link",
		zap.String("debt_id", debtID.String()),
		zap.Int64("amount", amount),
		zap.String("direction", string(direction)),
	)

	// Only process DEBIT transactions (payment going out)
	if direction == domain.DirectionDebit {
		amountFloat := float64(amount)
		p.logger.Info("processDebtLink: Processing DEBIT transaction - calling AddPayment",
			zap.String("debt_id", debtID.String()),
			zap.Float64("amount", amountFloat),
		)
		debt, err := p.debtService.AddPayment(ctx, debtID, amountFloat)
		if err != nil {
			p.logger.Error("processDebtLink: Failed to add payment to debt",
				zap.String("debt_id", debtID.String()),
				zap.Float64("amount", amountFloat),
				zap.Error(err),
			)
			return shared.ErrInternal.WithError(err)
		}

		p.logger.Info("processDebtLink: Successfully added payment to debt",
			zap.String("debt_id", debtID.String()),
			zap.Float64("amount", amountFloat),
			zap.Float64("debt_current_balance", debt.CurrentBalance),
			zap.Float64("debt_total_paid", debt.TotalPaid),
			zap.Float64("debt_percentage_paid", debt.PercentagePaid),
		)
	} else {
		p.logger.Debug("processDebtLink: Skipping non-DEBIT transaction",
			zap.String("debt_id", debtID.String()),
			zap.String("direction", string(direction)),
			zap.String("reason", "only DEBIT transactions can pay off debt"),
		)
	}

	return nil
}

// processIncomeProfileLink validates the income profile exists
// This is just for analytics linking, no amount updates
func (p *LinkProcessor) processIncomeProfileLink(ctx context.Context, userID uuid.UUID, profileID uuid.UUID) error {
	// Verify income profile exists and belongs to user
	_, err := p.incomeProfileService.GetIncomeProfile(ctx, userID.String(), profileID.String())
	if err != nil {
		p.logger.Error("Failed to get income profile for link",
			zap.String("profile_id", profileID.String()),
			zap.Error(err),
		)
		return shared.ErrNotFound.WithDetails("reason", "income profile not found")
	}

	p.logger.Debug("Linked transaction to income profile",
		zap.String("profile_id", profileID.String()),
	)

	return nil
}

// ValidateLinks validates all links before processing
func (p *LinkProcessor) ValidateLinks(ctx context.Context, userID uuid.UUID, links []domain.TransactionLink) error {
	for _, link := range links {
		linkID, err := uuid.Parse(link.ID)
		if err != nil {
			return shared.ErrBadRequest.WithDetails("reason", fmt.Sprintf("invalid link ID: %s", link.ID))
		}

		switch link.Type {
		case domain.LinkBudget:
			_, err := p.budgetService.GetBudgetByIDForUser(ctx, linkID, userID)
			if err != nil {
				return shared.ErrNotFound.WithDetails("reason", fmt.Sprintf("budget not found: %s", link.ID))
			}
		case domain.LinkDebt:
			debt, err := p.debtService.GetDebtByID(ctx, linkID)
			if err != nil {
				return shared.ErrNotFound.WithDetails("reason", fmt.Sprintf("debt not found: %s", link.ID))
			}
			_ = debt // debt doesn't have user check in interface
		case domain.LinkIncomeProfile:
			_, err := p.incomeProfileService.GetIncomeProfile(ctx, userID.String(), linkID.String())
			if err != nil {
				return shared.ErrNotFound.WithDetails("reason", fmt.Sprintf("income profile not found: %s", link.ID))
			}
		default:
			return shared.ErrBadRequest.WithDetails("reason", fmt.Sprintf("unknown link type: %s", link.Type))
		}
	}

	return nil
}
