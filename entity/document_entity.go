package entity

import "github.com/google/uuid"

type (
	Document struct {
		ID             uuid.UUID `json:"id" gorm:"type:uuid;primary_key"`
		RegistrationID string    `json:"registration_id" gorm:"not null"`
		FileStorageID  string    `json:"file_storage_id" gorm:"not null;unique"`
		Name           string    `json:"name" gorm:"not null"`
		DocumentType   string    `json:"document_type" gorm:"not null"`
		Registration   *Registration
		BaseModel
	}
)
