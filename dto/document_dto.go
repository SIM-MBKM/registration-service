package dto

type (
	DocumentRequest struct {
		RegistrationID string `json:"registration_id" form:"registration_id" binding:"required"`
		FileStorageID  string `json:"file_storage_id" form:"file_storage_id" binding:"required"`
		Name           string `json:"name" form:"name" binding:"required"`
		DocumentType   string `json:"document_type" form:"document_type" binding:"required"`
	}

	DocumentResponse struct {
		ID             string `json:"id"`
		RegistrationID string `json:"registration_id"`
		FileStorageID  string `json:"file_storage_id"`
		Name           string `json:"name"`
		DocumentType   string `json:"document_type"`
	}
)
