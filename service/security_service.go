package service

import (
	baseServiceHelpers "github.com/SIM-MBKM/mod-service/src/helpers"
)

type SecurityService struct {
	baseServiceHelper *baseServiceHelpers.Security
}

func NewSecurityService(hashMethod string, key string, cipherMode string) *SecurityService {
	return &SecurityService{
		baseServiceHelper: baseServiceHelpers.NewSecurity(hashMethod, key, cipherMode),
	}
}
