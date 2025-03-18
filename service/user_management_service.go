package service

import (
	baseService "github.com/SIM-MBKM/mod-service/src/service"
)

type UserManagementService struct {
	baseService *baseService.Service
}

func NewUserManagementService(baseURI string, asyncURIs []string) *UserManagementService {
	return &UserManagementService{
		baseService: baseService.NewService(baseURI, asyncURIs),
	}
}
