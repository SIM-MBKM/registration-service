package service

import (
	"errors"
	"strings"

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
	// split token
	tokenParts := strings.Split(token, " ")
	if len(tokenParts) != 2 {
		return nil, errors.New("invalid token")
	}

	token = tokenParts[1]

	res, err := s.baseService.Request(method, "matching-management/api/v1/matching/activity/"+activityID, nil, token)
	if err != nil {
		return nil, err
	}
	return res["data"], nil
}

func (s *MatchingManagementService) GetEquivalentsByRegistrationID(registrationID string, method string, token string) (interface{}, error) {
	// split token
	tokenParts := strings.Split(token, " ")
	if len(tokenParts) != 2 {
		return nil, errors.New("invalid token")
	}

	token = tokenParts[1]

	res, err := s.baseService.Request(method, "matching-management/api/v1/equivalent/registration/"+registrationID+"?noRecursion=1", nil, token)
	if err != nil {
		if err.Error() != "404 Not Found" {
			return nil, err
		}

		return nil, nil
	}

	return res["data"], nil
}
