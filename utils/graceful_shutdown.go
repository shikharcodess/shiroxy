package utils

import (
	"fmt"
	"os"
	"os/signal"
	"shiroxy/cmd/shiroxy/analytics"
	"shiroxy/cmd/shiroxy/domains"
	"shiroxy/pkg/color"
	"shiroxy/pkg/logger"
	"shiroxy/pkg/models"
	"syscall"
	"time"
)

func HandleGracefulShutdown(fromdefer bool, panicData interface{}, configuration *models.Config, storage *domains.Storage, logHandler *logger.Logger, analyticsConfiguration *analytics.AnalyticsConfiguration) {
	var sigs chan os.Signal
	if !fromdefer {
		sigs := make(chan os.Signal, 1)
		signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	}
	done := make(chan bool, 1)

	if !fromdefer {
		go func() {
			sig := <-sigs
			logHandler.LogWarning(fmt.Sprintf("Signal Received: %s", sig.String()), "ShutDown Handler", "👋")
			logHandler.LogWarning("Performing Cleanup And Data Persistance", "ShutDown Handler", "🤞")
			// Cleanup function called here
			err := cleanup()
			var isShutDownSuccessful bool
			if err != nil {
				isShutDownSuccessful = false
			} else {
				isShutDownSuccessful = true
			}

			done <- isShutDownSuccessful
		}()
	} else {
		logHandler.LogWarning(fmt.Sprintf("Signal Received: %s", panicData), "ShutDown Handler", "👋")
		logHandler.LogWarning("Performing Cleanup And Data Persistance", "ShutDown Handler", "🤞")
		// Cleanup function called here
		err := cleanup()
		var isShutDownSuccessful bool
		if err != nil {
			isShutDownSuccessful = false
		} else {
			isShutDownSuccessful = true
		}

		done <- isShutDownSuccessful
	}

	if !fromdefer {
		go func() {
			// <-done
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
			logHandler.LogError(fmt.Sprintf("Exiting Gracefully... | %s", signalText), "Shutdown Handler", signalEmoji)
			fmt.Println("")
			os.Exit(0)
		}()
	} else {
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
		logHandler.LogError(fmt.Sprintf("Exiting Gracefully... | %s", signalText), "Shutdown Handler", signalEmoji)
		fmt.Println("")
		os.Exit(0)
	}
}

func cleanup() error {

	f, err := os.Create("cleanup.txt")
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = f.WriteString("Cleanup completed at: " + time.Now().Format(time.RFC1123))
	if err != nil {
		return err
	}

	return nil
}
