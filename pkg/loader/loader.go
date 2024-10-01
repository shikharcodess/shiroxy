package loader

import (
	"fmt"
	"shiroxy/pkg/logger"
	"sync"
	"time"
)

type ProgressLoaderPayload struct {
	Lock           *sync.RWMutex
	Close          chan bool
	Failed         bool
	Content        string
	MainContent    string
	SuccessMessage string
	FailedMessage  string
}

func (c *ProgressLoaderPayload) CloseLoader(failed bool, content string, waitBefore int) {
	time.Sleep(time.Duration(waitBefore) * time.Millisecond)
	c.Lock.Lock()
	c.Failed = failed
	c.Content = content
	c.Lock.Unlock()
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
			loaderController.Lock.RLock()
			lC := *loaderController
			loaderController.Lock.RUnlock()
			for i := 0; i < len(lC.Content); i++ {
				if len(lC.Content) > 0 {
					fmt.Printf("\r%v%c", lC.MainContent, lC.Content[i%len(lC.Content)])
					time.Sleep(100 * time.Millisecond)
				} else {
					fmt.Print("\r", lC.MainContent)

				}
			}
		}
	}
}
