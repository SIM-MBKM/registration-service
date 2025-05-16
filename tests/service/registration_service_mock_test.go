package service_test

import (
	"context"
	"errors"
	"log"
	"mime/multipart"
	"reflect"
	"registration-service/dto"
	"registration-service/entity"
	"registration-service/helper"
	repository_mock "registration-service/mocks/repository"
	service_mock "registration-service/mocks/service"
	"registration-service/service"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"
)

type RegistrationServiceTestSuite struct {
	suite.Suite
	mockRegistrationRepo            *repository_mock.MockRegistrationRepository
	mockDocumentRepo                *repository_mock.MockDocumentRepository
	mockUserManagementService       *service_mock.MockUserManagementService
	mockActivityManagementService   *service_mock.MockActivityManagementService
	mockFileService                 *service_mock.MockFileService
	mockMatchingManagementService   *service_mock.MockMatchingManagementService
	mockMonitoringManagementService *service_mock.MockMonitoringManagementService
	service                         service.RegistrationService
}

func (suite *RegistrationServiceTestSuite) SetupTest() {
	// Create mocks for all dependencies
	suite.mockRegistrationRepo = new(repository_mock.MockRegistrationRepository)
	suite.mockDocumentRepo = new(repository_mock.MockDocumentRepository)
	suite.mockUserManagementService = new(service_mock.MockUserManagementService)
	suite.mockActivityManagementService = new(service_mock.MockActivityManagementService)
	suite.mockMatchingManagementService = new(service_mock.MockMatchingManagementService)
	suite.mockMonitoringManagementService = new(service_mock.MockMonitoringManagementService)

	// Create mock file service with initialized Storage
	suite.mockFileService = service_mock.NewMockFileService()

	// Create the registration service with the mocks
	mockService := &mockRegistrationService{
		registrationRepository:      suite.mockRegistrationRepo,
		documentRepository:          suite.mockDocumentRepo,
		userManagementService:       suite.mockUserManagementService,
		activityManagementService:   suite.mockActivityManagementService,
		fileService:                 suite.mockFileService,
		matchingManagementService:   suite.mockMatchingManagementService,
		monitoringManagementService: suite.mockMonitoringManagementService,
	}

	suite.service = mockService
}

type mockRegistrationService struct {
	registrationRepository      *repository_mock.MockRegistrationRepository
	documentRepository          *repository_mock.MockDocumentRepository
	userManagementService       *service_mock.MockUserManagementService
	activityManagementService   *service_mock.MockActivityManagementService
	fileService                 *service_mock.MockFileService
	matchingManagementService   *service_mock.MockMatchingManagementService
	monitoringManagementService *service_mock.MockMonitoringManagementService
}

func (s *mockRegistrationService) FindAllRegistrations(ctx context.Context, pagReq dto.PaginationRequest, filter dto.FilterRegistrationRequest, tx *gorm.DB, token string) ([]dto.GetRegistrationResponse, dto.PaginationResponse, error) {
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
			ApprovalStatus:            registration.ApprovalStatus,
			Documents:                 convertToDocumentResponse(registration.Document),
		})
	}

	return response, metaData, nil
}

func (s *mockRegistrationService) RegistrationsDataAccess(ctx context.Context, id string, token string, tx *gorm.DB) bool {
	var state bool
	state = false
	registration, err := s.registrationRepository.FindByID(ctx, id, tx)
	if err != nil {
		return false
	}

	userData := s.userManagementService.GetUserData("GET", token)
	if userData == nil {
		return false
	}

	var userID string
	if id, ok := userData["id"]; ok && id != nil {
		userID, ok = id.(string)
		if !ok {
			return false
		}
	} else {
		return false
	}

	var userRole string
	if role, ok := userData["role"]; ok && role != nil {
		userRole, ok = role.(string)
		if !ok {
			return false
		}
	} else {
		return false
	}

	var userEmail string
	if email, ok := userData["email"]; ok && email != nil {
		userEmail, ok = email.(string)
		if !ok {
			return false
		}
	} else {
		return false
	}

	if userRole == "MAHASISWA" {
		if registration.UserID == userID {
			state = true
		}
	} else if userRole == "DOSEN PEMBIMBING" {
		if registration.AcademicAdvisorEmail == userEmail {
			state = true
		}
	} else if userRole == "ADMIN" || userRole == "LO-MBKM" {
		state = true
	}

	return state
}

func (s *mockRegistrationService) FindRegistrationByID(ctx context.Context, id string, token string, tx *gorm.DB) (dto.GetRegistrationResponse, error) {
	access := s.RegistrationsDataAccess(ctx, id, token, tx)
	if !access {
		return dto.GetRegistrationResponse{}, errors.New("data not found")
	}

	// get matching data
	equivalents, err := s.matchingManagementService.GetEquivalentsByRegistrationID(id, "GET", token)
	if err != nil {
		return dto.GetRegistrationResponse{}, err
	}

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
		ApprovalStatus:            registration.ApprovalStatus,
		Documents:                 convertToDocumentResponse(registration.Document),
		Equivalents:               equivalents,
	}

	return response, nil
}

func (s *mockRegistrationService) CreateRegistration(ctx context.Context, registration dto.CreateRegistrationRequest, file *multipart.FileHeader, geoletter *multipart.FileHeader, tx *gorm.DB, token string) error {
	var registrationEntity entity.Registration
	var activitiesData []map[string]interface{}
	var usersData []map[string]interface{}
	var activitiesDataOld []map[string]interface{}
	// get user data
	userData := s.userManagementService.GetUserData("GET", token)
	if userData == nil {
		return errors.New("user data not found")
	}

	if registration.ActivityID != "" {
		activitiesData = s.activityManagementService.GetActivitiesData(map[string]interface{}{
			"activity_id":     registration.ActivityID,
			"program_type_id": "",
			"level_id":        "",
			"group_id":        "",
			"name":            "",
		}, "POST", token)
	}

	approvalStatus, ok := activitiesData[0]["approval_status"].(string)
	if !ok {
		return errors.New("approval status not found")
	}

	if approvalStatus != "APPROVED" {
		return errors.New("this activity is not open for registration")
	}

	if userData["id"] != "" {
		usersData = s.userManagementService.GetUserByFilter(map[string]interface{}{
			"user_id": userData["id"],
		}, "POST", token)
	}

	if len(activitiesData) == 0 || len(usersData) == 0 {
		return errors.New("data not found")
	}

	// Then safely extract values with type assertions and nil checks
	var activityName string
	if name, ok := activitiesData[0]["name"]; ok && name != nil {
		activityName, ok = name.(string)
		if !ok {
			return errors.New("Activity not found") // Default value if type assertion fails
		}
	} else {
		return errors.New("Activity not found") // Default value if key doesn't exist or is nil
	}

	// get registration by activity_id and user_nrp
	registrationByActivityIDAndNRP, err := s.registrationRepository.FindByActivityIDAndNRP(ctx, registration.ActivityID, userData["nrp"].(string), tx)
	if err != nil {
		if err.Error() != "record not found" {
			return errors.New("error getting registration by activity_id and user_nrp")
		}
	}

	if registrationByActivityIDAndNRP.ActivityID == registration.ActivityID {
		return errors.New("user already registered")
	}

	// get registration by user nrp and check the academic year (2024/2025)
	// because user can only register in one semester of academic year (2024/2025)
	// use
	registrationsByNRP, err := s.registrationRepository.FindByNRP(ctx, userData["nrp"].(string), tx)
	if err != nil {
		return errors.New("data not found")
	}

	if registrationsByNRP.ActivityID == registration.ActivityID {
		return errors.New("user already registered")
	} else if registrationsByNRP.ActivityID != registration.ActivityID && registrationsByNRP.ActivityID != "" {
		log.Println("REGISTRATION EXISTS BUT DIFFERENT ACTIVITY ID", registrationsByNRP.ActivityID, registration.ActivityID)
		// get activity data by registrationsByNRP.ActivityID
		if registrationsByNRP.ActivityID != "" {
			activitiesDataOld = s.activityManagementService.GetActivitiesData(map[string]interface{}{
				"activity_id":     registrationsByNRP.ActivityID,
				"program_type_id": "",
				"level_id":        "",
				"group_id":        "",
				"name":            "",
			}, "POST", token)
			log.Println("ACTIVITIES DATA OLD", activitiesDataOld)
		}

		// Check if activitiesDataOld has valid data
		if len(activitiesDataOld) == 0 {
			return errors.New("old activity data not found")
		}

		// Safely parse the old activity start date and duration
		var activityOldStartDate time.Time
		var activityOldMonthsDuration int

		// Handle start_date - could be string or another format
		if startDateVal, ok := activitiesDataOld[0]["start_period"]; ok && startDateVal != nil {
			// Try to parse as RFC3339 if it's a string
			if startDateStr, ok := startDateVal.(string); ok {
				parsedTime, err := time.Parse(time.RFC3339, startDateStr)
				if err != nil {
					return errors.New("invalid start_period format in old activity")
				}
				activityOldStartDate = parsedTime
			} else if startDate, ok := startDateVal.(time.Time); ok {
				// It's already a time.Time
				activityOldStartDate = startDate
			} else {
				return errors.New("invalid start_period type in old activity")
			}
		} else {
			return errors.New("start_period not found in old activity")
		}

		// Handle months_duration - could be float64 or int
		if durationVal, ok := activitiesDataOld[0]["months_duration"]; ok && durationVal != nil {
			switch v := durationVal.(type) {
			case int:
				activityOldMonthsDuration = v
			case float64:
				activityOldMonthsDuration = int(v)
			default:
				return errors.New("invalid months_duration type in old activity")
			}
		} else {
			return errors.New("months_duration not found in old activity")
		}

		activityOldEndDate := activityOldStartDate.AddDate(0, activityOldMonthsDuration, 0)

		// Similarly handle start_date for the new activity
		var activityNewStartDate time.Time
		if startDateVal, ok := activitiesData[0]["start_period"]; ok && startDateVal != nil {
			// Try to parse as RFC3339 if it's a string
			if startDateStr, ok := startDateVal.(string); ok {
				parsedTime, err := time.Parse(time.RFC3339, startDateStr)
				if err != nil {
					return errors.New("invalid start_period format in new activity")
				}
				activityNewStartDate = parsedTime
			} else if startDate, ok := startDateVal.(time.Time); ok {
				// It's already a time.Time
				activityNewStartDate = startDate
			} else {
				return errors.New("invalid start_period type in new activity")
			}
		} else {
			return errors.New("start_period not found in new activity")
		}

		// if activity new start date is after activity old end date then user can register
		if !activityNewStartDate.After(activityOldEndDate) {
			return errors.New("user already registered for an overlapping activity period")
		}
	}

	var userID string
	if id, ok := userData["id"]; ok && id != nil {
		userID, ok = id.(string)
		if !ok {
			return errors.New("User not found") // Default value if type assertion fails
		}
	} else {
		return errors.New("User not found") // Default value if key doesn't exist or is nil
	}

	var userNRP string
	if nrp, ok := usersData[0]["nrp"]; ok && nrp != nil {
		userNRP, ok = nrp.(string)
		if !ok {
			return errors.New("User not found") // Default value if type assertion fails
		}
	} else {
		return errors.New("User not found") // Default value if key doesn't exist or is nil
	}

	var userName string
	if name, ok := usersData[0]["name"]; ok && name != nil {
		userName, ok = name.(string)
		if !ok {
			return errors.New("User not found") // Default value if type assertion fails
		}
	} else {
		return errors.New("User not found") // Default value if key doesn't exist or is nil
	}

	// upload file
	result, err := s.fileService.Storage.GcsUpload(file, "sim_mbkm", "", "")
	if err != nil {
		return errors.New("failed to upload file")
	}

	geoletterResult, err := s.fileService.Storage.GcsUpload(geoletter, "sim_mbkm", "", "")
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

func (s *mockRegistrationService) UpdateRegistration(ctx context.Context, id string, registration dto.UpdateRegistrationDataRequest, token string, tx *gorm.DB) error {
	access := s.RegistrationsDataAccess(ctx, id, token, tx)
	if !access {
		return errors.New("data not found")
	}

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

func (s *mockRegistrationService) DeleteRegistration(ctx context.Context, id string, token string, tx *gorm.DB) error {
	access := s.RegistrationsDataAccess(ctx, id, token, tx)
	if !access {
		return errors.New("data not found")
	}

	// get registration by id
	registration, err := s.registrationRepository.FindByID(ctx, id, tx)
	if err != nil {
		return err
	}

	// delete document
	for _, document := range registration.Document {
		_, err = s.fileService.Storage.GcsDelete(document.FileStorageID, "sim_mbkm", "")
		if err != nil {
			return err
		}
	}

	err = s.registrationRepository.Destroy(ctx, id, tx)
	if err != nil {
		return err
	}

	return nil
}

func (s *mockRegistrationService) FindRegistrationByAdvisor(ctx context.Context, pagReq dto.PaginationRequest, filter dto.FilterRegistrationRequest, token string, tx *gorm.DB) ([]dto.GetRegistrationResponse, dto.PaginationResponse, error) {
	userEmail := s.ValidateAdvisor(ctx, token, tx)
	if userEmail == "" {
		return []dto.GetRegistrationResponse{}, dto.PaginationResponse{}, errors.New("Unauthorized")
	}

	filter.AcademicAdvisorEmail = userEmail

	// filter user data
	registrations, total, err := s.registrationRepository.Index(ctx, tx, pagReq, filter)
	if err != nil {
		return []dto.GetRegistrationResponse{}, dto.PaginationResponse{}, err
	}

	metaData := helper.MetaDataPagination(total, pagReq)

	var response []dto.GetRegistrationResponse
	for _, registration := range registrations {
		// get equivalent data
		equivalents, err := s.matchingManagementService.GetEquivalentsByRegistrationID(registration.ID.String(), "GET", token)
		if err != nil {
			return []dto.GetRegistrationResponse{}, dto.PaginationResponse{}, err
		}

		// get matching data
		matching, err := s.matchingManagementService.GetMatchingByActivityID(registration.ActivityID, "GET", token)
		if err != nil {
			return []dto.GetRegistrationResponse{}, dto.PaginationResponse{}, err
		}

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
			ApprovalStatus:            registration.ApprovalStatus,
			Equivalents:               equivalents,
			Matching:                  matching,
			Documents:                 convertToDocumentResponse(registration.Document),
		})
	}

	return response, metaData, nil
}

func (s *mockRegistrationService) FindRegistrationByLOMBKM(ctx context.Context, pagReq dto.PaginationRequest, filter dto.FilterRegistrationRequest, token string, tx *gorm.DB) ([]dto.GetRegistrationResponse, dto.PaginationResponse, error) {
	// filter user data
	registrations, total, err := s.registrationRepository.Index(ctx, tx, pagReq, filter)
	if err != nil {
		return []dto.GetRegistrationResponse{}, dto.PaginationResponse{}, err
	}

	metaData := helper.MetaDataPagination(total, pagReq)

	var response []dto.GetRegistrationResponse
	for _, registration := range registrations {
		// get equivalent data
		equivalents, err := s.matchingManagementService.GetEquivalentsByRegistrationID(registration.ID.String(), "GET", token)
		if err != nil {
			return []dto.GetRegistrationResponse{}, dto.PaginationResponse{}, err
		}

		// get matching data
		matching, err := s.matchingManagementService.GetMatchingByActivityID(registration.ActivityID, "GET", token)
		if err != nil {
			return []dto.GetRegistrationResponse{}, dto.PaginationResponse{}, err
		}

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
			ApprovalStatus:            registration.ApprovalStatus,
			Equivalents:               equivalents,
			Matching:                  matching,
			Documents:                 convertToDocumentResponse(registration.Document),
		})
	}

	return response, metaData, nil
}

func (s *mockRegistrationService) FindRegistrationByStudent(ctx context.Context, pagReq dto.PaginationRequest, filter dto.FilterRegistrationRequest, token string, tx *gorm.DB) ([]dto.GetRegistrationResponse, dto.PaginationResponse, error) {
	userNRP := s.ValidateStudent(ctx, token, tx)
	if userNRP == "" {
		return []dto.GetRegistrationResponse{}, dto.PaginationResponse{}, errors.New("Unauthorized")
	}

	filter.UserNRP = userNRP

	// filter user data
	registrations, total, err := s.registrationRepository.Index(ctx, tx, pagReq, filter)
	if err != nil {
		return []dto.GetRegistrationResponse{}, dto.PaginationResponse{}, err
	}

	metaData := helper.MetaDataPagination(total, pagReq)

	var response []dto.GetRegistrationResponse
	for _, registration := range registrations {
		equivalents, err := s.matchingManagementService.GetEquivalentsByRegistrationID(registration.ID.String(), "GET", token)
		if err != nil {
			return []dto.GetRegistrationResponse{}, dto.PaginationResponse{}, err
		}
		response = append(response, dto.GetRegistrationResponse{
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
			ApprovalStatus:            registration.ApprovalStatus,
			Documents:                 convertToDocumentResponse(registration.Document),
			Equivalents:               equivalents,
		})
	}

	return response, metaData, nil
}

func (s *mockRegistrationService) ValidateAdvisor(ctx context.Context, token string, tx *gorm.DB) string {
	var state bool
	state = false

	userData := s.userManagementService.GetUserData("GET", token)
	if userData == nil {
		return ""
	}

	var userRole string
	if role, ok := userData["role"]; ok && role != nil {
		userRole, ok = role.(string)
		if !ok {
			return ""
		}
	} else {
		return ""
	}

	var userEmail string
	if email, ok := userData["email"]; ok && email != nil {
		userEmail, ok = email.(string)
		if !ok {
			return ""
		}
	} else {
		return ""
	}

	if userRole == "DOSEN PEMBIMBING" {
		state = true
	}

	if state {
		return userEmail
	}

	return ""
}

func (s *mockRegistrationService) ValidateStudent(ctx context.Context, token string, tx *gorm.DB) string {
	var state bool
	state = false

	userData := s.userManagementService.GetUserData("GET", token)
	if userData == nil {
		return ""
	}

	var userRole string
	if role, ok := userData["role"]; ok && role != nil {
		userRole, ok = role.(string)
		if !ok {
			return ""
		}
	} else {
		return ""
	}

	var userNRP string
	if email, ok := userData["nrp"]; ok && email != nil {
		userNRP, ok = email.(string)
		if !ok {
			return ""
		}
	} else {
		return ""
	}

	if userRole == "MAHASISWA" {
		state = true
	}

	if state {
		return userNRP
	}

	return ""
}

func (s *mockRegistrationService) createReportSchedules(ctx context.Context, registration entity.Registration, token string) error {
	// get activity data
	activityData := s.activityManagementService.GetActivitiesData(map[string]interface{}{
		"activity_id":     registration.ActivityID,
		"program_type_id": "",
		"level_id":        "",
		"group_id":        "",
		"name":            "",
	}, "POST", token)

	// calculate how many times should upload the report schedule and week based on months_duration and start_period
	monthsDuration := int(activityData[0]["months_duration"].(float64))
	startPeriod := activityData[0]["start_period"].(string)

	// convert start_period to time
	startPeriodTime, err := time.Parse(time.RFC3339, startPeriod)
	if err != nil {
		log.Println("ERROR PARSE START PERIOD", err)
		return err
	}

	// Calculate total weeks (4 weeks per month)
	totalWeeks := monthsDuration * 4

	// Create weekly report schedules
	for week := 1; week <= totalWeeks; week++ {
		// Calculate start and end dates for each week
		startDate := startPeriodTime.AddDate(0, 0, (week-1)*7)
		endDate := startDate.AddDate(0, 0, 6) // 6 days after start date

		// create report schedule
		err = s.monitoringManagementService.CreateReportSchedule(map[string]interface{}{
			"registration_id":        registration.ID.String(),
			"user_id":                registration.UserID,
			"user_nrp":               registration.UserNRP,
			"academic_advisor_id":    registration.AcademicAdvisorID,
			"academic_advisor_email": registration.AcademicAdvisorEmail,
			"report_type":            "WEEKLY_REPORT",
			"week":                   week,
			"start_date":             startDate,
			"end_date":               endDate,
		}, "POST", token)

		if err != nil {
			return err
		}
	}

	// Create final report schedule
	finalReportStartDate := startPeriodTime
	finalReportEndDate := startPeriodTime.AddDate(0, monthsDuration, 0)

	err = s.monitoringManagementService.CreateReportSchedule(map[string]interface{}{
		"registration_id":        registration.ID.String(),
		"user_id":                registration.UserID,
		"user_nrp":               registration.UserNRP,
		"academic_advisor_id":    registration.AcademicAdvisorID,
		"academic_advisor_email": registration.AcademicAdvisorEmail,
		"report_type":            "FINAL_REPORT",
		"week":                   totalWeeks,
		"start_date":             finalReportStartDate,
		"end_date":               finalReportEndDate,
	}, "POST", token)

	if err != nil {
		return err
	}

	return nil
}

func (s *mockRegistrationService) AdvisorRegistrationApproval(ctx context.Context, token string, approval dto.ApprovalRequest, tx *gorm.DB) error {
	userEmail := s.ValidateAdvisor(ctx, token, tx)
	if userEmail == "" {
		return errors.New("Unauthorized")
	}

	for _, id := range approval.ID {
		registration, err := s.registrationRepository.FindByID(ctx, id, tx)

		if err != nil {
			log.Println("ERROR FIND REGISTRATION", err)
			return err
		}

		if registration.AcademicAdvisorEmail != userEmail {
			log.Println("UNAUTHORIZED")
			return errors.New("Unauthorized")
		}

		if registration.AcademicAdvisorValidation == "APPROVED" && approval.Status == "APPROVED" {
			return errors.New("Registration already approved")
		}

		if registration.AcademicAdvisorValidation == "REJECTED" && approval.Status == "REJECTED" {
			return errors.New("Registration already rejected")
		}

		if approval.Status == "APPROVED" {
			registration.AcademicAdvisorValidation = "APPROVED"
			if registration.LOValidation == "APPROVED" {
				registration.ApprovalStatus = true
				// Create report schedules when both validations are approved
				err = s.createReportSchedules(ctx, registration, token)
				if err != nil {
					return err
				}
			}
		}
		if approval.Status == "REJECTED" {
			registration.AcademicAdvisorValidation = "REJECTED"
			// if registration.LOValidation == "REJECTED" {
			registration.ApprovalStatus = false
			log.Println("REJECTED", registration)
			// }
		}

		log.Println("UPDATE", registration)
		err = s.registrationRepository.Update(ctx, id, registration, tx)
		if err != nil {
			return err
		}

	}

	return nil
}

func (s *mockRegistrationService) LORegistrationApproval(ctx context.Context, token string, approval dto.ApprovalRequest, tx *gorm.DB) error {
	for _, id := range approval.ID {
		registration, err := s.registrationRepository.FindByID(ctx, id, tx)

		if err != nil {
			return err
		}

		if registration.LOValidation == "APPROVED" && approval.Status == "APPROVED" {
			return errors.New("Registration already approved")
		}

		if registration.LOValidation == "REJECTED" && approval.Status == "REJECTED" {
			return errors.New("Registration already rejected")
		}

		if approval.Status == "APPROVED" {
			registration.LOValidation = "APPROVED"
			if registration.AcademicAdvisorValidation == "APPROVED" {
				registration.ApprovalStatus = true
				// Create report schedules when both validations are approved
				err = s.createReportSchedules(ctx, registration, token)
				if err != nil {
					return err
				}
			}
		}
		if approval.Status == "REJECTED" {
			registration.LOValidation = "REJECTED"
			// if registration.AcademicAdvisorValidation == "REJECTED" {
			// }
			registration.ApprovalStatus = false
		}

		err = s.registrationRepository.Update(ctx, id, registration, tx)
		if err != nil {
			return err
		}

	}

	return nil
}

func (s *mockRegistrationService) GetRegistrationTranscript(ctx context.Context, id string, token string, tx *gorm.DB) (dto.TranscriptResponse, error) {
	// Check if the user has access to this registration
	access := s.RegistrationsDataAccess(ctx, id, token, tx)
	if !access {
		return dto.TranscriptResponse{}, errors.New("data not found")
	}

	// Get registration data
	registration, err := s.registrationRepository.FindByID(ctx, id, tx)
	if err != nil {
		return dto.TranscriptResponse{}, err
	}

	// Check if registration is approved
	if !registration.ApprovalStatus {
		return dto.TranscriptResponse{}, errors.New("registration is not approved")
	}

	// Get transcript data from monitoring service
	transcriptData, err := s.monitoringManagementService.GetTranscriptByRegistrationID(id, token)
	if err != nil {
		return dto.TranscriptResponse{}, err
	}

	// Create response
	response := dto.TranscriptResponse{
		RegistrationID: registration.ID.String(),
		UserID:         registration.UserID,
		UserNRP:        registration.UserNRP,
		UserName:       registration.UserName,
		ActivityName:   registration.ActivityName,
		Semester:       registration.Semester,
		TotalSKS:       registration.TotalSKS,
		ApprovalStatus: registration.ApprovalStatus,
		TranscriptData: transcriptData,
	}

	return response, nil
}

func (s *mockRegistrationService) GetStudentRegistrationsWithTranscripts(ctx context.Context, pagReq dto.PaginationRequest, filter dto.FilterRegistrationRequest, token string, tx *gorm.DB) (dto.StudentTranscriptsResponse, dto.PaginationResponse, error) {
	// Validate student authentication and get NRP
	userNRP := s.ValidateStudent(ctx, token, tx)
	if userNRP == "" {
		return dto.StudentTranscriptsResponse{}, dto.PaginationResponse{}, errors.New("Unauthorized")
	}

	// Get user data
	userData := s.userManagementService.GetUserData("GET", token)
	if userData == nil {
		return dto.StudentTranscriptsResponse{}, dto.PaginationResponse{}, errors.New("User data not found")
	}

	// Set filter for this student
	filter.UserNRP = userNRP

	// Get all student registrations
	registrations, total, err := s.registrationRepository.Index(ctx, tx, pagReq, filter)
	if err != nil {
		return dto.StudentTranscriptsResponse{}, dto.PaginationResponse{}, err
	}

	// Create pagination metadata
	metaData := helper.MetaDataPagination(total, pagReq)

	// Prepare the response
	response := dto.StudentTranscriptsResponse{
		UserID:   userData["id"].(string),
		UserNRP:  userNRP,
		UserName: userData["name"].(string),
	}

	// Process each registration to fetch transcript data
	var studentTranscripts []dto.StudentTranscriptResponse
	for _, registration := range registrations {
		var transcriptData interface{} = nil

		// Only fetch transcript data if registration is approved
		if registration.ApprovalStatus {
			// Try to get transcript data
			transcriptData, err = s.monitoringManagementService.GetTranscriptByRegistrationID(registration.ID.String(), token)
			if err != nil {
				// Just log the error and continue, don't fail the whole request
				log.Printf("Error fetching transcript for registration %s: %v", registration.ID.String(), err)
			}
		} else {
			continue
		}

		studentTranscripts = append(studentTranscripts, dto.StudentTranscriptResponse{
			RegistrationID:            registration.ID.String(),
			ActivityID:                registration.ActivityID,
			ActivityName:              registration.ActivityName,
			Semester:                  registration.Semester,
			TotalSKS:                  registration.TotalSKS,
			ApprovalStatus:            registration.ApprovalStatus,
			LOValidation:              registration.LOValidation,
			AcademicAdvisorValidation: registration.AcademicAdvisorValidation,
			TranscriptData:            transcriptData,
		})
	}

	response.Registrations = studentTranscripts
	return response, metaData, nil
}

func (s *mockRegistrationService) GetStudentRegistrationsWithSyllabuses(ctx context.Context, pagReq dto.PaginationRequest, filter dto.FilterRegistrationRequest, token string, tx *gorm.DB) (dto.StudentSyllabusesResponse, dto.PaginationResponse, error) {
	// Validate student authentication and get NRP
	userNRP := s.ValidateStudent(ctx, token, tx)
	if userNRP == "" {
		return dto.StudentSyllabusesResponse{}, dto.PaginationResponse{}, errors.New("Unauthorized")
	}

	// Get user data
	userData := s.userManagementService.GetUserData("GET", token)
	if userData == nil {
		return dto.StudentSyllabusesResponse{}, dto.PaginationResponse{}, errors.New("User data not found")
	}

	// Set filter for this student
	filter.UserNRP = userNRP

	// Get all student registrations
	registrations, total, err := s.registrationRepository.Index(ctx, tx, pagReq, filter)
	if err != nil {
		return dto.StudentSyllabusesResponse{}, dto.PaginationResponse{}, err
	}

	// Create pagination metadata
	metaData := helper.MetaDataPagination(total, pagReq)

	// Prepare the response
	response := dto.StudentSyllabusesResponse{
		UserID:   userData["id"].(string),
		UserNRP:  userNRP,
		UserName: userData["name"].(string),
	}

	// Process each registration to fetch transcript data
	var studentSyllabuses []dto.StudentSyllabusResponse
	for _, registration := range registrations {
		var syllabusData interface{} = nil

		// Only fetch transcript data if registration is approved
		if registration.ApprovalStatus {
			// Try to get transcript data
			syllabusData, err = s.monitoringManagementService.GetSyllabusByRegistrationID(registration.ID.String(), token)
			if err != nil {
				// Just log the error and continue, don't fail the whole request
				log.Printf("Error fetching transcript for registration %s: %v", registration.ID.String(), err)
			}
		} else {
			continue
		}

		studentSyllabuses = append(studentSyllabuses, dto.StudentSyllabusResponse{
			RegistrationID:            registration.ID.String(),
			ActivityID:                registration.ActivityID,
			ActivityName:              registration.ActivityName,
			Semester:                  registration.Semester,
			TotalSKS:                  registration.TotalSKS,
			ApprovalStatus:            registration.ApprovalStatus,
			LOValidation:              registration.LOValidation,
			AcademicAdvisorValidation: registration.AcademicAdvisorValidation,
			SyllabusData:              syllabusData,
		})
	}

	response.Registrations = studentSyllabuses
	return response, metaData, nil
}

func (s *mockRegistrationService) FindRegistrationsWithMatching(ctx context.Context, pagReq dto.PaginationRequest, filter dto.FilterRegistrationRequest, token string, tx *gorm.DB) (dto.StudentRegistrationsWithMatchingResponse, dto.PaginationResponse, error) {
	// Validate student authentication and get NRP
	userNRP := s.ValidateStudent(ctx, token, tx)
	if userNRP == "" {
		return dto.StudentRegistrationsWithMatchingResponse{}, dto.PaginationResponse{}, errors.New("Unauthorized")
	}

	// Get user data
	userData := s.userManagementService.GetUserData("GET", token)
	if userData == nil {
		return dto.StudentRegistrationsWithMatchingResponse{}, dto.PaginationResponse{}, errors.New("User data not found")
	}

	// Set filter for this student
	filter.UserNRP = userNRP

	// Get all student registrations
	registrations, total, err := s.registrationRepository.Index(ctx, tx, pagReq, filter)
	if err != nil {
		return dto.StudentRegistrationsWithMatchingResponse{}, dto.PaginationResponse{}, err
	}

	// Create pagination metadata
	metaData := helper.MetaDataPagination(total, pagReq)

	// Prepare the response
	response := dto.StudentRegistrationsWithMatchingResponse{
		UserID:   userData["id"].(string),
		UserNRP:  userNRP,
		UserName: userData["name"].(string),
	}

	// Process each registration to fetch matching data
	var studentRegistrations []dto.StudentRegistrationWithMatchingResponse
	for _, registration := range registrations {
		// Get equivalents data
		equivalents, err := s.matchingManagementService.GetEquivalentsByRegistrationID(registration.ID.String(), "GET", token)
		if err != nil {
			// Just log the error and continue, don't fail the whole request
			log.Printf("Error fetching equivalents for registration %s: %v", registration.ID.String(), err)
		}

		// Get matching data
		matching, err := s.matchingManagementService.GetMatchingByActivityID(registration.ActivityID, "GET", token)
		if err != nil {
			// Just log the error and continue, don't fail the whole request
			log.Printf("Error fetching matching for activity %s: %v", registration.ActivityID, err)
		}

		studentRegistrations = append(studentRegistrations, dto.StudentRegistrationWithMatchingResponse{
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
			ApprovalStatus:            registration.ApprovalStatus,
			Documents:                 convertToDocumentResponse(registration.Document),
			Equivalents:               equivalents,
			Matching:                  matching,
		})
	}

	response.Registrations = studentRegistrations
	return response, metaData, nil
}

func (s *mockRegistrationService) CheckRegistrationEligibility(ctx context.Context, activityID string, token string, tx *gorm.DB) (dto.RegistrationEligibilityResponse, error) {
	// Get user data from token
	userData := s.userManagementService.GetUserData("GET", token)
	if userData == nil {
		return dto.RegistrationEligibilityResponse{
			Eligible: false,
			Message:  "User data not found",
		}, errors.New("user data not found")
	}

	// Ensure we have an NRP
	userNRP, ok := userData["nrp"].(string)
	if !ok || userNRP == "" {
		return dto.RegistrationEligibilityResponse{
			Eligible: false,
			Message:  "Invalid user NRP",
		}, errors.New("invalid user NRP")
	}

	// Check if the activity exists and is open for registration
	activitiesData := s.activityManagementService.GetActivitiesData(map[string]interface{}{
		"activity_id":     activityID,
		"program_type_id": "",
		"level_id":        "",
		"group_id":        "",
		"name":            "",
	}, "POST", token)

	if len(activitiesData) == 0 {
		return dto.RegistrationEligibilityResponse{
			Eligible: false,
			Message:  "Activity not found",
		}, errors.New("activity not found")
	}

	approvalStatus, ok := activitiesData[0]["approval_status"].(string)
	if !ok || approvalStatus != "APPROVED" {
		return dto.RegistrationEligibilityResponse{
			Eligible: false,
			Message:  "This activity is not open for registration",
		}, errors.New("this activity is not open for registration")
	}

	// Check if already registered for this activity
	registrationByActivityIDAndNRP, err := s.registrationRepository.FindByActivityIDAndNRP(ctx, activityID, userNRP, tx)
	if err != nil {
		if err.Error() != "record not found" {
			return dto.RegistrationEligibilityResponse{
				Eligible: false,
				Message:  "Error checking existing registration",
			}, errors.New("error getting registration by activity_id and user_nrp")
		}
	}

	if registrationByActivityIDAndNRP.ActivityID == activityID {
		return dto.RegistrationEligibilityResponse{
			Eligible: false,
			Message:  "User already registered for this activity",
		}, errors.New("user already registered")
	}

	// Check existing registrations for time conflicts
	registrationsByNRP, err := s.registrationRepository.FindByNRP(ctx, userNRP, tx)
	if err != nil {
		if err.Error() != "record not found" {
			return dto.RegistrationEligibilityResponse{
				Eligible: false,
				Message:  "Error checking existing registrations",
			}, errors.New("error checking existing registrations")
		}
	}

	if registrationsByNRP.ActivityID != "" && registrationsByNRP.ActivityID != activityID {
		// Get activity data for the existing registration
		activitiesDataOld := s.activityManagementService.GetActivitiesData(map[string]interface{}{
			"activity_id":     registrationsByNRP.ActivityID,
			"program_type_id": "",
			"level_id":        "",
			"group_id":        "",
			"name":            "",
		}, "POST", token)

		if len(activitiesDataOld) == 0 {
			return dto.RegistrationEligibilityResponse{
				Eligible: false,
				Message:  "Error checking existing activity data",
			}, errors.New("old activity data not found")
		}

		// Parse the old activity start date and duration
		var activityOldStartDate time.Time
		var activityOldMonthsDuration int

		// Handle start_date
		if startDateVal, ok := activitiesDataOld[0]["start_period"]; ok && startDateVal != nil {
			if startDateStr, ok := startDateVal.(string); ok {
				parsedTime, err := time.Parse(time.RFC3339, startDateStr)
				if err != nil {
					return dto.RegistrationEligibilityResponse{
						Eligible: false,
						Message:  "Invalid start period format in existing activity",
					}, errors.New("invalid start_period format in old activity")
				}
				activityOldStartDate = parsedTime
			} else if startDate, ok := startDateVal.(time.Time); ok {
				activityOldStartDate = startDate
			} else {
				return dto.RegistrationEligibilityResponse{
					Eligible: false,
					Message:  "Invalid start period data in existing activity",
				}, errors.New("invalid start_period type in old activity")
			}
		} else {
			return dto.RegistrationEligibilityResponse{
				Eligible: false,
				Message:  "Start period not found in existing activity",
			}, errors.New("start_period not found in old activity")
		}

		// Handle months_duration
		if durationVal, ok := activitiesDataOld[0]["months_duration"]; ok && durationVal != nil {
			switch v := durationVal.(type) {
			case int:
				activityOldMonthsDuration = v
			case float64:
				activityOldMonthsDuration = int(v)
			default:
				return dto.RegistrationEligibilityResponse{
					Eligible: false,
					Message:  "Invalid duration format in existing activity",
				}, errors.New("invalid months_duration type in old activity")
			}
		} else {
			return dto.RegistrationEligibilityResponse{
				Eligible: false,
				Message:  "Duration not found in existing activity",
			}, errors.New("months_duration not found in old activity")
		}

		activityOldEndDate := activityOldStartDate.AddDate(0, activityOldMonthsDuration, 0)

		// Get start date for the new activity
		var activityNewStartDate time.Time
		if startDateVal, ok := activitiesData[0]["start_period"]; ok && startDateVal != nil {
			if startDateStr, ok := startDateVal.(string); ok {
				parsedTime, err := time.Parse(time.RFC3339, startDateStr)
				if err != nil {
					return dto.RegistrationEligibilityResponse{
						Eligible: false,
						Message:  "Invalid start period format in new activity",
					}, errors.New("invalid start_period format in new activity")
				}
				activityNewStartDate = parsedTime
			} else if startDate, ok := startDateVal.(time.Time); ok {
				activityNewStartDate = startDate
			} else {
				return dto.RegistrationEligibilityResponse{
					Eligible: false,
					Message:  "Invalid start period data in new activity",
				}, errors.New("invalid start_period type in new activity")
			}
		} else {
			return dto.RegistrationEligibilityResponse{
				Eligible: false,
				Message:  "Start period not found in new activity",
			}, errors.New("start_period not found in new activity")
		}

		// Check for time conflicts
		if !activityNewStartDate.After(activityOldEndDate) {
			return dto.RegistrationEligibilityResponse{
				Eligible: false,
				Message:  "User already registered for an overlapping activity period",
			}, errors.New("user already registered for an overlapping activity period")
		}
	}

	// All checks passed, user is eligible to register
	return dto.RegistrationEligibilityResponse{
		Eligible: true,
		Message:  "User is eligible to register for this activity",
	}, nil
}

// TestRegistrationsDataAccessAuthorized tests successful authorization scenarios
func (suite *RegistrationServiceTestSuite) TestRegistrationsDataAccessAuthorized() {
	// Setup
	ctx := context.Background()
	id := uuid.New().String()
	token := "Bearer validToken"

	// Create mock registration
	mockRegistration := entity.Registration{
		ID:                   uuid.New(),
		UserID:               "user123",
		AcademicAdvisorEmail: "advisor@test.com",
	}

	// Mock repository behavior
	suite.mockRegistrationRepo.On("FindByID", ctx, id, mock.Anything).Return(mockRegistration, nil)

	// Test cases
	testCases := []struct {
		name     string
		userData map[string]interface{}
		expected bool
	}{
		{
			name: "Student with matching ID",
			userData: map[string]interface{}{
				"id":    "user123",
				"role":  "MAHASISWA",
				"email": "student@test.com",
			},
			expected: true,
		},
		{
			name: "Advisor with matching email",
			userData: map[string]interface{}{
				"id":    "advisor123",
				"role":  "DOSEN PEMBIMBING",
				"email": "advisor@test.com",
			},
			expected: true,
		},
		{
			name: "Admin role",
			userData: map[string]interface{}{
				"id":    "admin123",
				"role":  "ADMIN",
				"email": "admin@test.com",
			},
			expected: true,
		},
		{
			name: "LO-MBKM role",
			userData: map[string]interface{}{
				"id":    "lo123",
				"role":  "LO-MBKM",
				"email": "lo@test.com",
			},
			expected: true,
		},
	}

	for _, tc := range testCases {
		suite.T().Run(tc.name, func(t *testing.T) {
			// Mock user management service
			suite.mockUserManagementService.On("GetUserData", "GET", token).Return(tc.userData).Once()

			// Execute
			result := suite.service.RegistrationsDataAccess(ctx, id, token, nil)

			// Assert
			assert.Equal(t, tc.expected, result)
		})
	}
}

// TestRegistrationsDataAccessUnauthorized tests unauthorized scenarios
func (suite *RegistrationServiceTestSuite) TestRegistrationsDataAccessUnauthorized() {
	// Setup
	ctx := context.Background()
	id := uuid.New().String()
	token := "Bearer validToken"

	// Create mock registration
	mockRegistration := entity.Registration{
		ID:                   uuid.New(),
		UserID:               "user123",
		AcademicAdvisorEmail: "advisor@test.com",
	}

	// Mock repository behavior
	suite.mockRegistrationRepo.On("FindByID", ctx, id, mock.Anything).Return(mockRegistration, nil)

	// Test cases
	testCases := []struct {
		name     string
		userData map[string]interface{}
		expected bool
	}{
		{
			name: "Student with non-matching ID",
			userData: map[string]interface{}{
				"id":    "otheruser",
				"role":  "MAHASISWA",
				"email": "other@test.com",
			},
			expected: false,
		},
		{
			name: "Advisor with non-matching email",
			userData: map[string]interface{}{
				"id":    "advisor123",
				"role":  "DOSEN PEMBIMBING",
				"email": "otheradvisor@test.com",
			},
			expected: false,
		},
		{
			name:     "Missing user data",
			userData: nil,
			expected: false,
		},
	}

	for _, tc := range testCases {
		suite.T().Run(tc.name, func(t *testing.T) {
			// Mock user management service
			if tc.userData != nil {
				suite.mockUserManagementService.On("GetUserData", "GET", token).Return(tc.userData).Once()
			} else {
				suite.mockUserManagementService.On("GetUserData", "GET", token).Return(nil).Once()
			}

			// Execute
			result := suite.service.RegistrationsDataAccess(ctx, id, token, nil)

			// Assert
			assert.Equal(t, tc.expected, result)
		})
	}
}

// TestFindRegistrationByIDSuccess tests successful retrieval of registration
func (suite *RegistrationServiceTestSuite) TestFindRegistrationByIDSuccess() {
	// Setup
	ctx := context.Background()
	id := uuid.New().String()
	token := "Bearer validToken"

	// Mock registration
	mockRegistration := entity.Registration{
		ID:                        uuid.New(),
		ActivityID:                "activity123",
		ActivityName:              "Test Activity",
		UserID:                    "user123",
		UserNRP:                   "12345",
		UserName:                  "Test User",
		AdvisingConfirmation:      true,
		AcademicAdvisorID:         "advisor123",
		AcademicAdvisor:           "Test Advisor",
		AcademicAdvisorEmail:      "advisor@test.com",
		MentorName:                "Test Mentor",
		MentorEmail:               "mentor@test.com",
		LOValidation:              "PENDING",
		AcademicAdvisorValidation: "PENDING",
		Semester:                  1,
		TotalSKS:                  20,
	}

	// Mock equivalents data
	equivalentsData := map[string]interface{}{
		"course_id": "course123",
		"name":      "Course Name",
	}

	// Set up mocks
	suite.mockUserManagementService.On("GetUserData", "GET", token).Return(map[string]interface{}{
		"id":    "user123",
		"role":  "MAHASISWA",
		"email": "user@test.com",
	})
	suite.mockRegistrationRepo.On("FindByID", ctx, id, mock.Anything).Return(mockRegistration, nil)
	suite.mockMatchingManagementService.On("GetEquivalentsByRegistrationID", id, "GET", token).Return(equivalentsData, nil)

	// Execute
	result, err := suite.service.FindRegistrationByID(ctx, id, token, nil)

	// Assert
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), mockRegistration.ID.String(), result.ID)
	assert.Equal(suite.T(), mockRegistration.ActivityName, result.ActivityName)
	assert.Equal(suite.T(), mockRegistration.UserName, result.UserName)
	assert.Equal(suite.T(), equivalentsData, result.Equivalents)
}

// TestFindRegistrationByIDUnauthorized tests unauthorized access
func (suite *RegistrationServiceTestSuite) TestFindRegistrationByIDUnauthorized() {
	// Setup
	ctx := context.Background()
	id := uuid.New().String()
	token := "Bearer validToken"

	// Mock behavior for unauthorized
	suite.mockUserManagementService.On("GetUserData", "GET", token).Return(map[string]interface{}{
		"id":    "otheruser",
		"role":  "MAHASISWA",
		"email": "other@test.com",
	})
	suite.mockRegistrationRepo.On("FindByID", ctx, id, mock.Anything).Return(entity.Registration{UserID: "user123"}, nil)

	// Execute
	result, err := suite.service.FindRegistrationByID(ctx, id, token, nil)

	// Assert
	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), "data not found", err.Error())
	assert.Equal(suite.T(), dto.GetRegistrationResponse{}, result)
}

// TestFindRegistrationByIDNotFound tests not found error
func (suite *RegistrationServiceTestSuite) TestFindRegistrationByIDNotFound() {
	// Create mock repository
	ctx := context.Background()
	token := "Bearer test-token"
	id := uuid.New().String()

	// Mock RegistrationsDataAccess to return false (not found)
	suite.mockUserManagementService.On("GetUserData", "GET", token).Return(map[string]interface{}{
		"id":    "user123",
		"name":  "User Name",
		"role":  "MAHASISWA",
		"email": "user@example.com",
	})

	suite.mockRegistrationRepo.On("FindByID", ctx, id, mock.Anything).Return(entity.Registration{}, gorm.ErrRecordNotFound)

	// Call method
	_, err := suite.service.FindRegistrationByID(ctx, id, token, nil)

	// Assertions
	suite.Error(err)
	suite.Equal("data not found", err.Error())
}

// TestFindRegistrationByIDEquivalentsError tests error fetching equivalents
func (suite *RegistrationServiceTestSuite) TestFindRegistrationByIDEquivalentsError() {
	// Setup
	ctx := context.Background()
	id := uuid.New().String()
	token := "Bearer validToken"

	// Mock registration
	mockRegistration := entity.Registration{
		ID:           uuid.New(),
		ActivityID:   "activity123",
		ActivityName: "Test Activity",
		UserID:       "user123",
		UserName:     "Test User",
	}

	expectedError := errors.New("equivalents fetch error")

	// Set up mocks
	suite.mockUserManagementService.On("GetUserData", "GET", token).Return(map[string]interface{}{
		"id":    "user123",
		"role":  "MAHASISWA",
		"email": "user@test.com",
	})
	suite.mockRegistrationRepo.On("FindByID", ctx, id, mock.Anything).Return(mockRegistration, nil)
	suite.mockMatchingManagementService.On("GetEquivalentsByRegistrationID", id, "GET", token).Return(nil, expectedError)

	// Execute
	result, err := suite.service.FindRegistrationByID(ctx, id, token, nil)

	// Assert
	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), expectedError, err)
	assert.Equal(suite.T(), dto.GetRegistrationResponse{}, result)
}

// TestCreateRegistrationSuccess tests successful registration creation
func (suite *RegistrationServiceTestSuite) TestCreateRegistrationSuccess() {
	// Setup
	ctx := context.Background()
	token := "Bearer validToken"

	// Create a registration request
	registration := dto.CreateRegistrationRequest{
		ActivityID:           "activity123",
		AcademicAdvisorID:    "advisor123",
		AdvisingConfirmation: true,
		AcademicAdvisor:      "Test Advisor",
		AcademicAdvisorEmail: "advisor@example.com",
		MentorName:           "Test Mentor",
		MentorEmail:          "mentor@example.com",
		Semester:             1,
		TotalSKS:             20,
	}

	// Create test files
	file := &multipart.FileHeader{
		Filename: "test-file.pdf",
		Size:     1024,
	}
	geoletter := &multipart.FileHeader{
		Filename: "test-geoletter.pdf",
		Size:     1024,
	}

	// Mock user data
	userData := map[string]interface{}{
		"id":    "user123",
		"nrp":   "12345",
		"name":  "Test User",
		"role":  "MAHASISWA",
		"email": "student@example.com",
	}

	// Mock user filter data
	usersData := []map[string]interface{}{
		{
			"id":   "user123",
			"nrp":  "12345",
			"name": "Test User",
		},
	}

	// Mock activity data
	activityData := []map[string]interface{}{
		{
			"id":              "activity123",
			"name":            "Test Activity",
			"approval_status": "APPROVED",
			"start_period":    "2023-01-01T00:00:00Z",
			"months_duration": float64(3),
		},
	}

	// Mock file upload responses
	fileUploadResponse := &service_mock.FileStorageResponse{
		FileID:     "file-id-123",
		ObjectName: "test-file.pdf",
		Message:    "Upload successful",
	}

	geoletterUploadResponse := &service_mock.FileStorageResponse{
		FileID:     "file-id-456",
		ObjectName: "test-geoletter.pdf",
		Message:    "Upload successful",
	}

	// Setup mocks
	suite.mockUserManagementService.On("GetUserData", "GET", token).Return(userData)
	suite.mockUserManagementService.On("GetUserByFilter", mock.Anything, "POST", token).Return(usersData)
	suite.mockActivityManagementService.On("GetActivitiesData", mock.Anything, "POST", token).Return(activityData)

	// Check if registration exists
	suite.mockRegistrationRepo.On("FindByActivityIDAndNRP", ctx, "activity123", "12345", mock.Anything).
		Return(entity.Registration{}, errors.New("record not found"))

	// Return empty registration without error to avoid "data not found" error
	suite.mockRegistrationRepo.On("FindByNRP", ctx, "12345", mock.Anything).
		Return(entity.Registration{}, nil)

	// File uploads
	suite.mockFileService.Storage.On("GcsUpload", file, "sim_mbkm", "", "").
		Return(fileUploadResponse, nil)
	suite.mockFileService.Storage.On("GcsUpload", geoletter, "sim_mbkm", "", "").
		Return(geoletterUploadResponse, nil)

	// Registration creation
	suite.mockRegistrationRepo.On("Create", ctx, mock.AnythingOfType("entity.Registration"), mock.Anything).
		Return(entity.Registration{ID: uuid.New()}, nil)

	// Document creation
	suite.mockDocumentRepo.On("Create", ctx, mock.AnythingOfType("entity.Document"), mock.Anything).
		Return(entity.Document{ID: uuid.New()}, nil).Times(2)

	// Execute
	err := suite.service.CreateRegistration(ctx, registration, file, geoletter, nil, token)

	// Assert
	suite.NoError(err)

	// Verify expectations
	suite.mockUserManagementService.AssertExpectations(suite.T())
	suite.mockActivityManagementService.AssertExpectations(suite.T())
	suite.mockRegistrationRepo.AssertExpectations(suite.T())
	suite.mockDocumentRepo.AssertExpectations(suite.T())
	suite.mockFileService.Storage.AssertExpectations(suite.T())
}

// TestCreateRegistrationUserNotFound tests error when user is not found
func (suite *RegistrationServiceTestSuite) TestCreateRegistrationUserNotFound() {
	// Setup
	ctx := context.Background()
	token := "Bearer validToken"
	registrationReq := dto.CreateRegistrationRequest{
		ActivityID: "activity123",
	}

	// Mock file headers
	file := &multipart.FileHeader{}
	geoletter := &multipart.FileHeader{}

	// Mock user data - nil to simulate user not found
	suite.mockUserManagementService.On("GetUserData", "GET", token).Return(nil)

	// Execute
	err := suite.service.CreateRegistration(ctx, registrationReq, file, geoletter, nil, token)

	// Assert
	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), "user data not found", err.Error())
}

// TestCreateRegistrationActivityNotApproved tests error when activity is not approved
func (suite *RegistrationServiceTestSuite) TestCreateRegistrationActivityNotApproved() {
	// Setup
	ctx := context.Background()
	token := "Bearer validToken"
	registrationReq := dto.CreateRegistrationRequest{
		ActivityID: "activity123",
	}

	// Mock file headers
	file := &multipart.FileHeader{}
	geoletter := &multipart.FileHeader{}

	// Mock user data
	userData := map[string]interface{}{
		"id":   "user123",
		"nrp":  "12345",
		"name": "Test User",
	}

	// Mock activity data with not approved status
	activityData := []map[string]interface{}{
		{
			"id":              "activity123",
			"name":            "Test Activity",
			"approval_status": "PENDING",
		},
	}

	// Setup mocks
	suite.mockUserManagementService.On("GetUserData", "GET", token).Return(userData)
	suite.mockActivityManagementService.On("GetActivitiesData", mock.Anything, "POST", token).Return(activityData)

	// Execute
	err := suite.service.CreateRegistration(ctx, registrationReq, file, geoletter, nil, token)

	// Assert
	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), "this activity is not open for registration", err.Error())
}

// TestCreateRegistrationAlreadyRegistered tests error when user is already registered
func (suite *RegistrationServiceTestSuite) TestCreateRegistrationAlreadyRegistered() {
	// Setup
	ctx := context.Background()
	token := "Bearer validToken"
	registrationReq := dto.CreateRegistrationRequest{
		ActivityID: "activity123",
	}

	// Mock file headers
	file := &multipart.FileHeader{}
	geoletter := &multipart.FileHeader{}

	// Mock user data
	userData := map[string]interface{}{
		"id":   "user123",
		"nrp":  "12345",
		"name": "Test User",
	}

	// Mock activity data
	activityData := []map[string]interface{}{
		{
			"id":              "activity123",
			"name":            "Test Activity",
			"approval_status": "APPROVED",
		},
	}

	// Mock user by filter response
	userByFilter := []map[string]interface{}{
		{
			"id":   "user123",
			"nrp":  "12345",
			"name": "Test User",
		},
	}

	// Mock existing registration with same activity ID
	existingRegistration := entity.Registration{
		ActivityID: "activity123",
	}

	// Setup mocks
	suite.mockUserManagementService.On("GetUserData", "GET", token).Return(userData)
	suite.mockActivityManagementService.On("GetActivitiesData", mock.Anything, "POST", token).Return(activityData)
	suite.mockUserManagementService.On("GetUserByFilter", mock.Anything, "POST", token).Return(userByFilter)
	suite.mockRegistrationRepo.On("FindByActivityIDAndNRP", ctx, "activity123", "12345", mock.Anything).
		Return(existingRegistration, nil)

	// Execute
	err := suite.service.CreateRegistration(ctx, registrationReq, file, geoletter, nil, token)

	// Assert
	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), "user already registered", err.Error())
}

// TestCreateRegistrationFileUploadError tests error during file upload
func (suite *RegistrationServiceTestSuite) TestCreateRegistrationFileUploadError() {
	// Setup test data
	ctx := context.Background()
	token := "Bearer test-token"

	registration := dto.CreateRegistrationRequest{
		ActivityID:           "activity123",
		AcademicAdvisorID:    "advisor123",
		AdvisingConfirmation: true,
		AcademicAdvisor:      "Test Advisor",
		AcademicAdvisorEmail: "advisor@example.com",
		MentorName:           "Test Mentor",
		MentorEmail:          "mentor@example.com",
		Semester:             1,
		TotalSKS:             20,
	}

	file := &multipart.FileHeader{
		Filename: "test-file.pdf",
		Size:     1024,
	}

	geoletter := &multipart.FileHeader{
		Filename: "test-geoletter.pdf",
		Size:     1024,
	}

	// Mock user data
	userData := map[string]interface{}{
		"id":    "user123",
		"nrp":   "5022123456",
		"name":  "Test Student",
		"role":  "MAHASISWA",
		"email": "student@example.com",
	}

	suite.mockUserManagementService.On("GetUserData", "GET", token).Return(userData)

	// Mock activity data with approved status
	activityData := []map[string]interface{}{
		{
			"id":              "activity123",
			"name":            "Test Activity",
			"approval_status": "APPROVED",
			"start_period":    "2023-01-01T00:00:00Z",
			"months_duration": float64(3),
		},
	}

	suite.mockActivityManagementService.On("GetActivitiesData", mock.Anything, "POST", token).Return(activityData)

	// Mock user filter data
	usersData := []map[string]interface{}{
		{
			"id":   "user123",
			"nrp":  "5022123456",
			"name": "Test Student",
		},
	}

	suite.mockUserManagementService.On("GetUserByFilter", mock.Anything, "POST", token).Return(usersData)

	// Mock registration repository to check for existing registrations
	suite.mockRegistrationRepo.On("FindByActivityIDAndNRP", ctx, "activity123", "5022123456", mock.Anything).
		Return(entity.Registration{}, errors.New("record not found"))

	// Return empty registration without error to avoid "data not found" error
	suite.mockRegistrationRepo.On("FindByNRP", ctx, "5022123456", mock.Anything).
		Return(entity.Registration{}, nil)

	// Mock file upload error
	fileUploadError := errors.New("failed to upload file")

	// Mock file service directly with the Storage field
	suite.mockFileService.Storage.On("GcsUpload", file, "sim_mbkm", "", "").
		Return(nil, fileUploadError)

	// Call the method
	err := suite.service.CreateRegistration(ctx, registration, file, geoletter, nil, token)

	// Assertions
	suite.Error(err)
	suite.Equal("failed to upload file", err.Error())

	suite.mockUserManagementService.AssertExpectations(suite.T())
	suite.mockActivityManagementService.AssertExpectations(suite.T())
	suite.mockRegistrationRepo.AssertExpectations(suite.T())
	suite.mockFileService.Storage.AssertExpectations(suite.T())
}

// TestCreateRegistrationOverlappingPeriod tests error when registering for overlapping activity period
func (suite *RegistrationServiceTestSuite) TestCreateRegistrationOverlappingPeriod() {
	// Setup
	ctx := context.Background()
	token := "Bearer validToken"
	registrationReq := dto.CreateRegistrationRequest{
		ActivityID:           "activity123",
		AcademicAdvisorID:    "advisor123",
		AdvisingConfirmation: true,
		AcademicAdvisor:      "Test Advisor",
		AcademicAdvisorEmail: "advisor@example.com",
		MentorName:           "Test Mentor",
		MentorEmail:          "mentor@example.com",
		Semester:             1,
		TotalSKS:             20,
	}

	// Mock file headers
	file := &multipart.FileHeader{
		Filename: "test-file.pdf",
		Size:     1024,
	}
	geoletter := &multipart.FileHeader{
		Filename: "test-geoletter.pdf",
		Size:     1024,
	}

	// Mock user data
	userData := map[string]interface{}{
		"id":    "user123",
		"nrp":   "12345",
		"name":  "Test User",
		"role":  "MAHASISWA",
		"email": "student@example.com",
	}

	// Mock user filter data
	usersData := []map[string]interface{}{
		{
			"id":   "user123",
			"nrp":  "12345",
			"name": "Test User",
		},
	}

	// Mock activity data for new activity
	activityData := []map[string]interface{}{
		{
			"id":              "activity123",
			"name":            "Test Activity",
			"approval_status": "APPROVED",
			"start_period":    "2023-01-01T00:00:00Z",
			"months_duration": float64(3),
		},
	}

	// Mock activity data for existing activity with overlapping period
	existingActivityData := []map[string]interface{}{
		{
			"id":              "activity456",
			"name":            "Existing Activity",
			"start_period":    "2023-02-01T00:00:00Z",
			"months_duration": float64(4),
		},
	}

	// Mock existing registration with different activity ID
	existingRegistration := entity.Registration{
		ActivityID: "activity456",
	}

	// Setup mocks
	suite.mockUserManagementService.On("GetUserData", "GET", token).Return(userData)
	suite.mockUserManagementService.On("GetUserByFilter", mock.Anything, "POST", token).Return(usersData)

	// Mock activity data requests
	suite.mockActivityManagementService.On("GetActivitiesData", mock.MatchedBy(func(data map[string]interface{}) bool {
		return data["activity_id"] == "activity123"
	}), "POST", token).Return(activityData)

	suite.mockActivityManagementService.On("GetActivitiesData", mock.MatchedBy(func(data map[string]interface{}) bool {
		return data["activity_id"] == "activity456"
	}), "POST", token).Return(existingActivityData)

	// Registration repository mocks
	suite.mockRegistrationRepo.On("FindByActivityIDAndNRP", ctx, "activity123", "12345", mock.Anything).
		Return(entity.Registration{}, errors.New("record not found"))
	suite.mockRegistrationRepo.On("FindByNRP", ctx, "12345", mock.Anything).
		Return(existingRegistration, nil)

	// Execute
	err := suite.service.CreateRegistration(ctx, registrationReq, file, geoletter, nil, token)

	// Assert
	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), "user already registered for an overlapping activity period", err.Error())

	// Verify that our mocks were called as expected
	suite.mockUserManagementService.AssertExpectations(suite.T())
	suite.mockActivityManagementService.AssertExpectations(suite.T())
	suite.mockRegistrationRepo.AssertExpectations(suite.T())
}

// TestUpdateRegistrationSuccess tests successful registration update
func (suite *RegistrationServiceTestSuite) TestUpdateRegistrationSuccess() {
	// Setup
	ctx := context.Background()
	id := uuid.New().String()
	token := "Bearer validToken"

	// Update request
	updateReq := dto.UpdateRegistrationDataRequest{
		AdvisingConfirmation: true,
		MentorName:           "Updated Mentor",
		MentorEmail:          "updated@mentor.com",
		Semester:             2,
		TotalSKS:             24,
	}

	// Existing registration
	existingReg := entity.Registration{
		ID:                   uuid.New(),
		AdvisingConfirmation: false,
		MentorName:           "Original Mentor",
		MentorEmail:          "original@mentor.com",
		Semester:             1,
		TotalSKS:             20,
		UserID:               "user123",
	}

	// Mock authorized access
	suite.mockUserManagementService.On("GetUserData", "GET", token).Return(map[string]interface{}{
		"id":    "user123",
		"role":  "MAHASISWA",
		"email": "user@test.com",
	})
	suite.mockRegistrationRepo.On("FindByID", ctx, id, mock.Anything).Return(existingReg, nil).Times(2)
	suite.mockRegistrationRepo.On("Update", ctx, id, mock.AnythingOfType("entity.Registration"), mock.Anything).Return(nil)

	// Execute
	err := suite.service.UpdateRegistration(ctx, id, updateReq, token, nil)

	// Assert
	assert.NoError(suite.T(), err)
	suite.mockRegistrationRepo.AssertExpectations(suite.T())
}

// TestUpdateRegistrationUnauthorized tests unauthorized update
func (suite *RegistrationServiceTestSuite) TestUpdateRegistrationUnauthorized() {
	// Setup
	ctx := context.Background()
	id := uuid.New().String()
	token := "Bearer validToken"

	// Update request
	updateReq := dto.UpdateRegistrationDataRequest{
		MentorName: "Updated Mentor",
	}

	// Mock unauthorized access
	suite.mockUserManagementService.On("GetUserData", "GET", token).Return(map[string]interface{}{
		"id":    "otheruser",
		"role":  "MAHASISWA",
		"email": "other@test.com",
	})
	suite.mockRegistrationRepo.On("FindByID", ctx, id, mock.Anything).Return(entity.Registration{
		UserID: "user123",
	}, nil)

	// Execute
	err := suite.service.UpdateRegistration(ctx, id, updateReq, token, nil)

	// Assert
	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), "data not found", err.Error())
}

// TestUpdateRegistrationNotFound tests registration not found
func (suite *RegistrationServiceTestSuite) TestUpdateRegistrationNotFound() {
	// Setup test data
	ctx := context.Background()
	token := "Bearer test-token"
	id := uuid.New().String()
	updateData := dto.UpdateRegistrationDataRequest{
		MentorName:  "New Mentor",
		MentorEmail: "new.mentor@example.com",
	}

	// Mock user authentication
	suite.mockUserManagementService.On("GetUserData", "GET", token).Return(map[string]interface{}{
		"id":    "user123",
		"name":  "User Name",
		"role":  "MAHASISWA",
		"email": "user@example.com",
	})

	// Mock repository method to return not found error
	suite.mockRegistrationRepo.On("FindByID", ctx, id, mock.Anything).Return(entity.Registration{}, gorm.ErrRecordNotFound)

	// Call method
	err := suite.service.UpdateRegistration(ctx, id, updateData, token, nil)

	// Assertions
	suite.Error(err)
	suite.Equal("data not found", err.Error())
}

// TestUpdateRegistrationDatabaseError tests error during update
func (suite *RegistrationServiceTestSuite) TestUpdateRegistrationDatabaseError() {
	// Setup
	ctx := context.Background()
	id := uuid.New().String()
	token := "Bearer validToken"

	// Update request
	updateReq := dto.UpdateRegistrationDataRequest{
		MentorName: "Updated Mentor",
	}

	// Mock database error
	dbError := errors.New("database error")

	// Existing registration
	existingReg := entity.Registration{
		ID:         uuid.New(),
		MentorName: "Original Mentor",
		UserID:     "user123",
	}

	suite.mockUserManagementService.On("GetUserData", "GET", token).Return(map[string]interface{}{
		"id":    "user123",
		"role":  "MAHASISWA",
		"email": "user@test.com",
	})
	suite.mockRegistrationRepo.On("FindByID", ctx, id, mock.Anything).Return(existingReg, nil).Times(2)
	suite.mockRegistrationRepo.On("Update", ctx, id, mock.AnythingOfType("entity.Registration"), mock.Anything).Return(dbError)

	// Execute
	err := suite.service.UpdateRegistration(ctx, id, updateReq, token, nil)

	// Assert
	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), dbError, err)
}

// TestDeleteRegistrationSuccess tests successful registration deletion
func (suite *RegistrationServiceTestSuite) TestDeleteRegistrationSuccess() {
	// Setup test data
	ctx := context.Background()
	token := "Bearer test-token"
	id := uuid.New().String()

	// Mock user authentication
	suite.mockUserManagementService.On("GetUserData", "GET", token).Return(map[string]interface{}{
		"id":    "user123",
		"role":  "MAHASISWA",
		"email": "user@example.com",
	})

	// Create a test document
	document := entity.Document{
		ID:            uuid.New(),
		FileStorageID: "file-123",
		Name:          "test-document.pdf",
		DocumentType:  "Acceptence Letter",
	}

	// Create a test registration with document
	registration := entity.Registration{
		ID:       uuid.New(),
		UserID:   "user123",
		Document: []entity.Document{document},
	}

	// Mock repository methods
	suite.mockRegistrationRepo.On("FindByID", ctx, id, mock.Anything).Return(registration, nil)

	// Mock file service delete
	suite.mockFileService.Storage.On("GcsDelete", document.FileStorageID, "sim_mbkm", "").
		Return(&service_mock.FileStorageResponse{}, nil)

	// Mock repository destroy
	suite.mockRegistrationRepo.On("Destroy", ctx, id, mock.Anything).Return(nil)

	// Call the delete method
	err := suite.service.DeleteRegistration(ctx, id, token, nil)

	// Assertions
	suite.NoError(err)

	suite.mockUserManagementService.AssertExpectations(suite.T())
	suite.mockRegistrationRepo.AssertExpectations(suite.T())
	suite.mockFileService.Storage.AssertExpectations(suite.T())
}

// TestDeleteRegistrationUnauthorized tests unauthorized deletion
func (suite *RegistrationServiceTestSuite) TestDeleteRegistrationUnauthorized() {
	// Setup
	ctx := context.Background()
	id := uuid.New().String()
	token := "Bearer validToken"

	// Mock unauthorized access
	suite.mockUserManagementService.On("GetUserData", "GET", token).Return(map[string]interface{}{
		"id":    "otheruser",
		"role":  "MAHASISWA",
		"email": "other@test.com",
	})
	suite.mockRegistrationRepo.On("FindByID", ctx, id, mock.Anything).Return(entity.Registration{
		UserID: "user123",
	}, nil)

	// Execute
	err := suite.service.DeleteRegistration(ctx, id, token, nil)

	// Assert
	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), "data not found", err.Error())
}

// TestDeleteRegistrationNotFound tests registration not found
func (suite *RegistrationServiceTestSuite) TestDeleteRegistrationNotFound() {
	// Setup test data
	ctx := context.Background()
	token := "Bearer test-token"
	id := uuid.New().String()

	// Mock user authentication
	suite.mockUserManagementService.On("GetUserData", "GET", token).Return(map[string]interface{}{
		"id":    "user123",
		"name":  "User Name",
		"role":  "MAHASISWA",
		"email": "user@example.com",
	})

	// Mock FindByID to return not found error
	suite.mockRegistrationRepo.On("FindByID", ctx, id, mock.Anything).Return(entity.Registration{}, gorm.ErrRecordNotFound)

	// Call the delete method
	err := suite.service.DeleteRegistration(ctx, id, token, nil)

	// Assertions
	suite.Error(err)
	suite.Equal("data not found", err.Error())
}

// TestDeleteRegistrationFileDeleteError tests error during file deletion
func (suite *RegistrationServiceTestSuite) TestDeleteRegistrationFileDeleteError() {
	// Setup test data
	ctx := context.Background()
	token := "Bearer test-token"
	id := uuid.New().String()

	// Mock user authentication
	suite.mockUserManagementService.On("GetUserData", "GET", token).Return(map[string]interface{}{
		"id":    "user123",
		"role":  "MAHASISWA",
		"email": "user@example.com",
	})

	// Create a test document
	document := entity.Document{
		ID:            uuid.New(),
		FileStorageID: "file-123",
		Name:          "test-document.pdf",
		DocumentType:  "Acceptence Letter",
	}

	// Create a test registration with document
	registration := entity.Registration{
		ID:       uuid.New(),
		UserID:   "user123",
		Document: []entity.Document{document},
	}

	// Mock repository methods
	suite.mockRegistrationRepo.On("FindByID", ctx, id, mock.Anything).Return(registration, nil)

	// Mock file service to return an error during deletion
	fileDeleteError := errors.New("file delete error")
	suite.mockFileService.Storage.On("GcsDelete", document.FileStorageID, "sim_mbkm", "").
		Return(&service_mock.FileStorageResponse{}, fileDeleteError)

	// Call the delete method
	err := suite.service.DeleteRegistration(ctx, id, token, nil)

	// Assertions
	suite.Error(err)
	suite.Equal(fileDeleteError, err)

	suite.mockUserManagementService.AssertExpectations(suite.T())
	suite.mockRegistrationRepo.AssertExpectations(suite.T())
	suite.mockFileService.Storage.AssertExpectations(suite.T())
}

// TestAdvisorRegistrationApprovalSuccess tests successful approval by advisor
func (suite *RegistrationServiceTestSuite) TestAdvisorRegistrationApprovalSuccess() {
	// Setup
	ctx := context.Background()
	token := "Bearer validToken"
	registrationID := uuid.New().String()

	// Approval request
	approvalReq := dto.ApprovalRequest{
		Status: "APPROVED",
		ID:     []string{registrationID},
	}

	// Mock registration with pending status
	registration := entity.Registration{
		ID:                        uuid.New(),
		ActivityID:                "activity123",
		UserID:                    "user123",
		UserNRP:                   "12345",
		AcademicAdvisorEmail:      "advisor@test.com",
		LOValidation:              "APPROVED", // LO already approved
		AcademicAdvisorValidation: "PENDING",  // Advisor pending
		ApprovalStatus:            false,
	}

	// Mock activity data for report schedule
	activityData := []map[string]interface{}{
		{
			"activity_id":     "activity123",
			"name":            "Test Activity",
			"start_period":    "2023-01-01T00:00:00Z",
			"months_duration": float64(3),
		},
	}

	// Setup mocks
	suite.mockUserManagementService.On("GetUserData", "GET", token).Return(map[string]interface{}{
		"id":    "advisor123",
		"role":  "DOSEN PEMBIMBING",
		"email": "advisor@test.com",
	})
	suite.mockRegistrationRepo.On("FindByID", ctx, registrationID, mock.Anything).Return(registration, nil)
	suite.mockActivityManagementService.On("GetActivitiesData", mock.Anything, "POST", token).Return(activityData)
	suite.mockMonitoringManagementService.On("CreateReportSchedule", mock.Anything, "POST", token).Return(nil).Times(13) // 12 weekly + 1 final
	suite.mockRegistrationRepo.On("Update", ctx, registrationID, mock.AnythingOfType("entity.Registration"), mock.Anything).Return(nil)

	// Execute
	err := suite.service.AdvisorRegistrationApproval(ctx, token, approvalReq, nil)

	// Assert
	assert.NoError(suite.T(), err)
	suite.mockRegistrationRepo.AssertExpectations(suite.T())
	suite.mockMonitoringManagementService.AssertExpectations(suite.T())
}

// TestAdvisorRegistrationApprovalUnauthorized tests unauthorized approval by advisor
func (suite *RegistrationServiceTestSuite) TestAdvisorRegistrationApprovalUnauthorized() {
	// Setup
	ctx := context.Background()
	token := "Bearer validToken"
	registrationID := uuid.New().String()

	// Approval request
	approvalReq := dto.ApprovalRequest{
		Status: "APPROVED",
		ID:     []string{registrationID},
	}

	// Mock unauthorized access - not an advisor
	suite.mockUserManagementService.On("GetUserData", "GET", token).Return(map[string]interface{}{
		"id":    "student123",
		"role":  "MAHASISWA",
		"email": "student@test.com",
	})

	// Execute
	err := suite.service.AdvisorRegistrationApproval(ctx, token, approvalReq, nil)

	// Assert
	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), "Unauthorized", err.Error())
}

// TestAdvisorRegistrationApprovalWrongAdvisor tests approval by wrong advisor
func (suite *RegistrationServiceTestSuite) TestAdvisorRegistrationApprovalWrongAdvisor() {
	// Setup
	ctx := context.Background()
	token := "Bearer validToken"
	registrationID := uuid.New().String()

	// Approval request
	approvalReq := dto.ApprovalRequest{
		Status: "APPROVED",
		ID:     []string{registrationID},
	}

	// Mock registration with different advisor
	registration := entity.Registration{
		ID:                   uuid.New(),
		ActivityID:           "activity123",
		UserID:               "user123",
		UserNRP:              "12345",
		AcademicAdvisorEmail: "correct@advisor.com", // Different from token
	}

	// Mock authorized but wrong advisor
	suite.mockUserManagementService.On("GetUserData", "GET", token).Return(map[string]interface{}{
		"id":    "advisor123",
		"role":  "DOSEN PEMBIMBING",
		"email": "wrong@advisor.com",
	})
	suite.mockRegistrationRepo.On("FindByID", ctx, registrationID, mock.Anything).Return(registration, nil)

	// Execute
	err := suite.service.AdvisorRegistrationApproval(ctx, token, approvalReq, nil)

	// Assert
	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), "Unauthorized", err.Error())
}

// TestAdvisorRegistrationApprovalAlreadyApproved tests already approved registration
func (suite *RegistrationServiceTestSuite) TestAdvisorRegistrationApprovalAlreadyApproved() {
	// Setup
	ctx := context.Background()
	token := "Bearer validToken"
	registrationID := uuid.New().String()

	// Approval request
	approvalReq := dto.ApprovalRequest{
		Status: "APPROVED",
		ID:     []string{registrationID},
	}

	// Mock registration with already approved status
	registration := entity.Registration{
		ID:                        uuid.New(),
		ActivityID:                "activity123",
		UserID:                    "user123",
		UserNRP:                   "12345",
		AcademicAdvisorEmail:      "advisor@test.com",
		AcademicAdvisorValidation: "APPROVED", // Already approved
	}

	// Setup mocks
	suite.mockUserManagementService.On("GetUserData", "GET", token).Return(map[string]interface{}{
		"id":    "advisor123",
		"role":  "DOSEN PEMBIMBING",
		"email": "advisor@test.com",
	})
	suite.mockRegistrationRepo.On("FindByID", ctx, registrationID, mock.Anything).Return(registration, nil)

	// Execute
	err := suite.service.AdvisorRegistrationApproval(ctx, token, approvalReq, nil)

	// Assert
	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), "Registration already approved", err.Error())
}

// TestAdvisorRegistrationApprovalReject tests successful rejection by advisor
func (suite *RegistrationServiceTestSuite) TestAdvisorRegistrationApprovalReject() {
	// Setup
	ctx := context.Background()
	token := "Bearer validToken"
	registrationID := uuid.New().String()

	// Rejection request
	approvalReq := dto.ApprovalRequest{
		Status: "REJECTED",
		ID:     []string{registrationID},
	}

	// Mock registration with pending status
	registration := entity.Registration{
		ID:                        uuid.New(),
		ActivityID:                "activity123",
		UserID:                    "user123",
		UserNRP:                   "12345",
		AcademicAdvisorEmail:      "advisor@test.com",
		LOValidation:              "PENDING",
		AcademicAdvisorValidation: "PENDING",
		ApprovalStatus:            false,
	}

	// Setup mocks
	suite.mockUserManagementService.On("GetUserData", "GET", token).Return(map[string]interface{}{
		"id":    "advisor123",
		"role":  "DOSEN PEMBIMBING",
		"email": "advisor@test.com",
	})
	suite.mockRegistrationRepo.On("FindByID", ctx, registrationID, mock.Anything).Return(registration, nil)
	suite.mockRegistrationRepo.On("Update", ctx, registrationID, mock.AnythingOfType("entity.Registration"), mock.Anything).Return(nil)

	// Execute
	err := suite.service.AdvisorRegistrationApproval(ctx, token, approvalReq, nil)

	// Assert
	assert.NoError(suite.T(), err)
	suite.mockRegistrationRepo.AssertExpectations(suite.T())
}

// TestLORegistrationApprovalSuccess tests successful approval by LO-MBKM
func (suite *RegistrationServiceTestSuite) TestLORegistrationApprovalSuccess() {
	// Setup
	ctx := context.Background()
	token := "Bearer validToken"
	registrationID := uuid.New().String()

	// Approval request
	approvalReq := dto.ApprovalRequest{
		Status: "APPROVED",
		ID:     []string{registrationID},
	}

	// Mock registration with pending status
	registration := entity.Registration{
		ID:                        uuid.New(),
		ActivityID:                "activity123",
		UserID:                    "user123",
		UserNRP:                   "12345",
		AcademicAdvisorEmail:      "advisor@test.com",
		LOValidation:              "PENDING",
		AcademicAdvisorValidation: "APPROVED", // Advisor already approved
		ApprovalStatus:            false,
	}

	// Mock activity data for report schedule
	activityData := []map[string]interface{}{
		{
			"activity_id":     "activity123",
			"name":            "Test Activity",
			"start_period":    "2023-01-01T00:00:00Z",
			"months_duration": float64(3),
		},
	}

	// Setup mocks
	suite.mockRegistrationRepo.On("FindByID", ctx, registrationID, mock.Anything).Return(registration, nil)
	suite.mockActivityManagementService.On("GetActivitiesData", mock.Anything, "POST", token).Return(activityData)
	suite.mockMonitoringManagementService.On("CreateReportSchedule", mock.Anything, "POST", token).Return(nil).Times(13) // 12 weekly + 1 final
	suite.mockRegistrationRepo.On("Update", ctx, registrationID, mock.AnythingOfType("entity.Registration"), mock.Anything).Return(nil)

	// Execute
	err := suite.service.LORegistrationApproval(ctx, token, approvalReq, nil)

	// Assert
	assert.NoError(suite.T(), err)
	suite.mockRegistrationRepo.AssertExpectations(suite.T())
	suite.mockMonitoringManagementService.AssertExpectations(suite.T())
}

// TestLORegistrationApprovalAlreadyApproved tests already approved registration by LO
func (suite *RegistrationServiceTestSuite) TestLORegistrationApprovalAlreadyApproved() {
	// Setup
	ctx := context.Background()
	token := "Bearer validToken"
	registrationID := uuid.New().String()

	// Approval request
	approvalReq := dto.ApprovalRequest{
		Status: "APPROVED",
		ID:     []string{registrationID},
	}

	// Mock registration with already approved status
	registration := entity.Registration{
		ID:                        uuid.New(),
		ActivityID:                "activity123",
		UserID:                    "user123",
		UserNRP:                   "12345",
		LOValidation:              "APPROVED", // Already approved
		AcademicAdvisorValidation: "PENDING",
	}

	// Setup mocks
	suite.mockRegistrationRepo.On("FindByID", ctx, registrationID, mock.Anything).Return(registration, nil)

	// Execute
	err := suite.service.LORegistrationApproval(ctx, token, approvalReq, nil)

	// Assert
	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), "Registration already approved", err.Error())
}

// TestLORegistrationApprovalReject tests successful rejection by LO
func (suite *RegistrationServiceTestSuite) TestLORegistrationApprovalReject() {
	// Setup
	ctx := context.Background()
	token := "Bearer validToken"
	registrationID := uuid.New().String()

	// Rejection request
	approvalReq := dto.ApprovalRequest{
		Status: "REJECTED",
		ID:     []string{registrationID},
	}

	// Mock registration with pending status
	registration := entity.Registration{
		ID:                        uuid.New(),
		ActivityID:                "activity123",
		UserID:                    "user123",
		UserNRP:                   "12345",
		LOValidation:              "PENDING",
		AcademicAdvisorValidation: "PENDING",
		ApprovalStatus:            false,
	}

	// Setup mocks
	suite.mockRegistrationRepo.On("FindByID", ctx, registrationID, mock.Anything).Return(registration, nil)
	suite.mockRegistrationRepo.On("Update", ctx, registrationID, mock.AnythingOfType("entity.Registration"), mock.Anything).Return(nil)

	// Execute
	err := suite.service.LORegistrationApproval(ctx, token, approvalReq, nil)

	// Assert
	assert.NoError(suite.T(), err)
	suite.mockRegistrationRepo.AssertExpectations(suite.T())
}

// TestLORegistrationApprovalError tests error during approval by LO
func (suite *RegistrationServiceTestSuite) TestLORegistrationApprovalError() {
	// Setup
	ctx := context.Background()
	token := "Bearer validToken"
	registrationID := uuid.New().String()

	// Approval request
	approvalReq := dto.ApprovalRequest{
		Status: "APPROVED",
		ID:     []string{registrationID},
	}

	// Mock database error
	dbError := errors.New("database error")

	// Setup mocks
	suite.mockRegistrationRepo.On("FindByID", ctx, registrationID, mock.Anything).Return(entity.Registration{}, dbError)

	// Execute
	err := suite.service.LORegistrationApproval(ctx, token, approvalReq, nil)

	// Assert
	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), dbError, err)
}

func (suite *RegistrationServiceTestSuite) TestValidateStudentSuccess() {
	// Setup test data
	ctx := context.Background()
	token := "Bearer test-token"

	// Setup mock behavior
	userData := map[string]interface{}{
		"id":    "user123",
		"nrp":   "5022123456",
		"name":  "Test Student",
		"role":  "MAHASISWA",
		"email": "student@example.com",
	}

	suite.mockUserManagementService.On("GetUserData", "GET", token).Return(userData)

	// Call the method
	result := suite.service.ValidateStudent(ctx, token, nil)

	// Assertions
	suite.Equal("5022123456", result)
	suite.mockUserManagementService.AssertExpectations(suite.T())
}

func (suite *RegistrationServiceTestSuite) TestValidateStudentInvalidRole() {
	// Setup test data
	ctx := context.Background()
	token := "Bearer test-token"

	// Setup mock behavior
	userData := map[string]interface{}{
		"id":    "user123",
		"nrp":   "5022123456",
		"name":  "Test User",
		"role":  "ADMIN", // Not a student role
		"email": "admin@example.com",
	}

	suite.mockUserManagementService.On("GetUserData", "GET", token).Return(userData)

	// Call the method
	result := suite.service.ValidateStudent(ctx, token, nil)

	// Assertions
	suite.Equal("", result)
	suite.mockUserManagementService.AssertExpectations(suite.T())
}

func (suite *RegistrationServiceTestSuite) TestValidateStudentNilUserData() {
	// Setup test data
	ctx := context.Background()
	token := "Bearer test-token"

	// Setup mock behavior
	suite.mockUserManagementService.On("GetUserData", "GET", token).Return(nil)

	// Call the method
	result := suite.service.ValidateStudent(ctx, token, nil)

	// Assertions
	suite.Equal("", result)
	suite.mockUserManagementService.AssertExpectations(suite.T())
}

func (suite *RegistrationServiceTestSuite) TestValidateStudentMissingNRP() {
	// Setup test data
	ctx := context.Background()
	token := "Bearer test-token"

	// Setup mock behavior - missing nrp field
	userData := map[string]interface{}{
		"id":    "user123",
		"role":  "MAHASISWA",
		"name":  "Test Student",
		"email": "student@example.com",
	}

	suite.mockUserManagementService.On("GetUserData", "GET", token).Return(userData)

	// Call the method
	result := suite.service.ValidateStudent(ctx, token, nil)

	// Assertions
	suite.Equal("", result)
	suite.mockUserManagementService.AssertExpectations(suite.T())
}

func (suite *RegistrationServiceTestSuite) TestValidateAdvisorSuccess() {
	// Setup test data
	ctx := context.Background()
	token := "Bearer test-token"

	// Setup mock behavior
	userData := map[string]interface{}{
		"id":    "advisor123",
		"nrp":   "123456",
		"name":  "Test Advisor",
		"role":  "DOSEN PEMBIMBING",
		"email": "advisor@example.com",
	}

	suite.mockUserManagementService.On("GetUserData", "GET", token).Return(userData)

	// Call the method
	result := suite.service.ValidateAdvisor(ctx, token, nil)

	// Assertions
	suite.Equal("advisor@example.com", result)
	suite.mockUserManagementService.AssertExpectations(suite.T())
}

func (suite *RegistrationServiceTestSuite) TestValidateAdvisorInvalidRole() {
	// Setup test data
	ctx := context.Background()
	token := "Bearer test-token"

	// Setup mock behavior
	userData := map[string]interface{}{
		"id":    "user123",
		"nrp":   "5022123456",
		"name":  "Test User",
		"role":  "MAHASISWA", // Not an advisor role
		"email": "student@example.com",
	}

	suite.mockUserManagementService.On("GetUserData", "GET", token).Return(userData)

	// Call the method
	result := suite.service.ValidateAdvisor(ctx, token, nil)

	// Assertions
	suite.Equal("", result)
	suite.mockUserManagementService.AssertExpectations(suite.T())
}

func (suite *RegistrationServiceTestSuite) TestValidateAdvisorNilUserData() {
	// Setup test data
	ctx := context.Background()
	token := "Bearer test-token"

	// Setup mock behavior
	suite.mockUserManagementService.On("GetUserData", "GET", token).Return(nil)

	// Call the method
	result := suite.service.ValidateAdvisor(ctx, token, nil)

	// Assertions
	suite.Equal("", result)
	suite.mockUserManagementService.AssertExpectations(suite.T())
}

func (suite *RegistrationServiceTestSuite) TestValidateAdvisorMissingEmail() {
	// Setup test data
	ctx := context.Background()
	token := "Bearer test-token"

	// Setup mock behavior - missing email field
	userData := map[string]interface{}{
		"id":   "advisor123",
		"nrp":  "123456",
		"name": "Test Advisor",
		"role": "DOSEN PEMBIMBING",
	}

	suite.mockUserManagementService.On("GetUserData", "GET", token).Return(userData)

	// Call the method
	result := suite.service.ValidateAdvisor(ctx, token, nil)

	// Assertions
	suite.Equal("", result)
	suite.mockUserManagementService.AssertExpectations(suite.T())
}

func TestRegistrationServiceSuite(t *testing.T) {
	suite.Run(t, new(RegistrationServiceTestSuite))
}

func (suite *RegistrationServiceTestSuite) TestFindRegistrationByStudentSuccess() {
	// Setup test data
	ctx := context.Background()
	token := "Bearer test-token"
	pagReq := dto.PaginationRequest{
		Limit:  10,
		Offset: 0,
		URL:    "http://example.com/registrations",
	}
	filter := dto.FilterRegistrationRequest{
		ActivityName: "Test Activity",
	}

	// Mock ValidateStudent
	userData := map[string]interface{}{
		"id":    "user123",
		"nrp":   "5022123456",
		"name":  "Test Student",
		"role":  "MAHASISWA",
		"email": "student@example.com",
	}

	suite.mockUserManagementService.On("GetUserData", "GET", token).Return(userData)

	// Create mock registrations
	now := time.Now()
	mockRegistration := entity.Registration{
		ID:                        uuid.New(),
		ActivityID:                "activity1",
		ActivityName:              "Test Activity",
		UserID:                    "user123",
		UserNRP:                   "5022123456",
		UserName:                  "Test Student",
		AdvisingConfirmation:      true,
		AcademicAdvisorID:         "advisor1",
		AcademicAdvisor:           "Test Advisor",
		AcademicAdvisorEmail:      "advisor@example.com",
		MentorName:                "Test Mentor",
		MentorEmail:               "mentor@example.com",
		LOValidation:              "PENDING",
		AcademicAdvisorValidation: "PENDING",
		Semester:                  1,
		TotalSKS:                  20,
		ApprovalStatus:            false,
		BaseModel: entity.BaseModel{
			CreatedAt: &now,
			UpdatedAt: &now,
		},
	}

	// Set up expected filter with NRP
	expectedFilter := filter
	expectedFilter.UserNRP = "5022123456"

	// Mock repository behavior
	suite.mockRegistrationRepo.On("Index", ctx, mock.Anything, pagReq, expectedFilter).Return(
		[]entity.Registration{mockRegistration}, int64(1), nil)

	// Mock equivalent data
	equivalentData := map[string]interface{}{
		"id": "equiv1",
		"courses": []map[string]interface{}{
			{
				"id":   "course1",
				"name": "Course 1",
			},
		},
	}

	suite.mockMatchingManagementService.On("GetEquivalentsByRegistrationID",
		mockRegistration.ID.String(), "GET", token).Return(equivalentData, nil)

	// Call the method
	registrations, pagination, err := suite.service.FindRegistrationByStudent(ctx, pagReq, filter, token, nil)

	// Assertions
	suite.NoError(err)
	suite.Equal(1, len(registrations))
	suite.Equal(mockRegistration.ID.String(), registrations[0].ID)
	suite.Equal(mockRegistration.ActivityName, registrations[0].ActivityName)
	suite.Equal(mockRegistration.UserNRP, registrations[0].UserNRP)
	suite.Equal(equivalentData, registrations[0].Equivalents)
	suite.Equal(int64(1), pagination.Total)

	suite.mockUserManagementService.AssertExpectations(suite.T())
	suite.mockRegistrationRepo.AssertExpectations(suite.T())
	suite.mockMatchingManagementService.AssertExpectations(suite.T())
}

func (suite *RegistrationServiceTestSuite) TestFindRegistrationByStudentUnauthorized() {
	// Setup test data
	ctx := context.Background()
	token := "Bearer test-token"
	pagReq := dto.PaginationRequest{
		Limit:  10,
		Offset: 0,
	}
	filter := dto.FilterRegistrationRequest{}

	// Setup mock behavior for unauthorized user
	suite.mockUserManagementService.On("GetUserData", "GET", token).Return(nil)

	// Call the method
	registrations, pagination, err := suite.service.FindRegistrationByStudent(ctx, pagReq, filter, token, nil)

	// Assertions
	suite.Error(err)
	suite.Equal("Unauthorized", err.Error())
	suite.Empty(registrations)
	suite.Equal(dto.PaginationResponse{}, pagination)

	suite.mockUserManagementService.AssertExpectations(suite.T())
}

func (suite *RegistrationServiceTestSuite) TestFindRegistrationByStudentRepositoryError() {
	// Setup test data
	ctx := context.Background()
	token := "Bearer test-token"
	pagReq := dto.PaginationRequest{
		Limit:  10,
		Offset: 0,
	}
	filter := dto.FilterRegistrationRequest{}

	// Mock ValidateStudent
	userData := map[string]interface{}{
		"id":    "user123",
		"nrp":   "5022123456",
		"name":  "Test Student",
		"role":  "MAHASISWA",
		"email": "student@example.com",
	}

	suite.mockUserManagementService.On("GetUserData", "GET", token).Return(userData)

	// Set up expected filter with NRP
	expectedFilter := filter
	expectedFilter.UserNRP = "5022123456"

	// Mock repository error
	expectedError := errors.New("database error")
	suite.mockRegistrationRepo.On("Index", ctx, mock.Anything, pagReq, expectedFilter).Return(
		[]entity.Registration{}, int64(0), expectedError)

	// Call the method
	registrations, pagination, err := suite.service.FindRegistrationByStudent(ctx, pagReq, filter, token, nil)

	// Assertions
	suite.Error(err)
	suite.Equal(expectedError, err)
	suite.Empty(registrations)
	suite.Equal(dto.PaginationResponse{}, pagination)

	suite.mockUserManagementService.AssertExpectations(suite.T())
	suite.mockRegistrationRepo.AssertExpectations(suite.T())
}

func (suite *RegistrationServiceTestSuite) TestFindRegistrationByStudentEquivalentsError() {
	// Setup test data
	ctx := context.Background()
	token := "Bearer test-token"
	pagReq := dto.PaginationRequest{
		Limit:  10,
		Offset: 0,
		URL:    "http://example.com/registrations",
	}
	filter := dto.FilterRegistrationRequest{}

	// Mock ValidateStudent
	userData := map[string]interface{}{
		"id":    "user123",
		"nrp":   "5022123456",
		"name":  "Test Student",
		"role":  "MAHASISWA",
		"email": "student@example.com",
	}

	suite.mockUserManagementService.On("GetUserData", "GET", token).Return(userData)

	// Create mock registrations
	now := time.Now()
	mockRegistration := entity.Registration{
		ID:                        uuid.New(),
		ActivityID:                "activity1",
		ActivityName:              "Test Activity",
		UserID:                    "user123",
		UserNRP:                   "5022123456",
		UserName:                  "Test Student",
		AdvisingConfirmation:      true,
		AcademicAdvisorID:         "advisor1",
		AcademicAdvisor:           "Test Advisor",
		AcademicAdvisorEmail:      "advisor@example.com",
		MentorName:                "Test Mentor",
		MentorEmail:               "mentor@example.com",
		LOValidation:              "PENDING",
		AcademicAdvisorValidation: "PENDING",
		Semester:                  1,
		TotalSKS:                  20,
		ApprovalStatus:            false,
		BaseModel: entity.BaseModel{
			CreatedAt: &now,
			UpdatedAt: &now,
		},
	}

	// Set up expected filter with NRP
	expectedFilter := filter
	expectedFilter.UserNRP = "5022123456"

	// Mock repository behavior
	suite.mockRegistrationRepo.On("Index", ctx, mock.Anything, pagReq, expectedFilter).Return(
		[]entity.Registration{mockRegistration}, int64(1), nil)

	// Mock error from matching service
	expectedError := errors.New("matching service error")
	suite.mockMatchingManagementService.On("GetEquivalentsByRegistrationID",
		mockRegistration.ID.String(), "GET", token).Return(nil, expectedError)

	// Call the method
	registrations, pagination, err := suite.service.FindRegistrationByStudent(ctx, pagReq, filter, token, nil)

	// Assertions
	suite.Error(err)
	suite.Equal(expectedError, err)
	suite.Empty(registrations)
	suite.Equal(dto.PaginationResponse{}, pagination)

	suite.mockUserManagementService.AssertExpectations(suite.T())
	suite.mockRegistrationRepo.AssertExpectations(suite.T())
	suite.mockMatchingManagementService.AssertExpectations(suite.T())
}

func (suite *RegistrationServiceTestSuite) TestFindAllRegistrationsSuccess() {
	// Setup test data
	ctx := context.Background()
	token := "Bearer test-token"
	pagReq := dto.PaginationRequest{
		Limit:  10,
		Offset: 0,
		URL:    "http://example.com/registrations",
	}
	filter := dto.FilterRegistrationRequest{
		ActivityName: "Test Activity",
	}

	// Create mock registrations
	now := time.Now()
	regID := uuid.New()
	mockRegistration := entity.Registration{
		ID:                        regID,
		ActivityID:                "activity1",
		ActivityName:              "Test Activity",
		UserID:                    "user123",
		UserNRP:                   "5022123456",
		UserName:                  "Test Student",
		AdvisingConfirmation:      true,
		AcademicAdvisorID:         "advisor1",
		AcademicAdvisor:           "Test Advisor",
		AcademicAdvisorEmail:      "advisor@example.com",
		MentorName:                "Test Mentor",
		MentorEmail:               "mentor@example.com",
		LOValidation:              "PENDING",
		AcademicAdvisorValidation: "PENDING",
		Semester:                  1,
		TotalSKS:                  20,
		ApprovalStatus:            false,
		Document: []entity.Document{
			{
				ID:             uuid.New(),
				RegistrationID: regID.String(),
				FileStorageID:  "file1",
				Name:           "document1.pdf",
				DocumentType:   "Acceptence Letter",
			},
		},
		BaseModel: entity.BaseModel{
			CreatedAt: &now,
			UpdatedAt: &now,
		},
	}

	// Mock repository behavior
	suite.mockRegistrationRepo.On("Index", ctx, mock.Anything, pagReq, filter).Return(
		[]entity.Registration{mockRegistration}, int64(1), nil)

	// Call the method
	registrations, pagination, err := suite.service.FindAllRegistrations(ctx, pagReq, filter, nil, token)

	// Assertions
	suite.NoError(err)
	suite.Equal(1, len(registrations))
	suite.Equal(mockRegistration.ID.String(), registrations[0].ID)
	suite.Equal(mockRegistration.ActivityName, registrations[0].ActivityName)
	suite.Equal(mockRegistration.UserNRP, registrations[0].UserNRP)
	suite.Equal(int64(1), pagination.Total)

	// Document assertions
	suite.Equal(1, len(registrations[0].Documents))
	suite.Equal(mockRegistration.Document[0].ID.String(), registrations[0].Documents[0].ID)
	suite.Equal(mockRegistration.Document[0].Name, registrations[0].Documents[0].Name)

	suite.mockRegistrationRepo.AssertExpectations(suite.T())
}

func (suite *RegistrationServiceTestSuite) TestFindAllRegistrationsRepositoryError() {
	// Setup test data
	ctx := context.Background()
	token := "Bearer test-token"
	pagReq := dto.PaginationRequest{
		Limit:  10,
		Offset: 0,
	}
	filter := dto.FilterRegistrationRequest{}

	// Mock repository error
	expectedError := errors.New("database error")
	suite.mockRegistrationRepo.On("Index", ctx, mock.Anything, pagReq, filter).Return(
		[]entity.Registration{}, int64(0), expectedError)

	// Call the method
	registrations, pagination, err := suite.service.FindAllRegistrations(ctx, pagReq, filter, nil, token)

	// Assertions
	suite.Error(err)
	suite.Equal(expectedError, err)
	suite.Empty(registrations)
	suite.Equal(dto.PaginationResponse{}, pagination)

	suite.mockRegistrationRepo.AssertExpectations(suite.T())
}

func (suite *RegistrationServiceTestSuite) TestFindAllRegistrationsEmptyResults() {
	// Setup test data
	ctx := context.Background()
	token := "Bearer test-token"
	pagReq := dto.PaginationRequest{
		Limit:  10,
		Offset: 0,
		URL:    "http://example.com/registrations",
	}
	filter := dto.FilterRegistrationRequest{
		ActivityName: "Non-existent Activity",
	}

	// Mock repository behavior - empty results
	suite.mockRegistrationRepo.On("Index", ctx, mock.Anything, pagReq, filter).Return(
		[]entity.Registration{}, int64(0), nil)

	// Call the method
	registrations, pagination, err := suite.service.FindAllRegistrations(ctx, pagReq, filter, nil, token)

	// Assertions
	suite.NoError(err)
	suite.Empty(registrations)
	suite.Equal(int64(0), pagination.Total)

	suite.mockRegistrationRepo.AssertExpectations(suite.T())
}

func (suite *RegistrationServiceTestSuite) TestGetRegistrationTranscriptSuccess() {
	// Setup test data
	ctx := context.Background()
	token := "Bearer test-token"
	registrationID := uuid.New().String()

	// Mock RegistrationsDataAccess
	userData := map[string]interface{}{
		"id":    "user123",
		"nrp":   "5022123456",
		"name":  "Test Student",
		"role":  "MAHASISWA",
		"email": "student@example.com",
	}

	suite.mockUserManagementService.On("GetUserData", "GET", token).Return(userData)

	// Create mock registration with approval status = true
	now := time.Now()
	mockRegistration := entity.Registration{
		ID:                        uuid.MustParse(registrationID),
		ActivityID:                "activity1",
		ActivityName:              "Test Activity",
		UserID:                    "user123",
		UserNRP:                   "5022123456",
		UserName:                  "Test Student",
		AdvisingConfirmation:      true,
		AcademicAdvisorID:         "advisor1",
		AcademicAdvisor:           "Test Advisor",
		AcademicAdvisorEmail:      "advisor@example.com",
		MentorName:                "Test Mentor",
		MentorEmail:               "mentor@example.com",
		LOValidation:              "APPROVED",
		AcademicAdvisorValidation: "APPROVED",
		Semester:                  1,
		TotalSKS:                  20,
		ApprovalStatus:            true, // Registration is approved
		BaseModel: entity.BaseModel{
			CreatedAt: &now,
			UpdatedAt: &now,
		},
	}

	// Mock repository behavior
	suite.mockRegistrationRepo.On("FindByID", ctx, registrationID, mock.Anything).Return(mockRegistration, nil)

	// Mock transcript data
	transcriptData := map[string]interface{}{
		"scores": []map[string]interface{}{
			{
				"course_name": "Course 1",
				"grade":       "A",
				"score":       90,
			},
			{
				"course_name": "Course 2",
				"grade":       "B+",
				"score":       85,
			},
		},
	}

	suite.mockMonitoringManagementService.On("GetTranscriptByRegistrationID", registrationID, token).Return(transcriptData, nil)

	// Call the method
	transcript, err := suite.service.GetRegistrationTranscript(ctx, registrationID, token, nil)

	// Assertions
	suite.NoError(err)
	suite.Equal(registrationID, transcript.RegistrationID)
	suite.Equal(mockRegistration.UserID, transcript.UserID)
	suite.Equal(mockRegistration.UserNRP, transcript.UserNRP)
	suite.Equal(mockRegistration.UserName, transcript.UserName)
	suite.Equal(mockRegistration.ActivityName, transcript.ActivityName)
	suite.Equal(mockRegistration.Semester, transcript.Semester)
	suite.Equal(mockRegistration.TotalSKS, transcript.TotalSKS)
	suite.Equal(mockRegistration.ApprovalStatus, transcript.ApprovalStatus)
	suite.Equal(transcriptData, transcript.TranscriptData)

	suite.mockUserManagementService.AssertExpectations(suite.T())
	suite.mockRegistrationRepo.AssertExpectations(suite.T())
	suite.mockMonitoringManagementService.AssertExpectations(suite.T())
}

func (suite *RegistrationServiceTestSuite) TestGetStudentRegistrationsWithTranscriptsSuccess() {
	// Setup test data
	ctx := context.Background()
	token := "Bearer test-token"
	pagReq := dto.PaginationRequest{
		Limit:  10,
		Offset: 0,
		URL:    "http://example.com/registrations",
	}
	filter := dto.FilterRegistrationRequest{}

	// Mock ValidateStudent
	userData := map[string]interface{}{
		"id":    "user123",
		"nrp":   "5022123456",
		"name":  "Test Student",
		"role":  "MAHASISWA",
		"email": "student@example.com",
	}

	suite.mockUserManagementService.On("GetUserData", "GET", token).Return(userData)

	// Create mock registrations with approval status = true
	now := time.Now()
	mockRegistration := entity.Registration{
		ID:                        uuid.New(),
		ActivityID:                "activity1",
		ActivityName:              "Test Activity",
		UserID:                    "user123",
		UserNRP:                   "5022123456",
		UserName:                  "Test Student",
		AdvisingConfirmation:      true,
		AcademicAdvisorID:         "advisor1",
		AcademicAdvisor:           "Test Advisor",
		AcademicAdvisorEmail:      "advisor@example.com",
		MentorName:                "Test Mentor",
		MentorEmail:               "mentor@example.com",
		LOValidation:              "APPROVED",
		AcademicAdvisorValidation: "APPROVED",
		Semester:                  1,
		TotalSKS:                  20,
		ApprovalStatus:            true, // Registration is approved
		BaseModel: entity.BaseModel{
			CreatedAt: &now,
			UpdatedAt: &now,
		},
	}

	// Set up expected filter with NRP
	expectedFilter := filter
	expectedFilter.UserNRP = "5022123456"

	// Mock repository behavior
	suite.mockRegistrationRepo.On("Index", ctx, mock.Anything, pagReq, expectedFilter).Return(
		[]entity.Registration{mockRegistration}, int64(1), nil)

	// Mock equivalents data
	equivalentsData := map[string]interface{}{
		"id": "equiv1",
		"courses": []map[string]interface{}{
			{
				"id":   "course1",
				"name": "Course 1",
			},
		},
	}

	// Mock matching data
	matchingData := map[string]interface{}{
		"id":          "matching1",
		"activity_id": "activity1",
		"courses": []map[string]interface{}{
			{
				"id":   "external1",
				"name": "External Course 1",
			},
		},
	}

	suite.mockMatchingManagementService.On("GetEquivalentsByRegistrationID",
		mockRegistration.ID.String(), "GET", token).Return(equivalentsData, nil)
	suite.mockMatchingManagementService.On("GetMatchingByActivityID",
		mockRegistration.ActivityID, "GET", token).Return(matchingData, nil)

	// Call the method
	response, pagination, err := suite.service.FindRegistrationsWithMatching(ctx, pagReq, filter, token, nil)

	// Assertions
	suite.NoError(err)
	suite.Equal(userData["id"].(string), response.UserID)
	suite.Equal(userData["nrp"].(string), response.UserNRP)
	suite.Equal(userData["name"].(string), response.UserName)
	suite.Equal(1, len(response.Registrations))
	suite.Equal(mockRegistration.ID.String(), response.Registrations[0].ID)
	suite.Equal(mockRegistration.ActivityName, response.Registrations[0].ActivityName)
	suite.Equal(mockRegistration.UserName, response.Registrations[0].UserName)
	suite.Equal(equivalentsData, response.Registrations[0].Equivalents)
	suite.Equal(matchingData, response.Registrations[0].Matching)
	suite.Equal(int64(1), pagination.Total)

	suite.mockUserManagementService.AssertExpectations(suite.T())
	suite.mockRegistrationRepo.AssertExpectations(suite.T())
	suite.mockMatchingManagementService.AssertExpectations(suite.T())
}
