package main

import (
	"registration-service/config"
	"registration-service/helper"
	"registration-service/middleware"
	"registration-service/routes"

	storageService "github.com/SIM-MBKM/filestorage/storage"
	"github.com/SIM-MBKM/mod-service/src/helpers"
	baseServiceHelpers "github.com/SIM-MBKM/mod-service/src/helpers"
)

func main() {
	baseServiceHelpers.LoadEnv()
	secretKeyService := helpers.GetEnv("APP_KEY", "secret")
	port := helpers.GetEnv("GOLANG_PORT", "8088")

	userManagementServiceURI := helpers.GetEnv("USER_MANAGEMENT_BASE_URI", "http://localhost:8086")
	activityManagementServiceURI := helpers.GetEnv("ACTIVITY_MANAGEMENT_BASE_URI", "http://localhost:8087")

	db := config.SetupDatabaseConnection()

	config, err := storageService.LoadConfig()
	if err != nil {
		helper.PanicIfError(err)
	}

	cache := storageService.NewMemoryCache()

	tokenManager := storageService.NewCacheTokenManager(config, cache)

	registrationController, err := InitializeRegistration(db, config.SecretKey(secretKeyService), config.UserManagementbaseURI(userManagementServiceURI), config.BaseURI(activityManagementServiceURI), []string{"/async"}, config, tokenManager)

	if err != nil {
		helper.PanicIfError(err)
	}

	defer config.CloseDatabaseConnection(db)

	server := config.NewServer()
	server.Use(middleware.CORS())

	routes.RegistrationRoutes(server, registrationController)
	server.Run(":" + port)
}
