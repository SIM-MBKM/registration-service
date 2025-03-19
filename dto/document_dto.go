package dto

const (
	MESSAGE_DOCUMENT_GET_ALL_SUCCESS = "Get all documents success"
	MESSAGE_DOCUMENT_GET_SUCCESS     = "Get document success"
	MESSAGE_DOCUMENT_CREATE_SUCCESS  = "Create document success"
	MESSAGE_DOCUMENT_UPDATE_SUCCESS  = "Update document success"
	MESSAGE_DOCUMENT_DELETE_SUCCESS  = "Delete document success"
)

type (
	DocumentRequest struct {
		RegistrationID string `json:"registration_id" form:"registration_id" binding:"required"`
		Name           string `json:"name" form:"name" binding:"required"`
		DocumentType   string `json:"document_type" form:"document_type" binding:"required"`
	}

	UpdateDocumentRequest struct {
		RegistrationID string `json:"registration_id" form:"registration_id" binding:"required"`
	}

	DocumentResponse struct {
		ID             string `json:"id"`
		RegistrationID string `json:"registration_id"`
		FileStorageID  string `json:"file_storage_id"`
		Name           string `json:"name"`
		DocumentType   string `json:"document_type"`
	}
)
