package loader

import (
	"fmt"
	"shiroxy/pkg/logger"
	"time"
)

type ProgressLoaderPayload struct {
	Close          chan bool
	Failed         bool
	Content        string
	MainContent    string
	SuccessMessage string
	FailedMessage  string
}

func (c *ProgressLoaderPayload) CloseLoader(failed bool, content string, waitBefore int) {
	time.Sleep(time.Duration(waitBefore) * time.Millisecond)
	c.Failed = failed
	c.Content = content
	c.Close <- false
}

func ProgressLoader(loaderController *ProgressLoaderPayload) {
	var mainLoop bool = true
	for mainLoop {
		select {
		case <-loaderController.Close:
			mainLoop = false
			if loaderController.Failed {
				logger.RedPrintln(loaderController.FailedMessage)
			} else {
				logger.GreenPrintln(loaderController.SuccessMessage)
			}
		default:
			for i := 0; i < len(loaderController.Content); i++ {
				if len(loaderController.Content) > 0 {
					fmt.Printf("\r%v%c", loaderController.MainContent, loaderController.Content[i%len(loaderController.Content)])
					time.Sleep(100 * time.Millisecond)
				} else {
					fmt.Print("\r", loaderController.MainContent)

				}

			}
		}
	}
}
