package proxy_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"sync"
	"testing"
	"time"

	"shiroxy/cmd/shiroxy/proxy"
	"shiroxy/cmd/shiroxy/webhook"
)

// Mock Server struct for testing purposes
type MockServer struct {
	URL                           *url.URL
	Alive                         bool
	FireWebhookOnFirstHealthCheck bool
	HealthCheckUrl                string
	Lock                          *sync.Mutex
}

func TestNewHealthChecker(t *testing.T) {
	fmt.Println("TestNewHealthChecker")
	var wg sync.WaitGroup
	wg.Add(1)
	mockServers := &proxy.BackendServers{
		Servers: []*proxy.Server{
			{
				URL:                           mustParseURL("http://localhost:8080"),
				Alive:                         false,
				FireWebhookOnFirstHealthCheck: true,
				Lock:                          &sync.RWMutex{},
			},
		},
	}
	webhookHandler := &webhook.WebhookHandler{}
	hc := proxy.NewHealthChecker(mockServers, webhookHandler, 2*time.Second, &wg)

	fmt.Println("Started TestNewHealthChecker")

	if hc.Servers != mockServers {
		t.Errorf("expected servers to be %+v, got %+v", mockServers, hc.Servers)
	}
	if hc.HealthCheckTrigger != 2*time.Second {
		t.Errorf("expected healthCheckTrigger to be %v, got %v", 2*time.Second, hc.HealthCheckTrigger)
	}
}

func TestHealthChecker_StartHealthCheck(t *testing.T) {
	fmt.Println("TestHealthChecker_StartHealthCheck")
	var wg sync.WaitGroup
	// wg.Add(1)
	mockServers := &proxy.BackendServers{
		Servers: []*proxy.Server{
			{
				URL:                           mustParseURL("http://localhost:8080"),
				Alive:                         false,
				FireWebhookOnFirstHealthCheck: true,
				Lock:                          &sync.RWMutex{},
			},
		},
	}
	webhookHandler := &webhook.WebhookHandler{}
	hc := proxy.NewHealthChecker(mockServers, webhookHandler, 2*time.Second, &wg)

	hc.StartHealthCheck()
	time.Sleep(3 * time.Second) // Allow time for goroutine to run
	hc.StopHealthChecker()

	time.Sleep(10 * time.Second)

	// time.AfterFunc(5*time.Second, func() {
	// 	wg.Done()
	// })

	// wg.Done()
	// wg.Wait()

	fmt.Println("Finished TestHealthChecker_StartHealthCheck")

	// time.Sleep(5 * time.Second)
	// wg.Done()
}

func TestHealthChecker_CheckHealth(t *testing.T) {
	fmt.Println("TestHealthChecker_CheckHealth")
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	parsedURL, _ := url.Parse(server.URL)

	// Mock server and webhook handler
	mockServer := &proxy.Server{
		URL:                           parsedURL,
		Alive:                         false,
		FireWebhookOnFirstHealthCheck: true,
		Lock:                          &sync.RWMutex{},
	}
	webhookHandler := &webhook.WebhookHandler{}

	hc := proxy.NewHealthChecker(&proxy.BackendServers{
		Servers: []*proxy.Server{mockServer},
	}, webhookHandler, 2*time.Second, &sync.WaitGroup{})

	isHealthy := hc.CheckHealth(mockServer)

	if !isHealthy {
		t.Errorf("expected server to be healthy")
	}
	if !mockServer.Alive {
		t.Errorf("expected server Alive to be true, got %v", mockServer.Alive)
	}

	fmt.Println("TestHealthChecker_CheckHealth Finished")
}

func TestHealthChecker_UpdateTicker(t *testing.T) {
	fmt.Println("TestHealthChecker_UpdateTicker")
	var wg sync.WaitGroup
	wg.Add(1)
	mockServers := &proxy.BackendServers{
		Servers: []*proxy.Server{
			{
				URL:                           mustParseURL("http://localhost:8080"),
				Alive:                         false,
				FireWebhookOnFirstHealthCheck: true,
				Lock:                          &sync.RWMutex{},
			},
		},
	}
	webhookHandler := &webhook.WebhookHandler{}
	hc := proxy.NewHealthChecker(mockServers, webhookHandler, 2*time.Second, &wg)

	hc.StartHealthCheck()
	hc.UpdateTicker(1 * time.Second) // Change the ticker interval
	time.Sleep(2 * time.Second)      // Allow time for ticker to update

	// time.AfterFunc(5*time.Second, func() {
	// 	wg.Done()
	// })

	hc.StopHealthChecker()
	// wg.Wait()

	time.Sleep(10 * time.Second)
	fmt.Println("TestHealthChecker_UpdateTicker Finished")

	// time.Sleep(5 * time.Second)
	// wg.Done()
}

// Helper function to parse URL
func mustParseURL(rawURL string) *url.URL {
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		panic(err)
	}
	return parsedURL
}
