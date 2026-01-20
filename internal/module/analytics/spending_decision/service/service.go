package service

import (
	"context"
	spending_decision "personalfinancedss/internal/module/analytics/models/spending_decision"
	"personalfinancedss/internal/module/analytics/spending_decision/dto"

	"go.uber.org/zap"
)

type Service interface {
	ExecuteLargePurchase(ctx context.Context, input *dto.LargePurchaseInput) (*dto.LargePurchaseOutput, error)
}

type service struct {
	model  *spending_decision.LargePurchaseModel
	logger *zap.Logger
}

func NewService(logger *zap.Logger) Service {
	return &service{
		model:  spending_decision.NewSpendingDecisionModel(),
		logger: logger,
	}
}

func (s *service) ExecuteLargePurchase(ctx context.Context, input *dto.LargePurchaseInput) (*dto.LargePurchaseOutput, error) {
	s.logger.Info("Executing Large Purchase Analysis",
		zap.String("user_id", input.UserID),
		zap.String("item", input.ItemName),
		zap.Float64("amount", input.PurchaseAmount),
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

	output := result.(*dto.LargePurchaseOutput)
	s.logger.Info("Large Purchase Analysis completed",
		zap.String("decision", output.Recommendation.Decision),
	)

	return output, nil
}
