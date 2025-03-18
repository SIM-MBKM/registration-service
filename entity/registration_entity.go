package entity

import "github.com/google/uuid"

type (
	Registration struct {
		ID                        uuid.UUID `json:"id" gorm:"type:uuid;primary_key"`
		ActivityID                string    `json:"activity_id" gorm:"not null"`
		UserID                    string    `json:"user_id" gorm:"not null"`
		AdvisingConfirmation      bool      `json:"advising_confirmation" gorm:"not null"`
		AcademicAdvisor           string    `json:"academic_advisor" gorm:"not null"`
		AcademicAdvisorEmail      string    `json:"academic_advisor_email" gorm:"not null"`
		MentorName                string    `json:"mentor_name" gorm:"not null"`
		MentorEmail               string    `json:"mentor_email" gorm:"not null"`
		LOValidation              string    `json:"lo_validation" gorm:"not null"`
		AcademicAdvisorValidation string    `json:"academic_advisor_validation" gorm:"not null"`
		Semester                  string    `json:"semester" gorm:"not null"`
		TotalSKS                  int       `json:"total_sks" gorm:"not null"`
		ApprovalStatus            bool      `json:"is_approved" gorm:"not null"`
		Document                  []Document
		BaseModel
	}
)
