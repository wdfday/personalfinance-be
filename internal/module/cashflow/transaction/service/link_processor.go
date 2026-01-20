package service

import (
	"context"
	"fmt"

	budgetService "personalfinancedss/internal/module/cashflow/budget/service"
	debtService "personalfinancedss/internal/module/cashflow/debt/service"
	goalService "personalfinancedss/internal/module/cashflow/goal/service"
	incomeProfileService "personalfinancedss/internal/module/cashflow/income_profile/service"
	"personalfinancedss/internal/module/cashflow/transaction/domain"
	"personalfinancedss/internal/shared"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// LinkProcessor handles processing of transaction links to update related entities
type LinkProcessor struct {
	goalService          goalService.Service
	budgetService        budgetService.Service
	debtService          debtService.Service
	incomeProfileService incomeProfileService.Service
	logger               *zap.Logger
}

// NewLinkProcessor creates a new link processor
func NewLinkProcessor(
	goalSvc goalService.Service,
	budgetSvc budgetService.Service,
	debtSvc debtService.Service,
	incomeProfileSvc incomeProfileService.Service,
	logger *zap.Logger,
) *LinkProcessor {
	return &LinkProcessor{
		goalService:          goalSvc,
		budgetService:        budgetSvc,
		debtService:          debtSvc,
		incomeProfileService: incomeProfileSvc,
		logger:               logger,
	}
}

// ProcessLinks processes transaction links and updates related entities
// For GOAL: adds contribution to the goal
// For BUDGET: recalculates budget spending
// For DEBT: adds payment to the debt
// For INCOME_PROFILE: validates the link (no update needed, just for analytics)
func (p *LinkProcessor) ProcessLinks(ctx context.Context, userID uuid.UUID, amount int64, direction domain.Direction, links []domain.TransactionLink) error {
	if len(links) == 0 {
		return nil
	}

	for _, link := range links {
		if err := p.processLink(ctx, userID, amount, direction, link); err != nil {
			return err
		}
	}

	return nil
}

// processLink processes a single link
func (p *LinkProcessor) processLink(ctx context.Context, userID uuid.UUID, amount int64, direction domain.Direction, link domain.TransactionLink) error {
	linkID, err := uuid.Parse(link.ID)
	if err != nil {
		return shared.ErrBadRequest.WithDetails("reason", fmt.Sprintf("invalid link ID: %s", link.ID))
	}

	switch link.Type {
	case domain.LinkGoal:
		return p.processGoalLink(ctx, userID, linkID, amount, direction)
	case domain.LinkBudget:
		return p.processBudgetLink(ctx, userID, linkID)
	case domain.LinkDebt:
		return p.processDebtLink(ctx, linkID, amount, direction)
	case domain.LinkIncomeProfile:
		return p.processIncomeProfileLink(ctx, userID, linkID)
	default:
		return shared.ErrBadRequest.WithDetails("reason", fmt.Sprintf("unknown link type: %s", link.Type))
	}
}

// processGoalLink adds contribution to the goal
// Only CREDIT transactions add to the goal
func (p *LinkProcessor) processGoalLink(ctx context.Context, userID uuid.UUID, goalID uuid.UUID, amount int64, direction domain.Direction) error {
	// Verify goal exists and belongs to user
	goal, err := p.goalService.GetGoalByID(ctx, goalID)
	if err != nil {
		p.logger.Error("Failed to get goal for link",
			zap.String("goal_id", goalID.String()),
			zap.Error(err),
		)
		return shared.ErrNotFound.WithDetails("reason", "goal not found")
	}

	if goal.UserID != userID {
		return shared.ErrForbidden.WithDetails("reason", "goal does not belong to user")
	}

	// Only add contribution for CREDIT transactions (money coming in for the goal)
	if direction == domain.DirectionCredit {
		amountFloat := float64(amount)
		note := "Contribution from transaction"
		source := "transaction_link"
		if _, err := p.goalService.AddContribution(ctx, goalID, amountFloat, &note, source); err != nil {
			p.logger.Error("Failed to add contribution to goal",
				zap.String("goal_id", goalID.String()),
				zap.Float64("amount", amountFloat),
				zap.Error(err),
			)
			return shared.ErrInternal.WithError(err)
		}

		p.logger.Info("Added contribution to goal from transaction",
			zap.String("goal_id", goalID.String()),
			zap.Float64("amount", amountFloat),
		)
	}

	return nil
}

// processBudgetLink recalculates budget spending
func (p *LinkProcessor) processBudgetLink(ctx context.Context, userID uuid.UUID, budgetID uuid.UUID) error {
	// Verify budget exists and belongs to user
	budget, err := p.budgetService.GetBudgetByIDForUser(ctx, budgetID, userID)
	if err != nil {
		p.logger.Error("Failed to get budget for link",
			zap.String("budget_id", budgetID.String()),
			zap.Error(err),
		)
		return shared.ErrNotFound.WithDetails("reason", "budget not found")
	}

	if budget.UserID != userID {
		return shared.ErrForbidden.WithDetails("reason", "budget does not belong to user")
	}

	// Recalculate budget spending
	if err := p.budgetService.RecalculateBudgetSpendingForUser(ctx, budgetID, userID); err != nil {
		p.logger.Error("Failed to recalculate budget spending",
			zap.String("budget_id", budgetID.String()),
			zap.Error(err),
		)
		return shared.ErrInternal.WithError(err)
	}

	p.logger.Info("Recalculated budget spending from transaction link",
		zap.String("budget_id", budgetID.String()),
	)

	return nil
}

// processDebtLink adds payment to the debt
// Only DEBIT transactions pay off debt (money going out to pay debt)
func (p *LinkProcessor) processDebtLink(ctx context.Context, debtID uuid.UUID, amount int64, direction domain.Direction) error {
	// Only process DEBIT transactions (payment going out)
	if direction == domain.DirectionDebit {
		amountFloat := float64(amount)
		_, err := p.debtService.AddPayment(ctx, debtID, amountFloat)
		if err != nil {
			p.logger.Error("Failed to add payment to debt",
				zap.String("debt_id", debtID.String()),
				zap.Float64("amount", amountFloat),
				zap.Error(err),
			)
			return shared.ErrInternal.WithError(err)
		}

		p.logger.Info("Added payment to debt from transaction",
			zap.String("debt_id", debtID.String()),
			zap.Float64("amount", amountFloat),
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
		case domain.LinkGoal:
			goal, err := p.goalService.GetGoalByID(ctx, linkID)
			if err != nil {
				return shared.ErrNotFound.WithDetails("reason", fmt.Sprintf("goal not found: %s", link.ID))
			}
			if goal.UserID != userID {
				return shared.ErrForbidden.WithDetails("reason", "goal does not belong to user")
			}
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
