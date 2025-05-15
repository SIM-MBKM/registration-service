package service_mock

import (
	"context"
	"mime/multipart"
	"registration-service/dto"

	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

type MockDocumentService struct {
	mock.Mock
}

func NewMockDocumentService() *MockDocumentService {
	return &MockDocumentService{}
}

func (m *MockDocumentService) FindAllDocuments(ctx context.Context, pagReq dto.PaginationRequest, tx *gorm.DB) ([]dto.DocumentResponse, dto.PaginationResponse, error) {
	args := m.Called(ctx, pagReq, tx)
	return args.Get(0).([]dto.DocumentResponse), args.Get(1).(dto.PaginationResponse), args.Error(2)
}

func (m *MockDocumentService) FindDocumentById(ctx context.Context, id string, tx *gorm.DB) (dto.DocumentResponse, error) {
	args := m.Called(ctx, id, tx)
	return args.Get(0).(dto.DocumentResponse), args.Error(1)
}

func (m *MockDocumentService) CreateDocument(ctx context.Context, document dto.DocumentRequest, file *multipart.FileHeader, tx *gorm.DB) error {
	args := m.Called(ctx, document, file, tx)
	return args.Error(0)
}

func (m *MockDocumentService) UpdateDocument(ctx context.Context, id string, document dto.UpdateDocumentRequest, file *multipart.FileHeader, tx *gorm.DB) error {
	args := m.Called(ctx, id, document, file, tx)
	return args.Error(0)
}

func (m *MockDocumentService) DeleteDocument(ctx context.Context, id string, tx *gorm.DB) error {
	args := m.Called(ctx, id, tx)
	return args.Error(0)
}
