package service

import (
	"errors"
	"strings"

	baseService "github.com/SIM-MBKM/mod-service/src/service"
)

type BrokerService struct {
	baseService *baseService.Service
}

const (
	SEND_NOTIFICATION = "broker-service/api/v1/send-notification"
)

func NewBrokerService(baseURI string, asyncURIs []string) *BrokerService {
	return &BrokerService{
		baseService: baseService.NewService(baseURI, asyncURIs),
	}
}

func (s *BrokerService) SendNotification(data map[string]interface{}, method string, token string) error {
	// split token
	tokenParts := strings.Split(token, " ")
	if len(tokenParts) != 2 {
		return nil
	}

	token = tokenParts[1]

	res, err := s.baseService.Request(method, SEND_NOTIFICATION, data, token)
	if err != nil {
		return err
	}

	res, ok := res["data"].(map[string]interface{})
	if !ok {
		return errors.New("invalid response")
	}

	status, _ := res["status"].(string)
	if status != "success" {
		return errors.New("failed to send notification")
	}

	return nil
}
