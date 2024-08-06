package routes

import (
	"shiroxy/cmd/shiroxy/analytics"
	"shiroxy/cmd/shiroxy/api/controllers"
	"shiroxy/cmd/shiroxy/api/middlewares"
	"shiroxy/cmd/shiroxy/domains"
	"shiroxy/cmd/shiroxy/proxy"
	"shiroxy/cmd/shiroxy/webhook"

	"github.com/gin-gonic/gin"
)

func AnalyticsRoutes(router *gin.RouterGroup, analyticsHandler *analytics.AnalyticsConfiguration, middleware *middlewares.Middlewares, storage *domains.Storage, webhookHandler *webhook.WebhookHandler, healthChecker *proxy.HealthChecker) {
	analyticsController := controllers.AnalyticsController{
		Storage:          storage,
		AnalyticsHandler: analyticsHandler,
		Middlewares:      middleware,
		WebhookHandler:   webhookHandler,
		HealthChecker:    healthChecker,
	}
	domain := router.Group("/analytics")
	// domain.Use(middleware.CheckAccess())

	domain.GET("/domains", analyticsController.FetchDomainAnalytics)
	domain.GET("/systems", analyticsController.FetchSystemAnalytics)
}
