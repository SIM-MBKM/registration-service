package routes

import (
	"registration-service/controller"
	"registration-service/middleware"
	"registration-service/service"

	"github.com/gin-gonic/gin"
)

func RegistrationRoutes(router *gin.Engine, programTypeController controller.RegistrationController, userService service.UserManagementService) {
	registrationServiceRoute := router.Group("/registration-management/api/v1")
	{
		registrationRoutes := registrationServiceRoute.Group("/registration")
		{
			registrationRoutes.POST("/all", middleware.AuthorizationRole(userService, []string{"ADMIN", "LO-MBKM"}), programTypeController.GetAllRegistrations)
			registrationRoutes.GET("/:id", programTypeController.GetRegistrationByID)
			registrationRoutes.POST("/", programTypeController.CreateRegistration)
			registrationRoutes.PUT("/:id", programTypeController.UpdateRegistration)
			registrationRoutes.DELETE("/:id", programTypeController.DeleteRegistration)
			registrationRoutes.POST("/advisor", middleware.AuthorizationRole(userService, []string{"DOSEN PEMBIMBING"}), programTypeController.GetRegistrationsByAdvisor)
			registrationRoutes.POST("/student", middleware.AuthorizationRole(userService, []string{"MAHASISWA"}), programTypeController.GetRegistrationsByStudent)
			registrationRoutes.POST("/approval", middleware.AuthorizationRole(userService, []string{"ADMIN", "LO-MBKM", "DOSEN PEMBIMBING"}), programTypeController.ApproveRegistration)
		}
	}
}
