package service

import (
	"log"

	baseService "github.com/SIM-MBKM/mod-service/src/service"
)

type ActivityManagementService struct {
	baseService *baseService.Service
}

const (
	GET_ACTIVITIY_FILTER_ENDPOINT = "activity-management/api/v1/activity/filter"
)

func NewActivityManagementService(baseURI string, asyncURIs []string) *ActivityManagementService {
	return &ActivityManagementService{
		baseService: baseService.NewService(baseURI, asyncURIs),
	}
}

func (s *ActivityManagementService) GetActivitiesData(data map[string]interface{}, method string, token string) []map[string]interface{} {
	res, err := s.baseService.Request(method, GET_ACTIVITIY_FILTER_ENDPOINT, data, token)

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
			"id":              activity["id"],
			"name":            activity["name"],
			"start_period":    activity["start_period"],
			"months_duration": activity["months_duration"],
			"approval_status": activity["approval_status"],
		})
	}

	return activitiesData
}
