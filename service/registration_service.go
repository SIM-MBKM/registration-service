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
	"time"

	storageService "github.com/SIM-MBKM/filestorage/storage"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type registrationService struct {
	registrationRepository      repository.RegistrationRepository
	documentRepository          repository.DocumentRepository
	userManagementService       *UserManagementService
	activityManagementService   *ActivityManagementService
	fileService                 *FileService
	matchingManagementService   *MatchingManagementService
	monitoringManagementService *MonitoringManagementService
}

type RegistrationService interface {
	FindAllRegistrations(ctx context.Context, pagReq dto.PaginationRequest, filter dto.FilterRegistrationRequest, tx *gorm.DB, token string) ([]dto.GetRegistrationResponse, dto.PaginationResponse, error)
	FindRegistrationByID(ctx context.Context, id string, token string, tx *gorm.DB) (dto.GetRegistrationResponse, error)
	CreateRegistration(ctx context.Context, registration dto.CreateRegistrationRequest, file *multipart.FileHeader, geoletter *multipart.FileHeader, tx *gorm.DB, token string) error
	UpdateRegistration(ctx context.Context, id string, registration dto.UpdateRegistrationDataRequest, token string, tx *gorm.DB) error
	DeleteRegistration(ctx context.Context, id string, token string, tx *gorm.DB) error
	RegistrationsDataAccess(ctx context.Context, id string, token string, tx *gorm.DB) bool
	FindRegistrationByAdvisor(ctx context.Context, pagReq dto.PaginationRequest, filter dto.FilterRegistrationRequest, token string, tx *gorm.DB) ([]dto.GetRegistrationResponse, dto.PaginationResponse, error)
	FindRegistrationByStudent(ctx context.Context, pagReq dto.PaginationRequest, filter dto.FilterRegistrationRequest, token string, tx *gorm.DB) ([]dto.GetRegistrationResponse, dto.PaginationResponse, error)
	ValidateAdvisor(ctx context.Context, token string, tx *gorm.DB) string
	ValidateStudent(ctx context.Context, token string, tx *gorm.DB) string
	AdvisorRegistrationApproval(ctx context.Context, token string, approval dto.ApprovalRequest, tx *gorm.DB) error
	LORegistrationApproval(ctx context.Context, token string, approval dto.ApprovalRequest, tx *gorm.DB) error
}

func NewRegistrationService(registrationRepository repository.RegistrationRepository, documentRepository repository.DocumentRepository, secretKey string, userManagementbaseURI string, activityManagementbaseURI string, matchingManagementbaseURI string, monitoringManagementbaseURI string, asyncURIs []string, config *storageService.Config, tokenManager *storageService.CacheTokenManager) RegistrationService {
	return &registrationService{
		registrationRepository:      registrationRepository,
		documentRepository:          documentRepository,
		userManagementService:       NewUserManagementService(userManagementbaseURI, asyncURIs),
		activityManagementService:   NewActivityManagementService(activityManagementbaseURI, asyncURIs),
		matchingManagementService:   NewMatchingManagementService(matchingManagementbaseURI, asyncURIs),
		monitoringManagementService: NewMonitoringManagementService(monitoringManagementbaseURI, asyncURIs),
		fileService:                 NewFileService(config, tokenManager),
	}
}

func (s *registrationService) LORegistrationApproval(ctx context.Context, token string, approval dto.ApprovalRequest, tx *gorm.DB) error {
	registration, err := s.registrationRepository.FindByID(ctx, approval.ID, tx)

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
		}
	}
	if approval.Status == "REJECTED" {
		registration.LOValidation = "REJECTED"
		if registration.AcademicAdvisorValidation == "REJECTED" {
			registration.ApprovalStatus = false
		}
	}

	if registration.ApprovalStatus == true {
		// get activity data
		activityData := s.activityManagementService.GetActivitiesData(map[string]interface{}{
			"activity_id":     registration.ActivityID,
			"program_type_id": "",
			"level_id":        "",
			"group_id":        "",
			"name":            "",
		}, "POST", token)

		userData := s.userManagementService.GetUserData("GET", token)
		if userData["id"] != registration.UserID {
			return errors.New("Unauthorized")
		}

		// calculate how many times should upload the report schedule and week based on months_duration and start_period
		monthsDuration := int(activityData[0]["months_duration"].(float64))

		startPeriod := activityData[0]["start_period"].(string)

		// convert start_period to time
		startPeriodTime, err := time.Parse(time.RFC3339, startPeriod)
		if err != nil {
			log.Println("ERROR PARSE START PERIOD", err)
			return err
		}

		for i := 0; i < monthsDuration; i++ {
			// create report schedule
			// start_date
			startDate := startPeriodTime.AddDate(0, 0, i*7)
			endDate := startPeriodTime.AddDate(0, monthsDuration, 0)
			week := i + 1

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

		// create final report
		err = s.monitoringManagementService.CreateReportSchedule(map[string]interface{}{
			"registration_id":        registration.ID.String(),
			"user_id":                registration.UserID,
			"user_nrp":               registration.UserNRP,
			"academic_advisor_id":    registration.AcademicAdvisorID,
			"academic_advisor_email": registration.AcademicAdvisorEmail,
			"report_type":            "FINAL_REPORT",
			"week":                   monthsDuration,
			"start_date":             startPeriodTime,
			"end_date":               startPeriodTime.AddDate(0, monthsDuration, 7),
		}, "POST", token)

		err = s.registrationRepository.Update(ctx, approval.ID, registration, tx)

		if err != nil {
			return err
		}
	}

	err = s.registrationRepository.Update(ctx, approval.ID, registration, tx)

	if err != nil {
		return err
	}

	return nil
}

func (s *registrationService) AdvisorRegistrationApproval(ctx context.Context, token string, approval dto.ApprovalRequest, tx *gorm.DB) error {
	userEmail := s.ValidateAdvisor(ctx, token, tx)
	if userEmail == "" {
		return errors.New("Unauthorized")
	}

	registration, err := s.registrationRepository.FindByID(ctx, approval.ID, tx)

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
		}
	}
	if approval.Status == "REJECTED" {
		registration.AcademicAdvisorValidation = "REJECTED"
		if registration.LOValidation == "REJECTED" {
			registration.ApprovalStatus = false
		}
	}

	if registration.ApprovalStatus == true {
		// get activity data
		activityData := s.activityManagementService.GetActivitiesData(map[string]interface{}{
			"activity_id":     registration.ActivityID,
			"program_type_id": "",
			"level_id":        "",
			"group_id":        "",
			"name":            "",
		}, "POST", token)

		userData := s.userManagementService.GetUserData("GET", token)
		if userData["id"] != registration.UserID {
			return errors.New("Unauthorized")
		}

		// calculate how many times should upload the report schedule and week based on months_duration and start_period
		monthsDuration := int(activityData[0]["months_duration"].(float64))

		startPeriod := activityData[0]["start_period"].(string)

		// convert start_period to time
		startPeriodTime, err := time.Parse(time.RFC3339, startPeriod)
		if err != nil {
			log.Println("ERROR PARSE START PERIOD", err)
			return err
		}

		for i := 0; i < monthsDuration; i++ {
			// create report schedule
			// start_date
			startDate := startPeriodTime.AddDate(0, 0, i*7)
			endDate := startPeriodTime.AddDate(0, monthsDuration, 0)
			week := i + 1

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

		// create final report
		err = s.monitoringManagementService.CreateReportSchedule(map[string]interface{}{
			"registration_id":        registration.ID.String(),
			"user_id":                registration.UserID,
			"user_nrp":               registration.UserNRP,
			"academic_advisor_id":    registration.AcademicAdvisorID,
			"academic_advisor_email": registration.AcademicAdvisorEmail,
			"report_type":            "FINAL_REPORT",
			"week":                   monthsDuration,
			"start_date":             startPeriodTime,
			"end_date":               startPeriodTime.AddDate(0, monthsDuration, 7),
		}, "POST", token)

		err = s.registrationRepository.Update(ctx, approval.ID, registration, tx)

		if err != nil {
			return err
		}
	}

	if err != nil {
		return err
	}

	return nil
}

func (s *registrationService) FindRegistrationByStudent(ctx context.Context, pagReq dto.PaginationRequest, filter dto.FilterRegistrationRequest, token string, tx *gorm.DB) ([]dto.GetRegistrationResponse, dto.PaginationResponse, error) {
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

func (s *registrationService) FindRegistrationByAdvisor(ctx context.Context, pagReq dto.PaginationRequest, filter dto.FilterRegistrationRequest, token string, tx *gorm.DB) ([]dto.GetRegistrationResponse, dto.PaginationResponse, error) {
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

func (s *registrationService) ValidateStudent(ctx context.Context, token string, tx *gorm.DB) string {
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

func (s *registrationService) ValidateAdvisor(ctx context.Context, token string, tx *gorm.DB) string {
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

func (s *registrationService) RegistrationsDataAccess(ctx context.Context, id string, token string, tx *gorm.DB) bool {
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
			ApprovalStatus:            registration.ApprovalStatus,
			Documents:                 convertToDocumentResponse(registration.Document),
		})
	}

	return response, metaData, nil
}

func (s *registrationService) FindRegistrationByID(ctx context.Context, id string, token string, tx *gorm.DB) (dto.GetRegistrationResponse, error) {
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

func (s *registrationService) CreateRegistration(ctx context.Context, registration dto.CreateRegistrationRequest, file *multipart.FileHeader, geoletter *multipart.FileHeader, tx *gorm.DB, token string) error {
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
		return errors.New("data not found")
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
	} else if registrationsByNRP.ActivityID != registration.ActivityID {
		// get activity data by registrationsByNRP.ActivityID
		if registrationsByNRP.ActivityID != "" {
			activitiesDataOld = s.activityManagementService.GetActivitiesData(map[string]interface{}{
				"activity_id":     registrationsByNRP.ActivityID,
				"program_type_id": "",
				"level_id":        "",
				"group_id":        "",
				"name":            "",
			}, "POST", token)
		}

		activityOldStartDate := activitiesDataOld[0]["start_date"].(time.Time)
		activityOldMonthsDuration := activitiesDataOld[0]["months_duration"].(int)
		activityOldEndDate := activityOldStartDate.AddDate(0, activityOldMonthsDuration, 0)

		// if activity new start date is after activity old end date then user can register
		if !activitiesData[0]["start_date"].(time.Time).After(activityOldEndDate) {
			return errors.New("user already registered")
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

func (s *registrationService) UpdateRegistration(ctx context.Context, id string, registration dto.UpdateRegistrationDataRequest, token string, tx *gorm.DB) error {
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

func (s *registrationService) DeleteRegistration(ctx context.Context, id string, token string, tx *gorm.DB) error {
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
		_, err = s.fileService.storage.GcsDelete(document.FileStorageID, "sim_mbkm", "")
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
