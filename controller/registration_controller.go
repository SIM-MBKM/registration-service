package controller

import (
	"net/http"
	"registration-service/dto"
	"registration-service/helper"
	"registration-service/service"

	"github.com/gin-gonic/gin"
)

type registrationController struct {
	registrationService service.RegistrationService
}

type RegistrationController interface {
	GetAllRegistrations(ctx *gin.Context)
	GetRegistrationByID(ctx *gin.Context)
	CreateRegistration(ctx *gin.Context)
	UpdateRegistration(ctx *gin.Context)
	DeleteRegistration(ctx *gin.Context)
	GetRegistrationsByAdvisor(ctx *gin.Context)
	GetRegistrationsByLOMBKM(ctx *gin.Context)
	GetRegistrationsByStudent(ctx *gin.Context)
	ApproveRegistration(ctx *gin.Context)
	GetRegistrationTranscript(ctx *gin.Context)
	GetStudentRegistrationsWithTranscripts(ctx *gin.Context)
	GetStudentRegistrationsWithSyllabuses(ctx *gin.Context)
	GetStudentRegistrationsWithMatching(ctx *gin.Context)
	CheckRegistrationEligibility(ctx *gin.Context)
	GetTotalRegistrationByAdvisorEmail(ctx *gin.Context)
}

func NewRegistrationController(registrationService service.RegistrationService) RegistrationController {
	return &registrationController{
		registrationService: registrationService,
	}
}

func (c *registrationController) GetTotalRegistrationByAdvisorEmail(ctx *gin.Context) {
	token := ctx.GetHeader("Authorization")
	if token == "" {
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, dto.Response{
			Status:  dto.STATUS_ERROR,
			Message: dto.MESSAGE_UNAUTHORIZED,
		})
		return
	}

	registrationCount, err := c.registrationService.FindTotalRegistrationByAdvisorEmail(ctx, token, nil)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, dto.Response{
			Status:  dto.STATUS_ERROR,
			Message: err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, dto.Response{
		Status:  dto.STATUS_SUCCESS,
		Message: dto.MESSAGE_REGISTRATION_GET_TOTAL_SUCCESS,
		Data:    registrationCount,
	})
}

func (c *registrationController) ApproveRegistration(ctx *gin.Context) {
	var request dto.ApprovalRequest
	err := ctx.ShouldBindJSON(&request)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, dto.Response{
			Status:  dto.STATUS_ERROR,
			Message: err.Error(),
		})
		return
	}

	token := ctx.GetHeader("Authorization")
	if token == "" {
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, dto.Response{
			Status:  dto.STATUS_ERROR,
			Message: dto.MESSAGE_UNAUTHORIZED,
		})
		return
	}

	userRole, exists := ctx.Get("userRole")

	if !exists {
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, dto.Response{
			Status:  dto.STATUS_ERROR,
			Message: dto.MESSAGE_UNAUTHORIZED,
		})
		return
	}

	userRoleStr, ok := userRole.(string)
	if !ok {
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, dto.Response{
			Status:  dto.STATUS_ERROR,
			Message: dto.MESSAGE_UNAUTHORIZED,
		})
		return
	}

	if userRoleStr == "DOSEN PEMBIMBING" {
		err := c.registrationService.AdvisorRegistrationApproval(ctx, token, request, nil)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, dto.Response{
				Status:  dto.STATUS_ERROR,
				Message: dto.MESSAGE_REGISTRATION_UPDATE_ERROR,
			})
			return
		}
	} else if userRoleStr == "ADMIN" {
		err := c.registrationService.LORegistrationApproval(ctx, token, request, nil)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, dto.Response{
				Status:  dto.STATUS_ERROR,
				Message: dto.MESSAGE_REGISTRATION_UPDATE_ERROR,
			})
			return
		}
	} else if userRoleStr == "LO-MBKM" {
		err := c.registrationService.LORegistrationApproval(ctx, token, request, nil)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, dto.Response{
				Status:  dto.STATUS_ERROR,
				Message: dto.MESSAGE_REGISTRATION_UPDATE_ERROR,
			})
			return
		}
	} else {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, dto.Response{
			Status:  dto.STATUS_ERROR,
			Message: dto.MESSAGE_UNAUTHORIZED,
		})
		return
	}

	ctx.AbortWithStatusJSON(http.StatusOK, dto.Response{
		Status:  dto.STATUS_SUCCESS,
		Message: dto.MESSAGE_REGISTRATION_UPDATE_SUCCESS,
	})
	return
}

func (c *registrationController) GetRegistrationsByStudent(ctx *gin.Context) {
	var request dto.FilterRegistrationRequest
	err := ctx.ShouldBindJSON(&request)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, dto.Response{
			Status:  dto.STATUS_ERROR,
			Message: err.Error(),
		})
		return
	}

	// get header token
	token := ctx.GetHeader("Authorization")
	if token == "" {
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, dto.Response{
			Status:  dto.STATUS_ERROR,
			Message: dto.MESSAGE_UNAUTHORIZED,
		})
		return
	}

	pagReq := helper.Pagination(ctx)
	registrations, metaData, err := c.registrationService.FindRegistrationByStudent(ctx, pagReq, request, token, nil)

	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, dto.Response{
			Status:  dto.STATUS_ERROR,
			Message: err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, dto.Response{
		Message:            dto.MESSAGE_REGISTRATION_GET_ALL_SUCCESS,
		Status:             dto.STATUS_SUCCESS,
		Data:               registrations,
		PaginationResponse: &metaData,
	})
}

func (c *registrationController) GetRegistrationsByAdvisor(ctx *gin.Context) {
	var request dto.FilterRegistrationRequest
	err := ctx.ShouldBindJSON(&request)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, dto.Response{
			Status:  dto.STATUS_ERROR,
			Message: err.Error(),
		})
		return
	}

	// get header token
	token := ctx.GetHeader("Authorization")
	if token == "" {
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, dto.Response{
			Status:  dto.STATUS_ERROR,
			Message: dto.MESSAGE_UNAUTHORIZED,
		})
		return
	}

	pagReq := helper.Pagination(ctx)
	registrations, metaData, err := c.registrationService.FindRegistrationByAdvisor(ctx, pagReq, request, token, nil)

	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, dto.Response{
			Status:  dto.STATUS_ERROR,
			Message: err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, dto.Response{
		Message:            dto.MESSAGE_REGISTRATION_GET_ALL_SUCCESS,
		Status:             dto.STATUS_SUCCESS,
		Data:               registrations,
		PaginationResponse: &metaData,
	})
}

func (c *registrationController) GetRegistrationsByLOMBKM(ctx *gin.Context) {
	var request dto.FilterRegistrationRequest
	err := ctx.ShouldBindJSON(&request)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, dto.Response{
			Status:  dto.STATUS_ERROR,
			Message: err.Error(),
		})
		return
	}

	// get header token
	token := ctx.GetHeader("Authorization")
	if token == "" {
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, dto.Response{
			Status:  dto.STATUS_ERROR,
			Message: dto.MESSAGE_UNAUTHORIZED,
		})
		return
	}

	pagReq := helper.Pagination(ctx)
	registrations, metaData, err := c.registrationService.FindRegistrationByLOMBKM(ctx, pagReq, request, token, nil)

	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, dto.Response{
			Status:  dto.STATUS_ERROR,
			Message: err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, dto.Response{
		Message:            dto.MESSAGE_REGISTRATION_GET_ALL_SUCCESS,
		Status:             dto.STATUS_SUCCESS,
		Data:               registrations,
		PaginationResponse: &metaData,
	})
}

func (c *registrationController) GetAllRegistrations(ctx *gin.Context) {
	var request dto.FilterRegistrationRequest
	err := ctx.ShouldBindJSON(&request)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, dto.Response{
			Status:  dto.STATUS_ERROR,
			Message: err.Error(),
		})
		return
	}

	// get header token
	token := ctx.GetHeader("Authorization")
	if token == "" {
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, dto.Response{
			Status:  dto.STATUS_ERROR,
			Message: dto.MESSAGE_UNAUTHORIZED,
		})
		return
	}

	pagReq := helper.Pagination(ctx)
	registrations, metaData, err := c.registrationService.FindAllRegistrations(ctx, pagReq, request, nil, token)

	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, dto.Response{
			Status:  dto.STATUS_ERROR,
			Message: err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, dto.Response{
		Message:            dto.MESSAGE_REGISTRATION_GET_ALL_SUCCESS,
		Status:             dto.STATUS_SUCCESS,
		Data:               registrations,
		PaginationResponse: &metaData,
	})
}

func (c *registrationController) CreateRegistration(ctx *gin.Context) {
	var request dto.CreateRegistrationRequest
	err := ctx.ShouldBind(&request)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, dto.Response{
			Status:  dto.STATUS_ERROR,
			Message: err.Error(),
		})
		return
	}

	file, err := ctx.FormFile("file")
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, dto.Response{
			Status:  dto.STATUS_ERROR,
			Message: err.Error(),
		})
		return
	}
	geoletter, err := ctx.FormFile("geoletter")
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, dto.Response{
			Status:  dto.STATUS_ERROR,
			Message: err.Error(),
		})
		return
	}

	// get header token
	token := ctx.GetHeader("Authorization")
	if token == "" {
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, dto.Response{
			Status:  dto.STATUS_ERROR,
			Message: dto.MESSAGE_UNAUTHORIZED,
		})
		return
	}

	err = c.registrationService.CreateRegistration(ctx, request, file, geoletter, nil, token)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, dto.Response{
			Status:  dto.STATUS_ERROR,
			Message: err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusCreated, dto.Response{
		Message: dto.MESSAGE_REGISTRATION_CREATE_SUCCESS,
		Status:  dto.STATUS_SUCCESS,
	})
}

func (c *registrationController) GetRegistrationByID(ctx *gin.Context) {
	token := ctx.GetHeader("Authorization")
	if token == "" {
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, dto.Response{
			Status:  dto.STATUS_ERROR,
			Message: dto.MESSAGE_UNAUTHORIZED,
		})
		return
	}

	id := ctx.Param("id")
	activity, err := c.registrationService.FindRegistrationByID(ctx, id, token, nil)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, dto.Response{
			Status:  dto.STATUS_ERROR,
			Message: err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, dto.Response{
		Message: dto.MESSAGE_REGISTRATION_GET_SUCCESS,
		Status:  dto.STATUS_SUCCESS,
		Data:    activity,
	})
}

func (c *registrationController) UpdateRegistration(ctx *gin.Context) {
	token := ctx.GetHeader("Authorization")
	if token == "" {
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, dto.Response{
			Status:  dto.STATUS_ERROR,
			Message: dto.MESSAGE_UNAUTHORIZED,
		})
		return
	}

	id := ctx.Param("id")
	var request dto.UpdateRegistrationDataRequest
	err := ctx.ShouldBindJSON(&request)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, dto.Response{
			Status:  dto.STATUS_ERROR,
			Message: err.Error(),
		})
		return
	}

	err = c.registrationService.UpdateRegistration(ctx, id, request, token, nil)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, dto.Response{
			Status:  dto.STATUS_ERROR,
			Message: err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, dto.Response{
		Message: dto.MESSAGE_REGISTRATION_UPDATE_SUCCESS,
		Status:  dto.STATUS_SUCCESS,
	})
}

func (c *registrationController) DeleteRegistration(ctx *gin.Context) {
	token := ctx.GetHeader("Authorization")
	if token == "" {
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, dto.Response{
			Status:  dto.STATUS_ERROR,
			Message: dto.MESSAGE_UNAUTHORIZED,
		})
		return
	}

	id := ctx.Param("id")
	err := c.registrationService.DeleteRegistration(ctx, id, token, nil)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, dto.Response{
			Status:  dto.STATUS_ERROR,
			Message: err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, dto.Response{
		Message: dto.MESSAGE_REGISTRATION_DELETE_SUCCESS,
		Status:  dto.STATUS_SUCCESS,
	})
}

func (c *registrationController) GetRegistrationTranscript(ctx *gin.Context) {
	registrationID := ctx.Param("id")
	if registrationID == "" {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, dto.Response{
			Status:  dto.STATUS_ERROR,
			Message: "ID is required",
		})
		return
	}

	// get header token
	token := ctx.GetHeader("Authorization")
	if token == "" {
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, dto.Response{
			Status:  dto.STATUS_ERROR,
			Message: dto.MESSAGE_UNAUTHORIZED,
		})
		return
	}

	transcript, err := c.registrationService.GetRegistrationTranscript(ctx, registrationID, token, nil)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, dto.Response{
			Status:  dto.STATUS_ERROR,
			Message: err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, dto.Response{
		Message: dto.MESSAGE_REGISTRATION_TRANSCRIPT_SUCCESS,
		Status:  dto.STATUS_SUCCESS,
		Data:    transcript,
	})
}

func (c *registrationController) GetStudentRegistrationsWithTranscripts(ctx *gin.Context) {
	var request dto.FilterRegistrationRequest
	err := ctx.ShouldBindJSON(&request)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, dto.Response{
			Status:  dto.STATUS_ERROR,
			Message: err.Error(),
		})
		return
	}

	// get header token
	token := ctx.GetHeader("Authorization")
	if token == "" {
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, dto.Response{
			Status:  dto.STATUS_ERROR,
			Message: dto.MESSAGE_UNAUTHORIZED,
		})
		return
	}

	pagReq := helper.Pagination(ctx)
	transcripts, metaData, err := c.registrationService.GetStudentRegistrationsWithTranscripts(ctx, pagReq, request, token, nil)

	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, dto.Response{
			Status:  dto.STATUS_ERROR,
			Message: err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, dto.Response{
		Message:            dto.MESSAGE_REGISTRATION_TRANSCRIPT_SUCCESS,
		Status:             dto.STATUS_SUCCESS,
		Data:               transcripts,
		PaginationResponse: &metaData,
	})
}

func (c *registrationController) GetStudentRegistrationsWithSyllabuses(ctx *gin.Context) {
	var request dto.FilterRegistrationRequest
	err := ctx.ShouldBindJSON(&request)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, dto.Response{
			Status:  dto.STATUS_ERROR,
			Message: err.Error(),
		})
		return
	}

	// get header token
	token := ctx.GetHeader("Authorization")
	if token == "" {
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, dto.Response{
			Status:  dto.STATUS_ERROR,
			Message: dto.MESSAGE_UNAUTHORIZED,
		})
		return
	}

	pagReq := helper.Pagination(ctx)
	transcripts, metaData, err := c.registrationService.GetStudentRegistrationsWithSyllabuses(ctx, pagReq, request, token, nil)

	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, dto.Response{
			Status:  dto.STATUS_ERROR,
			Message: err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, dto.Response{
		Message:            dto.MESSAGE_REGISTRATION_TRANSCRIPT_SUCCESS,
		Status:             dto.STATUS_SUCCESS,
		Data:               transcripts,
		PaginationResponse: &metaData,
	})
}

func (c *registrationController) GetStudentRegistrationsWithMatching(ctx *gin.Context) {
	var request dto.FilterRegistrationRequest
	err := ctx.ShouldBindJSON(&request)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, dto.Response{
			Status:  dto.STATUS_ERROR,
			Message: err.Error(),
		})
		return
	}

	// get header token
	token := ctx.GetHeader("Authorization")
	if token == "" {
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, dto.Response{
			Status:  dto.STATUS_ERROR,
			Message: dto.MESSAGE_UNAUTHORIZED,
		})
		return
	}

	pagReq := helper.Pagination(ctx)
	registrations, metaData, err := c.registrationService.FindRegistrationsWithMatching(ctx, pagReq, request, token, nil)

	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, dto.Response{
			Status:  dto.STATUS_ERROR,
			Message: err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, dto.Response{
		Message:            dto.MESSAGE_REGISTRATION_MATCHING_SUCCESS,
		Status:             dto.STATUS_SUCCESS,
		Data:               registrations,
		PaginationResponse: &metaData,
	})
}

func (c *registrationController) CheckRegistrationEligibility(ctx *gin.Context) {
	activityID := ctx.Query("activity_id")
	if activityID == "" {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, dto.Response{
			Status:  dto.STATUS_ERROR,
			Message: "Activity ID is required",
		})
		return
	}

	// Get header token
	token := ctx.GetHeader("Authorization")
	if token == "" {
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, dto.Response{
			Status:  dto.STATUS_ERROR,
			Message: dto.MESSAGE_UNAUTHORIZED,
		})
		return
	}

	eligibility, _ := c.registrationService.CheckRegistrationEligibility(ctx, activityID, token, nil)

	// Even if there's an error, we still want to return the eligibility response
	// since it contains valuable information about why the registration isn't eligible
	ctx.JSON(http.StatusOK, dto.Response{
		Message: dto.MESSAGE_REGISTRATION_ELIGIBILITY_SUCCESS,
		Status:  dto.STATUS_SUCCESS,
		Data:    eligibility,
	})
}
