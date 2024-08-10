package routes

import (
	"shiroxy/cmd/shiroxy/api/controllers"
	"shiroxy/cmd/shiroxy/types"

	"github.com/gin-gonic/gin"
)

func BackendsRoutes(router *gin.RouterGroup, context *types.APIContext) {
	backendsController := controllers.BackendController{
		Context: context,
	}
	backend := router.Group("/backends")
	backend.GET("/", backendsController.FetchAllBackendServers)
	backend.GET("/register", backendsController.RegisterNewBackendServer)
	backend.GET("/remove", backendsController.RemoveBackendServer)

}
