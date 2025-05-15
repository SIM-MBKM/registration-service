package repository_mock

import (
	"context"

	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

// MockBaseRepository is a mock implementation of repository.BaseRepository
type MockBaseRepository struct {
	mock.Mock
}

// BeginTx mocks the BeginTx method
func (m *MockBaseRepository) BeginTx(ctx context.Context) (*gorm.DB, error) {
	args := m.Called(ctx)
	return args.Get(0).(*gorm.DB), args.Error(1)
}

// CommitTx mocks the CommitTx method
func (m *MockBaseRepository) CommitTx(ctx context.Context, tx *gorm.DB) (*gorm.DB, error) {
	args := m.Called(ctx, tx)
	return args.Get(0).(*gorm.DB), args.Error(1)
}

// RollbackTx mocks the RollbackTx method
func (m *MockBaseRepository) RollbackTx(ctx context.Context, tx *gorm.DB) {
	m.Called(ctx, tx)
}
