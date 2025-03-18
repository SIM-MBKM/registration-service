package dto

type (
	PaginationResponse struct {
		CurrentPage  int    `json:"current_page"`
		FirstPageUrl string `json:"first_page_url"`
		LastPage     int64  `json:"last_page"`
		LastPageUrl  string `json:"last_page_url"`
		NextPageUrl  string `json:"next_page_url"`
		PerPage      int    `json:"per_page"`
		PrevPageUrl  string `json:"prev_page_url"`
		To           int    `json:"to"`
		Total        int64  `json:"total"`
		TotalPages   int64  `json:"total_pages"`
	}

	PaginationRequest struct {
		URL    string `json:"url"`
		Offset int    `json:"offset"`
		Limit  int    `json:"limit"`
	}
)
