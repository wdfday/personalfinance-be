package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTransaction_TableName(t *testing.T) {
	tx := Transaction{}
	assert.Equal(t, "transactions", tx.TableName())
}
