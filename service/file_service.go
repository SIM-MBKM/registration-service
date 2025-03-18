package service

import (
	storageService "github.com/SIM-MBKM/filestorage/storage"
)

type FileService struct {
	storage *storageService.FileStorageManager
}

func NewFileService(config *storageService.Config, tokenManager *storageService.CacheTokenManager) *FileService {
	return &FileService{
		storage: storageService.NewFileStorageManager(config, tokenManager),
	}
}
