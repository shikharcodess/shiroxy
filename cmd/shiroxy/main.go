// Shiroxy main entry point
// Author: @ShikharY10

package main

import (
	"fmt"
	"log"
	"shiroxy/cmd/shiroxy/analytics"
	"shiroxy/cmd/shiroxy/api"
	"shiroxy/cmd/shiroxy/domains"
	"shiroxy/cmd/shiroxy/proxy"
	"shiroxy/pkg/cli"
	"shiroxy/pkg/logger"
	"shiroxy/pkg/shutdown"
	"sync"
	"time"
)

var ACME_SERVER_URL string
var INSECURE_SKIP_VERIFY string
var VERSION string

func main() {
	wg := sync.WaitGroup{}

	// Starting Logger
	logHandler, err := logger.StartLogger(nil)
	if err != nil {
		log.Fatal(err)
	}

	// Reading Configuration
	configuration, err := cli.Execute()
	if err != nil {
		log.Fatal(err)
	}

	// Injecting logger with configuration
	logHandler.InjectLogConfig(&configuration.Logging)
	if err != nil {
		log.Fatal(err)
	}

	// Starting storage service for storing domain and user data
	storageHandler, err := domains.InitializeStorage(&configuration.Default.Storage, ACME_SERVER_URL, INSECURE_SKIP_VERIFY, &wg)
	if err != nil {
		logHandler.LogError(err.Error(), "Startup", "main")
	}

	var collectionInterval int
	if configuration.Default.Analytics.CollectionInterval == 0 {
		collectionInterval = 10
	} else {
		collectionInterval = configuration.Default.Analytics.CollectionInterval
	}

	analyticsConfiguration, err := analytics.StartAnalytics(time.Duration(time.Second*time.Duration(collectionInterval)), logHandler, &wg)
	if err != nil {
		logHandler.LogError(err.Error(), "Startup", "main")
	}

	// Loading data that was persisted during last shutdown or failover
	shutdown.LoadShutdownPersistence(*logHandler, configuration, storageHandler, analyticsConfiguration)

	// Starting service that take care of graceful shutdown by doing cleanup and data persistance
	wg.Add(1)
	go func() {
		defer wg.Done()
		shutdown.HandleGracefulShutdown(false, nil, configuration, storageHandler, logHandler, analyticsConfiguration, &wg)
	}()
	defer func() {
		fmt.Println("Shutdown defer Fired...")
		if r := recover(); r != nil {
			shutdown.HandleGracefulShutdown(true, r, configuration, storageHandler, logHandler, analyticsConfiguration, &wg)
		}
	}()

	laodBalancer, err := proxy.StartShiroxyHandler(configuration, storageHandler, logHandler, &wg)
	if err != nil {
		panic(err)
	}

	api.StartShiroxyAPI(*configuration, laodBalancer.HealthChecker, storageHandler, analyticsConfiguration, logHandler, &wg)

	wg.Wait()
}
