package repository_test

import (
	"context"
	"errors"
	"registration-service/dto"
	"registration-service/entity"
	"registration-service/mocks"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

// Helper to create a mock document entity
func createMockDocument() entity.Document {
	now := time.Now()
	return entity.Document{
		ID:             uuid.New(),
		RegistrationID: uuid.New().String(),
		FileStorageID:  "file-storage-123",
		Name:           "document1.pdf",
		DocumentType:   "Acceptence Letter",
		BaseModel: entity.BaseModel{
			CreatedAt: &now,
			UpdatedAt: &now,
		},
	}
}

func TestDocumentRepository_Create(t *testing.T) {
	// Create mock repository
	mockRepo := new(mocks.MockDocumentRepository)

	// Setup test data
	ctx := context.Background()
	document := createMockDocument()

	// Setup mock behavior
	mockRepo.On("Create", ctx, document, mock.Anything).Return(document, nil)

	// Call method
	result, err := mockRepo.Create(ctx, document, nil)

	// Assertions
	assert.NoError(t, err)
	assert.Equal(t, document, result)
	mockRepo.AssertExpectations(t)
}

func TestDocumentRepository_Create_Error(t *testing.T) {
	// Create mock repository
	mockRepo := new(mocks.MockDocumentRepository)

	// Setup test data
	ctx := context.Background()
	document := createMockDocument()
	expectedError := errors.New("database error")

	// Setup mock behavior
	mockRepo.On("Create", ctx, document, mock.Anything).Return(entity.Document{}, expectedError)

	// Call method
	_, err := mockRepo.Create(ctx, document, nil)

	// Assertions
	assert.Error(t, err)
	assert.Equal(t, expectedError, err)
	mockRepo.AssertExpectations(t)
}

func TestDocumentRepository_FindByID(t *testing.T) {
	// Create mock repository
	mockRepo := new(mocks.MockDocumentRepository)

	// Setup test data
	ctx := context.Background()
	document := createMockDocument()
	id := document.ID.String()

	// Setup mock behavior
	mockRepo.On("FindByID", ctx, id, mock.Anything).Return(document, nil)

	// Call method
	result, err := mockRepo.FindByID(ctx, id, nil)

	// Assertions
	assert.NoError(t, err)
	assert.Equal(t, document, result)
	mockRepo.AssertExpectations(t)
}

func TestDocumentRepository_FindByID_NotFound(t *testing.T) {
	// Create mock repository
	mockRepo := new(mocks.MockDocumentRepository)

	// Setup test data
	ctx := context.Background()
	id := uuid.New().String()
	mockError := gorm.ErrRecordNotFound

	// Setup mock behavior
	mockRepo.On("FindByID", ctx, id, mock.Anything).Return(entity.Document{}, mockError)

	// Call method
	result, err := mockRepo.FindByID(ctx, id, nil)

	// Assertions
	assert.Error(t, err)
	assert.Equal(t, gorm.ErrRecordNotFound, err)
	assert.Equal(t, entity.Document{}, result)
	mockRepo.AssertExpectations(t)
}

func TestDocumentRepository_Index(t *testing.T) {
	// Create mock repository
	mockRepo := new(mocks.MockDocumentRepository)

	// Setup test data
	ctx := context.Background()
	pagReq := dto.PaginationRequest{
		Limit:  10,
		Offset: 0,
	}
	documents := []entity.Document{createMockDocument(), createMockDocument()}
	expectedTotal := int64(2)

	// Setup mock behavior
	mockRepo.On("Index", ctx, pagReq, mock.Anything).Return(documents, expectedTotal, nil)

	// Call method
	results, total, err := mockRepo.Index(ctx, pagReq, nil)

	// Assertions
	assert.NoError(t, err)
	assert.Equal(t, expectedTotal, total)
	assert.Equal(t, documents, results)
	assert.Len(t, results, 2)
	mockRepo.AssertExpectations(t)
}

func TestDocumentRepository_Update(t *testing.T) {
	// Create mock repository
	mockRepo := new(mocks.MockDocumentRepository)

	// Setup test data
	ctx := context.Background()
	document := createMockDocument()
	id := document.ID.String()

	// Setup mock behavior
	mockRepo.On("Update", ctx, id, document, mock.Anything).Return(nil)

	// Call method
	err := mockRepo.Update(ctx, id, document, nil)

	// Assertions
	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestDocumentRepository_Update_Error(t *testing.T) {
	// Create mock repository
	mockRepo := new(mocks.MockDocumentRepository)

	// Setup test data
	ctx := context.Background()
	document := createMockDocument()
	id := document.ID.String()
	expectedError := errors.New("update error")

	// Setup mock behavior
	mockRepo.On("Update", ctx, id, document, mock.Anything).Return(expectedError)

	// Call method
	err := mockRepo.Update(ctx, id, document, nil)

	// Assertions
	assert.Error(t, err)
	assert.Equal(t, expectedError, err)
	mockRepo.AssertExpectations(t)
}

func TestDocumentRepository_DeleteByID(t *testing.T) {
	// Create mock repository
	mockRepo := new(mocks.MockDocumentRepository)

	// Setup test data
	ctx := context.Background()
	id := uuid.New().String()

	// Setup mock behavior
	mockRepo.On("DeleteByID", ctx, id, mock.Anything).Return(nil)

	// Call method
	err := mockRepo.DeleteByID(ctx, id, nil)

	// Assertions
	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestDocumentRepository_DeleteByID_Error(t *testing.T) {
	// Create mock repository
	mockRepo := new(mocks.MockDocumentRepository)

	// Setup test data
	ctx := context.Background()
	id := uuid.New().String()
	expectedError := errors.New("delete error")

	// Setup mock behavior
	mockRepo.On("DeleteByID", ctx, id, mock.Anything).Return(expectedError)

	// Call method
	err := mockRepo.DeleteByID(ctx, id, nil)

	// Assertions
	assert.Error(t, err)
	assert.Equal(t, expectedError, err)
	mockRepo.AssertExpectations(t)
}

func TestDocumentRepository_FindTotal(t *testing.T) {
	// Create mock repository
	mockRepo := new(mocks.MockDocumentRepository)

	// Setup test data
	ctx := context.Background()
	expectedTotal := int64(25)

	// Setup mock behavior
	mockRepo.On("FindTotal", ctx, mock.Anything).Return(expectedTotal, nil)

	// Call method
	total, err := mockRepo.FindTotal(ctx, nil)

	// Assertions
	assert.NoError(t, err)
	assert.Equal(t, expectedTotal, total)
	mockRepo.AssertExpectations(t)
}

func TestDocumentRepository_FindTotal_Error(t *testing.T) {
	// Create mock repository
	mockRepo := new(mocks.MockDocumentRepository)

	// Setup test data
	ctx := context.Background()
	expectedError := errors.New("database error")

	// Setup mock behavior
	mockRepo.On("FindTotal", ctx, mock.Anything).Return(int64(0), expectedError)

	// Call method
	total, err := mockRepo.FindTotal(ctx, nil)

	// Assertions
	assert.Error(t, err)
	assert.Equal(t, expectedError, err)
	assert.Equal(t, int64(0), total)
	mockRepo.AssertExpectations(t)
}

func TestDocumentRepository_GetAll(t *testing.T) {
	// Create mock repository
	mockRepo := new(mocks.MockDocumentRepository)

	// Setup test data
	documents := []entity.Document{createMockDocument(), createMockDocument()}

	// Setup mock behavior
	mockRepo.On("GetAll").Return(documents, nil)

	// Call method
	results, err := mockRepo.GetAll()

	// Assertions
	assert.NoError(t, err)
	assert.Equal(t, documents, results)
	assert.Len(t, results, 2)
	mockRepo.AssertExpectations(t)
}

func TestDocumentRepository_GetAll_Error(t *testing.T) {
	// Create mock repository
	mockRepo := new(mocks.MockDocumentRepository)

	// Setup test data
	expectedError := errors.New("database error")

	// Setup mock behavior
	mockRepo.On("GetAll").Return([]entity.Document{}, expectedError)

	// Call method
	results, err := mockRepo.GetAll()

	// Assertions
	assert.Error(t, err)
	assert.Equal(t, expectedError, err)
	assert.Empty(t, results)
	mockRepo.AssertExpectations(t)
}
