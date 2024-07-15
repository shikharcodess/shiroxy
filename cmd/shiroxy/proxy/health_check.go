package proxy

import (
	"net/http"
	"sync"
	"time"
)

type healthChecker struct {
	Servers               []*Server
	healthCheckTrigger    time.Duration
	stop                  chan bool
	changeTriggerInterval chan time.Duration
	wg                    *sync.WaitGroup
}

func NewHealthChecker(server []*Server, triggerInterval time.Duration, wg *sync.WaitGroup) *healthChecker {
	return &healthChecker{

		Servers:               server,
		healthCheckTrigger:    triggerInterval,
		stop:                  make(chan bool, 1),
		changeTriggerInterval: make(chan time.Duration, 1),
		wg:                    wg,
	}
}

func (hc *healthChecker) StartHealthCheck() {
	hc.wg.Add(1)
	go func() {
		defer hc.wg.Done()
		ticker := time.NewTicker(time.Second * 5)

		for {
			select {
			case <-ticker.C:
				for _, server := range hc.Servers {
					client := &http.Client{
						Timeout: 2 * time.Second,
					}
					resp, err := client.Head(server.URL.String())
					if err != nil || resp.StatusCode != http.StatusOK {
						server.Alive = false
					} else {
						server.Alive = true
					}
				}

			case newDuration := <-hc.changeTriggerInterval:
				ticker.Stop()
				ticker = time.NewTicker(newDuration)
			case <-hc.stop:
				ticker.Stop()
				return
			}
		}
	}()
}

func (hc *healthChecker) StopHealthChecker() {
	hc.stop <- true
}

func (hc *healthChecker) UpdateTicker(d time.Duration) {
	hc.changeTriggerInterval <- d
}
