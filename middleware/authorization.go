package middleware

import (
	"log"
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
		log.Println("Authorization Role Response:", res)

		var userRole string
		if role, ok := res["role"]; ok && role != nil {
			userRole, ok = role.(string)
			if !ok {
				log.Println("Error: role is not a string")
				c.AbortWithStatusJSON(http.StatusUnauthorized, dto.Response{
					Status:  dto.STATUS_ERROR,
					Message: dto.MESSAGE_UNAUTHORIZED,
				})
				return
			}
		} else {
			log.Println("Error: role not found in response")
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
			log.Println("Error: user does not have the required role")
			c.AbortWithStatusJSON(http.StatusUnauthorized, dto.Response{
				Status:  dto.STATUS_ERROR,
				Message: dto.MESSAGE_FORBIDDEN,
			})
			return
		}

		// save role to context
		c.Set("userRole", userRole)

		c.Next()
	}
}
