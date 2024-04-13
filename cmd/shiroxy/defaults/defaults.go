package defaults

import (
	"fmt"
	"log"
	challenger "shiroxy/pkg/certificate/challenge"
	"shiroxy/pkg/configuration"
	"shiroxy/pkg/loader"
	"time"
)

func ExecuteDefault(configuration *configuration.Config) {
	if configuration.Default.Enable_Dns_Challenge_Solver {
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
			fmt.Println("closing loader on success")
			loaderController.CloseLoader(false, "", 1000)
		}(&loaderController)
	}
}
