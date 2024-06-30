package routes

import (
	"shiroxy/cmd/shiroxy/api/controllers"
	"shiroxy/cmd/shiroxy/api/middlewares"
	"shiroxy/cmd/shiroxy/domains"

	"github.com/gin-gonic/gin"
)

func DomainRoutes(router *gin.Engine, middleware *middlewares.Middlewares, storage *domains.Storage) {
	domainController := controllers.DomainController{
		Storage:     storage,
		Middlewares: middleware,
	}
	domain := router.Group("/domain")
	domain.Use(middleware.CheckAccess())

	domain.POST("/register", domainController.RegisterDomain)
	domain.POST("/forcessl", domainController.ForceSSL)
	domain.POST("/")
}
