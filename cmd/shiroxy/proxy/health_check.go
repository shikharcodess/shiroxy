// Author: @ShikharY10
// Docs Author: ChatGPT

// Package proxy provides health-check functionalities for monitoring the status of backend servers.
package proxy

import (
	"net/http"
	"shiroxy/cmd/shiroxy/webhook" // Custom package for handling webhooks.
	"sync"
	"time"
)

// HealthChecker is responsible for periodically checking the health of backend servers and
// triggering webhooks based on the server's status.
type HealthChecker struct {
	webhookHandler        *webhook.WebhookHandler // Handler to fire webhooks on server health changes.
	Servers               *BackendServers         // List of backend servers to monitor.
	healthCheckTrigger    time.Duration           // Interval for periodic health checks.
	stop                  chan bool               // Channel to signal stopping the health checks.
	changeTriggerInterval chan time.Duration      // Channel to update the health check interval.
	wg                    *sync.WaitGroup         // WaitGroup for synchronizing goroutines.
	lock                  *sync.Mutex             // Mutex to control concurrent access to shared resources.
}

// NewHealthChecker creates a new HealthChecker instance for monitoring backend server health.
// Parameters:
//   - server: *BackendServers, list of servers to be monitored.
//   - webhookHandler: *webhook.WebhookHandler, handler for firing webhooks on health changes.
//   - triggerInterval: time.Duration, interval between health checks.
//   - wg: *sync.WaitGroup, for synchronization with other goroutines.
//
// Returns:
//   - *HealthChecker: the initialized HealthChecker instance.
func NewHealthChecker(server *BackendServers, webhookHandler *webhook.WebhookHandler, triggerInterval time.Duration, wg *sync.WaitGroup) *HealthChecker {
	return &HealthChecker{
		webhookHandler:        webhookHandler,
		Servers:               server,
		healthCheckTrigger:    triggerInterval,
		stop:                  make(chan bool),          // Channel to control stopping health checks.
		changeTriggerInterval: make(chan time.Duration), // Channel for changing the health check interval.
		wg:                    wg,
		lock:                  &sync.Mutex{}, // Mutex to ensure thread-safe operations.
	}
}

// StartHealthCheck initiates periodic health checks for all backend servers in a separate goroutine.
// The interval of health checks can be modified dynamically, and it stops when a stop signal is received.
func (hc *HealthChecker) StartHealthCheck() {
	hc.wg.Add(1) // Increment WaitGroup counter.
	go func() {
		defer hc.wg.Done()                              // Decrement WaitGroup counter when done.
		ticker := time.NewTicker(hc.healthCheckTrigger) // Create a ticker with the initial health check interval.

		for {
			select {
			case <-ticker.C:
				// On every tick, perform health checks on all servers.
				for _, server := range hc.Servers.Servers {
					hc.wg.Add(1)              // Increment WaitGroup counter for each server health check.
					go hc.CheckHealth(server) // Check server health in a separate goroutine.
				}

			case newDuration := <-hc.changeTriggerInterval:
				// If a new interval duration is received, update the ticker.
				ticker.Stop()
				ticker = time.NewTicker(newDuration)

			case <-hc.stop:
				// If stop signal is received, stop the ticker and exit the loop.
				ticker.Stop()
				return
			}
		}
	}()
}

// CheckHealth checks the health of a given server by sending an HTTP HEAD request.
// If the server is unhealthy, it triggers a webhook and updates the server's status.
// Parameters:
//   - server: *Server, the server whose health is to be checked.
//
// Returns:
//   - bool: true if the server is healthy, false otherwise.
func (hc *HealthChecker) CheckHealth(server *Server) bool {
	defer hc.wg.Done()                            // Decrement WaitGroup counter when done.
	client := &http.Client{}                      // Create a new HTTP client for making the health check request.
	resp, err := client.Head(server.URL.String()) // Send an HTTP HEAD request to the server's URL.

	if err != nil || resp.StatusCode != http.StatusOK {
		// If an error occurs or the server responds with a non-OK status, mark the server as unhealthy.
		server.Alive = false
		if server.FireWebhookOnFirstHealthCheck {
			// Fire a webhook for server registration success if it's the first health check.
			data := map[string]string{
				"host": server.URL.Host,
			}
			hc.webhookHandler.Fire("backendserver.register.success", data)
			server.FireWebhookOnFirstHealthCheck = false
		}
		if resp != nil {
			resp.Body.Close() // Close the response body if it's not nil.
		}
		return false
	} else {
		// If the server is healthy, mark it as alive.
		server.Alive = true
		if !server.FireWebhookOnFirstHealthCheck {
			// Fire a webhook for server registration failure if it's the first health check.
			data := map[string]string{
				"host": server.URL.Host,
			}
			hc.webhookHandler.Fire("backendserver.register.failed", data)
			server.FireWebhookOnFirstHealthCheck = false
		}
		if resp != nil {
			resp.Body.Close() // Close the response body if it's not nil.
		}
		return true
	}
}

// StopHealthChecker stops the health checks by sending a signal to the stop channel.
func (hc *HealthChecker) StopHealthChecker() {
	hc.stop <- true // Send stop signal to the health check goroutine.
}

// UpdateTicker dynamically changes the health check interval.
// Parameters:
//   - d: time.Duration, the new interval duration for health checks.
func (hc *HealthChecker) UpdateTicker(d time.Duration) {
	hc.changeTriggerInterval <- d // Send the new duration to the health check goroutine.
}
