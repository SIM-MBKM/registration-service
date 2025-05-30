package service

import (
	"log"
	"strings"

	baseService "github.com/SIM-MBKM/mod-service/src/service"
)

type UserManagementService struct {
	baseService *baseService.Service
}

const (
	GET_USER_BY_ID_ENDPOINT          = "api/v1/user/service/by-user-id"
	GET_USER_BY_FILTER_ENDPOINT      = "api/v1/user/service/users-all"
	GET_USER_ROLE_ENDPOINT           = "api/v1/user/service/users/me/role"
	GET_USER_DATA_ENDPOINT           = "api/v1/user/service/users/me"
	GET_DOSEN_DATA_BY_EMAIL_ENDPOINT = "api/v1/user/service/by-email/"
)

func NewUserManagementService(baseURI string, asyncURIs []string) *UserManagementService {
	return &UserManagementService{
		baseService: baseService.NewService(baseURI, asyncURIs),
	}
}

// create function to get user by id
func (s *UserManagementService) GetUserData(method string, token string) map[string]interface{} {
	// split token
	tokenParts := strings.Split(token, " ")
	if len(tokenParts) != 2 {
		return nil
	}

	token = tokenParts[1]

	res, err := s.baseService.Request(method, GET_USER_DATA_ENDPOINT, nil, token)
	if err != nil {
		return nil
	}

	users, ok := res["data"].(map[string]interface{})
	if !ok {
		return nil
	}

	var usersData map[string]interface{}
	usersData = map[string]interface{}{
		"id":    users["id"],
		"nrp":   users["nrp"],
		"name":  users["name"],
		"role":  users["role"],
		"email": users["email"],
	}
	return usersData
}

func (s *UserManagementService) GetUserByFilter(data map[string]interface{}, method string, token string) []map[string]interface{} {
	// split token
	tokenParts := strings.Split(token, " ")
	if len(tokenParts) != 2 {
		return nil
	}

	token = tokenParts[1]

	res, err := s.baseService.Request(method, GET_USER_BY_FILTER_ENDPOINT, data, token)
	if err != nil {
		return nil
	}

	// First, get the data as []interface{}
	usersInterface, ok := res["data"].([]interface{})
	if !ok {
		log.Println("Failed to convert data to []interface{}")
		return nil
	}

	var usersData []map[string]interface{}
	for _, userInterface := range usersInterface {
		user, ok := userInterface.(map[string]interface{})
		if !ok {
			log.Println("Failed to convert user to map[string]interface{}")
			continue
		}

		usersData = append(usersData, map[string]interface{}{
			"id":    user["id"],
			"nrp":   user["nrp"],
			"name":  user["name"],
			"email": user["email"],
		})
	}
	return usersData
}

func (s *UserManagementService) GetUserRole(method string, token string) map[string]interface{} {
	// split token
	tokenParts := strings.Split(token, " ")
	if len(tokenParts) != 2 {
		return nil
	}

	token = tokenParts[1]

	res, err := s.baseService.Request(method, GET_USER_ROLE_ENDPOINT, nil, token)
	if err != nil {
		return nil
	}

	roles, ok := res["data"].(map[string]interface{})
	if !ok {
		return nil
	}

	var rolesData map[string]interface{}
	rolesData = map[string]interface{}{
		"role": roles["role"],
	}
	return rolesData
}

func (s *UserManagementService) GetDosenDataByEmail(email string, method string, token string) map[string]interface{} {
	// split token
	tokenParts := strings.Split(token, " ")
	if len(tokenParts) != 2 {
		return nil
	}

	token = tokenParts[1]

	endpoint := GET_DOSEN_DATA_BY_EMAIL_ENDPOINT + email

	res, err := s.baseService.Request(method, endpoint, nil, token)
	if err != nil {
		return nil
	}

	dosen, ok := res["data"].(map[string]interface{})
	if !ok {
		return nil
	}

	var dosenData map[string]interface{}
	dosenData = map[string]interface{}{
		"id":           dosen["id"],
		"auth_user_id": dosen["auth_user_id"],
		"nip":          dosen["nrp"],
		"name":         dosen["name"],
		"email":        dosen["email"],
	}
	return dosenData
}
