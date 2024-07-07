package defaults

import (
	"log"
	challenger "shiroxy/pkg/certificate/challenge"
	"shiroxy/pkg/loader"
	"shiroxy/pkg/logger"
	"shiroxy/pkg/models"
	"time"
)

func ExecuteDefault(configuration *models.Config, logHandler *logger.Logger) error {
	if configuration.Default.EnableDnsChallengeSolver {
		// Starting DNS challeng solver server
		loaderController := loader.ProgressLoaderPayload{
			Close:          make(chan bool),
			Failed:         false,
			Content:        `-\|/`,
			MainContent:    "* Starting DNS Challenge Solver Server : ",
			SuccessMessage: "Success 🎉",
			FailedMessage:  "Failed 😔",
		}
		go loader.ProgressLoader(&loaderController)
		time.Sleep(1 * time.Second)
		go func(loaderController *loader.ProgressLoaderPayload) {

			err := challenger.StartChallenger("5001", loaderController)
			if err != nil {
				loaderController.CloseLoader(true, err.Error(), 1000)
				log.Fatal(err)
			}
			loaderController.CloseLoader(false, "", 1000)
		}(&loaderController)
	}
	return nil
}
