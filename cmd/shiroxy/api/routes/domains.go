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

	domain.POST("/", domainController.RegisterDomain)
	domain.PATCH("/:domain", domainController.UpdateDomain)
	domain.PATCH("/:domain/retryssl", domainController.ForceSSL)
	domain.GET("/:domain", domainController.FetchDomainInfo)
	domain.DELETE("/:domain", domainController.RemoveDomain)

	return nil
}
