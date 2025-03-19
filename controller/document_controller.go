package controller

import (
	"net/http"
	"registration-service/dto"
	"registration-service/helper"
	"registration-service/service"

	"github.com/gin-gonic/gin"
)

type documentController struct {
	documentService service.DocumentService
}

type DocumentController interface {
	GetAllDocuments(ctx *gin.Context)
	GetDocumentByID(ctx *gin.Context)
	CreateDocument(ctx *gin.Context)
	UpdateDocument(ctx *gin.Context)
	DeleteDocument(ctx *gin.Context)
}

func NewProgramTypeController(documentService service.DocumentService) DocumentController {
	return &documentController{documentService: documentService}
}

func (c *documentController) GetAllDocuments(ctx *gin.Context) {
	pagReq := helper.Pagination(ctx)

	documents, metaData, err := c.documentService.FindAllDocuments(ctx, pagReq, nil)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, dto.Response{
			Status:  dto.STATUS_ERROR,
			Message: err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, dto.Response{
		Message:            dto.MESSAGE_DOCUMENT_GET_ALL_SUCCESS,
		Status:             dto.STATUS_SUCCESS,
		Data:               documents,
		PaginationResponse: &metaData,
	})
}

func (c *documentController) GetDocumentByID(ctx *gin.Context) {
	id := ctx.Param("id")
	document, err := c.documentService.FindDocumentById(ctx, id, nil)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, dto.Response{
			Status:  dto.STATUS_ERROR,
			Message: err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, dto.Response{
		Message: dto.MESSAGE_DOCUMENT_GET_SUCCESS,
		Status:  dto.STATUS_SUCCESS,
		Data:    document,
	})
}

func (c *documentController) CreateDocument(ctx *gin.Context) {

	file, err := ctx.FormFile("file")
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, dto.Response{
			Status:  dto.STATUS_ERROR,
			Message: err.Error(),
		})
		return
	}

	var request dto.DocumentRequest
	err = ctx.ShouldBindJSON(&request)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, dto.Response{
			Status:  dto.STATUS_ERROR,
			Message: err.Error(),
		})
		return
	}

	err = c.documentService.CreateDocument(ctx, request, file, nil)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, dto.Response{
			Status:  dto.STATUS_ERROR,
			Message: err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusCreated, dto.Response{
		Message: dto.MESSAGE_DOCUMENT_CREATE_SUCCESS,
		Status:  dto.STATUS_SUCCESS,
	})
}

func (c *documentController) UpdateDocument(ctx *gin.Context) {
	id := ctx.Param("id")

	file, err := ctx.FormFile("file")
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, dto.Response{
			Status:  dto.STATUS_ERROR,
			Message: err.Error(),
		})
		return
	}

	var request dto.UpdateDocumentRequest
	err = ctx.ShouldBindJSON(&request)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, dto.Response{
			Status:  dto.STATUS_ERROR,
			Message: err.Error(),
		})
		return
	}

	err = c.documentService.UpdateDocument(ctx, id, request, file, nil)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, dto.Response{
			Status:  dto.STATUS_ERROR,
			Message: err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, dto.Response{
		Message: dto.MESSAGE_DOCUMENT_UPDATE_SUCCESS,
		Status:  dto.STATUS_SUCCESS,
	})
}

func (c *documentController) DeleteDocument(ctx *gin.Context) {
	id := ctx.Param("id")
	err := c.documentService.DeleteDocument(ctx, id, nil)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, dto.Response{
			Status:  dto.STATUS_ERROR,
			Message: err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, dto.Response{
		Message: dto.MESSAGE_DOCUMENT_DELETE_SUCCESS,
		Status:  dto.STATUS_SUCCESS,
	})
}
