package routes

import (
	"registration-service/controller"

	"github.com/gin-gonic/gin"
)

func RegistrationRoutes(router *gin.Engine, programTypeController controller.RegistrationController) {
	registrationServiceRoute := router.Group("/registration-management/api")
	{
		registrationRoutes := registrationServiceRoute.Group("/registration")
		{
			registrationRoutes.GET("/", programTypeController.GetAllRegistrations)
			registrationRoutes.GET("/:id", programTypeController.GetRegistrationByID)
			registrationRoutes.POST("/", programTypeController.CreateRegistration)
			registrationRoutes.PUT("/:id", programTypeController.UpdateRegistration)
			registrationRoutes.DELETE("/:id", programTypeController.DeleteRegistration)
		}
	}
}
