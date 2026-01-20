package service

import (
	"context"

	"personalfinancedss/internal/module/analytics/budget_allocation/dto"
	budget_allocation "personalfinancedss/internal/module/analytics/models/budget_allocation"

	"go.uber.org/zap"
)

// Service interface for budget allocation operations
type Service interface {
	// ExecuteBudgetAllocation runs the budget allocation model with given input
	ExecuteBudgetAllocation(ctx context.Context, input *dto.BudgetAllocationModelInput) (*dto.BudgetAllocationModelOutput, error)
}

// service implements Service interface using MBMS pattern
type service struct {
	model  *budget_allocation.BudgetAllocationModel
	logger *zap.Logger
}

// NewService creates a new budget allocation service
func NewService(model *budget_allocation.BudgetAllocationModel, logger *zap.Logger) Service {
	return &service{
		model:  model,
		logger: logger,
	}
}

// ExecuteBudgetAllocation runs the budget allocation model
func (s *service) ExecuteBudgetAllocation(ctx context.Context, input *dto.BudgetAllocationModelInput) (*dto.BudgetAllocationModelOutput, error) {
	s.logger.Info("Executing Budget Allocation model",
		zap.String("user_id", input.UserID.String()),
		zap.Int("year", input.Year),
		zap.Int("month", input.Month))

	// Validate input
	if err := s.model.Validate(ctx, input); err != nil {
		s.logger.Error("Budget allocation validation failed", zap.Error(err))
		return nil, err
	}

	// Execute model
	result, err := s.model.Execute(ctx, input)
	if err != nil {
		s.logger.Error("Budget allocation execution failed", zap.Error(err))
		return nil, err
	}

	output := result.(*dto.BudgetAllocationModelOutput)

	s.logger.Info("Budget allocation execution completed",
		zap.String("user_id", input.UserID.String()),
		zap.Bool("is_feasible", output.IsFeasible),
		zap.Int("scenarios_count", len(output.Scenarios)),
		zap.Int64("computation_time_ms", output.Metadata.ComputationTime))

	return output, nil
}
