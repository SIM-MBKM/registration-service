package service

import (
	baseService "github.com/SIM-MBKM/mod-service/src/service"
)

type ActivityManagementService struct {
	baseService *baseService.Service
}

func NewActivityManagementService(baseURI string, asyncURIs []string) *ActivityManagementService {
	return &ActivityManagementService{
		baseService: baseService.NewService(baseURI, asyncURIs),
	}
}
