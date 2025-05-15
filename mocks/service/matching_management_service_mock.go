package service_mock

import (
	"github.com/stretchr/testify/mock"
)

type MockMatchingManagementService struct {
	mock.Mock
}

func NewMockMatchingManagementService() *MockMatchingManagementService {
	return &MockMatchingManagementService{}
}

func (m *MockMatchingManagementService) GetMatchingByActivityID(activityID string, method string, token string) (interface{}, error) {
	args := m.Called(activityID, method, token)
	return args.Get(0), args.Error(1)
}

func (m *MockMatchingManagementService) GetEquivalentsByRegistrationID(registrationID string, method string, token string) (interface{}, error) {
	args := m.Called(registrationID, method, token)
	return args.Get(0), args.Error(1)
}
