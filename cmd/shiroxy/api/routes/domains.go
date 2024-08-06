package routes

import (
	"shiroxy/cmd/shiroxy/api/controllers"
	"shiroxy/cmd/shiroxy/api/middlewares"
	"shiroxy/cmd/shiroxy/domains"
	"shiroxy/cmd/shiroxy/webhook"

	"github.com/gin-gonic/gin"
)

func DomainRoutes(router *gin.RouterGroup, middleware *middlewares.Middlewares, storage *domains.Storage, webhookHandler *webhook.WebhookHandler) {
	domainController := controllers.DomainController{
		Storage:        storage,
		Middlewares:    middleware,
		WebhookHandler: webhookHandler,
	}
	domain := router.Group("/domain")
	// domain.Use(middleware.CheckAccess())

	domain.POST("/register", domainController.RegisterDomain)
	domain.POST("/forcessl", domainController.ForceSSL)
	domain.POST("/")
}
