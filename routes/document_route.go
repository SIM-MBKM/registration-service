package routes

import (
	"registration-service/controller"

	"github.com/gin-gonic/gin"
)

func DocumentRoutes(router *gin.Engine, documentController controller.DocumentController) {
	documentServiceRoute := router.Group("/registration-management/api")
	{
		documentRoutes := documentServiceRoute.Group("/document")
		{
			documentRoutes.GET("/", documentController.GetAllDocuments)
			documentRoutes.GET("/:id", documentController.GetDocumentByID)
			documentRoutes.POST("/", documentController.CreateDocument)
			documentRoutes.PUT("/:id", documentController.UpdateDocument)
			documentRoutes.DELETE("/:id", documentController.DeleteDocument)
		}
	}
}
