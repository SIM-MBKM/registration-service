package dto

const (
	// StatusSuccess is a constant for success status
	STATUS_SUCCESS = "success"
	// StatusError is a constant for error status
	STATUS_ERROR         = "error"
	MESSAGE_UNAUTHORIZED = "Unauthorized"
	MESSAGE_FORBIDDEN    = "Forbidden"
)

type Response struct {
	Message string      `json:"message"`
	Status  string      `json:"status"`
	Data    interface{} `json:"data,omitempty"`
	// diisi oleh ResponseMeta
	*PaginationResponse
}

type ResponseMeta struct {
	AfterCursor  *string `json:"after_cursor"`
	BeforeCursor *string `json:"before_cursor"`
}
