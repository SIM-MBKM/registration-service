package main

import (
	localConfig "registration-service/config"
	"registration-service/helper"
	"registration-service/middleware"
	"registration-service/routes"
	"registration-service/service"

	storageService "github.com/SIM-MBKM/filestorage/storage"
	"github.com/SIM-MBKM/mod-service/src/helpers"
	baseServiceHelpers "github.com/SIM-MBKM/mod-service/src/helpers"
)

func main() {
	baseServiceHelpers.LoadEnv()
	secretKeyService := helpers.GetEnv("APP_KEY", "secret")
	port := helpers.GetEnv("GOLANG_PORT", "8088")

	userManagementServiceURI := helpers.GetEnv("USER_MANAGEMENT_BASE_URI", "http://localhost:8086")
	activityManagementServiceURI := helpers.GetEnv("ACTIVITY_MANAGEMENT_BASE_URI", "http://localhost:8088")
	matchingManagementServiceURI := helpers.GetEnv("MATCHING_MANAGEMENT_BASE_URI", "http://localhost:8087")

	db := localConfig.SetupDatabaseConnection()

	config, err := storageService.LoadConfig()
	if err != nil {
		helper.PanicIfError(err)
	}

	cache := storageService.NewMemoryCache()

	tokenManager := storageService.NewCacheTokenManager(config, cache)

	registrationController, err := InitializeRegistration(db, localConfig.SecretKey(secretKeyService), localConfig.UserManagementbaseURI(userManagementServiceURI), localConfig.ActivityManagementbaseURI(activityManagementServiceURI), localConfig.MatchingManagementbaseURI(matchingManagementServiceURI), []string{"/async"}, config, tokenManager)

	if err != nil {
		helper.PanicIfError(err)
	}

	documentController, err := InitializeDocument(db, config, tokenManager)

	if err != nil {
		helper.PanicIfError(err)
	}

	defer localConfig.CloseDatabaseConnection(db)

	server := localConfig.NewServer()
	server.Use(middleware.CORS())

	userService := service.NewUserManagementService(userManagementServiceURI, []string{"/async"})

	routes.RegistrationRoutes(server, registrationController, *userService)
	routes.DocumentRoutes(server, documentController)
	server.Run(":" + port)
}
