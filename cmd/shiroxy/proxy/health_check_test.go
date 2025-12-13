package proxy_test

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"shiroxy/cmd/shiroxy/proxy"
	"shiroxy/cmd/shiroxy/webhook"
	"sync"
	"testing"
	"time"
)

func TestHealthChecker(t *testing.T) {
	// Set up the WaitGroup and defer waiting to make sure it waits for all checks.
	var wg sync.WaitGroup
	defer wg.Wait()

	// Set up mock webhook handler.
	webhookHandler := &webhook.WebhookHandler{}

	// Set up mock backend server.
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Respond with a 200 OK if the server is healthy, else respond with a 500 error.
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	// Create mock backend servers.
	url, _ := url.Parse(server.URL)
	backendServers := &proxy.BackendServers{
		Servers: []*proxy.Server{
			{
				URL:                           url,
				HealthCheckUrl:                server.URL,
				FireWebhookOnFirstHealthCheck: true,
				Alive:                         false,
				Lock:                          &sync.RWMutex{},
			},
		},
	}

	// Initialize HealthChecker.
	hc := proxy.NewHealthChecker(backendServers, webhookHandler, 1*time.Second, &wg)

	// Start health checks.
	hc.StartHealthCheck()

	// Stop the health checker after a short time.
	time.AfterFunc(3*time.Second, func() {
		hc.StopHealthChecker()
	})

	// Add time to allow the health checker to perform its checks.
	time.Sleep(5 * time.Second)
}

func TestHealthCheckerStop(t *testing.T) {
	// Set up the WaitGroup and defer waiting to make sure it waits for all checks.
	var wg sync.WaitGroup
	defer wg.Wait()

	// Set up mock webhook handler.
	webhookHandler := &webhook.WebhookHandler{}

	// Set up mock backend server.
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Respond with a 500 error to simulate an unhealthy server.
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()
	url, _ := url.Parse(server.URL)

	// Create mock backend servers.
	backendServers := &proxy.BackendServers{
		Servers: []*proxy.Server{
			{
				URL:                           url,
				HealthCheckUrl:                server.URL,
				FireWebhookOnFirstHealthCheck: true,
				Alive:                         true,
				Lock:                          &sync.RWMutex{},
			},
		},
	}

	// Initialize HealthChecker.
	hc := proxy.NewHealthChecker(backendServers, webhookHandler, 1*time.Second, &wg)

	// Start health checks.
	hc.StartHealthCheck()

	// Stop the health checker after a short time.
	time.AfterFunc(3*time.Second, func() {
		hc.StopHealthChecker()
	})

	// Add time to allow the health checker to perform its checks.
	time.Sleep(5 * time.Second)
}
