package mocks

import (
	"context"
	"registration-service/dto"
	"registration-service/entity"

	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

// MockDocumentRepository is a mock implementation of repository.DocumentRepository
type MockDocumentRepository struct {
	mock.Mock
}

// Index mocks the Index method
func (m *MockDocumentRepository) Index(ctx context.Context, pagReq dto.PaginationRequest, tx *gorm.DB) ([]entity.Document, int64, error) {
	args := m.Called(ctx, pagReq, tx)
	return args.Get(0).([]entity.Document), args.Get(1).(int64), args.Error(2)
}

// Create mocks the Create method
func (m *MockDocumentRepository) Create(ctx context.Context, document entity.Document, tx *gorm.DB) (entity.Document, error) {
	args := m.Called(ctx, document, tx)
	return args.Get(0).(entity.Document), args.Error(1)
}

// Update mocks the Update method
func (m *MockDocumentRepository) Update(ctx context.Context, id string, document entity.Document, tx *gorm.DB) error {
	args := m.Called(ctx, id, document, tx)
	return args.Error(0)
}

// FindByID mocks the FindByID method
func (m *MockDocumentRepository) FindByID(ctx context.Context, id string, tx *gorm.DB) (entity.Document, error) {
	args := m.Called(ctx, id, tx)
	return args.Get(0).(entity.Document), args.Error(1)
}

// DeleteByID mocks the DeleteByID method
func (m *MockDocumentRepository) DeleteByID(ctx context.Context, id string, tx *gorm.DB) error {
	args := m.Called(ctx, id, tx)
	return args.Error(0)
}

// FindTotal mocks the FindTotal method
func (m *MockDocumentRepository) FindTotal(ctx context.Context, tx *gorm.DB) (int64, error) {
	args := m.Called(ctx, tx)
	return args.Get(0).(int64), args.Error(1)
}

// GetAll mocks the GetAll method
func (m *MockDocumentRepository) GetAll() ([]entity.Document, error) {
	args := m.Called()
	return args.Get(0).([]entity.Document), args.Error(1)
}
