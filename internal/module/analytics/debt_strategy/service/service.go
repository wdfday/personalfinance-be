package service

import (
	"context"
	"personalfinancedss/internal/module/analytics/debt_strategy/dto"
	debt_strategy "personalfinancedss/internal/module/analytics/models/debt_strategy"

	"go.uber.org/zap"
)

type Service interface {
	ExecuteDebtStrategy(ctx context.Context, input *dto.DebtStrategyInput) (*dto.DebtStrategyOutput, error)
}

type service struct {
	model  *debt_strategy.DebtStrategyModel
	logger *zap.Logger
}

func NewService(logger *zap.Logger) Service {
	return &service{
		model:  debt_strategy.NewDebtStrategyModel(),
		logger: logger,
	}
}

func (s *service) ExecuteDebtStrategy(ctx context.Context, input *dto.DebtStrategyInput) (*dto.DebtStrategyOutput, error) {
	s.logger.Info("Executing debt strategy analysis",
		zap.String("user_id", input.UserID),
		zap.Int("debt_count", len(input.Debts)),
		zap.Float64("budget", input.TotalDebtBudget),
	)

	if err := s.model.Validate(ctx, input); err != nil {
		s.logger.Warn("Validation failed", zap.Error(err))
		return nil, err
	}

	result, err := s.model.Execute(ctx, input)
	if err != nil {
		s.logger.Error("Execution failed", zap.Error(err))
		return nil, err
	}

	output := result.(*dto.DebtStrategyOutput)

	s.logger.Info("Debt strategy analysis completed",
		zap.String("recommended_strategy", string(output.RecommendedStrategy)),
		zap.Int("months_to_debt_free", output.MonthsToDebtFree),
		zap.Float64("total_interest", output.TotalInterest),
	)

	return output, nil
}
