package proxy

import (
	"net/http"
	"shiroxy/cmd/shiroxy/webhook"
	"sync"
	"time"
)

type HealthChecker struct {
	webhookHandler        *webhook.WebhookHandler
	Servers               *BackendServers
	healthCheckTrigger    time.Duration
	stop                  chan bool
	changeTriggerInterval chan time.Duration
	wg                    *sync.WaitGroup
	lock                  *sync.Mutex
}

func NewHealthChecker(server *BackendServers, webhookHandler *webhook.WebhookHandler, triggerInterval time.Duration, wg *sync.WaitGroup) *HealthChecker {
	return &HealthChecker{
		webhookHandler:        webhookHandler,
		Servers:               server,
		healthCheckTrigger:    triggerInterval,
		stop:                  make(chan bool),
		changeTriggerInterval: make(chan time.Duration),
		wg:                    wg,
		lock:                  &sync.Mutex{},
	}
}

func (hc *HealthChecker) StartHealthCheck() {
	hc.wg.Add(1)
	go func() {
		defer hc.wg.Done()
		ticker := time.NewTicker(hc.healthCheckTrigger)

		for {
			select {
			case <-ticker.C:
				for _, server := range hc.Servers.Servers {
					hc.wg.Add(1)
					go hc.CheckHealth(server)
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

func (hc *HealthChecker) CheckHealth(server *Server) bool {
	defer hc.wg.Done()
	client := &http.Client{}
	resp, err := client.Head(server.URL.String())

	if err != nil || resp.StatusCode != http.StatusOK {
		server.Alive = false
		if server.FireWebhookOnFirstHealthCheck {
			data := map[string]string{
				"host": server.URL.Host,
			}
			hc.webhookHandler.Fire("backendserver.register.success", data)
			server.FireWebhookOnFirstHealthCheck = false
		}
		if resp != nil {
			resp.Body.Close()
		}
		return false
	} else {
		server.Alive = true
		if !server.FireWebhookOnFirstHealthCheck {
			data := map[string]string{
				"host": server.URL.Host,
			}
			hc.webhookHandler.Fire("backendserver.register.failed", data)
			server.FireWebhookOnFirstHealthCheck = false
		}
		if resp != nil {
			resp.Body.Close()
		}
		return true
	}
}

func (hc *HealthChecker) StopHealthChecker() {
	hc.stop <- true
}

func (hc *HealthChecker) UpdateTicker(d time.Duration) {
	hc.changeTriggerInterval <- d
}
