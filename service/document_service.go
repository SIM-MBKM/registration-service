package service

import (
	"context"
	"errors"
	"mime/multipart"
	"reflect"
	"registration-service/dto"
	"registration-service/entity"
	"registration-service/helper"
	"registration-service/repository"

	storageService "github.com/SIM-MBKM/filestorage/storage"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type DocumentUpdate struct {
	RegistrationID string
	FileStorageID  string
}

type documentService struct {
	documentRepository     repository.DocumentRepository
	registrationRepository repository.RegistrationRepository
	fileService            *FileService
}

type DocumentService interface {
	FindAllDocuments(ctx context.Context, pagReq dto.PaginationRequest, tx *gorm.DB) ([]dto.DocumentResponse, dto.PaginationResponse, error)
	FindDocumentById(ctx context.Context, id string, tx *gorm.DB) (dto.DocumentResponse, error)
	CreateDocument(ctx context.Context, document dto.DocumentRequest, file *multipart.FileHeader, tx *gorm.DB) error
	UpdateDocument(ctx context.Context, id string, document dto.UpdateDocumentRequest, file *multipart.FileHeader, tx *gorm.DB) error
	DeleteDocument(ctx context.Context, id string, tx *gorm.DB) error
}

func NewDocumentService(documentRepository repository.DocumentRepository, registrationRepository repository.RegistrationRepository, config *storageService.Config, tokenManager *storageService.CacheTokenManager) DocumentService {
	return &documentService{
		documentRepository:     documentRepository,
		registrationRepository: registrationRepository,
		fileService:            NewFileService(config, tokenManager),
	}
}

func (s *documentService) FindAllDocuments(ctx context.Context, pagReq dto.PaginationRequest, tx *gorm.DB) ([]dto.DocumentResponse, dto.PaginationResponse, error) {
	documents, total_data, err := s.documentRepository.Index(ctx, pagReq, tx)

	metaData := helper.MetaDataPagination(total_data, pagReq)

	if err != nil {
		return nil, dto.PaginationResponse{}, err
	}

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

func (s *documentService) FindDocumentById(ctx context.Context, id string, tx *gorm.DB) (dto.DocumentResponse, error) {
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

func (s *documentService) CreateDocument(ctx context.Context, document dto.DocumentRequest, file *multipart.FileHeader, tx *gorm.DB) error {
	// upload file
	result, err := s.fileService.storage.GcsUpload(file, "sim_mbkm", "", "")
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

func (s *documentService) UpdateDocument(ctx context.Context, id string, document dto.UpdateDocumentRequest, file *multipart.FileHeader, tx *gorm.DB) error {
	res, err := s.documentRepository.FindByID(ctx, id, tx)
	if err != nil {
		return err
	}

	registration, err := s.registrationRepository.FindByID(ctx, string(document.RegistrationID), tx)
	if err != nil {
		return err
	}

	result, err := s.fileService.storage.GcsUpload(file, "sim_mbkm", "", "")
	if err != nil {
		return errors.New("failed to upload file")
	}

	documentUpdate := DocumentUpdate{
		RegistrationID: registration.ID.String(),
		FileStorageID:  result.FileID,
	}

	// Create documentEntity with original ID
	documentEntity := entity.Document{
		ID: res.ID,
	}

	// Get reflection values
	resValue := reflect.ValueOf(res)
	reqValue := reflect.ValueOf(documentUpdate)
	entityValue := reflect.ValueOf(&documentEntity).Elem()

	// Iterate through fields of the request type
	for i := 0; i < reqValue.Type().NumField(); i++ {
		reqField := reqValue.Type().Field(i)
		reqFieldValue := reqValue.Field(i)

		// Find corresponding field in the entity
		entityField := entityValue.FieldByName(reqField.Name)

		// Find corresponding field in the original result
		resField := resValue.FieldByName(reqField.Name)

		// Check if the field exists and can be set
		if entityField.IsValid() && entityField.CanSet() {
			// If the request field is not zero, use its value
			if !reflect.DeepEqual(reqFieldValue.Interface(), reflect.Zero(reqFieldValue.Type()).Interface()) {
				entityField.Set(reqFieldValue)
			} else {
				// Otherwise, use the original value
				entityField.Set(resField)
			}
		}
	}

	// Perform the update
	err = s.documentRepository.Update(ctx, id, documentEntity, tx)
	if err != nil {
		return err
	}

	return nil
}

func (s *documentService) DeleteDocument(ctx context.Context, id string, tx *gorm.DB) error {
	res, err := s.documentRepository.FindByID(ctx, id, tx)
	if err != nil {
		return err
	}

	_, err = s.fileService.storage.GcsDelete(res.FileStorageID, "sim_mbkm", "")
	if err != nil {
		return err
	}

	err = s.documentRepository.DeleteByID(ctx, id, tx)
	if err != nil {
		return err
	}

	return nil
}
