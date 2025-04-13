//go:build wireinject
// +build wireinject

package main

import (
	"registration-service/config"
	"registration-service/controller"
	"registration-service/repository"
	"registration-service/service"

	storageService "github.com/SIM-MBKM/filestorage/storage"
	"github.com/google/wire"
	"gorm.io/gorm"
)

func ProvideRegistrationRepository(db *gorm.DB) repository.RegistrationRepository {
	return repository.NewRegistrationRepository(db)
}

func ProvideDocumentRepository(db *gorm.DB) repository.DocumentRepository {
	return repository.NewDocumentRepository(db)
}

func ProvideRegistrationService(
	registrationRepository repository.RegistrationRepository,
	documentRepository repository.DocumentRepository,
	secretKey config.SecretKey,
	userManagementbaseURI config.UserManagementbaseURI,
	activityManagementbaseURI config.ActivityManagementbaseURI,
	matchingManagementbaseURI config.MatchingManagementbaseURI,
	asyncURIs config.AsyncURIs,
	config *storageService.Config,
	tokenManager *storageService.CacheTokenManager,
) service.RegistrationService {
	return service.NewRegistrationService(registrationRepository, documentRepository, string(secretKey), string(userManagementbaseURI), string(activityManagementbaseURI), string(matchingManagementbaseURI), []string(asyncURIs), config, tokenManager)
}

func ProvideRegistrationController(registrationService service.RegistrationService) controller.RegistrationController {
	return controller.NewRegistrationController(registrationService)
}

var RegistrationSet = wire.NewSet(
	ProvideRegistrationRepository,
	ProvideDocumentRepository,
	ProvideRegistrationService,
	ProvideRegistrationController,
)

func InitializeRegistration(
	db *gorm.DB,
	secretKey config.SecretKey,
	userManagementbaseURI config.UserManagementbaseURI,
	activityManagementbaseURI config.ActivityManagementbaseURI,
	matchingManagementbaseURI config.MatchingManagementbaseURI,
	asyncURIs config.AsyncURIs,
	config *storageService.Config,
	tokenManager *storageService.CacheTokenManager,
) (controller.RegistrationController, error) {
	wire.Build(RegistrationSet)
	return nil, nil
}

func ProvideDocumentService(
	documentRepository repository.DocumentRepository,
	registrationRepository repository.RegistrationRepository,
	config *storageService.Config,
	tokenManager *storageService.CacheTokenManager,
) service.DocumentService {
	return service.NewDocumentService(documentRepository, registrationRepository, config, tokenManager)
}

func ProvideDocumentController(documentService service.DocumentService) controller.DocumentController {
	return controller.NewDocumentController(documentService)
}

var DocumentSet = wire.NewSet(
	ProvideDocumentRepository,
	ProvideRegistrationRepository,
	ProvideDocumentService,
	ProvideDocumentController,
)

func InitializeDocument(
	db *gorm.DB,
	config *storageService.Config,
	tokenManager *storageService.CacheTokenManager,
) (controller.DocumentController, error) {
	wire.Build(DocumentSet)
	return nil, nil
}
