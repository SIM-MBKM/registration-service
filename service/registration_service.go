package service

import (
	"context"
	"errors"
	"log"
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

type registrationService struct {
	registrationRepository    repository.RegistrationRepository
	documentRepository        repository.DocumentRepository
	userManagementService     *UserManagementService
	activityManagementService *ActivityManagementService
	fileService               *FileService
}

type RegistrationService interface {
	FindAllRegistrations(ctx context.Context, pagReq dto.PaginationRequest, filter dto.FilterRegistrationRequest, tx *gorm.DB, token string) ([]dto.GetRegistrationResponse, dto.PaginationResponse, error)
	FindRegistrationByID(ctx context.Context, id string, tx *gorm.DB) (dto.GetRegistrationResponse, error)
	CreateRegistration(ctx context.Context, registration dto.CreateRegistrationRequest, file *multipart.FileHeader, geoletter *multipart.FileHeader, tx *gorm.DB, token string) error
	UpdateRegistration(ctx context.Context, id string, registration dto.UpdateRegistrationDataRequest, tx *gorm.DB) error
	DeleteRegistration(ctx context.Context, id string, tx *gorm.DB) error
	GetActivitiesData(data map[string]interface{}, method string, endpoint string, token string) []map[string]interface{}
	GetUserByFilter(data map[string]interface{}, method string, endpoint string, token string) []map[string]interface{}
	GetUserData(method string, endpoint string, token string) map[string]interface{}
}

func NewRegistrationService(registrationRepository repository.RegistrationRepository, documentRepository repository.DocumentRepository, secretKey string, userManagementbaseURI string, activityManagementbaseURI string, asyncURIs []string, config *storageService.Config, tokenManager *storageService.CacheTokenManager) RegistrationService {
	return &registrationService{
		registrationRepository:    registrationRepository,
		documentRepository:        documentRepository,
		userManagementService:     NewUserManagementService(userManagementbaseURI, asyncURIs),
		activityManagementService: NewActivityManagementService(activityManagementbaseURI, asyncURIs),
		fileService:               NewFileService(config, tokenManager),
	}
}

func (s *registrationService) GetActivitiesData(data map[string]interface{}, method string, endpoint string, token string) []map[string]interface{} {
	res, err := s.activityManagementService.baseService.Request(method, endpoint, data, token)
	log.Println("DATA", res)
	if err != nil {
		return nil
	}

	// First, get the data as []interface{}
	activitiesInterface, ok := res["data"].([]interface{})
	if !ok {
		log.Println("Failed to convert data to []interface{}")
		return nil
	}

	// Then, convert each item to map[string]interface{}
	var activitiesData []map[string]interface{}
	for _, activityInterface := range activitiesInterface {
		activity, ok := activityInterface.(map[string]interface{})
		if !ok {
			log.Println("Failed to convert activity to map[string]interface{}")
			continue
		}

		activitiesData = append(activitiesData, map[string]interface{}{
			"id":   activity["id"],
			"name": activity["name"],
		})
	}

	return activitiesData
}

func (s *registrationService) GetUserByFilter(data map[string]interface{}, method string, endpoint string, token string) []map[string]interface{} {
	res, err := s.userManagementService.baseService.Request(method, endpoint, data, token)
	if err != nil {
		return nil
	}

	// First, get the data as []interface{}
	usersInterface, ok := res["data"].([]interface{})
	if !ok {
		log.Println("Failed to convert data to []interface{}")
		return nil
	}

	var usersData []map[string]interface{}
	for _, userInterface := range usersInterface {
		user, ok := userInterface.(map[string]interface{})
		if !ok {
			log.Println("Failed to convert user to map[string]interface{}")
			continue
		}

		usersData = append(usersData, map[string]interface{}{
			"id":   user["id"],
			"nrp":  user["nrp"],
			"name": user["name"],
		})
	}
	return usersData
}
func (s *registrationService) GetUserData(method string, endpoint string, token string) map[string]interface{} {
	res, err := s.userManagementService.baseService.Request(method, endpoint, nil, token)
	if err != nil {
		return nil
	}

	users, ok := res["data"].(map[string]interface{})
	if !ok {
		return nil
	}

	var usersData map[string]interface{}
	usersData = map[string]interface{}{
		"id":   users["id"],
		"nrp":  users["nrp"],
		"name": users["name"],
	}
	return usersData
}

func (s *registrationService) FindAllRegistrations(ctx context.Context, pagReq dto.PaginationRequest, filter dto.FilterRegistrationRequest, tx *gorm.DB, token string) ([]dto.GetRegistrationResponse, dto.PaginationResponse, error) {
	registrations, total, err := s.registrationRepository.Index(ctx, tx, pagReq, filter)
	if err != nil {
		return nil, dto.PaginationResponse{}, err
	}

	metaData := helper.MetaDataPagination(total, pagReq)

	var response []dto.GetRegistrationResponse
	for _, registration := range registrations {
		response = append(response, dto.GetRegistrationResponse{
			ID:                        registration.ID.String(),
			ActivityID:                registration.ActivityID,
			ActivityName:              registration.ActivityName,
			UserID:                    registration.UserID,
			UserNRP:                   registration.UserNRP,
			AdvisingConfirmation:      registration.AdvisingConfirmation,
			AcademicAdvisor:           registration.AcademicAdvisor,
			AcademicAdvisorEmail:      registration.AcademicAdvisorEmail,
			MentorName:                registration.MentorName,
			MentorEmail:               registration.MentorEmail,
			LOValidation:              registration.LOValidation,
			AcademicAdvisorValidation: registration.AcademicAdvisorValidation,
			Semester:                  registration.Semester,
			TotalSKS:                  registration.TotalSKS,
			Documents:                 convertToDocumentResponse(registration.Document),
		})
	}

	return response, metaData, nil
}

func (s *registrationService) FindRegistrationByID(ctx context.Context, id string, tx *gorm.DB) (dto.GetRegistrationResponse, error) {
	registration, err := s.registrationRepository.FindByID(ctx, id, tx)
	if err != nil {
		return dto.GetRegistrationResponse{}, err
	}

	response := dto.GetRegistrationResponse{
		ID:                        registration.ID.String(),
		ActivityID:                registration.ActivityID,
		ActivityName:              registration.ActivityName,
		UserID:                    registration.UserID,
		UserNRP:                   registration.UserNRP,
		UserName:                  registration.UserName,
		AdvisingConfirmation:      registration.AdvisingConfirmation,
		AcademicAdvisor:           registration.AcademicAdvisor,
		AcademicAdvisorEmail:      registration.AcademicAdvisorEmail,
		MentorName:                registration.MentorName,
		MentorEmail:               registration.MentorEmail,
		LOValidation:              registration.LOValidation,
		AcademicAdvisorValidation: registration.AcademicAdvisorValidation,
		Semester:                  registration.Semester,
		TotalSKS:                  registration.TotalSKS,
		Documents:                 convertToDocumentResponse(registration.Document),
	}

	return response, nil
}

func (s *registrationService) CreateRegistration(ctx context.Context, registration dto.CreateRegistrationRequest, file *multipart.FileHeader, geoletter *multipart.FileHeader, tx *gorm.DB, token string) error {
	var registrationEntity entity.Registration
	var activitiesData []map[string]interface{}
	var usersData []map[string]interface{}

	// get user data
	userData := s.GetUserData("GET", "user-management-service/api/v1/user/role", token)
	if userData == nil {
		return errors.New("user data not found")
	}

	if registration.ActivityID != "" {
		activitiesData = s.GetActivitiesData(map[string]interface{}{
			"activity_id":     registration.ActivityID,
			"program_type_id": "",
			"level_id":        "",
			"group_id":        "",
			"name":            "",
		}, "POST", "activity-management/api/activity/filter", token)
	}

	if userData["id"] != "" {
		usersData = s.GetUserByFilter(map[string]interface{}{
			"user_id": userData["id"],
		}, "POST", "user-management-service/api/v1/user/filter", token)
	}

	if len(activitiesData) == 0 || len(usersData) == 0 {
		return errors.New("data not found")
	}
	// Then safely extract values with type assertions and nil checks
	var activityName string
	if name, ok := activitiesData[0]["name"]; ok && name != nil {
		activityName, ok = name.(string)
		if !ok {
			activityName = "" // Default value if type assertion fails
		}
	} else {
		activityName = "" // Default value if key doesn't exist or is nil
	}

	var userID string
	if id, ok := userData["id"]; ok && id != nil {
		userID, ok = id.(string)
		if !ok {
			userID = "" // Default value if type assertion fails
		}
	} else {
		userID = "" // Default value if key doesn't exist or is nil
	}

	var userNRP string
	if nrp, ok := usersData[0]["nrp"]; ok && nrp != nil {
		userNRP, ok = nrp.(string)
		if !ok {
			userNRP = "" // Default value if type assertion fails
		}
	} else {
		userNRP = "" // Default value if key doesn't exist or is nil
	}

	var userName string
	if name, ok := usersData[0]["name"]; ok && name != nil {
		userName, ok = name.(string)
		if !ok {
			userName = "" // Default value if type assertion fails
		}
	} else {
		userName = "" // Default value if key doesn't exist or is nil
	}

	// upload file
	result, err := s.fileService.storage.GcsUpload(file, "sim_mbkm", "", "")
	if err != nil {
		return errors.New("failed to upload file")
	}

	geoletterResult, err := s.fileService.storage.GcsUpload(geoletter, "sim_mbkm", "", "")
	if err != nil {
		return errors.New("failed to upload file")
	}

	loValidation := "PENDING"
	academicAdvisorValidation := "PENDING"

	registrationEntity = entity.Registration{
		ID:                        uuid.New(),
		ActivityID:                registration.ActivityID,
		ActivityName:              activityName,
		UserID:                    userID,
		UserNRP:                   userNRP,
		UserName:                  userName,
		AdvisingConfirmation:      registration.AdvisingConfirmation,
		AcademicAdvisorID:         registration.AcademicAdvisorID,
		AcademicAdvisor:           registration.AcademicAdvisor,
		AcademicAdvisorEmail:      registration.AcademicAdvisorEmail,
		MentorName:                registration.MentorName,
		MentorEmail:               registration.MentorEmail,
		LOValidation:              loValidation,
		AcademicAdvisorValidation: academicAdvisorValidation,
		Semester:                  registration.Semester,
		TotalSKS:                  registration.TotalSKS,
	}

	_, err = s.registrationRepository.Create(ctx, registrationEntity, tx)
	if err != nil {
		return err
	}

	// Create document entity
	documentEntity := entity.Document{
		ID:             uuid.New(),
		RegistrationID: registrationEntity.ID.String(),
		Name:           file.Filename,
		FileStorageID:  result.FileID,
		DocumentType:   "Acceptence Letter",
	}

	_, err = s.documentRepository.Create(ctx, documentEntity, tx)
	if err != nil {
		return err
	}

	geoletterEntity := entity.Document{
		ID:             uuid.New(),
		RegistrationID: registrationEntity.ID.String(),
		Name:           geoletter.Filename,
		FileStorageID:  geoletterResult.FileID,
		DocumentType:   "Geoletter",
	}

	_, err = s.documentRepository.Create(ctx, geoletterEntity, tx)
	if err != nil {
		return err
	}

	return nil
}

func (s *registrationService) UpdateRegistration(ctx context.Context, id string, registration dto.UpdateRegistrationDataRequest, tx *gorm.DB) error {
	// Find existing program type
	res, err := s.registrationRepository.FindByID(ctx, id, tx)
	if err != nil {
		return err
	}

	// Create programTypeEntity with original ID
	registrationEntity := entity.Registration{
		ID: res.ID,
	}

	// Get reflection values
	resValue := reflect.ValueOf(res)
	reqValue := reflect.ValueOf(registration)
	entityValue := reflect.ValueOf(&registrationEntity).Elem()

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
	err = s.registrationRepository.Update(ctx, id, registrationEntity, tx)
	if err != nil {
		return err
	}

	return nil
}

func (s *registrationService) DeleteRegistration(ctx context.Context, id string, tx *gorm.DB) error {
	err := s.registrationRepository.Destroy(ctx, id, tx)
	if err != nil {
		return err
	}

	return nil
}

func convertToDocumentResponse(documents []entity.Document) []dto.DocumentResponse {
	var documentResponses []dto.DocumentResponse
	for _, document := range documents {
		documentResponses = append(documentResponses, dto.DocumentResponse{
			ID:             document.ID.String(),
			RegistrationID: document.RegistrationID,
			Name:           document.Name,
			FileStorageID:  document.FileStorageID,
			DocumentType:   document.DocumentType,
		})
	}
	return documentResponses
}
