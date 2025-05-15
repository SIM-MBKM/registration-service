package service_mock

import (
	"mime/multipart"

	"github.com/stretchr/testify/mock"
)

type FileStorageResponse struct {
	FileID     string
	ObjectName string
	Message    string
}

type MockStorageInterface struct {
	mock.Mock
}

func (m *MockStorageInterface) GcsUpload(file *multipart.FileHeader, projectID, bucketName, objectName string) (*FileStorageResponse, error) {
	args := m.Called(file, projectID, bucketName, objectName)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*FileStorageResponse), args.Error(1)
}

func (m *MockStorageInterface) GcsDelete(fileID, projectID, bucketName string) (*FileStorageResponse, error) {
	args := m.Called(fileID, projectID, bucketName)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*FileStorageResponse), args.Error(1)
}

type MockFileService struct {
	mock.Mock
	storage *MockStorageInterface
}

func NewMockFileService() *MockFileService {
	return &MockFileService{
		storage: &MockStorageInterface{},
	}
}
