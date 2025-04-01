package middleware

import (
	"net/http"
	"registration-service/dto"
	"registration-service/service"

	"github.com/gin-gonic/gin"
)

func AuthorizationRole(userService service.UserManagementService, role []string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// get header token
		token := c.GetHeader("Authorization")
		if token == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, dto.Response{
				Status:  dto.STATUS_ERROR,
				Message: dto.MESSAGE_UNAUTHORIZED,
			})
			return
		}

		res := userService.GetUserRole("GET", token)

		var userRole string
		if role, ok := res["role"]; ok && role != nil {
			userRole, ok = role.(string)
			if !ok {
				c.AbortWithStatusJSON(http.StatusUnauthorized, dto.Response{
					Status:  dto.STATUS_ERROR,
					Message: dto.MESSAGE_UNAUTHORIZED,
				})
				return
			}
		} else {
			c.AbortWithStatusJSON(http.StatusUnauthorized, dto.Response{
				Status:  dto.STATUS_ERROR,
				Message: dto.MESSAGE_UNAUTHORIZED,
			})
			return
		}

		// checking if userRole is in role
		isRole := false
		for _, r := range role {
			if userRole == r {
				isRole = true
				break
			}
		}

		if !isRole {
			c.AbortWithStatusJSON(http.StatusUnauthorized, dto.Response{
				Status:  dto.STATUS_ERROR,
				Message: dto.MESSAGE_FORBIDDEN,
			})
			return
		}

		c.Next()
	}
}
