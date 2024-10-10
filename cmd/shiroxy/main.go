// Shiroxy main entry point
// Author: @ShikharY10

package main

import (
	"fmt"
	"log"
	"os"
	"runtime/debug"
	"shiroxy/cmd/shiroxy/analytics"
	"shiroxy/cmd/shiroxy/api"
	"shiroxy/cmd/shiroxy/domains"
	"shiroxy/cmd/shiroxy/proxy"
	"shiroxy/cmd/shiroxy/webhook"
	"shiroxy/pkg/cli"
	"shiroxy/pkg/logger"
	"shiroxy/pkg/shutdown"
	"sync"
	"time"
)

var ACME_SERVER_URL string      // URL of the ACME server for certificate generation.
var INSECURE_SKIP_VERIFY string // Flag to skip SSL verification (for testing purposes).
var VERSION string              // Shiroxy application version.
var MODE string

// main is the entry point of the Shiroxy application. It sets up logging, reads configuration,
// initializes storage, starts analytics, handles graceful shutdown, and starts the API and proxy services.
func main() {
	if MODE == "" {
		MODE = "dev"
	}

	os.Setenv("SHIROXY_ENVIRONMENT", MODE)
	wg := sync.WaitGroup{} // WaitGroup to manage concurrency and wait for goroutines to finish.

	// Starting Logger
	logHandler, err := logger.StartLogger(nil)
	if err != nil {
		log.Fatal(err) // Terminate the application if the logger fails to start.
	}

	// Reading Configuration
	configuration, err := cli.Execute()
	if err != nil {
		log.Fatal(err) // Terminate if configuration reading fails.
	}

	// Injecting logger with configuration
	logHandler.InjectLogConfig(&configuration.Logging)
	if err != nil {
		log.Fatal(err) // Terminate if logger configuration injection fails.
	}

	// Starting storage service for storing domain and user data
	storageHandler, err := domains.InitializeStorage(&configuration.Default.Storage, ACME_SERVER_URL, INSECURE_SKIP_VERIFY, &wg)
	if err != nil {
		logHandler.LogError(err.Error(), "Startup", "main")
	}

	// Set analytics collection interval; default to 10 if not specified in the configuration.
	var collectionInterval int
	if configuration.Default.Analytics.CollectionInterval == 0 {
		collectionInterval = 10
	} else {
		collectionInterval = configuration.Default.Analytics.CollectionInterval
	}

	// Starting analytics service
	analyticsConfiguration, err := analytics.StartAnalytics(time.Duration(time.Second*time.Duration(collectionInterval)), logHandler, &wg)
	if err != nil {
		logHandler.LogError(err.Error(), "Startup", "main")
	}

	// Loading data that was persisted during the last shutdown or failover
	shutdown.LoadShutdownPersistence(*logHandler, configuration, storageHandler, analyticsConfiguration)

	// Starting service that handles graceful shutdown, performing cleanup and data persistence
	wg.Add(1)
	go func() {
		defer wg.Done()
		shutdown.HandleGracefulShutdown(false, nil, configuration, storageHandler, logHandler, analyticsConfiguration, &wg)
	}()

	// Deferred function to handle any panic that occurs in the main function, enabling graceful shutdown.
	defer func() {
		if r := recover(); r != nil {
			if configuration.Default.Mode == "dev" {
				fmt.Printf("Panic occurred: %v\n", r)
				debug.PrintStack()
			}
			shutdown.HandleGracefulShutdown(true, r, configuration, storageHandler, logHandler, analyticsConfiguration, &wg)
		}
	}()

	// Starting the webhook handler service
	webhookHandler, err := webhook.StartWebhookHandler(configuration.Webhook, logHandler, &wg, storageHandler.WebhookSecret)
	if err != nil {
		logHandler.LogError(err.Error(), "Webhook", "main")
	}

	// Starting the proxy load balancer (Shiroxy handler)
	laodBalancer, err := proxy.StartShiroxyHandler(configuration, storageHandler, webhookHandler, logHandler, &wg)
	if err != nil {
		panic(err) // Panic if the load balancer fails to start.
	}

	// Starting the Shiroxy API service
	api.StartShiroxyAPI(configuration, laodBalancer, storageHandler, analyticsConfiguration, logHandler, webhookHandler, &wg)

	// Wait for all goroutines to finish
	wg.Wait()
}
