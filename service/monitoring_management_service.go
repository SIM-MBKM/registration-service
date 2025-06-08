package service

import (
	"errors"
	"fmt"
	"log"
	"strings"

	baseService "github.com/SIM-MBKM/mod-service/src/service"
)

type MonitoringManagementService struct {
	baseService *baseService.Service
}

const (
	// MatchingManagementServiceBaseURI is the base URI for the matching management service
	CREATE_REPORT_SCHEDULE                  = "monitoring-service/api/v1/report-schedules/"
	GET_REPORT_SCHEDULES_BY_REGISTRATION_ID = "monitoring-service/api/v1/report-schedules/registrations/"
	GET_TRANSCRIPT_BY_REGISTRATION_ID       = "monitoring-service/api/v1/transcripts/registrations/"
	GET_SYLLABUS_BY_REGISTRATION_ID         = "monitoring-service/api/v1/syllabuses/registrations/"
)

func NewMonitoringManagementService(baseURI string, asyncURIs []string) *MonitoringManagementService {
	return &MonitoringManagementService{
		baseService: baseService.NewService(baseURI, asyncURIs),
	}
}

func (s *MonitoringManagementService) GetReportSchedulesByRegistrationID(registrationID string, token string) (interface{}, error) {
	// split token
	tokenParts := strings.Split(token, " ")
	if len(tokenParts) != 2 {
		return nil, errors.New("invalid token")
	}

	token = tokenParts[1]

	endpoint := fmt.Sprintf("%s%s/report-schedules", GET_REPORT_SCHEDULES_BY_REGISTRATION_ID, registrationID)
	res, err := s.baseService.Request("GET", endpoint, nil, token)
	log.Println("GetReportSchedulesByRegistrationID", res)
	if err != nil {
		log.Println("Error in GetReportSchedulesByRegistrationID:", err)
		return nil, err
	}
	result, ok := res["data"].(interface{})
	if !ok {
		log.Println("Failed to convert response to map in GetReportSchedulesByRegistrationID")
		return nil, errors.New("failed to convert to map")
	}

	return result, nil
}

func (s *MonitoringManagementService) GetSyllabusByRegistrationID(registrationID string, token string) (interface{}, error) {
	// split token
	tokenParts := strings.Split(token, " ")
	if len(tokenParts) != 2 {
		return nil, errors.New("invalid token")
	}

	token = tokenParts[1]

	endpoint := fmt.Sprintf("%s%s", GET_SYLLABUS_BY_REGISTRATION_ID, registrationID)
	res, err := s.baseService.Request("GET", endpoint, nil, token)
	if err != nil {
		return nil, err
	}
	result, ok := res["data"].(interface{})
	if !ok {
		return nil, errors.New("failed to convert to map")
	}

	return result, nil
}

func (s *MonitoringManagementService) GetTranscriptByRegistrationID(registrationID string, token string) (interface{}, error) {

	// split token
	tokenParts := strings.Split(token, " ")
	if len(tokenParts) != 2 {
		return nil, errors.New("invalid token")
	}

	token = tokenParts[1]

	endpoint := fmt.Sprintf("%s%s", GET_TRANSCRIPT_BY_REGISTRATION_ID, registrationID)
	res, err := s.baseService.Request("GET", endpoint, nil, token)
	if err != nil {
		return nil, err
	}
	result, ok := res["data"].(interface{})
	if !ok {
		return nil, errors.New("failed to convert to map")
	}

	return result, nil
}

func (s *MonitoringManagementService) CreateReportSchedule(data map[string]interface{}, method string, token string) error {
	// split token
	tokenParts := strings.Split(token, " ")
	if len(tokenParts) != 2 {
		return nil
	}

	token = tokenParts[1]

	res, err := s.baseService.Request("POST", CREATE_REPORT_SCHEDULE, data, token)
	if err != nil {
		return nil
	}
	resInterface, ok := res["status"].([]interface{})
	if !ok {
		log.Println("FAILED TO CONVERT TO INTERFACE")
		return errors.New("failed to convert to interface")
	}

	resString, ok := resInterface[0].(string)
	if !ok {
		log.Println("FAILED TO CONVERT TO STRING")
		return errors.New("failed to convert to string")
	}

	if resString == "success" {
		log.Println("SUCCESS TO CREATE REPORT SCHEDULE")
		return nil
	}

	return errors.New("failed to create report schedule")
}
