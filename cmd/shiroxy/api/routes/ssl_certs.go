package routes

import (
	"shiroxy/cmd/shiroxy/api/controllers"
	"shiroxy/cmd/shiroxy/domains"

	"github.com/gin-gonic/gin"
)

func SSLRoutes(router *gin.Engine, storage *domains.Storage) {
	dnsChallengeSolver := controllers.GetDNSChallengeSolver(storage)
	user := router.Group("/")
	{
		user.GET("/.well-known/acme-challenge/:filename", dnsChallengeSolver.SolveDNSChallenge)
	}
}
