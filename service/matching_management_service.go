package service

import (
	baseService "github.com/SIM-MBKM/mod-service/src/service"
)

type MatchingManagementService struct {
	baseService *baseService.Service
}

const (
	// MatchingManagementServiceBaseURI is the base URI for the matching management service
	GET_MATCHING_BY_ACTIVITY_ID = "matching-management/api/v1/matching/activity/"
)

func NewMatchingManagementService(baseURI string, asyncURIs []string) *MatchingManagementService {
	return &MatchingManagementService{
		baseService: baseService.NewService(baseURI, asyncURIs),
	}
}

func (s *MatchingManagementService) GetMatchingByActivityID(activityID string, method string, token string) (interface{}, error) {
	res, err := s.baseService.Request(method, "matching-management/api/matching/activity/"+activityID, nil, token)
	if err != nil {
		return nil, err
	}
	return res["data"], nil
}

func (s *MatchingManagementService) GetMatchingByRegistrationID(registrationID string, method string, token string) (interface{}, error) {
	res, err := s.baseService.Request(method, "matching-management/api/equivalent/registration/"+registrationID+"?noRecursion=1", nil, token)
	if err != nil {
		return nil, err
	}

	return res["data"], nil
}
