package routes

import (
	"shiroxy/cmd/shiroxy/api/controllers"
	"shiroxy/cmd/shiroxy/api/middlewares"
	"shiroxy/cmd/shiroxy/types"

	"github.com/gin-gonic/gin"
)

func AnalyticsRoutes(router *gin.RouterGroup, apiContext *types.APIContext) error {
	analyticsMiddleware, err := middlewares.InitializeMiddleware(apiContext.LogHandler, "")
	if err != nil {
		return err
	}
	analyticsController := controllers.AnalyticsController{
		Context:     apiContext,
		Middlewares: analyticsMiddleware,
	}
	domain := router.Group("/analytics")
	// domain.Use(middleware.CheckAccess())

	domain.GET("/domains", analyticsController.FetchDomainAnalytics)
	domain.GET("/systems", analyticsController.FetchSystemAnalytics)

	return nil
}
