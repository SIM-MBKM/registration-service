package service

import (
	"errors"
	"log"

	baseService "github.com/SIM-MBKM/mod-service/src/service"
)

type MonitoringManagementService struct {
	baseService *baseService.Service
}

const (
	// MatchingManagementServiceBaseURI is the base URI for the matching management service
	CREATE_REPORT_SCHEDULE = "monitoring-service/api/v1/report-schedules/"
)

func NewMonitoringManagementService(baseURI string, asyncURIs []string) *MonitoringManagementService {
	return &MonitoringManagementService{
		baseService: baseService.NewService(baseURI, asyncURIs),
	}
}

func (s *MonitoringManagementService) CreateReportSchedule(data map[string]interface{}, method string, token string) error {
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
