package service

import (
	"context"

	"personalfinancedss/internal/module/analytics/debt_tradeoff/dto"
	tradeoff "personalfinancedss/internal/module/analytics/models/tradeoff"

	"go.uber.org/zap"
)

type Service interface {
	ExecuteTradeoff(ctx context.Context, input *dto.TradeoffInput) (*dto.TradeoffOutput, error)
}

type service struct {
	model  *tradeoff.TradeoffModel
	logger *zap.Logger
}

func NewService(logger *zap.Logger) Service {
	return &service{
		model:  tradeoff.NewTradeoffModel(),
		logger: logger,
	}
}

func (s *service) ExecuteTradeoff(ctx context.Context, input *dto.TradeoffInput) (*dto.TradeoffOutput, error) {
	if err := s.model.Validate(ctx, input); err != nil {
		s.logger.Error("Tradeoff validation failed", zap.Error(err))
		return nil, err
	}

	result, err := s.model.Execute(ctx, input)
	if err != nil {
		s.logger.Error("Tradeoff execution failed", zap.Error(err))
		return nil, err
	}

	return result.(*dto.TradeoffOutput), nil
}
