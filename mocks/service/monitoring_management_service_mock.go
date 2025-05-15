package service_mock

import (
	"github.com/stretchr/testify/mock"
)

type MockMonitoringManagementService struct {
	mock.Mock
}

func NewMockMonitoringManagementService() *MockMonitoringManagementService {
	return &MockMonitoringManagementService{}
}

func (m *MockMonitoringManagementService) GetSyllabusByRegistrationID(registrationID string, token string) (interface{}, error) {
	args := m.Called(registrationID, token)
	return args.Get(0), args.Error(1)
}

func (m *MockMonitoringManagementService) GetTranscriptByRegistrationID(registrationID string, token string) (interface{}, error) {
	args := m.Called(registrationID, token)
	return args.Get(0), args.Error(1)
}

func (m *MockMonitoringManagementService) CreateReportSchedule(data map[string]interface{}, method string, token string) error {
	args := m.Called(data, method, token)
	return args.Error(0)
}
