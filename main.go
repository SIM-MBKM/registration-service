package main

import (
	"log"
	localConfig "registration-service/config"
	"registration-service/helper"
	"registration-service/middleware"
	"registration-service/routes"
	"registration-service/service"
	"strconv"

	storageService "github.com/SIM-MBKM/filestorage/storage"
	"github.com/SIM-MBKM/mod-service/src/helpers"
	baseServiceHelpers "github.com/SIM-MBKM/mod-service/src/helpers"
	securityMiddleware "github.com/SIM-MBKM/mod-service/src/middleware"
)

func main() {
	baseServiceHelpers.LoadEnv()
	secretKeyService := baseServiceHelpers.GetEnv("APP_KEY", "secret")
	port := baseServiceHelpers.GetEnv("GOLANG_PORT", "8088")

	expireSeconds, _ := strconv.ParseInt(baseServiceHelpers.GetEnv("APP_KEY_EXPIRE_SECONDS", "9999"), 10, 64)

	userManagementServiceURI := helpers.GetEnv("USER_MANAGEMENT_BASE_URI", "http://localhost:8086")
	log.Println("User Management Service URI:", userManagementServiceURI)
	activityManagementServiceURI := helpers.GetEnv("ACTIVITY_MANAGEMENT_BASE_URI", "http://localhost:8088")
	log.Println("Activity Management Service URI:", activityManagementServiceURI)
	matchingManagementServiceURI := helpers.GetEnv("MATCHING_MANAGEMENT_BASE_URI", "http://localhost:8087")
	log.Println("Matching Management Service URI:", matchingManagementServiceURI)
	monitoringManagementServiceURI := helpers.GetEnv("MONITORING_MANAGEMENT_BASE_URI", "http://localhost:8089")
	log.Println("Monitoring Management Service URI:", monitoringManagementServiceURI)
	brokerbaseURI := helpers.GetEnv("BROKER_BASE_URI", "http://localhost:8099")
	log.Println("Broker Base URI:", brokerbaseURI)

	db := localConfig.SetupDatabaseConnection()

	config, err := storageService.LoadConfig()
	if err != nil {
		helper.PanicIfError(err)
	}

	cache := storageService.NewMemoryCache()

	tokenManager := storageService.NewCacheTokenManager(config, cache)

	registrationController, err := InitializeRegistration(db, localConfig.SecretKey(secretKeyService), localConfig.UserManagementbaseURI(userManagementServiceURI), localConfig.ActivityManagementbaseURI(activityManagementServiceURI), localConfig.MatchingManagementbaseURI(matchingManagementServiceURI), localConfig.MonitoringManagementbaseURI(monitoringManagementServiceURI), localConfig.BrokerbaseURI(brokerbaseURI), []string{"/async"}, config, tokenManager)

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
	server.Use(securityMiddleware.AccessKeyMiddleware(secretKeyService, expireSeconds))

	userService := service.NewUserManagementService(userManagementServiceURI, []string{"/async"})

	routes.RegistrationRoutes(server, registrationController, *userService)
	routes.DocumentRoutes(server, documentController)
	server.Run(":" + port)
}
