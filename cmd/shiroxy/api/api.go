package api

import (
	"fmt"
	"shiroxy/cmd/shiroxy/analytics"
	"shiroxy/cmd/shiroxy/api/routes"
	"shiroxy/cmd/shiroxy/domains"
	"shiroxy/cmd/shiroxy/proxy"
	"shiroxy/cmd/shiroxy/types"
	"shiroxy/cmd/shiroxy/webhook"
	"shiroxy/pkg/logger"
	"shiroxy/pkg/models"
	"sync"

	"github.com/gin-gonic/gin"
)

// type ShiroxyAPI struct {
// 	healthChecker    *proxy.HealthChecker
// 	domainStorage    *domains.Storage
// 	analyticsHandler *analytics.AnalyticsConfiguration
// 	logHandler       *logger.Logger
// }

func StartShiroxyAPI(config models.Config, loadBalancer *proxy.LoadBalancer, domainStorage *domains.Storage, analyticsHandler *analytics.AnalyticsConfiguration, loghandler *logger.Logger, webhookHandler *webhook.WebhookHandler, wg *sync.WaitGroup) {
	apiContext := types.APIContext{
		WaitGroup:        wg,
		WebhookHandler:   webhookHandler,
		DomainStorage:    domainStorage,
		AnalyticsHandler: analyticsHandler,
		LogHandler:       loghandler,
		LoadBalancer:     loadBalancer,
	}

	gin.SetMode(gin.ReleaseMode)

	account := gin.Accounts{}
	account[config.Default.User.Email] = config.Default.User.Secret

	newRouter := gin.New()

	router := newRouter.Group("/v1")
	router.GET("/health", func(ctx *gin.Context) {
		ctx.JSON(200, gin.H{
			"status": "healthy",
		})
	})

	router.Use(gin.BasicAuth(account))

	routes.DomainRoutes(router, &apiContext)
	routes.AnalyticsRoutes(router, &apiContext)
	routes.BackendsRoutes(router, &apiContext)

	// Todo: remove this in final version ===============
	router.GET("/auth", func(ctx *gin.Context) {
		ctx.JSON(200, gin.H{
			"status": "auth verified",
		})
	})
	// ===================================================

	wg.Add(1)
	go func() {
		defer wg.Done()
		var adminAPIPort string
		if config.Default.AdminAPI.Port == "" {
			adminAPIPort = "2210"
		} else {
			adminAPIPort = config.Default.AdminAPI.Port
		}

		err := newRouter.Run(fmt.Sprintf(":%s", adminAPIPort))
		if err != nil {
			loghandler.LogError(fmt.Sprintf("error starting admin API, %s", err.Error()), "ADMIN API", "ERROR")
		}
	}()

	// DNS Solver Listener

	wg.Add(1)
	go func() {
		defer wg.Done()
		dnsRouter := gin.Default()
		routes.SSLRoutes(dnsRouter, domainStorage)
		err := dnsRouter.Run(":5002")
		if err != nil {
			loghandler.LogError(fmt.Sprintf("error starting admin API, %s", err.Error()), "ADMIN API", "ERROR")
		}
	}()
}
