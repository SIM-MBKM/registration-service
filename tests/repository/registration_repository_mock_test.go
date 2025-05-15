package repository_test

import (
	"context"
	"errors"
	"registration-service/dto"
	"registration-service/entity"
	repository_mock "registration-service/mocks/repository"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

// This file contains additional tests that use mock objects for dependencies
// instead of mocking the database with sqlmock

// Helper to create a mock registration entity
func createMockRegistration() entity.Registration {
	now := time.Now()
	return entity.Registration{
		ID:                        uuid.New(),
		ActivityID:                "activity1",
		ActivityName:              "Activity 1",
		UserID:                    "user1",
		UserName:                  "User 1",
		UserNRP:                   "12345",
		AdvisingConfirmation:      true,
		AcademicAdvisorID:         "advisor1",
		AcademicAdvisor:           "Advisor 1",
		AcademicAdvisorEmail:      "advisor@example.com",
		MentorName:                "Mentor 1",
		MentorEmail:               "mentor@example.com",
		LOValidation:              "PENDING",
		AcademicAdvisorValidation: "PENDING",
		Semester:                  1,
		TotalSKS:                  20,
		ApprovalStatus:            false,
		BaseModel: entity.BaseModel{
			CreatedAt: &now,
			UpdatedAt: &now,
		},
	}
}

// Test error scenarios using mocks
func TestRegistrationRepository_CreateWithTransactionError(t *testing.T) {
	// Create mock repository
	mockRepo := new(repository_mock.MockRegistrationRepository)

	// Setup test data
	ctx := context.Background()
	registration := createMockRegistration()
	expectedError := errors.New("transaction error")

	// Setup mock behavior
	mockRepo.On("Create", ctx, registration, mock.Anything).Return(entity.Registration{}, expectedError)

	// Call method
	result, err := mockRepo.Create(ctx, registration, nil)

	// Assertions
	assert.Error(t, err)
	assert.Equal(t, expectedError, err)
	assert.Equal(t, entity.Registration{}, result)
	mockRepo.AssertExpectations(t)
}

func TestRegistrationRepository_FindByID_NotFound(t *testing.T) {
	// Create mock repository
	mockRepo := new(repository_mock.MockRegistrationRepository)

	// Setup test data
	ctx := context.Background()
	id := uuid.New().String()
	mockError := gorm.ErrRecordNotFound

	// Setup mock behavior
	mockRepo.On("FindByID", ctx, id, mock.Anything).Return(entity.Registration{}, mockError)

	// Call method
	registration, err := mockRepo.FindByID(ctx, id, nil)

	// Assertions
	assert.Error(t, err)
	assert.Equal(t, gorm.ErrRecordNotFound, err)
	assert.Equal(t, entity.Registration{}, registration)
	mockRepo.AssertExpectations(t)
}

func TestRegistrationRepository_Index_WithFilters(t *testing.T) {
	// Create mock repository
	mockRepo := new(repository_mock.MockRegistrationRepository)

	// Setup test data
	ctx := context.Background()
	filter := dto.FilterRegistrationRequest{
		ActivityName:         "Test Activity",
		UserNRP:              "12345",
		AcademicAdvisorEmail: "advisor@example.com",
		ApprovalStatus:       true,
	}
	pagReq := dto.PaginationRequest{
		Limit:  10,
		Offset: 0,
	}

	expectedRegistrations := []entity.Registration{createMockRegistration()}
	expectedTotal := int64(1)

	// Setup mock behavior
	mockRepo.On("Index", ctx, mock.Anything, pagReq, filter).Return(expectedRegistrations, expectedTotal, nil)

	// Call method
	registrations, total, err := mockRepo.Index(ctx, nil, pagReq, filter)

	// Assertions
	assert.NoError(t, err)
	assert.Equal(t, expectedRegistrations, registrations)
	assert.Equal(t, expectedTotal, total)
	mockRepo.AssertExpectations(t)
}

func TestRegistrationRepository_Update_Error(t *testing.T) {
	// Create mock repository
	mockRepo := new(repository_mock.MockRegistrationRepository)

	// Setup test data
	ctx := context.Background()
	id := uuid.New().String()
	registration := createMockRegistration()
	expectedError := errors.New("update error")

	// Setup mock behavior
	mockRepo.On("Update", ctx, id, registration, mock.Anything).Return(expectedError)

	// Call method
	err := mockRepo.Update(ctx, id, registration, nil)

	// Assertions
	assert.Error(t, err)
	assert.Equal(t, expectedError, err)
	mockRepo.AssertExpectations(t)
}

func TestRegistrationRepository_Destroy_Error(t *testing.T) {
	// Create mock repository
	mockRepo := new(repository_mock.MockRegistrationRepository)

	// Setup test data
	ctx := context.Background()
	id := uuid.New().String()
	expectedError := errors.New("destroy error")

	// Setup mock behavior
	mockRepo.On("Destroy", ctx, id, mock.Anything).Return(expectedError)

	// Call method
	err := mockRepo.Destroy(ctx, id, nil)

	// Assertions
	assert.Error(t, err)
	assert.Equal(t, expectedError, err)
	mockRepo.AssertExpectations(t)
}

func TestRegistrationRepository_FindByNRP_Error(t *testing.T) {
	// Create mock repository
	mockRepo := new(repository_mock.MockRegistrationRepository)

	// Setup test data
	ctx := context.Background()
	nrp := "12345"
	expectedError := errors.New("nrp search error")

	// Setup mock behavior
	mockRepo.On("FindByNRP", ctx, nrp, mock.Anything).Return(entity.Registration{}, expectedError)

	// Call method
	_, err := mockRepo.FindByNRP(ctx, nrp, nil)

	// Assertions
	assert.Error(t, err)
	assert.Equal(t, expectedError, err)
	mockRepo.AssertExpectations(t)
}

func TestRegistrationRepository_FindByActivityIDAndNRP_Error(t *testing.T) {
	// Create mock repository
	mockRepo := new(repository_mock.MockRegistrationRepository)

	// Setup test data
	ctx := context.Background()
	activityID := "activity1"
	nrp := "12345"
	expectedError := errors.New("activity and nrp search error")

	// Setup mock behavior
	mockRepo.On("FindByActivityIDAndNRP", ctx, activityID, nrp, mock.Anything).Return(entity.Registration{}, expectedError)

	// Call method
	_, err := mockRepo.FindByActivityIDAndNRP(ctx, activityID, nrp, nil)

	// Assertions
	assert.Error(t, err)
	assert.Equal(t, expectedError, err)
	mockRepo.AssertExpectations(t)
}

func TestRegistrationRepository_FindTotal_Error(t *testing.T) {
	// Create mock repository
	mockRepo := new(repository_mock.MockRegistrationRepository)

	// Setup test data
	ctx := context.Background()
	filter := dto.FilterRegistrationRequest{}
	expectedError := errors.New("count error")

	// Setup mock behavior
	mockRepo.On("FindTotal", ctx, filter, mock.Anything).Return(int64(0), expectedError)

	// Call method
	_, err := mockRepo.FindTotal(ctx, filter, nil)

	// Assertions
	assert.Error(t, err)
	assert.Equal(t, expectedError, err)
	mockRepo.AssertExpectations(t)
}

func TestRegistrationRepository_FindRegistrationByAdvisiorEmail_Error(t *testing.T) {
	// Create mock repository
	mockRepo := new(repository_mock.MockRegistrationRepository)

	// Setup test data
	ctx := context.Background()
	email := "advisor@example.com"
	expectedError := errors.New("advisor email search error")

	// Setup mock behavior
	mockRepo.On("FindRegistrationByAdvisiorEmail", ctx, email, mock.Anything).Return(entity.Registration{}, expectedError)

	// Call method
	_, err := mockRepo.FindRegistrationByAdvisiorEmail(ctx, email, nil)

	// Assertions
	assert.Error(t, err)
	assert.Equal(t, expectedError, err)
	mockRepo.AssertExpectations(t)
}
