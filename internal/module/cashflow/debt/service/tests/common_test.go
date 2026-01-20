package tests

import (
	"personalfinancedss/internal/module/cashflow/debt/service"

	"go.uber.org/zap"
)

func setupService() (service.Service, *MockRepository) {
	mockRepo := new(MockRepository)
	logger := zap.NewNop()
	svc := service.NewService(mockRepo, logger)
	return svc, mockRepo
}
