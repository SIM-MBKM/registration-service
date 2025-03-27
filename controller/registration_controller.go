package controller

import (
	"log"
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
}

func NewRegistrationController(registrationService service.RegistrationService) RegistrationController {
	return &registrationController{
		registrationService: registrationService,
	}
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
	id := ctx.Param("id")
	activity, err := c.registrationService.FindRegistrationByID(ctx, id, nil)
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

	err = c.registrationService.UpdateRegistration(ctx, id, request, nil)
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
	id := ctx.Param("id")
	err := c.registrationService.DeleteRegistration(ctx, id, nil)
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
