package service_mock

import (
	"context"
	"mime/multipart"
	"registration-service/dto"

	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

type MockRegistrationService struct {
	mock.Mock
}

func NewMockRegistrationService() *MockRegistrationService {
	return &MockRegistrationService{}
}

func (m *MockRegistrationService) FindAllRegistrations(ctx context.Context, pagReq dto.PaginationRequest, filter dto.FilterRegistrationRequest, tx *gorm.DB, token string) ([]dto.GetRegistrationResponse, dto.PaginationResponse, error) {
	args := m.Called(ctx, pagReq, filter, tx, token)
	return args.Get(0).([]dto.GetRegistrationResponse), args.Get(1).(dto.PaginationResponse), args.Error(2)
}

func (m *MockRegistrationService) FindRegistrationByID(ctx context.Context, id string, token string, tx *gorm.DB) (dto.GetRegistrationResponse, error) {
	args := m.Called(ctx, id, token, tx)
	return args.Get(0).(dto.GetRegistrationResponse), args.Error(1)
}

func (m *MockRegistrationService) CreateRegistration(ctx context.Context, registration dto.CreateRegistrationRequest, file *multipart.FileHeader, geoletter *multipart.FileHeader, tx *gorm.DB, token string) error {
	args := m.Called(ctx, registration, file, geoletter, tx, token)
	return args.Error(0)
}

func (m *MockRegistrationService) UpdateRegistration(ctx context.Context, id string, registration dto.UpdateRegistrationDataRequest, token string, tx *gorm.DB) error {
	args := m.Called(ctx, id, registration, token, tx)
	return args.Error(0)
}

func (m *MockRegistrationService) DeleteRegistration(ctx context.Context, id string, token string, tx *gorm.DB) error {
	args := m.Called(ctx, id, token, tx)
	return args.Error(0)
}

func (m *MockRegistrationService) RegistrationsDataAccess(ctx context.Context, id string, token string, tx *gorm.DB) bool {
	args := m.Called(ctx, id, token, tx)
	return args.Bool(0)
}

func (m *MockRegistrationService) FindRegistrationByAdvisor(ctx context.Context, pagReq dto.PaginationRequest, filter dto.FilterRegistrationRequest, token string, tx *gorm.DB) ([]dto.GetRegistrationResponse, dto.PaginationResponse, error) {
	args := m.Called(ctx, pagReq, filter, token, tx)
	return args.Get(0).([]dto.GetRegistrationResponse), args.Get(1).(dto.PaginationResponse), args.Error(2)
}

func (m *MockRegistrationService) FindRegistrationByLOMBKM(ctx context.Context, pagReq dto.PaginationRequest, filter dto.FilterRegistrationRequest, token string, tx *gorm.DB) ([]dto.GetRegistrationResponse, dto.PaginationResponse, error) {
	args := m.Called(ctx, pagReq, filter, token, tx)
	return args.Get(0).([]dto.GetRegistrationResponse), args.Get(1).(dto.PaginationResponse), args.Error(2)
}

func (m *MockRegistrationService) FindRegistrationByStudent(ctx context.Context, pagReq dto.PaginationRequest, filter dto.FilterRegistrationRequest, token string, tx *gorm.DB) ([]dto.GetRegistrationResponse, dto.PaginationResponse, error) {
	args := m.Called(ctx, pagReq, filter, token, tx)
	return args.Get(0).([]dto.GetRegistrationResponse), args.Get(1).(dto.PaginationResponse), args.Error(2)
}

func (m *MockRegistrationService) ValidateAdvisor(ctx context.Context, token string, tx *gorm.DB) string {
	args := m.Called(ctx, token, tx)
	return args.String(0)
}

func (m *MockRegistrationService) ValidateStudent(ctx context.Context, token string, tx *gorm.DB) string {
	args := m.Called(ctx, token, tx)
	return args.String(0)
}

func (m *MockRegistrationService) AdvisorRegistrationApproval(ctx context.Context, token string, approval dto.ApprovalRequest, tx *gorm.DB) error {
	args := m.Called(ctx, token, approval, tx)
	return args.Error(0)
}

func (m *MockRegistrationService) LORegistrationApproval(ctx context.Context, token string, approval dto.ApprovalRequest, tx *gorm.DB) error {
	args := m.Called(ctx, token, approval, tx)
	return args.Error(0)
}

func (m *MockRegistrationService) GetRegistrationTranscript(ctx context.Context, id string, token string, tx *gorm.DB) (dto.TranscriptResponse, error) {
	args := m.Called(ctx, id, token, tx)
	return args.Get(0).(dto.TranscriptResponse), args.Error(1)
}

func (m *MockRegistrationService) GetStudentRegistrationsWithTranscripts(ctx context.Context, pagReq dto.PaginationRequest, filter dto.FilterRegistrationRequest, token string, tx *gorm.DB) (dto.StudentTranscriptsResponse, dto.PaginationResponse, error) {
	args := m.Called(ctx, pagReq, filter, token, tx)
	return args.Get(0).(dto.StudentTranscriptsResponse), args.Get(1).(dto.PaginationResponse), args.Error(2)
}

func (m *MockRegistrationService) GetStudentRegistrationsWithSyllabuses(ctx context.Context, pagReq dto.PaginationRequest, filter dto.FilterRegistrationRequest, token string, tx *gorm.DB) (dto.StudentSyllabusesResponse, dto.PaginationResponse, error) {
	args := m.Called(ctx, pagReq, filter, token, tx)
	return args.Get(0).(dto.StudentSyllabusesResponse), args.Get(1).(dto.PaginationResponse), args.Error(2)
}

func (m *MockRegistrationService) FindRegistrationsWithMatching(ctx context.Context, pagReq dto.PaginationRequest, filter dto.FilterRegistrationRequest, token string, tx *gorm.DB) (dto.StudentRegistrationsWithMatchingResponse, dto.PaginationResponse, error) {
	args := m.Called(ctx, pagReq, filter, token, tx)
	return args.Get(0).(dto.StudentRegistrationsWithMatchingResponse), args.Get(1).(dto.PaginationResponse), args.Error(2)
}

func (m *MockRegistrationService) CheckRegistrationEligibility(ctx context.Context, activityID string, token string, tx *gorm.DB) (dto.RegistrationEligibilityResponse, error) {
	args := m.Called(ctx, activityID, token, tx)
	return args.Get(0).(dto.RegistrationEligibilityResponse), args.Error(1)
}
