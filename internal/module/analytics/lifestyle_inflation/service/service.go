package service

import (
	"context"
	"personalfinancedss/internal/module/analytics/lifestyle_inflation/dto"
	lifestyle_inflation "personalfinancedss/internal/module/analytics/models/lifestyle_inflation"

	"go.uber.org/zap"
)

type Service interface {
	Execute(ctx context.Context, input *dto.Lifestyle_inflationInput) (*dto.Lifestyle_inflationOutput, error)
}

type service struct {
	model  *lifestyle_inflation.Lifestyle_inflationModel
	logger *zap.Logger
}

func NewService(logger *zap.Logger) Service {
	return &service{
		model:  lifestyle_inflation.NewLifestyle_inflationModel(),
		logger: logger,
	}
}

func (s *service) Execute(ctx context.Context, input *dto.Lifestyle_inflationInput) (*dto.Lifestyle_inflationOutput, error) {
	result, err := s.model.Execute(ctx, input)
	if err != nil {
		return nil, err
	}
	return result.(*dto.Lifestyle_inflationOutput), nil
}
