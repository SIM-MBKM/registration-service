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
	GetRegistrationTranscript(ctx context.Context, id string, token string, tx *gorm.DB) (dto.TranscriptResponse, error)
	GetStudentRegistrationsWithTranscripts(ctx context.Context, pagReq dto.PaginationRequest, filter dto.FilterRegistrationRequest, token string, tx *gorm.DB) (dto.StudentTranscriptsResponse, dto.PaginationResponse, error)
	GetStudentRegistrationsWithSyllabuses(ctx context.Context, pagReq dto.PaginationRequest, filter dto.FilterRegistrationRequest, token string, tx *gorm.DB) (dto.StudentSyllabusesResponse, dto.PaginationResponse, error)
	FindRegistrationsWithMatching(ctx context.Context, pagReq dto.PaginationRequest, filter dto.FilterRegistrationRequest, token string, tx *gorm.DB) (dto.StudentRegistrationsWithMatchingResponse, dto.PaginationResponse, error)
	CheckRegistrationEligibility(ctx context.Context, activityID string, token string, tx *gorm.DB) (dto.RegistrationEligibilityResponse, error)
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

func (s *registrationService) createReportSchedules(ctx context.Context, registration entity.Registration, token string) error {
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

func (s *registrationService) LORegistrationApproval(ctx context.Context, token string, approval dto.ApprovalRequest, tx *gorm.DB) error {
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
			if registration.AcademicAdvisorValidation == "REJECTED" {
				registration.ApprovalStatus = false
			}
		}

		err = s.registrationRepository.Update(ctx, id, registration, tx)
		if err != nil {
			return err
		}

	}

	return nil
}

func (s *registrationService) AdvisorRegistrationApproval(ctx context.Context, token string, approval dto.ApprovalRequest, tx *gorm.DB) error {
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
			if registration.LOValidation == "REJECTED" {
				registration.ApprovalStatus = false
			}
		}

		err = s.registrationRepository.Update(ctx, id, registration, tx)
		if err != nil {
			return err
		}

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

func (s *registrationService) GetRegistrationTranscript(ctx context.Context, id string, token string, tx *gorm.DB) (dto.TranscriptResponse, error) {
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

func (s *registrationService) GetStudentRegistrationsWithTranscripts(ctx context.Context, pagReq dto.PaginationRequest, filter dto.FilterRegistrationRequest, token string, tx *gorm.DB) (dto.StudentTranscriptsResponse, dto.PaginationResponse, error) {
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

func (s *registrationService) GetStudentRegistrationsWithSyllabuses(ctx context.Context, pagReq dto.PaginationRequest, filter dto.FilterRegistrationRequest, token string, tx *gorm.DB) (dto.StudentSyllabusesResponse, dto.PaginationResponse, error) {
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

func (s *registrationService) FindRegistrationsWithMatching(ctx context.Context, pagReq dto.PaginationRequest, filter dto.FilterRegistrationRequest, token string, tx *gorm.DB) (dto.StudentRegistrationsWithMatchingResponse, dto.PaginationResponse, error) {
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

func (s *registrationService) CheckRegistrationEligibility(ctx context.Context, activityID string, token string, tx *gorm.DB) (dto.RegistrationEligibilityResponse, error) {
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
