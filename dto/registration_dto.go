package dto

const (
	MESSAGE_REGISTRATION_GET_ALL_SUCCESS = "Get all activities success"
	MESSAGE_REGISTRATION_GET_SUCCESS     = "Get registration success"
	MESSAGE_REGISTRATION_CREATE_SUCCESS  = "Create registration success"
	MESSAGE_REGISTRATION_UPDATE_SUCCESS  = "Update registration success"
	MESSAGE_REGISTRATION_DELETE_SUCCESS  = "Delete registration success"
	MESSAGE_REGISTRATION_UPDATE_ERROR    = "Update registration failed"
)

type (
	GetRegistrationResponse struct {
		ID                        string                 `json:"id"`
		ActivityID                string                 `json:"activity_id"`
		UserID                    string                 `json:"user_id"`
		UserNRP                   string                 `json:"user_nrp"`
		UserName                  string                 `json:"user_name"`
		AdvisingConfirmation      bool                   `json:"advising_confirmation"`
		AcademicAdvisor           string                 `json:"academic_advisor"`
		AcademicAdvisorEmail      string                 `json:"academic_advisor_email"`
		MentorName                string                 `json:"mentor_name"`
		MentorEmail               string                 `json:"mentor_email"`
		LOValidation              string                 `json:"lo_validation"`
		AcademicAdvisorValidation string                 `json:"academic_advisor_validation"`
		Semester                  string                 `json:"semester"`
		TotalSKS                  int                    `json:"total_sks"`
		ActivityName              string                 `json:"activity_name"`
		Documents                 []DocumentResponse     `json:"documents"`
		Matching                  map[string]interface{} `json:"matching"`
	}

	FilterRegistrationRequest struct {
		ActivityName              string `json:"activity_name"`
		UserName                  string `json:"user_name"`
		UserNRP                   string `json:"user_nrp"`
		AcademicAdvisorEmail      string `json:"academic_advisor_email"`
		ApprovalStatus            bool   `json:"approval_status"`
		LOValidation              string `json:"lo_validation"`
		AcademicAdvisorValidation string `json:"academic_advisor_validation"`
	}

	FilterDataRequest struct {
		ActivityID                []string `json:"activity_id"`
		UserID                    []string `json:"user_id"`
		AcademicAdvisor           string   `json:"academic_advisor"`
		ApprovalStatus            bool     `json:"approval_status"`
		LOValidation              string   `json:"lo_validation"`
		AcademicAdvisorValidation string   `json:"academic_advisor_validation"`
	}

	CreateRegistrationRequest struct {
		ActivityID           string `form:"activity_id" binding:"required"`
		AcademicAdvisorID    string `form:"academic_advisor_id" binding:"required"`
		AdvisingConfirmation bool   `form:"advising_confirmation" binding:"required"`
		AcademicAdvisor      string `form:"academic_advisor" binding:"required"` // This field doesn't match what's in your form
		AcademicAdvisorEmail string `form:"academic_advisor_email" binding:"required"`
		MentorName           string `form:"mentor_name" binding:"required"`
		MentorEmail          string `form:"mentor_email" binding:"required"`
		Semester             string `form:"semester" binding:"required"`
		TotalSKS             int    `form:"total_sks" binding:"required"`
	}

	ApprovalRequest struct {
		Status string `json:"status" binding:"required"`
		ID     string `json:"id" binding:"required"`
	}

	UpdateRegistrationDataRequest struct {
		AdvisingConfirmation bool   `json:"advising_confirmation"`
		AcademicAdvisor      string `json:"academic_advisor"`
		AcademicAdvisorEmail string `json:"academic_advisor_email"`
		MentorName           string `json:"mentor_name"`
		MentorEmail          string `json:"mentor_email"`
		Semester             string `json:"semester"`
		TotalSKS             int    `json:"total_sks"`
	}
)
