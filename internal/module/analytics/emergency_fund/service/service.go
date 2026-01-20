package service

import (
	"context"
	"personalfinancedss/internal/module/analytics/emergency_fund/dto"
	emergency_fund "personalfinancedss/internal/module/analytics/models/emergency_fund"

	"go.uber.org/zap"
)

type Service interface {
	Execute(ctx context.Context, input *dto.Emergency_fundInput) (*dto.Emergency_fundOutput, error)
}

type service struct {
	model  *emergency_fund.Emergency_fundModel
	logger *zap.Logger
}

func NewService(logger *zap.Logger) Service {
	return &service{
		model:  emergency_fund.NewEmergency_fundModel(),
		logger: logger,
	}
}

func (s *service) Execute(ctx context.Context, input *dto.Emergency_fundInput) (*dto.Emergency_fundOutput, error) {
	result, err := s.model.Execute(ctx, input)
	if err != nil {
		return nil, err
	}
	return result.(*dto.Emergency_fundOutput), nil
}
