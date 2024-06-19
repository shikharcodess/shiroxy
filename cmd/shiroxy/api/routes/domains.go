package routes

import (
	"shiroxy/cmd/shiroxy/api/controllers"

	"github.com/gin-gonic/gin"
)

func DomainRoutes(router *gin.Engine) {
	domainController := controllers.DomainController{}
	domain := router.Group("/domain")
	domain.POST("/register", domainController.RegisterDomain)
}
