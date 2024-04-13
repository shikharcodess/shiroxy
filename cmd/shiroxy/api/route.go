package shiroxyhttp

import (
	"github.com/gin-gonic/gin"
)

func UserRoutes(router *gin.Engine) {
	user := router.Group("/user")
	{
		user.GET("/.well-known/acme-challenge/:filename", CreateChallenge)
	}
}
