package routes

import (
	"shiroxy/cmd/shiroxy/api/controllers"
	"shiroxy/cmd/shiroxy/api/middlewares"
	"shiroxy/cmd/shiroxy/types"

	"github.com/gin-gonic/gin"
)

func DomainRoutes(router *gin.RouterGroup, apiContext *types.APIContext) error {
	domainMiddleware, err := middlewares.InitializeMiddleware(apiContext.LogHandler, "")
	if err != nil {
		return err
	}

	domainController := controllers.DomainController{
		Middlewares: domainMiddleware,
		Context:     apiContext,
	}
	domain := router.Group("/domain")
	// domain.Use(middleware.CheckAccess())

	domain.POST("/register", domainController.RegisterDomain)
	domain.POST("/forcessl", domainController.ForceSSL)
	domain.POST("/")

	return nil
}
