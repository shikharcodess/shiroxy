package api

import (
	"fmt"
	"shiroxy/cmd/shiroxy/analytics"
	"shiroxy/cmd/shiroxy/api/middlewares"
	"shiroxy/cmd/shiroxy/api/routes"
	"shiroxy/cmd/shiroxy/domains"
	"shiroxy/cmd/shiroxy/proxy"
	"shiroxy/pkg/logger"
	"shiroxy/pkg/models"
	"sync"

	"github.com/gin-gonic/gin"
)

type shiroxyAPI struct {
	healthChecker    *proxy.HealthChecker
	domainStorage    *domains.Storage
	analyticsHandler *analytics.AnalyticsConfiguration
	logHandler       *logger.Logger
}

func StartShiroxyAPI(config models.Config, healthChecker *proxy.HealthChecker, domainStorage *domains.Storage, analyticsHandler *analytics.AnalyticsConfiguration, loghandler *logger.Logger, wg *sync.WaitGroup) {
	apiHndler := shiroxyAPI{
		healthChecker:    healthChecker,
		domainStorage:    domainStorage,
		analyticsHandler: analyticsHandler,
		logHandler:       loghandler,
	}

	middleware, err := middlewares.InitializeMiddleware(apiHndler.logHandler, "api")
	if err != nil {
		loghandler.LogError("failed to start shiroxy dynamic API", "api", "error")
		return
	}

	gin.SetMode(gin.ReleaseMode)

	account := gin.Accounts{}
	account[config.Default.User.Email] = config.Default.User.Secret

	router := gin.New()

	router.GET("/health", func(ctx *gin.Context) {
		ctx.JSON(200, gin.H{
			"status": "healthy",
		})
	})

	router.Use(gin.BasicAuth(account))

	routes.DomainRoutes(router, middleware, domainStorage)
	// routes.SSLRoutes(router, domainStorage)

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

		err := router.Run(fmt.Sprintf(":%s", adminAPIPort))
		if err != nil {
			loghandler.LogError(fmt.Sprintf("error starting admin API, %s", err.Error()), "ADMIN API", "ERROR")
		}
	}()

	// DNS Solver Listener

	wg.Add(1)
	go func() {
		dnsRouter := gin.Default()
		routes.SSLRoutes(dnsRouter, domainStorage)
		err := dnsRouter.Run(":5002")
		if err != nil {
			loghandler.LogError(fmt.Sprintf("error starting admin API, %s", err.Error()), "ADMIN API", "ERROR")
		}
	}()
}
