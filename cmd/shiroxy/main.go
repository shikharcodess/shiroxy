// Shiroxy main entry point
// Author: @ShikharY10

package main

import (
	"fmt"
	"log"
	"shiroxy/cmd/shiroxy/analytics"
	"shiroxy/cmd/shiroxy/domains"
	"shiroxy/cmd/shiroxy/proxy"
	"shiroxy/pkg/cli"
	"shiroxy/pkg/logger"
	"shiroxy/pkg/shutdown"
	"sync"
	"time"
)

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
	storageHandler, err := domains.InitializeStorage(&configuration.Default.Storage)
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

	proxy.StartShiroxyHandler(configuration, storageHandler, logHandler, &wg)

	wg.Wait()
}
