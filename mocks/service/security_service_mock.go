package service_mock

import (
	"github.com/stretchr/testify/mock"
)

type MockSecurityService struct {
	mock.Mock
}

func NewMockSecurityService() *MockSecurityService {
	return &MockSecurityService{}
}

func (m *MockSecurityService) ValidateWithHash(plain string, hash string) bool {
	args := m.Called(plain, hash)
	return args.Bool(0)
}

func (m *MockSecurityService) HashAndSalt(plain string) string {
	args := m.Called(plain)
	return args.String(0)
}
