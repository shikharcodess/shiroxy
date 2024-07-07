package shutdown

import (
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
	"time"

	"google.golang.org/protobuf/proto"
)

func HandleGracefulShutdown(fromdefer bool, panicData interface{}, configuration *models.Config, storage *domains.Storage, logHandler *logger.Logger, analyticsConfiguration *analytics.AnalyticsConfiguration, wg *sync.WaitGroup) {
	// Channel to listen for OS signals
	var sigs chan os.Signal
	if !fromdefer {
		sigs = make(chan os.Signal, 1)
		signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	}

	// Channel to indicate shutdown status
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

		// Cleanup function called here
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
				Datetime: time.Now().String(),
				Domains:  domainMetadataArray,
			}

			storageData, err := proto.Marshal(&dataPersistance)
			if err != nil {
				logHandler.LogError(err.Error(), "Shutdown", "")
			}

			shutdownMetadata := ShutdownMetadata{
				DomainMetadata: string(storageData),
				SystemData:     string(analyticsJsonMarshaledData),
			}

			shutdownMarshaledMetadata, err := proto.Marshal(&shutdownMetadata)
			if err != nil {
				logHandler.LogError(err.Error(), "Shutdown", "")
			}

			err = cleanup(fmt.Sprintf("%s/persistance.shiroxy", configuration.Default.DataPersistancePath), string(shutdownMarshaledMetadata))
			if err != nil {
				logHandler.Log(err.Error(), "Shutdown", "Error")
			}
			isShutDownSuccessful := err == nil
			done <- isShutDownSuccessful
		}()
	}()

	// Wait for cleanup to complete
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
