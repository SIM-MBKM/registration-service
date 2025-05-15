package service_mock

import (
	"github.com/stretchr/testify/mock"
)

type MockActivityManagementService struct {
	mock.Mock
}

func NewMockActivityManagementService() *MockActivityManagementService {
	return &MockActivityManagementService{}
}

func (m *MockActivityManagementService) GetActivitiesData(data map[string]interface{}, method string, token string) []map[string]interface{} {
	args := m.Called(data, method, token)
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).([]map[string]interface{})
} 