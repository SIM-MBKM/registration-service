package dto

const (
	MESSAGE_REGISTRATION_GET_ALL_SUCCESS     = "Get all activities success"
	MESSAGE_REGISTRATION_GET_SUCCESS         = "Get registration success"
	MESSAGE_REGISTRATION_CREATE_SUCCESS      = "Create registration success"
	MESSAGE_REGISTRATION_UPDATE_SUCCESS      = "Update registration success"
	MESSAGE_REGISTRATION_DELETE_SUCCESS      = "Delete registration success"
	MESSAGE_REGISTRATION_UPDATE_ERROR        = "Update registration failed"
	MESSAGE_REGISTRATION_TRANSCRIPT_SUCCESS  = "Get registration transcript success"
	MESSAGE_REGISTRATION_MATCHING_SUCCESS    = "Get registration with matching success"
	MESSAGE_REGISTRATION_ELIGIBILITY_SUCCESS = "Check registration eligibility success"
	MESSAGE_REGISTRATION_GET_TOTAL_SUCCESS   = "Get total registration success"
)

type (
	GetRegistrationResponse struct {
		ID                        string             `json:"id"`
		ActivityID                string             `json:"activity_id"`
		UserID                    string             `json:"user_id"`
		UserNRP                   string             `json:"user_nrp"`
		UserName                  string             `json:"user_name"`
		AdvisingConfirmation      bool               `json:"advising_confirmation"`
		AcademicAdvisor           string             `json:"academic_advisor"`
		AcademicAdvisorEmail      string             `json:"academic_advisor_email"`
		MentorName                string             `json:"mentor_name"`
		MentorEmail               string             `json:"mentor_email"`
		LOValidation              string             `json:"lo_validation"`
		AcademicAdvisorValidation string             `json:"academic_advisor_validation"`
		Semester                  int                `json:"semester"`
		TotalSKS                  int                `json:"total_sks"`
		ActivityName              string             `json:"activity_name"`
		ApprovalStatus            bool               `json:"approval_status"`
		Documents                 []DocumentResponse `json:"documents"`
		Equivalents               interface{}        `json:"equivalents"`
		Matching                  interface{}        `json:"matching"`
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
		Semester             int    `form:"semester" binding:"required"`
		TotalSKS             int    `form:"total_sks" binding:"required"`
	}

	ApprovalRequest struct {
		Status string   `json:"status" binding:"required"`
		ID     []string `json:"id" binding:"required"`
	}

	UpdateRegistrationDataRequest struct {
		AdvisingConfirmation bool   `json:"advising_confirmation"`
		AcademicAdvisor      string `json:"academic_advisor"`
		AcademicAdvisorEmail string `json:"academic_advisor_email"`
		MentorName           string `json:"mentor_name"`
		MentorEmail          string `json:"mentor_email"`
		Semester             int    `json:"semester"`
		TotalSKS             int    `json:"total_sks"`
	}

	TranscriptResponse struct {
		RegistrationID string      `json:"registration_id"`
		UserID         string      `json:"user_id"`
		UserNRP        string      `json:"user_nrp"`
		UserName       string      `json:"user_name"`
		ActivityName   string      `json:"activity_name"`
		Semester       int         `json:"semester"`
		TotalSKS       int         `json:"total_sks"`
		ApprovalStatus bool        `json:"approval_status"`
		TranscriptData interface{} `json:"transcript_data"`
	}

	StudentTranscriptResponse struct {
		RegistrationID            string      `json:"registration_id"`
		ActivityID                string      `json:"activity_id"`
		ActivityName              string      `json:"activity_name"`
		Semester                  int         `json:"semester"`
		TotalSKS                  int         `json:"total_sks"`
		ApprovalStatus            bool        `json:"approval_status"`
		LOValidation              string      `json:"lo_validation"`
		AcademicAdvisorValidation string      `json:"academic_advisor_validation"`
		TranscriptData            interface{} `json:"transcript_data"`
	}

	StudentSyllabusResponse struct {
		RegistrationID            string      `json:"registration_id"`
		ActivityID                string      `json:"activity_id"`
		ActivityName              string      `json:"activity_name"`
		Semester                  int         `json:"semester"`
		TotalSKS                  int         `json:"total_sks"`
		ApprovalStatus            bool        `json:"approval_status"`
		LOValidation              string      `json:"lo_validation"`
		AcademicAdvisorValidation string      `json:"academic_advisor_validation"`
		SyllabusData              interface{} `json:"syllabus_data"`
	}

	StudentTranscriptsResponse struct {
		UserID        string                      `json:"user_id"`
		UserNRP       string                      `json:"user_nrp"`
		UserName      string                      `json:"user_name"`
		Registrations []StudentTranscriptResponse `json:"registrations"`
	}

	StudentSyllabusesResponse struct {
		UserID        string                    `json:"user_id"`
		UserNRP       string                    `json:"user_nrp"`
		UserName      string                    `json:"user_name"`
		Registrations []StudentSyllabusResponse `json:"registrations"`
	}

	StudentRegistrationWithMatchingResponse struct {
		ID                        string             `json:"id"`
		ActivityID                string             `json:"activity_id"`
		UserID                    string             `json:"user_id"`
		UserNRP                   string             `json:"user_nrp"`
		UserName                  string             `json:"user_name"`
		AdvisingConfirmation      bool               `json:"advising_confirmation"`
		AcademicAdvisor           string             `json:"academic_advisor"`
		AcademicAdvisorEmail      string             `json:"academic_advisor_email"`
		MentorName                string             `json:"mentor_name"`
		MentorEmail               string             `json:"mentor_email"`
		LOValidation              string             `json:"lo_validation"`
		AcademicAdvisorValidation string             `json:"academic_advisor_validation"`
		Semester                  int                `json:"semester"`
		TotalSKS                  int                `json:"total_sks"`
		ActivityName              string             `json:"activity_name"`
		ApprovalStatus            bool               `json:"approval_status"`
		Documents                 []DocumentResponse `json:"documents"`
		Equivalents               interface{}        `json:"equivalents"`
		Matching                  interface{}        `json:"matching"`
	}

	StudentRegistrationsWithMatchingResponse struct {
		UserID        string                                    `json:"user_id"`
		UserNRP       string                                    `json:"user_nrp"`
		UserName      string                                    `json:"user_name"`
		Registrations []StudentRegistrationWithMatchingResponse `json:"registrations"`
	}

	RegistrationEligibilityResponse struct {
		Eligible bool   `json:"eligible"`
		Message  string `json:"message"`
	}
)
