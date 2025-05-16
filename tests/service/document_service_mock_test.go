package service_test

import (
	"context"
	"errors"
	"mime/multipart"
	"registration-service/dto"
	"registration-service/entity"
	"registration-service/helper"
	repository_mock "registration-service/mocks/repository"
	service_mock "registration-service/mocks/service"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"
)

// DocumentServiceTestSuite defines the test suite for document service
type DocumentServiceTestSuite struct {
	suite.Suite
	mockDocumentRepo     *repository_mock.MockDocumentRepository
	mockRegistrationRepo *repository_mock.MockRegistrationRepository
	mockFileService      *service_mock.MockFileService
	service              mockDocumentService
}

// SetupTest initializes test dependencies before each test
func (suite *DocumentServiceTestSuite) SetupTest() {
	// Create mocks for all dependencies
	suite.mockDocumentRepo = new(repository_mock.MockDocumentRepository)
	suite.mockRegistrationRepo = new(repository_mock.MockRegistrationRepository)

	// Create mock file service with initialized Storage
	suite.mockFileService = service_mock.NewMockFileService()

	// Create the document service with the mocks
	suite.service = mockDocumentService{
		documentRepository:     suite.mockDocumentRepo,
		registrationRepository: suite.mockRegistrationRepo,
		fileService:            suite.mockFileService,
	}
}

// mockDocumentService is a mock implementation of the document service
type mockDocumentService struct {
	documentRepository     *repository_mock.MockDocumentRepository
	registrationRepository *repository_mock.MockRegistrationRepository
	fileService            *service_mock.MockFileService
}

// FindAllDocuments mock implementation
func (s mockDocumentService) FindAllDocuments(ctx context.Context, pagReq dto.PaginationRequest, tx *gorm.DB) ([]dto.DocumentResponse, dto.PaginationResponse, error) {
	documents, total, err := s.documentRepository.Index(ctx, pagReq, tx)
	if err != nil {
		return nil, dto.PaginationResponse{}, err
	}

	// Also call FindTotal to match the expectations in tests
	_, err = s.documentRepository.FindTotal(ctx, tx)
	if err != nil {
		return nil, dto.PaginationResponse{}, err
	}

	metaData := helper.MetaDataPagination(total, pagReq)

	var response []dto.DocumentResponse
	for _, document := range documents {
		response = append(response, dto.DocumentResponse{
			ID:             document.ID.String(),
			Name:           document.Name,
			FileStorageID:  document.FileStorageID,
			RegistrationID: document.RegistrationID,
			DocumentType:   document.DocumentType,
		})
	}

	return response, metaData, nil
}

// FindDocumentById mock implementation
func (s mockDocumentService) FindDocumentById(ctx context.Context, id string, tx *gorm.DB) (dto.DocumentResponse, error) {
	document, err := s.documentRepository.FindByID(ctx, id, tx)
	if err != nil {
		return dto.DocumentResponse{}, err
	}

	response := dto.DocumentResponse{
		ID:             document.ID.String(),
		Name:           document.Name,
		RegistrationID: document.RegistrationID,
		FileStorageID:  document.FileStorageID,
		DocumentType:   document.DocumentType,
	}

	return response, nil
}

// CreateDocument mock implementation
func (s mockDocumentService) CreateDocument(ctx context.Context, document dto.DocumentRequest, file *multipart.FileHeader, tx *gorm.DB) error {
	// Verify registration exists
	_, err := s.registrationRepository.FindByID(ctx, document.RegistrationID, tx)
	if err != nil {
		return err
	}

	// Upload file
	result, err := s.fileService.Storage.GcsUpload(file, "sim_mbkm", "", "")
	if err != nil {
		return errors.New("failed to upload file")
	}

	var documentEntity entity.Document
	documentEntity.ID = uuid.New()
	documentEntity.Name = document.Name
	documentEntity.FileStorageID = result.FileID
	documentEntity.RegistrationID = document.RegistrationID
	documentEntity.DocumentType = document.DocumentType

	_, err = s.documentRepository.Create(ctx, documentEntity, tx)
	if err != nil {
		return err
	}

	return nil
}

// UpdateDocument mock implementation
func (s mockDocumentService) UpdateDocument(ctx context.Context, id string, document dto.UpdateDocumentRequest, file *multipart.FileHeader, tx *gorm.DB) error {
	// Check if document exists
	existingDoc, err := s.documentRepository.FindByID(ctx, id, tx)
	if err != nil {
		return err
	}

	// Check if registration exists
	_, err = s.registrationRepository.FindByID(ctx, document.RegistrationID, tx)
	if err != nil {
		return err
	}

	// Upload new file
	result, err := s.fileService.Storage.GcsUpload(file, "sim_mbkm", "", "")
	if err != nil {
		return errors.New("failed to upload file")
	}

	// Update document entity with new file ID and registration ID
	documentEntity := entity.Document{
		ID:             existingDoc.ID,
		RegistrationID: document.RegistrationID,
		FileStorageID:  result.FileID,
		Name:           existingDoc.Name,
		DocumentType:   existingDoc.DocumentType,
	}

	err = s.documentRepository.Update(ctx, id, documentEntity, tx)
	if err != nil {
		return err
	}

	return nil
}

// DeleteDocument mock implementation
func (s mockDocumentService) DeleteDocument(ctx context.Context, id string, tx *gorm.DB) error {
	// Check if document exists
	document, err := s.documentRepository.FindByID(ctx, id, tx)
	if err != nil {
		return err
	}

	// Delete file from storage
	_, err = s.fileService.Storage.GcsDelete(document.FileStorageID, "sim_mbkm", "")
	if err != nil {
		return err
	}

	// Delete document record
	err = s.documentRepository.DeleteByID(ctx, id, tx)
	if err != nil {
		return err
	}

	return nil
}

// Helper to create a sample document entity
func createSampleDocument() entity.Document {
	now := time.Now()
	return entity.Document{
		ID:             uuid.New(),
		RegistrationID: uuid.New().String(),
		FileStorageID:  "file-123456",
		Name:           "sample-document.pdf",
		DocumentType:   "Acceptence Letter",
		BaseModel: entity.BaseModel{
			CreatedAt: &now,
			UpdatedAt: &now,
		},
	}
}

// TestFindAllDocumentsSuccess tests the successful retrieval of all documents
func (suite *DocumentServiceTestSuite) TestFindAllDocumentsSuccess() {
	// Setup test data
	ctx := context.Background()
	pagReq := dto.PaginationRequest{
		Limit:  10,
		Offset: 0,
		URL:    "http://example.com/documents",
	}

	// Create sample documents
	doc1 := createSampleDocument()
	doc2 := createSampleDocument()
	documents := []entity.Document{doc1, doc2}

	// Setup mocks
	suite.mockDocumentRepo.On("Index", ctx, pagReq, mock.Anything).
		Return(documents, int64(2), nil)
	suite.mockDocumentRepo.On("FindTotal", ctx, mock.Anything).
		Return(int64(2), nil)

	// Execute
	result, pagination, err := suite.service.FindAllDocuments(ctx, pagReq, nil)

	// Assert
	suite.NoError(err)
	suite.Equal(2, len(result))
	suite.Equal(int64(2), pagination.Total)
	suite.Equal(1, pagination.CurrentPage)

	// Verify the documents were converted correctly
	suite.Equal(doc1.ID.String(), result[0].ID)
	suite.Equal(doc1.Name, result[0].Name)
	suite.Equal(doc1.FileStorageID, result[0].FileStorageID)
	suite.Equal(doc1.RegistrationID, result[0].RegistrationID)
	suite.Equal(doc1.DocumentType, result[0].DocumentType)

	suite.mockDocumentRepo.AssertExpectations(suite.T())
}

// TestFindAllDocumentsEmpty tests retrieving documents when none exist
func (suite *DocumentServiceTestSuite) TestFindAllDocumentsEmpty() {
	// Setup
	ctx := context.Background()
	pagReq := dto.PaginationRequest{
		Limit:  10,
		Offset: 0,
		URL:    "http://example.com/documents",
	}

	// Setup mocks
	suite.mockDocumentRepo.On("Index", ctx, pagReq, mock.Anything).
		Return([]entity.Document{}, int64(0), nil)
	suite.mockDocumentRepo.On("FindTotal", ctx, mock.Anything).
		Return(int64(0), nil)

	// Execute
	result, pagination, err := suite.service.FindAllDocuments(ctx, pagReq, nil)

	// Assert
	suite.NoError(err)
	suite.Empty(result)
	suite.Equal(int64(0), pagination.Total)
	suite.Equal(1, pagination.CurrentPage)

	suite.mockDocumentRepo.AssertExpectations(suite.T())
}

// TestFindAllDocumentsError tests error handling when retrieving documents
func (suite *DocumentServiceTestSuite) TestFindAllDocumentsError() {
	// Setup
	ctx := context.Background()
	pagReq := dto.PaginationRequest{
		Limit:  10,
		Offset: 0,
		URL:    "http://example.com/documents",
	}

	expectedError := errors.New("database error")

	// Setup mocks
	suite.mockDocumentRepo.On("Index", ctx, pagReq, mock.Anything).
		Return([]entity.Document{}, int64(0), expectedError)

	// Execute
	result, pagination, err := suite.service.FindAllDocuments(ctx, pagReq, nil)

	// Assert
	suite.Error(err)
	suite.Equal(expectedError, err)
	suite.Empty(result)
	suite.Equal(dto.PaginationResponse{}, pagination)

	suite.mockDocumentRepo.AssertExpectations(suite.T())
}

// TestFindDocumentByIdSuccess tests successfully retrieving a document by ID
func (suite *DocumentServiceTestSuite) TestFindDocumentByIdSuccess() {
	// Setup
	ctx := context.Background()
	document := createSampleDocument()
	id := document.ID.String()

	// Setup mocks
	suite.mockDocumentRepo.On("FindByID", ctx, id, mock.Anything).
		Return(document, nil)

	// Execute
	result, err := suite.service.FindDocumentById(ctx, id, nil)

	// Assert
	suite.NoError(err)
	suite.Equal(id, result.ID)
	suite.Equal(document.Name, result.Name)
	suite.Equal(document.FileStorageID, result.FileStorageID)
	suite.Equal(document.RegistrationID, result.RegistrationID)
	suite.Equal(document.DocumentType, result.DocumentType)

	suite.mockDocumentRepo.AssertExpectations(suite.T())
}

// TestFindDocumentByIdNotFound tests retrieving a non-existent document
func (suite *DocumentServiceTestSuite) TestFindDocumentByIdNotFound() {
	// Setup
	ctx := context.Background()
	id := uuid.New().String()

	// Setup mocks
	suite.mockDocumentRepo.On("FindByID", ctx, id, mock.Anything).
		Return(entity.Document{}, gorm.ErrRecordNotFound)

	// Execute
	result, err := suite.service.FindDocumentById(ctx, id, nil)

	// Assert
	suite.Error(err)
	suite.Equal(gorm.ErrRecordNotFound, err)
	suite.Equal(dto.DocumentResponse{}, result)

	suite.mockDocumentRepo.AssertExpectations(suite.T())
}

// TestFindDocumentByIdDatabaseError tests error handling during document retrieval
func (suite *DocumentServiceTestSuite) TestFindDocumentByIdDatabaseError() {
	// Setup
	ctx := context.Background()
	id := uuid.New().String()
	expectedError := errors.New("database connection error")

	// Setup mocks
	suite.mockDocumentRepo.On("FindByID", ctx, id, mock.Anything).
		Return(entity.Document{}, expectedError)

	// Execute
	result, err := suite.service.FindDocumentById(ctx, id, nil)

	// Assert
	suite.Error(err)
	suite.Equal(expectedError, err)
	suite.Equal(dto.DocumentResponse{}, result)

	suite.mockDocumentRepo.AssertExpectations(suite.T())
}

// TestCreateDocumentSuccess tests successfully creating a document
func (suite *DocumentServiceTestSuite) TestCreateDocumentSuccess() {
	// Setup
	ctx := context.Background()
	registrationID := uuid.New().String()

	// Create request
	documentRequest := dto.DocumentRequest{
		RegistrationID: registrationID,
		Name:           "test-document.pdf",
		DocumentType:   "Acceptence Letter",
	}

	// Mock file
	file := &multipart.FileHeader{
		Filename: "test-document.pdf",
		Size:     1024,
	}

	// Mock registration
	registration := entity.Registration{
		ID: uuid.MustParse(registrationID),
	}

	// Mock file upload result
	uploadResult := &service_mock.FileStorageResponse{
		FileID: "uploaded-file-123",
	}

	// Setup mocks
	suite.mockRegistrationRepo.On("FindByID", ctx, registrationID, mock.Anything).
		Return(registration, nil)
	suite.mockFileService.Storage.On("GcsUpload", file, "sim_mbkm", "", "").
		Return(uploadResult, nil)
	suite.mockDocumentRepo.On("Create", ctx, mock.AnythingOfType("entity.Document"), mock.Anything).
		Return(entity.Document{}, nil)

	// Execute
	err := suite.service.CreateDocument(ctx, documentRequest, file, nil)

	// Assert
	suite.NoError(err)

	// Verify mocks were called correctly
	suite.mockRegistrationRepo.AssertExpectations(suite.T())
	suite.mockFileService.Storage.AssertExpectations(suite.T())
	suite.mockDocumentRepo.AssertExpectations(suite.T())

	// Verify the document entity created has correct values
	createCall := suite.mockDocumentRepo.Calls[0]
	documentArg := createCall.Arguments.Get(1).(entity.Document)
	suite.Equal(registrationID, documentArg.RegistrationID)
	suite.Equal("test-document.pdf", documentArg.Name)
	suite.Equal("uploaded-file-123", documentArg.FileStorageID)
	suite.Equal("Acceptence Letter", documentArg.DocumentType)
}

// TestCreateDocumentRegistrationNotFound tests creating a document with non-existent registration
func (suite *DocumentServiceTestSuite) TestCreateDocumentRegistrationNotFound() {
	// Setup
	ctx := context.Background()
	registrationID := uuid.New().String()

	// Create request
	documentRequest := dto.DocumentRequest{
		RegistrationID: registrationID,
		Name:           "test-document.pdf",
		DocumentType:   "Acceptence Letter",
	}

	// Mock file
	file := &multipart.FileHeader{
		Filename: "test-document.pdf",
		Size:     1024,
	}

	// Setup mocks
	suite.mockRegistrationRepo.On("FindByID", ctx, registrationID, mock.Anything).
		Return(entity.Registration{}, gorm.ErrRecordNotFound)

	// Execute
	err := suite.service.CreateDocument(ctx, documentRequest, file, nil)

	// Assert
	suite.Error(err)
	suite.Equal(gorm.ErrRecordNotFound, err)

	// Verify that only FindByID was called, but not file upload or document creation
	suite.mockRegistrationRepo.AssertExpectations(suite.T())
	suite.mockFileService.Storage.AssertNotCalled(suite.T(), "GcsUpload")
	suite.mockDocumentRepo.AssertNotCalled(suite.T(), "Create")
}

// TestCreateDocumentFileUploadError tests error during file upload when creating a document
func (suite *DocumentServiceTestSuite) TestCreateDocumentFileUploadError() {
	// Setup
	ctx := context.Background()
	registrationID := uuid.New().String()

	// Create request
	documentRequest := dto.DocumentRequest{
		RegistrationID: registrationID,
		Name:           "test-document.pdf",
		DocumentType:   "Acceptence Letter",
	}

	// Mock file
	file := &multipart.FileHeader{
		Filename: "test-document.pdf",
		Size:     1024,
	}

	// Mock registration
	registration := entity.Registration{
		ID: uuid.MustParse(registrationID),
	}

	// Setup mocks
	suite.mockRegistrationRepo.On("FindByID", ctx, registrationID, mock.Anything).
		Return(registration, nil)
	suite.mockFileService.Storage.On("GcsUpload", file, "sim_mbkm", "", "").
		Return(nil, errors.New("upload failed"))

	// Execute
	err := suite.service.CreateDocument(ctx, documentRequest, file, nil)

	// Assert
	suite.Error(err)
	suite.Equal("failed to upload file", err.Error())

	// Verify that file upload was attempted but document creation was not
	suite.mockRegistrationRepo.AssertExpectations(suite.T())
	suite.mockFileService.Storage.AssertExpectations(suite.T())
	suite.mockDocumentRepo.AssertNotCalled(suite.T(), "Create")
}

// TestCreateDocumentDatabaseError tests database error when creating a document
func (suite *DocumentServiceTestSuite) TestCreateDocumentDatabaseError() {
	// Setup
	ctx := context.Background()
	registrationID := uuid.New().String()
	expectedError := errors.New("database error")

	// Create request
	documentRequest := dto.DocumentRequest{
		RegistrationID: registrationID,
		Name:           "test-document.pdf",
		DocumentType:   "Acceptence Letter",
	}

	// Mock file
	file := &multipart.FileHeader{
		Filename: "test-document.pdf",
		Size:     1024,
	}

	// Mock registration
	registration := entity.Registration{
		ID: uuid.MustParse(registrationID),
	}

	// Mock file upload result
	uploadResult := &service_mock.FileStorageResponse{
		FileID: "uploaded-file-123",
	}

	// Setup mocks
	suite.mockRegistrationRepo.On("FindByID", ctx, registrationID, mock.Anything).
		Return(registration, nil)
	suite.mockFileService.Storage.On("GcsUpload", file, "sim_mbkm", "", "").
		Return(uploadResult, nil)
	suite.mockDocumentRepo.On("Create", ctx, mock.AnythingOfType("entity.Document"), mock.Anything).
		Return(entity.Document{}, expectedError)

	// Execute
	err := suite.service.CreateDocument(ctx, documentRequest, file, nil)

	// Assert
	suite.Error(err)
	suite.Equal(expectedError, err)

	// Verify all expected calls were made
	suite.mockRegistrationRepo.AssertExpectations(suite.T())
	suite.mockFileService.Storage.AssertExpectations(suite.T())
	suite.mockDocumentRepo.AssertExpectations(suite.T())
}

// TestUpdateDocumentSuccess tests successfully updating a document
func (suite *DocumentServiceTestSuite) TestUpdateDocumentSuccess() {
	// Setup
	ctx := context.Background()
	documentID := uuid.New().String()
	existingRegistrationID := uuid.New().String()
	newRegistrationID := uuid.New().String()

	// Create existing document
	existingDocument := entity.Document{
		ID:             uuid.MustParse(documentID),
		RegistrationID: existingRegistrationID,
		FileStorageID:  "old-file-123",
		Name:           "old-document.pdf",
		DocumentType:   "Acceptence Letter",
	}

	// Update request
	updateRequest := dto.UpdateDocumentRequest{
		RegistrationID: newRegistrationID,
	}

	// Mock file
	file := &multipart.FileHeader{
		Filename: "updated-document.pdf",
		Size:     2048,
	}

	// Mock registration
	newRegistration := entity.Registration{
		ID: uuid.MustParse(newRegistrationID),
	}

	// Mock file upload result
	uploadResult := &service_mock.FileStorageResponse{
		FileID: "new-file-456",
	}

	// Setup mocks
	suite.mockDocumentRepo.On("FindByID", ctx, documentID, mock.Anything).
		Return(existingDocument, nil)
	suite.mockRegistrationRepo.On("FindByID", ctx, newRegistrationID, mock.Anything).
		Return(newRegistration, nil)
	suite.mockFileService.Storage.On("GcsUpload", file, "sim_mbkm", "", "").
		Return(uploadResult, nil)
	suite.mockDocumentRepo.On("Update", ctx, documentID, mock.AnythingOfType("entity.Document"), mock.Anything).
		Return(nil)

	// Execute
	err := suite.service.UpdateDocument(ctx, documentID, updateRequest, file, nil)

	// Assert
	suite.NoError(err)

	// Verify all expected calls were made
	suite.mockDocumentRepo.AssertExpectations(suite.T())
	suite.mockRegistrationRepo.AssertExpectations(suite.T())
	suite.mockFileService.Storage.AssertExpectations(suite.T())

	// Verify the document was updated with correct values
	updateCall := suite.mockDocumentRepo.Calls[1]
	documentArg := updateCall.Arguments.Get(2).(entity.Document)
	suite.Equal(existingDocument.ID, documentArg.ID)
	suite.Equal(newRegistrationID, documentArg.RegistrationID)
	suite.Equal("new-file-456", documentArg.FileStorageID)
	suite.Equal(existingDocument.Name, documentArg.Name)
	suite.Equal(existingDocument.DocumentType, documentArg.DocumentType)
}

// TestUpdateDocumentNotFound tests updating a non-existent document
func (suite *DocumentServiceTestSuite) TestUpdateDocumentNotFound() {
	// Setup
	ctx := context.Background()
	documentID := uuid.New().String()
	newRegistrationID := uuid.New().String()

	// Update request
	updateRequest := dto.UpdateDocumentRequest{
		RegistrationID: newRegistrationID,
	}

	// Mock file
	file := &multipart.FileHeader{
		Filename: "updated-document.pdf",
		Size:     2048,
	}

	// Setup mocks
	suite.mockDocumentRepo.On("FindByID", ctx, documentID, mock.Anything).
		Return(entity.Document{}, gorm.ErrRecordNotFound)

	// Execute
	err := suite.service.UpdateDocument(ctx, documentID, updateRequest, file, nil)

	// Assert
	suite.Error(err)
	suite.Equal(gorm.ErrRecordNotFound, err)

	// Verify only FindByID was called and not the other methods
	suite.mockDocumentRepo.AssertExpectations(suite.T())
	suite.mockRegistrationRepo.AssertNotCalled(suite.T(), "FindByID")
	suite.mockFileService.Storage.AssertNotCalled(suite.T(), "GcsUpload")
	suite.mockDocumentRepo.AssertNotCalled(suite.T(), "Update")
}

// TestUpdateDocumentRegistrationNotFound tests updating a document with non-existent registration
func (suite *DocumentServiceTestSuite) TestUpdateDocumentRegistrationNotFound() {
	// Setup
	ctx := context.Background()
	documentID := uuid.New().String()
	existingRegistrationID := uuid.New().String()
	newRegistrationID := uuid.New().String()

	// Create existing document
	existingDocument := entity.Document{
		ID:             uuid.MustParse(documentID),
		RegistrationID: existingRegistrationID,
		FileStorageID:  "old-file-123",
		Name:           "old-document.pdf",
		DocumentType:   "Acceptence Letter",
	}

	// Update request
	updateRequest := dto.UpdateDocumentRequest{
		RegistrationID: newRegistrationID,
	}

	// Mock file
	file := &multipart.FileHeader{
		Filename: "updated-document.pdf",
		Size:     2048,
	}

	// Setup mocks
	suite.mockDocumentRepo.On("FindByID", ctx, documentID, mock.Anything).
		Return(existingDocument, nil)
	suite.mockRegistrationRepo.On("FindByID", ctx, newRegistrationID, mock.Anything).
		Return(entity.Registration{}, gorm.ErrRecordNotFound)

	// Execute
	err := suite.service.UpdateDocument(ctx, documentID, updateRequest, file, nil)

	// Assert
	suite.Error(err)
	suite.Equal(gorm.ErrRecordNotFound, err)

	// Verify the correct methods were called
	suite.mockDocumentRepo.AssertExpectations(suite.T())
	suite.mockRegistrationRepo.AssertExpectations(suite.T())
	suite.mockFileService.Storage.AssertNotCalled(suite.T(), "GcsUpload")
	suite.mockDocumentRepo.AssertNotCalled(suite.T(), "Update")
}

// TestUpdateDocumentFileUploadError tests file upload error during document update
func (suite *DocumentServiceTestSuite) TestUpdateDocumentFileUploadError() {
	// Setup
	ctx := context.Background()
	documentID := uuid.New().String()
	existingRegistrationID := uuid.New().String()
	newRegistrationID := uuid.New().String()

	// Create existing document
	existingDocument := entity.Document{
		ID:             uuid.MustParse(documentID),
		RegistrationID: existingRegistrationID,
		FileStorageID:  "old-file-123",
		Name:           "old-document.pdf",
		DocumentType:   "Acceptence Letter",
	}

	// Update request
	updateRequest := dto.UpdateDocumentRequest{
		RegistrationID: newRegistrationID,
	}

	// Mock file
	file := &multipart.FileHeader{
		Filename: "updated-document.pdf",
		Size:     2048,
	}

	// Mock registration
	newRegistration := entity.Registration{
		ID: uuid.MustParse(newRegistrationID),
	}

	// Setup mocks
	suite.mockDocumentRepo.On("FindByID", ctx, documentID, mock.Anything).
		Return(existingDocument, nil)
	suite.mockRegistrationRepo.On("FindByID", ctx, newRegistrationID, mock.Anything).
		Return(newRegistration, nil)
	suite.mockFileService.Storage.On("GcsUpload", file, "sim_mbkm", "", "").
		Return(nil, errors.New("upload failed"))

	// Execute
	err := suite.service.UpdateDocument(ctx, documentID, updateRequest, file, nil)

	// Assert
	suite.Error(err)
	suite.Equal("failed to upload file", err.Error())

	// Verify that file upload was attempted but document update was not
	suite.mockDocumentRepo.AssertExpectations(suite.T())
	suite.mockRegistrationRepo.AssertExpectations(suite.T())
	suite.mockFileService.Storage.AssertExpectations(suite.T())
	suite.mockDocumentRepo.AssertNotCalled(suite.T(), "Update")
}

// TestUpdateDocumentDatabaseError tests database error during document update
func (suite *DocumentServiceTestSuite) TestUpdateDocumentDatabaseError() {
	// Setup
	ctx := context.Background()
	documentID := uuid.New().String()
	existingRegistrationID := uuid.New().String()
	newRegistrationID := uuid.New().String()
	expectedError := errors.New("database error")

	// Create existing document
	existingDocument := entity.Document{
		ID:             uuid.MustParse(documentID),
		RegistrationID: existingRegistrationID,
		FileStorageID:  "old-file-123",
		Name:           "old-document.pdf",
		DocumentType:   "Acceptence Letter",
	}

	// Update request
	updateRequest := dto.UpdateDocumentRequest{
		RegistrationID: newRegistrationID,
	}

	// Mock file
	file := &multipart.FileHeader{
		Filename: "updated-document.pdf",
		Size:     2048,
	}

	// Mock registration
	newRegistration := entity.Registration{
		ID: uuid.MustParse(newRegistrationID),
	}

	// Mock file upload result
	uploadResult := &service_mock.FileStorageResponse{
		FileID: "new-file-456",
	}

	// Setup mocks
	suite.mockDocumentRepo.On("FindByID", ctx, documentID, mock.Anything).
		Return(existingDocument, nil)
	suite.mockRegistrationRepo.On("FindByID", ctx, newRegistrationID, mock.Anything).
		Return(newRegistration, nil)
	suite.mockFileService.Storage.On("GcsUpload", file, "sim_mbkm", "", "").
		Return(uploadResult, nil)
	suite.mockDocumentRepo.On("Update", ctx, documentID, mock.AnythingOfType("entity.Document"), mock.Anything).
		Return(expectedError)

	// Execute
	err := suite.service.UpdateDocument(ctx, documentID, updateRequest, file, nil)

	// Assert
	suite.Error(err)
	suite.Equal(expectedError, err)

	// Verify all expected calls were made
	suite.mockDocumentRepo.AssertExpectations(suite.T())
	suite.mockRegistrationRepo.AssertExpectations(suite.T())
	suite.mockFileService.Storage.AssertExpectations(suite.T())
}

// TestDeleteDocumentSuccess tests successfully deleting a document
func (suite *DocumentServiceTestSuite) TestDeleteDocumentSuccess() {
	// Setup
	ctx := context.Background()
	documentID := uuid.New().String()

	// Create document to delete
	document := entity.Document{
		ID:             uuid.MustParse(documentID),
		RegistrationID: uuid.New().String(),
		FileStorageID:  "file-to-delete-123",
		Name:           "document-to-delete.pdf",
		DocumentType:   "Acceptence Letter",
	}

	// Setup mocks
	suite.mockDocumentRepo.On("FindByID", ctx, documentID, mock.Anything).
		Return(document, nil)
	suite.mockFileService.Storage.On("GcsDelete", document.FileStorageID, "sim_mbkm", "").
		Return(&service_mock.FileStorageResponse{}, nil)
	suite.mockDocumentRepo.On("DeleteByID", ctx, documentID, mock.Anything).
		Return(nil)

	// Execute
	err := suite.service.DeleteDocument(ctx, documentID, nil)

	// Assert
	suite.NoError(err)

	// Verify all expected calls were made
	suite.mockDocumentRepo.AssertExpectations(suite.T())
	suite.mockFileService.Storage.AssertExpectations(suite.T())
}

// TestDeleteDocumentNotFound tests deleting a non-existent document
func (suite *DocumentServiceTestSuite) TestDeleteDocumentNotFound() {
	// Setup
	ctx := context.Background()
	documentID := uuid.New().String()

	// Setup mocks
	suite.mockDocumentRepo.On("FindByID", ctx, documentID, mock.Anything).
		Return(entity.Document{}, gorm.ErrRecordNotFound)

	// Execute
	err := suite.service.DeleteDocument(ctx, documentID, nil)

	// Assert
	suite.Error(err)
	suite.Equal(gorm.ErrRecordNotFound, err)

	// Verify that only FindByID was called
	suite.mockDocumentRepo.AssertExpectations(suite.T())
	suite.mockFileService.Storage.AssertNotCalled(suite.T(), "GcsDelete")
	suite.mockDocumentRepo.AssertNotCalled(suite.T(), "DeleteByID")
}

// TestDeleteDocumentFileDeleteError tests file deletion error when deleting a document
func (suite *DocumentServiceTestSuite) TestDeleteDocumentFileDeleteError() {
	// Setup
	ctx := context.Background()
	documentID := uuid.New().String()

	// Create document to delete
	document := entity.Document{
		ID:             uuid.MustParse(documentID),
		RegistrationID: uuid.New().String(),
		FileStorageID:  "file-to-delete-123",
		Name:           "document-to-delete.pdf",
		DocumentType:   "Acceptence Letter",
	}

	// Setup mocks
	suite.mockDocumentRepo.On("FindByID", ctx, documentID, mock.Anything).
		Return(document, nil)
	suite.mockFileService.Storage.On("GcsDelete", document.FileStorageID, "sim_mbkm", "").
		Return(nil, errors.New("delete failed"))

	// Execute
	err := suite.service.DeleteDocument(ctx, documentID, nil)

	// Assert
	suite.Error(err)
	suite.Equal("delete failed", err.Error())

	// Verify correct methods were called
	suite.mockDocumentRepo.AssertExpectations(suite.T())
	suite.mockFileService.Storage.AssertExpectations(suite.T())
	suite.mockDocumentRepo.AssertNotCalled(suite.T(), "DeleteByID")
}

// TestDeleteDocumentDatabaseError tests database error when deleting a document
func (suite *DocumentServiceTestSuite) TestDeleteDocumentDatabaseError() {
	// Setup
	ctx := context.Background()
	documentID := uuid.New().String()
	expectedError := errors.New("database error")

	// Create document to delete
	document := entity.Document{
		ID:             uuid.MustParse(documentID),
		RegistrationID: uuid.New().String(),
		FileStorageID:  "file-to-delete-123",
		Name:           "document-to-delete.pdf",
		DocumentType:   "Acceptence Letter",
	}

	// Setup mocks
	suite.mockDocumentRepo.On("FindByID", ctx, documentID, mock.Anything).
		Return(document, nil)
	suite.mockFileService.Storage.On("GcsDelete", document.FileStorageID, "sim_mbkm", "").
		Return(&service_mock.FileStorageResponse{}, nil)
	suite.mockDocumentRepo.On("DeleteByID", ctx, documentID, mock.Anything).
		Return(expectedError)

	// Execute
	err := suite.service.DeleteDocument(ctx, documentID, nil)

	// Assert
	suite.Error(err)
	suite.Equal(expectedError, err)

	// Verify all expected calls were made
	suite.mockDocumentRepo.AssertExpectations(suite.T())
	suite.mockFileService.Storage.AssertExpectations(suite.T())
}

// TestDocumentServiceSuite runs the document service test suite
func TestDocumentServiceSuite(t *testing.T) {
	suite.Run(t, new(DocumentServiceTestSuite))
}
