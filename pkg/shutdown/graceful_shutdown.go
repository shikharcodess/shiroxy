// Author: @ShikharY10
// Docs Author: ChatGPT

// Package shutdown provides utilities for handling graceful shutdown of applications,
// including data persistence and resource cleanup.
package shutdown

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"shiroxy/cmd/shiroxy/analytics"
	"shiroxy/cmd/shiroxy/domains"
	"shiroxy/pkg/color"
	"shiroxy/pkg/logger"
	"shiroxy/pkg/models"
	"sync"
	"syscall"

	"google.golang.org/protobuf/proto"
)

// HandleGracefulShutdown handles the application's graceful shutdown process,
// including cleanup, data persistence, and logging.
// Parameters:
//   - fromdefer: bool, indicates if the function is called from a deferred panic recovery.
//   - panicData: interface{}, contains data related to the panic if called during recovery.
//   - configuration: *models.Config, application configuration settings.
//   - storage: *domains.Storage, domain metadata storage instance.
//   - logHandler: *logger.Logger, logger instance for logging messages.
//   - analyticsConfiguration: *analytics.AnalyticsConfiguration, analytics configuration instance for managing analytics data.
//   - wg: *sync.WaitGroup, synchronization primitive to wait for all cleanup goroutines.
func HandleGracefulShutdown(fromdefer bool, panicData interface{}, configuration *models.Config, storage *domains.Storage, logHandler *logger.Logger, analyticsConfiguration *analytics.AnalyticsConfiguration, wg *sync.WaitGroup) {
	// Channel to listen for OS signals (e.g., SIGINT, SIGTERM).
	var sigs chan os.Signal
	if !fromdefer {
		sigs = make(chan os.Signal, 1)
		signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM) // Notify channel on interrupt/termination signals.
	}

	// Channel to indicate the completion status of the shutdown process.
	done := make(chan bool, 1)

	// Add to the WaitGroup to ensure all cleanup goroutines are completed.
	wg.Add(1)
	go func() {
		defer wg.Done()
		if !fromdefer {
			// If not called from a deferred panic recovery, wait for an OS signal.
			sig := <-sigs
			logHandler.Log(fmt.Sprintf("Signal Received: %s", sig.String()), "ShutDown", "👋")
		} else {
			// If called during panic recovery, log the panic data.
			logHandler.Log(fmt.Sprintf("Panic Recovered: %v", panicData), "ShutDown", "👋")
		}
		logHandler.Log("Performing Cleanup And Data Persistence...", "ShutDown", "🤞")

		// Perform the cleanup process in a separate goroutine.
		wg.Add(1)
		go func() {
			defer wg.Done()
			// Retrieve analytics data and stop analytics.
			analyticsData := <-analyticsConfiguration.ReadAnalyticsData
			analyticsConfiguration.StopAnalytics()

			// Marshal analytics data to JSON format.
			analyticsJsonMarshaledData, err := json.Marshal(analyticsData)
			if err != nil {
				logHandler.LogError(err.Error(), "Shutdown", "")
			}

			// Collect domain metadata for persistence.
			var domainMetadataArray []*domains.DomainMetadata
			if storage.DomainMetadata != nil {
				for _, domainMetadata := range storage.DomainMetadata {
					domainMetadataArray = append(domainMetadataArray, domainMetadata)
				}
			}

			// Create data persistence object.
			dataPersistance := domains.DataPersistance{
				Datetime: "", // Timestamp (if needed) can be set here.
				Domains:  domainMetadataArray,
			}

			// Marshal domain data into a protobuf format.
			storageData, err := proto.Marshal(&dataPersistance)
			if err != nil {
				logHandler.LogError(err.Error(), "Shutdown", "")
			}

			// Create a shutdown metadata object containing analytics and domain information.
			shutdownMetadata := ShutdownMetadata{
				DomainMetadata: storageData,
				SystemData:     analyticsJsonMarshaledData,
				WebhookSecret:  storage.WebhookSecret,
			}

			// Marshal the shutdown metadata into protobuf format.
			shutdownMarshaledMetadata, err := proto.Marshal(&shutdownMetadata)
			if err != nil {
				logHandler.LogError(err.Error(), "Shutdown", "")
			}

			// Encode the shutdown metadata to a Base64 string.
			base64EncodedData := base64.StdEncoding.EncodeToString(shutdownMarshaledMetadata)

			// Write the encoded data to a file for persistence.
			err = cleanup(fmt.Sprintf("%s/persistence.shiroxy", configuration.Default.DataPersistancePath), base64EncodedData)
			if err != nil {
				logHandler.Log(err.Error(), "Shutdown", "Error")
			}
			isShutDownSuccessful := err == nil
			done <- isShutDownSuccessful // Send the success status to the `done` channel.
		}()
	}()

	// Wait for the cleanup process to complete and retrieve the shutdown status.
	closeSignal := <-done
	var signalText string
	var signalEmoji string
	if closeSignal {
		signalText = fmt.Sprint(color.ColorGreen, "Success", color.ColorReset)
		signalEmoji = "😇"
	} else {
		signalText = fmt.Sprint(color.ColorRed, "Failed", color.ColorReset)
		signalEmoji = "😞"
	}
	logHandler.LogError(fmt.Sprintf("Exiting Gracefully... | %s", signalText), "Shutdown", signalEmoji)
	fmt.Println("")

	// Exit the application.
	os.Exit(0)
}

// cleanup writes the provided data to the specified file path for persistence.
// Parameters:
//   - filePath: string, the path of the file to write to.
//   - data: string, the data to write to the file.
//
// Returns:
//   - error: error if the file creation or writing fails.
func cleanup(filePath, data string) error {
	// Create the file at the specified path.
	f, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer f.Close()

	// Write the data to the file.
	_, err = f.WriteString(data)
	if err != nil {
		return err
	}

	return nil
}
