package mocks

import (
	"context"
	"registration-service/dto"
	"registration-service/entity"

	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

// MockRegistrationRepository is a mock implementation of repository.RegistrationRepository
type MockRegistrationRepository struct {
	mock.Mock
}

// Index mocks the Index method
func (m *MockRegistrationRepository) Index(ctx context.Context, tx *gorm.DB, pagReq dto.PaginationRequest, filter dto.FilterRegistrationRequest) ([]entity.Registration, int64, error) {
	args := m.Called(ctx, tx, pagReq, filter)
	return args.Get(0).([]entity.Registration), args.Get(1).(int64), args.Error(2)
}

// Create mocks the Create method
func (m *MockRegistrationRepository) Create(ctx context.Context, registration entity.Registration, tx *gorm.DB) (entity.Registration, error) {
	args := m.Called(ctx, registration, tx)
	return args.Get(0).(entity.Registration), args.Error(1)
}

// Update mocks the Update method
func (m *MockRegistrationRepository) Update(ctx context.Context, id string, registration entity.Registration, tx *gorm.DB) error {
	args := m.Called(ctx, id, registration, tx)
	return args.Error(0)
}

// FindByID mocks the FindByID method
func (m *MockRegistrationRepository) FindByID(ctx context.Context, id string, tx *gorm.DB) (entity.Registration, error) {
	args := m.Called(ctx, id, tx)
	return args.Get(0).(entity.Registration), args.Error(1)
}

// Destroy mocks the Destroy method
func (m *MockRegistrationRepository) Destroy(ctx context.Context, id string, tx *gorm.DB) error {
	args := m.Called(ctx, id, tx)
	return args.Error(0)
}

// FilterSubQuery mocks the FilterSubQuery method
func (m *MockRegistrationRepository) FilterSubQuery(ctx context.Context, tx *gorm.DB, filter dto.FilterRegistrationRequest) *gorm.DB {
	args := m.Called(ctx, tx, filter)
	return args.Get(0).(*gorm.DB)
}

// FindTotal mocks the FindTotal method
func (m *MockRegistrationRepository) FindTotal(ctx context.Context, filter dto.FilterRegistrationRequest, tx *gorm.DB) (int64, error) {
	args := m.Called(ctx, filter, tx)
	return args.Get(0).(int64), args.Error(1)
}

// FindRegistrationByAdvisiorEmail mocks the FindRegistrationByAdvisiorEmail method
func (m *MockRegistrationRepository) FindRegistrationByAdvisiorEmail(ctx context.Context, email string, tx *gorm.DB) (entity.Registration, error) {
	args := m.Called(ctx, email, tx)
	return args.Get(0).(entity.Registration), args.Error(1)
}

// FindByNRP mocks the FindByNRP method
func (m *MockRegistrationRepository) FindByNRP(ctx context.Context, nrp string, tx *gorm.DB) (entity.Registration, error) {
	args := m.Called(ctx, nrp, tx)
	return args.Get(0).(entity.Registration), args.Error(1)
}

// FindByActivityIDAndNRP mocks the FindByActivityIDAndNRP method
func (m *MockRegistrationRepository) FindByActivityIDAndNRP(ctx context.Context, activityID string, nrp string, tx *gorm.DB) (entity.Registration, error) {
	args := m.Called(ctx, activityID, nrp, tx)
	return args.Get(0).(entity.Registration), args.Error(1)
}
