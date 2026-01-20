package tests

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestDebtDeleter_DeleteDebt_Success(t *testing.T) {
	svc, mockRepo := setupService()
	ctx := context.Background()
	id := uuid.New()

	mockRepo.On("Delete", ctx, id).Return(nil)

	err := svc.DeleteDebt(ctx, id)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestDebtDeleter_DeleteDebt_Error(t *testing.T) {
	svc, mockRepo := setupService()
	ctx := context.Background()
	id := uuid.New()

	mockRepo.On("Delete", ctx, id).Return(errors.New("db error"))

	err := svc.DeleteDebt(ctx, id)

	assert.Error(t, err)
	mockRepo.AssertExpectations(t)
}
