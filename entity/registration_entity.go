package entity

import "github.com/google/uuid"

type (
	Registration struct {
		ID                        uuid.UUID `json:"id" gorm:"type:uuid;primary_key"`
		ActivityID                string    `json:"activity_id" gorm:"not null"`
		ActivityName              string    `json:"activity_name" gorm:"not null"`
		UserID                    string    `json:"user_id" gorm:"not null"`
		UserName                  string    `json:"user_name" gorm:"not null"`
		UserNRP                   string    `json:"user_nrp" gorm:"not null"`
		AdvisingConfirmation      bool      `json:"advising_confirmation" gorm:"not null"`
		AcademicAdvisorID         string    `json:"academic_advisor_id" gorm:"not null"`
		AcademicAdvisor           string    `json:"academic_advisor" gorm:"not null"`
		AcademicAdvisorEmail      string    `json:"academic_advisor_email" gorm:"not null"`
		MentorName                string    `json:"mentor_name" gorm:"not null"`
		MentorEmail               string    `json:"mentor_email" gorm:"not null"`
		LOValidation              string    `json:"lo_validation" gorm:"not null"`
		AcademicAdvisorValidation string    `json:"academic_advisor_validation" gorm:"not null"`
		Semester                  int       `json:"semester" gorm:"not null"`
		TotalSKS                  int       `json:"total_sks" gorm:"not null"`
		ApprovalStatus            bool      `json:"approval_status" gorm:"not null"`
		Document                  []Document
		BaseModel
	}

	RegistrationCount struct {
		Total                                int64 `json:"total"`
		TotalApproved                        int64 `json:"total_approved"`
		TotalPercentageFromLastMonth         int64 `json:"total_percentage_from_last_month"`
		TotalApprovedPercentageFromLastMonth int64 `json:"total_approved_percentage_from_last_month"`
	}
)
