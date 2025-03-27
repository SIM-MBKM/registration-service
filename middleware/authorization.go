package middleware

import (
	"registration-service/service"

	"github.com/gin-gonic/gin"
)

func Authorization(userService service.UserManagementService) gin.HandlerFunc {
	return func(c *gin.Context) {

		// if userID != film.CreatorID {
		// 	c.AbortWithStatusJSON(http.StatusUnauthorized, dto.Response{
		// 		Status:  dto.STATUS_ERROR,
		// 		Message: dto.MESSAGE_UNAUTHORIZED,
		// 	})
		// }
		c.Next()
	}
}
