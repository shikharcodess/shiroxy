package routes

import (
	"shiroxy/cmd/shiroxy/api/controllers"
	"shiroxy/cmd/shiroxy/api/middlewares"
	"shiroxy/cmd/shiroxy/types"

	"github.com/gin-gonic/gin"
)

func BackendsRoutes(router *gin.RouterGroup, context *types.APIContext) error {
	middleware, err := middlewares.InitializeMiddleware(context.LogHandler, "")
	if err != nil {
		return err
	}

	backendsController := controllers.BackendController{
		Context:     context,
		Middlewares: middleware,
	}
	backend := router.Group("/backends")
	backend.GET("/", backendsController.FetchAllBackendServers)
	backend.GET("/register", backendsController.RegisterNewBackendServer)
	backend.GET("/remove", backendsController.RemoveBackendServer)

	return nil
}
