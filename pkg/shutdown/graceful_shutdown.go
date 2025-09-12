// Author: @ShikharY10
// Docs Author: Copilot

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

// HandleGracefulShutdown manages the graceful shutdown process of the application.
// It handles both signal-based and panic-based shutdowns, performs cleanup, persists data,
// and logs the shutdown status.
//
// Parameters:
//   - fromdefer: Indicates if the shutdown was triggered from a deferred panic recovery.
//   - panicData: The data recovered from a panic, if any.
//   - configuration: Application configuration containing persistence paths and settings.
//   - storage: Storage object containing domain metadata and webhook secrets.
//   - logHandler: Logger instance for logging shutdown events and errors.
//   - analyticsConfiguration: Analytics configuration for reading and stopping analytics data.
//   - wg: WaitGroup to synchronize goroutines during shutdown.
//
// The function listens for OS signals (SIGINT, SIGTERM) unless triggered by a panic,
// marshals and persists domain and analytics data, and logs the outcome before exiting.
//
// cleanup writes the provided data string to the specified file path.
// Returns an error if file creation or writing fails.
func HandleGracefulShutdown(fromdefer bool, panicData interface{}, configuration *models.Config, storage *domains.Storage, logHandler *logger.Logger, analyticsConfiguration *analytics.AnalyticsConfiguration, wg *sync.WaitGroup) {
	shiroxyEnvionment := os.Getenv("SHIROXY_ENVIRONMENT")
	if shiroxyEnvionment == "" {
		shiroxyEnvionment = "dev"
	}

	var sigs chan os.Signal
	if !fromdefer {
		sigs = make(chan os.Signal, 1)
		signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	}

	done := make(chan bool, 1)

	wg.Add(1)
	go func() {
		defer wg.Done()
		if !fromdefer {
			sig := <-sigs
			logHandler.Log(fmt.Sprintf("Signal Received: %s", sig.String()), "ShutDown", "👋")
		} else {
			logHandler.Log(fmt.Sprintf("Panic Recovered: %v", panicData), "ShutDown", "👋")
		}
		logHandler.Log("Performing Cleanup And Data Persistence...", "ShutDown", "🤞")

		wg.Add(1)
		go func() {
			defer wg.Done()
			analyticsData := <-analyticsConfiguration.ReadAnalyticsData
			analyticsConfiguration.StopAnalytics()

			analyticsJsonMarshaledData, err := json.Marshal(analyticsData)
			if err != nil {
				logHandler.LogError(err.Error(), "Shutdown", "")
			}

			var domainMetadataArray []*domains.DomainMetadata
			if storage.DomainMetadata != nil {
				for _, domainMetadata := range storage.DomainMetadata {
					domainMetadataArray = append(domainMetadataArray, domainMetadata)
				}
			}

			dataPersistance := domains.DataPersistance{
				Datetime: "",
				Domains:  domainMetadataArray,
			}

			storageData, err := proto.Marshal(&dataPersistance)
			if err != nil {
				logHandler.LogError(err.Error(), "Shutdown", "")
			}

			shutdownMetadata := ShutdownMetadata{
				DomainMetadata: storageData,
				SystemData:     analyticsJsonMarshaledData,
				WebhookSecret:  storage.WebhookSecret,
			}

			shutdownMarshaledMetadata, err := proto.Marshal(&shutdownMetadata)
			if err != nil {
				logHandler.LogError(err.Error(), "Shutdown", "")
			}

			base64EncodedData := base64.StdEncoding.EncodeToString(shutdownMarshaledMetadata)

			err = cleanup(fmt.Sprintf("%s/%s-persistence.shiroxy", configuration.Default.DataPersistancePath, shiroxyEnvionment), base64EncodedData)
			if err != nil {
				logHandler.Log(err.Error(), "Shutdown", "Error")
			}
			isShutDownSuccessful := err == nil
			done <- isShutDownSuccessful
		}()
	}()

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

	os.Exit(0)
}

// cleanup creates or truncates the file at the specified filePath and writes the provided data string to it.
// It returns an error if the file cannot be created or written to.
func cleanup(filePath, data string) error {
	f, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = f.WriteString(data)
	if err != nil {
		return err
	}

	return nil
}
