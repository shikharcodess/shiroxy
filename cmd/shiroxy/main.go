// Shiroxy main entry point
// Author: @ShikharY10

package main

import (
	"log"
	"shiroxy/cmd/shiroxy/analytics"
	"shiroxy/cmd/shiroxy/storage"
	"shiroxy/pkg/cli"
	"shiroxy/pkg/logger"
	"shiroxy/utils"
	"time"
)

func main() {
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
	storageHandler, err := storage.InitializeStorage(&configuration.Default.Storage)
	if err != nil {
		logHandler.LogError(err.Error(), "Startup", "main")
	}

	analyticsConfiguration, err := analytics.StartAnalytics(time.Duration(time.Second * time.Duration(configuration.Default.Analytics.CollectionInterval)))
	if err != nil {
		logHandler.LogError(err.Error(), "Startup", "main")
	}

	// Starting service that take care of graceful shutdown by doing cleanup and data persistance
	utils.HandleGracefulShutdown(false, nil, configuration, storageHandler, logHandler, analyticsConfiguration)
	defer func() {
		if r := recover(); r != nil {
			utils.HandleGracefulShutdown(true, r, configuration, storageHandler, logHandler, analyticsConfiguration)
		}
	}()

}
