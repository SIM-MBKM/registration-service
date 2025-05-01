package routes

import (
	"registration-service/controller"
	"registration-service/middleware"
	"registration-service/service"

	"github.com/gin-gonic/gin"
)

func RegistrationRoutes(router *gin.Engine, programTypeController controller.RegistrationController, userService service.UserManagementService) {
	registrationServiceRoute := router.Group("/registration-management/api/v1/registration")
	{
		registrationServiceRoute.POST("/all", middleware.AuthorizationRole(userService, []string{"ADMIN", "LO-MBKM"}), programTypeController.GetAllRegistrations)
		registrationServiceRoute.GET("/:id", programTypeController.GetRegistrationByID)
		registrationServiceRoute.POST("", programTypeController.CreateRegistration)
		registrationServiceRoute.PUT("/:id", programTypeController.UpdateRegistration)
		registrationServiceRoute.DELETE("/:id", programTypeController.DeleteRegistration)
		registrationServiceRoute.POST("/advisor", middleware.AuthorizationRole(userService, []string{"DOSEN PEMBIMBING"}), programTypeController.GetRegistrationsByAdvisor)
		registrationServiceRoute.POST("/student", middleware.AuthorizationRole(userService, []string{"MAHASISWA"}), programTypeController.GetRegistrationsByStudent)
		registrationServiceRoute.POST("/approval", middleware.AuthorizationRole(userService, []string{"ADMIN", "LO-MBKM", "DOSEN PEMBIMBING"}), programTypeController.ApproveRegistration)
		registrationServiceRoute.GET("/:id/transcript", programTypeController.GetRegistrationTranscript)
		registrationServiceRoute.POST("/student/transcripts", middleware.AuthorizationRole(userService, []string{"MAHASISWA"}), programTypeController.GetStudentRegistrationsWithTranscripts)
		registrationServiceRoute.POST("/student/syllabuses", middleware.AuthorizationRole(userService, []string{"MAHASISWA"}), programTypeController.GetStudentRegistrationsWithSyllabuses)
		registrationServiceRoute.POST("/student/matching", middleware.AuthorizationRole(userService, []string{"MAHASISWA"}), programTypeController.GetStudentRegistrationsWithMatching)
		registrationServiceRoute.GET("/check-eligibility", middleware.AuthorizationRole(userService, []string{"MAHASISWA"}), programTypeController.CheckRegistrationEligibility)
	}
}
